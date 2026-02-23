package evaluator

import (
	"math/big"

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

	var result value.Value = value.NULL
	for _, child := range node.Children {
		result = e.Eval(child, env)
		if value.IsError(result) {
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
			if value.IsError(result) {
				return result
			}
			continue
		}

		result = e.applyPipe(child, result, env)
		if value.IsError(result) {
			return result
		}
	}

	return result
}

func (e *Evaluator) evalUnary(node *syntax.Node, env *value.Env) value.Value {
	var prefixes []string
	var operand value.Value

	for _, child := range node.Children {
		if child.Type == "TERMINAL" {
			if child.Value == "not" || child.Value == "-" {
				prefixes = append(prefixes, child.Value)
				continue
			}
		}
		operand = e.Eval(child, env)
		if value.IsError(operand) {
			return operand
		}
	}

	if operand == nil {
		return value.NULL
	}

	for i := len(prefixes) - 1; i >= 0; i-- {
		switch prefixes[i] {
		case "not":
			if value.IsTruthy(operand) {
				operand = value.FALSE
			} else {
				operand = value.TRUE
			}
		case "-":
			if num, ok := operand.(*value.Number); ok {
				neg := new(big.Rat).Neg(num.Val)
				operand = &value.Number{Val: neg}
			} else {
				return e.makeError(node, env, "cannot negate %s", operand.Type())
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
			if value.IsError(result) {
				return result
			}
			continue
		}

		switch child.Type {
		case "call":
			args := e.evalCallArgs(child, env)
			for _, a := range args {
				if value.IsError(a) {
					return a
				}
			}
			result = e.applyFunction(result, args, child, env)
		case "index":
			result = e.evalIndex(result, child, env)
		case "access":
			result = e.evalAccess(result, child, env)
		}

		if value.IsError(result) {
			return result
		}
	}

	return result
}

func (e *Evaluator) applyPipe(node *syntax.Node, arg value.Value, env *value.Env) value.Value {
	fn := e.evalToFunction(node, env)
	if value.IsError(fn) {
		return fn
	}

	args := []value.Value{arg}
	args = append(args, e.collectCallArgs(node, env)...)

	return e.applyFunction(fn, args, node, env)
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
		if child.Type == "TERMINAL" && isOperator(child.Value) {
			op = child.Value
			continue
		}

		val := e.Eval(child, env)
		if value.IsError(val) {
			return val
		}

		if result == nil {
			result = val
		} else if op != "" {
			result = e.applyOperator(op, result, val, node, env)
			if value.IsError(result) {
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
	return value.NULL
}

func (e *Evaluator) evalToFunction(node *syntax.Node, env *value.Env) value.Value {
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

func (e *Evaluator) collectCallArgs(node *syntax.Node, env *value.Env) []value.Value {
	for _, child := range node.Children {
		if child.Type == "call" {
			return e.evalCallArgs(child, env)
		}
	}
	return nil
}
