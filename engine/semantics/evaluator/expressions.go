package evaluator

import (
	"aiki/engine"

	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
)

func (e *Evaluator) evalExpr(node *syntax.Node, env *value.Env) value.Value {
	switch node.Type {
	case "pipe_expr":
		return e.evalPipe(node, env)
	case "infix_expr":
		return e.evalInfix(node, env)
	case "unary_expr":
		return e.evalUnary(node, env)
	case "postfix_expr":
		return e.evalPostfix(node, env)
	}

	if len(node.Children) == 1 {
		return e.Eval(node.Children[0], env)
	}

	var result value.Value = value.EMPTY
	for _, child := range node.Children {
		result = e.Eval(child, env)
		if shouldHalt(result) {
			return result
		}
	}
	return result
}

func (e *Evaluator) evalPipe(node *syntax.Node, env *value.Env) value.Value {
	var result value.Value

	for _, child := range node.Children {
		if child.Type == "TERMINAL" && child.Value == "|>" {
			continue
		}

		if result == nil {
			result = e.Eval(child, env)
			if shouldHalt(result) {
				return result
			}
			// Short-circuit on shaped error
			if value.IsShapedError(result) {
				return result
			}
			continue
		}

		result = e.applyPipe(child, result, env)
		if shouldHalt(result) {
			return result
		}
		// Short-circuit on shaped error
		if value.IsShapedError(result) {
			return result
		}
	}

	return result
}

func (e *Evaluator) evalUnary(node *syntax.Node, env *value.Env) value.Value {
	var operand value.Value

	for _, child := range node.Children {
		if child.Type == "TERMINAL" && (child.Value == "not" || child.Value == "-") {
			continue
		}
		operand = e.Eval(child, env)
		if shouldHalt(operand) {
			return operand
		}
	}

	if operand == nil {
		return value.EMPTY
	}

	// Prefix operators apply inside-out. Walk the immutable children backward
	// instead of collecting a temporary prefix slice.
	for i := len(node.Children) - 1; i >= 0; i-- {
		child := node.Children[i]
		if child.Type != "TERMINAL" {
			continue
		}
		switch child.Value {
		case "not":
			if value.IsTruthy(operand) {
				operand = value.FALSE
			} else {
				operand = value.TRUE
			}
		case "-":
			if num, ok := operand.(*value.Number); ok {
				e.semanticHit(engine.SemanticArithmetic, node, env)
				operand = num.Neg()
			} else {
				return e.makeFault(node, env, "cannot negate %s", operand.Type())
			}
		}
	}

	return operand
}

func (e *Evaluator) evalPostfix(node *syntax.Node, env *value.Env) value.Value {
	var result value.Value

	for _, child := range node.Children {
		if result == nil {
			result = e.Eval(child, env)
			if shouldHalt(result) {
				return result
			}
			continue
		}

		switch child.Type {
		case "call":
			args := e.evalCallArgsFor(result, child, env, nil, false)
			for _, a := range args.Values {
				if shouldHalt(a) {
					e.releaseEvaluatedArgs(args)
					return a
				}
			}
			result = e.applyEvaluatedFunction(result, args, child, env)
		case "index":
			result = e.evalIndex(result, child, env)
		case "access":
			result = e.evalAccess(result, child, env)
		}

		if shouldHalt(result) {
			return result
		}
	}

	return result
}

func (e *Evaluator) applyPipe(node *syntax.Node, arg value.Value, env *value.Env) value.Value {
	fn := e.evalToFunction(node, env)
	if shouldHalt(fn) {
		return fn
	}

	args := e.evalCallArgsFor(fn, callNode(node), env, arg, true)
	return e.applyEvaluatedFunction(fn, args, node, env)
}

func (e *Evaluator) evalInfix(node *syntax.Node, env *value.Env) value.Value {
	var result value.Value
	var op string

	for _, child := range node.Children {
		if child.Type == "BINOP" {
			for _, c := range child.Children {
				if c.Type == "TERMINAL" {
					op = c.Value
				}
			}
			continue
		}
		if child.Type == "TERMINAL" && e.isBinaryOperator(child.Value) {
			op = child.Value
			continue
		}

		if result != nil && op != "" && isLazyLogicalOperator(op) {
			var handled bool
			result, handled = e.applyLazyLogicalOperator(op, result, child, node, env)
			if !handled {
				return e.makeFault(node, env, "unknown logical operator: %s", op)
			}
			if shouldHalt(result) {
				return result
			}
			op = ""
			continue
		}

		val := e.Eval(child, env)
		if shouldHalt(val) {
			return val
		}

		if result == nil {
			result = val
		} else if op != "" {
			result = e.applyOperator(op, result, val, node, env)
			if shouldHalt(result) {
				return result
			}
			op = ""
		}
	}

	return result
}

func (e *Evaluator) evalPrimary(node *syntax.Node, env *value.Env) value.Value {
	for _, child := range node.Children {
		switch child.Type {
		case "NUMBER", "STRING", "RUNE", "SYMBOL", "NAME":
			return e.Eval(child, env)
		case "TERMINAL":
			if child.Value == "true" {
				return value.TRUE
			}
			if child.Value == "false" {
				return value.FALSE
			}
			if child.Value == "(" {
				// Find expr inside parens
				for _, c := range node.Children {
					if c.Type == "expr" || c.Type == "pipe_expr" {
						return e.Eval(c, env)
					}
				}
			}
		case "list_literal", "func_literal", "expr", "pipe_expr":
			return e.Eval(child, env)
		}
	}
	return value.EMPTY
}

func (e *Evaluator) evalToFunction(node *syntax.Node, env *value.Env) value.Value {
	// For postfix_expr like "str.trim()", we need to evaluate "str.trim" (not just "str")
	// Walk through primary and accesses, stopping before any call
	if node.Type == "postfix_expr" || node.Type == "infix_expr" || node.Type == "unary_expr" || node.Type == "pipe_expr" {
		return e.evalPostfixToFunction(node, env)
	}
	for _, child := range node.Children {
		if child.Type == "NAME" {
			return e.evalName(child, env)
		}
		if child.Type != "TERMINAL" && child.Type != "call" {
			return e.evalToFunction(child, env)
		}
	}
	return e.Eval(node, env)
}

// evalPostfixToFunction evaluates a postfix expression up to but not including the call.
// For "str.trim()", returns the trim function. For "str.foo.bar()", returns bar.
func (e *Evaluator) evalPostfixToFunction(node *syntax.Node, env *value.Env) value.Value {
	// Unwrap to postfix_expr
	for node.Type == "infix_expr" || node.Type == "unary_expr" || node.Type == "pipe_expr" {
		for _, child := range node.Children {
			if child.Type != "TERMINAL" {
				node = child
				break
			}
		}
	}

	if node.Type != "postfix_expr" {
		// Fall back to regular eval for primary
		return e.evalToFunction(node, env)
	}

	var result value.Value
	for _, child := range node.Children {
		// Skip the call - we want the function, not the result of calling it
		if child.Type == "call" {
			break
		}

		if result == nil {
			// Evaluate primary (first element)
			result = e.Eval(child, env)
			if shouldHalt(result) {
				return result
			}
			continue
		}

		// Handle access (module.method)
		if child.Type == "access" {
			result = e.evalAccess(result, child, env)
			if shouldHalt(result) {
				return result
			}
		}
		// Handle index (though unusual in pipe context)
		if child.Type == "index" {
			result = e.evalIndex(result, child, env)
			if shouldHalt(result) {
				return result
			}
		}
	}

	if result == nil {
		return e.Eval(node, env)
	}
	return result
}
