package evaluator

import (
	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
)

// tailCallValue is an internal sentinel used to implement proper tail calls.
// It must never escape to user code.
type tailCallValue struct {
	Fn   *value.Function
	Args []value.Value
	Node *syntax.Node // call site for line attribution
}

func (t *tailCallValue) Type() value.Type { return value.Type("tailcall") }
func (t *tailCallValue) Inspect() string  { return "<tailcall>" }

// evalTail evaluates a node in tail position context.
// It returns either a normal value, a *value.Return, an *value.Error, or a *tailCallValue.
func (e *Evaluator) evalTail(node *syntax.Node, env *value.Env) value.Value {
	switch node.Type {
	case "expr":
		// expr is a wrapper production. Evaluate its first non terminal child in tail context.
		for _, ch := range node.Children {
			if ch.Type == "TERMINAL" {
				continue
			}
			return e.evalTail(ch, env)
		}
		return value.EMPTY
	case "primary":
		// primary is also a wrapper in many paths.
		for _, ch := range node.Children {
			if ch.Type == "TERMINAL" {
				continue
			}
			return e.evalTail(ch, env)
		}
		return value.EMPTY
	case "block":
		return e.evalBlockTail(node, env)
	case "return_stmt":
		// return itself exits the current block; evaluate its expression in tail context
		for _, child := range node.Children {
			if child.Type == "TERMINAL" {
				continue
			}
			v := e.evalTail(child, env)
			if value.IsError(v) {
				return v
			}
			return &value.Return{Val: v}
		}
		return &value.Return{Val: value.EMPTY}
	case "expr_stmt":
		if len(node.Children) == 0 {
			return value.EMPTY
		}
		return e.evalTail(node.Children[0], env)
	case "if_stmt":
		return e.evalIfTail(node, env)
	case "match_stmt":
		return e.evalMatchTail(node, env)
	case "pipe_expr":
		return e.evalPipeTail(node, env)
	case "postfix_expr":
		// try to convert last call into a tail jump
		if tc := e.tryTailCall(node, env); tc != nil {
			return tc
		}
		return e.Eval(node, env)
	default:
		// for other nodes, evaluate normally
		return e.Eval(node, env)
	}
}

func (e *Evaluator) evalBlockTail(node *syntax.Node, env *value.Env) value.Value {
	// Evaluate all statements except the last in normal mode.
	// Evaluate the last statement in tail mode.
	var stmts []*syntax.Node
	for _, child := range node.Children {
		if child.Type == "TERMINAL" {
			continue
		}
		stmts = append(stmts, child)
	}
	if len(stmts) == 0 {
		return value.EMPTY
	}
	for i := 0; i < len(stmts)-1; i++ {
		res := e.Eval(stmts[i], env)
		if ret, ok := res.(*value.Return); ok {
			return ret
		}
		if value.IsError(res) {
			return res
		}
	}
	// Tail evaluate last
	last := e.evalTail(stmts[len(stmts)-1], env)
	return last
}

func (e *Evaluator) evalIfTail(node *syntax.Node, env *value.Env) value.Value {
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
		return e.makeError(node, env, "if: invalid syntax")
	}

	condVal := e.Eval(cond, env)
	if value.IsError(condVal) {
		return condVal
	}

	if value.IsTruthy(condVal) {
		return e.evalTail(thenBlock, env)
	}
	if elseBlock != nil {
		return e.evalTail(elseBlock, env)
	}
	return value.EMPTY
}

func (e *Evaluator) evalMatchTail(node *syntax.Node, env *value.Env) value.Value {
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
		return e.makeError(node, env, "match: missing expression")
	}

	matchVal := e.Eval(matchExpr, env)
	if value.IsError(matchVal) {
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
			return e.evalTail(block, matchEnv)
		}
	}

	return value.EMPTY
}

func (e *Evaluator) evalPipeTail(node *syntax.Node, env *value.Env) value.Value {
	// Evaluate pipe left to right. Only the final application may be in tail position.
	var parts []*syntax.Node
	for _, child := range node.Children {
		if child.Type == "TERMINAL" && child.Value == "|>" {
			continue
		}
		parts = append(parts, child)
	}
	if len(parts) == 0 {
		return value.EMPTY
	}
	// first part normal
	result := e.Eval(parts[0], env)
	if value.IsError(result) {
		return result
	}
	for i := 1; i < len(parts); i++ {
		// last application in tail context
		if i == len(parts)-1 {
			fn := e.evalToFunction(parts[i], env)
			if value.IsError(fn) {
				return fn
			}
			args := []value.Value{result}
			args = append(args, e.collectCallArgs(parts[i], env)...)
			// only tail optimize user functions
			if uf, ok := fn.(*value.Function); ok {
				return &tailCallValue{Fn: uf, Args: args, Node: parts[i]}
			}
			return e.applyFunction(fn, args, parts[i], env)
		}
		// non tail intermediate step
		result = e.applyPipe(parts[i], result, env)
		if value.IsError(result) {
			return result
		}
	}
	return result
}

func (e *Evaluator) tryTailCall(node *syntax.Node, env *value.Env) value.Value {
	// Tail call if the last postfix op is a call and there are no ops after it.
	if len(node.Children) < 2 {
		return nil
	}
	last := node.Children[len(node.Children)-1]
	if last.Type != "call" {
		return nil
	}
	// evaluate callee and any preceding postfix ops except last call
	result := e.Eval(node.Children[0], env)
	if value.IsError(result) {
		return result
	}
	for _, child := range node.Children[1 : len(node.Children)-1] {
		switch child.Type {
		case "index":
			result = e.evalIndex(result, child, env)
		case "access":
			result = e.evalAccess(result, child, env)
		case "call":
			// intermediate call is not tail, evaluate normally
			args := e.evalCallArgs(child, env)
			for _, a := range args {
				if value.IsError(a) {
					return a
				}
			}
			result = e.applyFunction(result, args, child, env)
		}
		if value.IsError(result) {
			return result
		}
	}
	args := e.evalCallArgs(last, env)
	for _, a := range args {
		if value.IsError(a) {
			return a
		}
	}
	if uf, ok := result.(*value.Function); ok {
		return &tailCallValue{Fn: uf, Args: args, Node: last}
	}
	return nil
}
