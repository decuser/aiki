package evaluator

import (
	"aiki/engine"
	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
)

// tryTailCall evaluates a return expression that is a call in tail position.
// If it is a user function call, it returns a TailCall marker to be handled by applyUserFunction.
// Builtin calls are executed normally.
func (e *Evaluator) tryTailCall(expr *syntax.Node, env *value.Env) (value.Value, bool) {
	if expr == nil || expr.Type != "postfix_expr" {
		return nil, false
	}

	// Identify the last non terminal child.
	var last *syntax.Node
	for i := len(expr.Children) - 1; i >= 0; i-- {
		c := expr.Children[i]
		if c.Type != "TERMINAL" {
			last = c
			break
		}
	}
	if last == nil || last.Type != "call" {
		return nil, false
	}

	// Evaluate base expression.
	var result value.Value
	var base *syntax.Node
	for _, c := range expr.Children {
		if c.Type == "TERMINAL" {
			continue
		}
		// base is the first non postfix node
		if c.Type != "call" && c.Type != "access" && c.Type != "index" {
			base = c
			break
		}
	}
	if base == nil {
		return nil, false
	}
	result = e.Eval(base, env)
	if value.IsError(result) {
		return result, true
	}

	// Apply postfixes up to, but excluding, the last call.
	for _, c := range expr.Children {
		if c.Type == "TERMINAL" || c == base {
			continue
		}
		switch c.Type {
		case "access":
			result = e.evalAccess(result, c, env)
			if value.IsError(result) {
				return result, true
			}
		case "index":
			result = e.evalIndex(result, c, env)
			if value.IsError(result) {
				return result, true
			}
		case "call":
			// Stop at the last call, handled below
			if c == last {
				goto DONE
			}
			// Intermediate calls must be executed normally
			args := e.evalCallArgs(c, env)
			for _, a := range args {
				if value.IsError(a) {
					return a, true
				}
			}
			result = e.applyFunction(result, args, c, env)
			if value.IsError(result) {
				return result, true
			}
		}
	}
DONE:
	args := e.evalCallArgs(last, env)
	for _, a := range args {
		if value.IsError(a) {
			return a, true
		}
	}

	// Only user function calls need the tail call marker.
	if _, ok := result.(*value.Function); ok {
		pos := engine.Position{File: env.GetFile(), Line: last.Pos.Line, Col: last.Pos.Col}
		return &value.TailCall{Fn: result, Args: args, Pos: pos}, true
	}

	// Builtins in tail position do not need special handling.
	return e.applyFunction(result, args, last, env), true
}
