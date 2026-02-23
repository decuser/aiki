package eval

import (
	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
)

func (e *Evaluator) evalExpr(node *syntax.Node, env *value.Env) value.Value {
	children := e.nonTerminalChildren(node)

	if len(children) == 0 {
		return value.NULL
	}
	if len(children) == 1 {
		return e.Eval(children[0], env)
	}

	for _, child := range node.Children {
		if child.Type == "TERMINAL" && child.Value == "|>" {
			return e.evalPipe(node, env)
		}
	}

	return e.evalInfix(node, env)
}

func (e *Evaluator) evalPipe(node *syntax.Node, env *value.Env) value.Value {
	children := e.nonTerminalChildren(node)
	if len(children) == 0 {
		return value.NULL
	}

	result := e.Eval(children[0], env)
	if value.IsError(result) {
		return result
	}

	for i := 1; i < len(children); i++ {
		result = e.applyPipe(children[i], result, env)
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
	children := e.nonTerminalChildren(node)
	if len(children) == 0 {
		return value.NULL
	}

	result := e.Eval(children[0], env)
	if value.IsError(result) {
		return result
	}

	i := 1
	for _, child := range node.Children {
		if child.Type == "TERMINAL" && isOperator(child.Value) {
			if i >= len(children) {
				break
			}
			right := e.Eval(children[i], env)
			if value.IsError(right) {
				return right
			}
			result = e.applyOperator(child.Value, result, right, node, env)
			if value.IsError(result) {
				return result
			}
			i++
		}
	}

	return result
}

func (e *Evaluator) evalPrimary(node *syntax.Node, env *value.Env) value.Value {
	children := e.nonTerminalChildren(node)
	if len(children) == 0 {
		return value.NULL
	}

	for _, child := range node.Children {
		if child.Type == "TERMINAL" && child.Value == "(" {
			for _, c := range children {
				if c.Type == "expr" {
					return e.Eval(c, env)
				}
			}
		}
	}

	result := e.Eval(children[0], env)
	if value.IsError(result) {
		return result
	}

	for i := 1; i < len(children); i++ {
		child := children[i]
		switch child.Type {
		case "call":
			args := e.evalCallArgs(child, env)
			if len(args) == 1 && value.IsError(args[0]) {
				return args[0]
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
