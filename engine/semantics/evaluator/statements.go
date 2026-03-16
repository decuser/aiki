package evaluator

import (
	"fmt"
	"os"
	"strings"

	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
)

func (e *Evaluator) evalProgram(node *syntax.Node, env *value.Env) value.Value {
	// Push main frame at program entry
	env.PushFrame("<main>", 1, env.GetScope())
	defer env.PopFrame()

	var result value.Value = value.EMPTY
	for _, child := range node.Children {
		// Skip TERMINAL nodes (semicolons, newlines)
		if child.Type == "TERMINAL" || child.Type == "NEWLINE" {
			continue
		}
		result = e.Eval(child, env)
		if ret, ok := result.(*value.Return); ok {
			return e.resolveTailCalls(ret.Val, env)
		}
		if shouldHalt(result) {
			return e.resolveTailCalls(result, env)
		}
	}
	return e.resolveTailCalls(result, env)
}

func (e *Evaluator) evalStatement(node *syntax.Node, env *value.Env) value.Value {
	if len(node.Children) > 0 {
		return e.Eval(node.Children[0], env)
	}
	return value.EMPTY
}

func (e *Evaluator) evalExprStmt(node *syntax.Node, env *value.Env) value.Value {
	if len(node.Children) > 0 {
		return e.Eval(node.Children[0], env)
	}
	return value.EMPTY
}

func (e *Evaluator) evalPackage(node *syntax.Node, env *value.Env) value.Value {
	// Extract package name from STRING child
	for _, child := range node.Children {
		if child.Type == "STRING" {
			name := child.Value
			// Remove quotes
			if len(name) >= 2 && name[0] == '"' && name[len(name)-1] == '"' {
				name = name[1 : len(name)-1]
			}
			env.SetPackageName(name)
			return value.EMPTY
		}
	}
	return e.makeFault(node, env, "package: missing name")
}

func (e *Evaluator) evalLet(node *syntax.Node, env *value.Env) value.Value {
	var name string
	var valNode *syntax.Node
	var isShape bool
	var shapeFields []string

	for i, child := range node.Children {
		if child.Type == "NAME" && name == "" {
			name = child.Value
		}
		if child.Type == "SHAPE" {
			isShape = true
			name = strings.TrimPrefix(child.Value, "@")
		}
		if child.Type == "field" {
			for _, f := range child.Children {
				if f.Type == "NAME" {
					shapeFields = append(shapeFields, f.Value)
				}
				if f.Type == "SHAPE" {
					shapeFields = append(shapeFields, f.Value)
				}
			}
		}
		if child.Type == "TERMINAL" && child.Value == "=" && i+1 < len(node.Children) {
			valNode = node.Children[i+1]
		}
	}

	if isShape {
		env.DefineShape(&value.ShapeDef{Name: name, Fields: shapeFields})
		return value.EMPTY
	}

	if name == "" {
		return e.makeFault(node, env, "let: missing name")
	}
	if valNode == nil {
		return e.makeFault(node, env, "let: missing value")
	}

	if e.runtime != nil && e.runtime.HasBuiltin(name, env.GetScope()) {
		return e.makeFault(node, env, "cannot shadow builtin: %s", name)
	}

	// Warn if shadowing a prelude binding at top level (not inside functions/blocks)
	// Only warn for user scope (ScopeUser), not for libs (ScopePrelude)
	if env.GetScope() == value.ScopeUser {
		if preludeEnv := env.GetPreludeEnv(); preludeEnv != nil {
			if _, exists := preludeEnv.Get(name); exists {
				// Only warn if env is directly enclosed by prelude (top-level user scope)
				if env.Outer() == preludeEnv {
					fmt.Fprintf(os.Stderr, "warning: shadowing prelude function '%s'\n", name)
				}
			}
		}
	}

	val := e.Eval(valNode, env)
	if shouldHalt(val) {
		return val
	}

	if fn, ok := val.(*value.Function); ok {
		fn.Name = name
	}

	env.Set(name, val)
	return value.EMPTY
}

