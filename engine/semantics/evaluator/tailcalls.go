package evaluator

import (
	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
	"strings"
)

// tailCallValue is an internal sentinel used to implement proper tail calls.
// It must never escape to user code.
type tailCallValue struct {
	Fn       *value.Function
	Args     []value.Value
	ArgFrame *argFrame
	Node     *syntax.Node // call site for line attribution
}

func (t *tailCallValue) Type() value.Type { return value.Type("tailcall") }
func (t *tailCallValue) Inspect() string  { return "<tailcall>" }

func singleWrappedChild(node *syntax.Node) (*syntax.Node, bool) {
	var only *syntax.Node
	for _, ch := range node.Children {
		if ch.Type == "TERMINAL" {
			if ch.Value != "" && ch.Value != "(" && ch.Value != ")" {
				return nil, false
			}
			continue
		}
		if only != nil {
			return nil, false
		}
		only = ch
	}
	return only, only != nil
}

// evalTail evaluates a node in tail position context.
// It returns either a normal value, a *value.Return, an *value.Error, or a *tailCallValue.
func (e *Evaluator) evalTail(node *syntax.Node, env *value.Env) value.Value {

	// Unwrap simple expression wrapper nodes that contain a single non terminal child
	// and only harmless terminals like parentheses.
	if strings.HasSuffix(node.Type, "_expr") || node.Type == "expr" || node.Type == "primary" {
		if child, ok := singleWrappedChild(node); ok {
			return e.evalTail(child, env)
		}
	}

	switch node.Type {
	case "block":
		return e.evalBlockTail(node, env)
	case "return_stmt":
		// return itself exits the current block; evaluate its expression in tail context
		for _, child := range node.Children {
			if child.Type == "TERMINAL" {
				continue
			}
			v := e.evalTail(child, env)
			if shouldHalt(v) {
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
	case "select_stmt":
		return e.evalSelectTail(node, env)
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
	// Keep one statement pending. Once the next statement is seen, the pending
	// statement is known not to be tail and can be evaluated normally. This
	// avoids building a temporary statement slice for every function body.
	var pending *syntax.Node
	for _, child := range node.Children {
		if child.Type == "TERMINAL" {
			continue
		}
		if pending != nil {
			res := e.Eval(pending, env)
			if ret, ok := res.(*value.Return); ok {
				return ret
			}
			if shouldHalt(res) {
				return res
			}
		}
		pending = child
	}
	if pending == nil {
		return value.EMPTY
	}
	return e.evalTail(pending, env)
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
		return e.makeFault(node, env, "if: invalid syntax")
	}

	condVal := e.Eval(cond, env)
	if shouldHalt(condVal) {
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
	for i, child := range node.Children {
		if child.Type == "TERMINAL" && child.Value == "match" && i+1 < len(node.Children) {
			matchExpr = node.Children[i+1]
			break
		}
	}

	if matchExpr == nil {
		return e.makeFault(node, env, "match: missing expression")
	}

	matchVal := e.Eval(matchExpr, env)
	if shouldHalt(matchVal) {
		return matchVal
	}

	var pattern *syntax.Node
	for _, child := range node.Children {
		switch child.Type {
		case "pattern":
			pattern = child
		case "block":
			if pattern == nil {
				continue
			}
			bindings := patternBindingMap(pattern)
			if e.matchPattern(pattern, matchVal, bindings, env) {
				matchEnv := value.NewEnclosedEnv(env)
				for k, v := range bindings {
					matchEnv.Set(k, v)
				}
				return e.evalTail(child, matchEnv)
			}
			pattern = nil
		}
	}

	return value.EMPTY
}

func (e *Evaluator) evalPipeTail(node *syntax.Node, env *value.Env) value.Value {
	// Evaluate pipe left to right. Count parts without materializing a temporary
	// slice, then traverse the immutable AST directly.
	partCount := 0
	for _, child := range node.Children {
		if child.Type == "TERMINAL" && child.Value == "|>" {
			continue
		}
		partCount++
	}
	if partCount == 0 {
		return value.EMPTY
	}

	partIndex := 0
	var result value.Value
	for _, part := range node.Children {
		if part.Type == "TERMINAL" && part.Value == "|>" {
			continue
		}

		if partIndex == 0 {
			result = e.Eval(part, env)
			if shouldHalt(result) || value.IsShapedError(result) {
				return result
			}
			partIndex++
			continue
		}

		if partIndex == partCount-1 {
			fn := e.evalToFunction(part, env)
			if shouldHalt(fn) {
				return fn
			}
			args := e.evalCallArgsFor(fn, callNode(part), env, result, true)
			if uf, ok := fn.(*value.Function); ok {
				return e.tailCallWithArgs(uf, args, part, env)
			}
			return e.applyEvaluatedFunction(fn, args, part, env)
		}

		result = e.applyPipe(part, result, env)
		if shouldHalt(result) || value.IsShapedError(result) {
			return result
		}
		partIndex++
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
	if shouldHalt(result) {
		return result
	}
	for _, child := range node.Children[1 : len(node.Children)-1] {
		switch child.Type {
		case "index":
			result = e.evalIndex(result, child, env)
		case "access":
			result = e.evalAccess(result, child, env)
		case "call":
			// Intermediate call is not tail, but the callee is already known so
			// argument lifetime can be selected before values are evaluated.
			args := e.evalCallArgsFor(result, child, env, nil, false)
			for _, a := range args.Values {
				if shouldHalt(a) {
					e.releaseEvaluatedArgs(args)
					return a
				}
			}
			result = e.applyEvaluatedFunction(result, args, child, env)
		}
		if shouldHalt(result) {
			return result
		}
	}
	args := e.evalCallArgsFor(result, last, env, nil, false)
	for _, a := range args.Values {
		if shouldHalt(a) {
			e.releaseEvaluatedArgs(args)
			return a
		}
	}
	if uf, ok := result.(*value.Function); ok {
		return e.tailCallWithArgs(uf, args, last, env)
	}
	e.releaseEvaluatedArgs(args)
	return nil
}