func (e *Evaluator) evalAssign(node *syntax.Node, env *value.Env) value.Value {
	var name string
	var valNode *syntax.Node

	for i, child := range node.Children {
		if child.Type == "NAME" && name == "" {
			name = child.Value
		}
		if child.Type == "TERMINAL" && child.Value == "=" && i+1 < len(node.Children) {
			valNode = node.Children[i+1]
		}
	}

	if name == "" || valNode == nil {
		return e.makeFault(node, env, "assign: invalid statement")
	}

	if _, ok := env.Get(name); !ok {
		return e.makeFault(node, env, "undefined variable: %s", name)
	}

	// Block assignment to prelude bindings - use let to shadow instead
	// Only applies to user code (env != preludeEnv), not prelude/test code
	if preludeEnv := env.GetPreludeEnv(); preludeEnv != nil && env != preludeEnv {
		if _, exists := preludeEnv.Get(name); exists {
			// Check if there's a binding in a non-prelude scope that shadows it
			found := false
			for e := env; e != nil && e != preludeEnv; e = e.Outer() {
				if e.HasOwn(name) {
					found = true
					break
				}
			}
			if !found {
				return e.makeFault(node, env, "cannot assign to prelude function '%s'; use let to shadow", name)
			}
		}
	}

	val := e.Eval(valNode, env)
	if shouldHalt(val) {
		return val
	}

	if !env.Update(name, val) {
		return e.makeFault(node, env, "cannot assign to: %s", name)
	}

	return value.EMPTY
}

func (e *Evaluator) evalReturn(node *syntax.Node, env *value.Env) value.Value {
	for _, child := range node.Children {
		if child.Type == "TERMINAL" {
			continue
		}
		val := e.evalTail(child, env)
		if shouldHalt(val) {
			return val
		}
		return &value.Return{Val: val}
	}
	return &value.Return{Val: value.EMPTY}
}

func (e *Evaluator) evalIf(node *syntax.Node, env *value.Env) value.Value {
	var cond *syntax.Node
	var thenBlock *syntax.Node
	var elseBlock *syntax.Node

	for i, child := range node.Children {
		if child.Type == "TERMINAL" && child.Value == "if" {
			if i+1 < len(node.Children) {
				cond = node.Children[i+1]
			}
		}
		if child.Type == "block" {
			if thenBlock == nil {
				thenBlock = child
			} else {
				elseBlock = child
			}
		}
		if child.Type == "if_stmt" {
			elseBlock = child
		}
	}

	if cond == nil || thenBlock == nil {
		return e.makeFault(node, env, "if: invalid syntax")
	}

	condVal := e.Eval(cond, env)
	if shouldHalt(condVal) {
		return condVal
	}

	if value.IsTruthy(condVal) {
		return e.Eval(thenBlock, env)
	} else if elseBlock != nil {
		return e.Eval(elseBlock, env)
	}

	return value.EMPTY
}

func (e *Evaluator) evalWhile(node *syntax.Node, env *value.Env) value.Value {
	var cond *syntax.Node
	var body *syntax.Node

	for i, child := range node.Children {
		if child.Type == "TERMINAL" && child.Value == "while" {
			if i+1 < len(node.Children) {
				cond = node.Children[i+1]
			}
		}
		if child.Type == "block" {
			body = child
		}
	}

	if cond == nil || body == nil {
		return e.makeFault(node, env, "while: invalid syntax")
	}

	var result value.Value = value.EMPTY
	for {
		condVal := e.Eval(cond, env)
		if shouldHalt(condVal) {
			return condVal
		}
		if !value.IsTruthy(condVal) {
			break
		}
		result = e.Eval(body, env)
		if ret, ok := result.(*value.Return); ok {
			return ret
		}
		if shouldHalt(result) {
			return result
		}
	}
	return result
}

func (e *Evaluator) evalMatch(node *syntax.Node, env *value.Env) value.Value {
	var matchExpr *syntax.Node
	var cases []*syntax.Node

	for i, child := range node.Children {
		if child.Type == "TERMINAL" && child.Value == "match" {
			if i+1 < len(node.Children) {
				matchExpr = node.Children[i+1]
			}
		}
		if child.Type == "pattern" || child.Type == "block" {
			cases = append(cases, child)
		}
	}

	if matchExpr == nil {
		return e.makeFault(node, env, "match: missing expression")
	}

	matchVal := e.Eval(matchExpr, env)
	if shouldHalt(matchVal) {
		return matchVal
	}

	for i := 0; i+1 < len(cases); i += 2 {
		pattern := cases[i]
		block := cases[i+1]

		bindings := make(map[string]value.Value)
		if e.matchPattern(pattern, matchVal, bindings, env) {
			matchEnv := value.NewEnclosedEnv(env)
			for k, v := range bindings {
				matchEnv.Set(k, v)
			}
			return e.Eval(block, matchEnv)
		}
	}

	return value.EMPTY
}

func (e *Evaluator) evalBlock(node *syntax.Node, env *value.Env) value.Value {
	var result value.Value = value.EMPTY
	for _, child := range node.Children {
		if child.Type == "TERMINAL" || child.Type == "NEWLINE" {
			continue
		}
		result = e.Eval(child, env)
		if ret, ok := result.(*value.Return); ok {
			return ret
		}
		if shouldHalt(result) {
			return result
		}
	}
	return result
}
