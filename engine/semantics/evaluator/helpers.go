package eval

import (
	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
)

func (e *Evaluator) nonTerminalChildren(node *syntax.Node) []*syntax.Node {
	var result []*syntax.Node
	for _, child := range node.Children {
		if child.Type != "TERMINAL" {
			result = append(result, child)
		}
	}
	return result
}

func (e *Evaluator) makeError(node *syntax.Node, env *value.Env, format string, args ...interface{}) *value.Error {
	return value.NewErrorAt(
		env.GetFile(),
		node.Pos.Line,
		env.GetSourceLine(node.Pos.Line),
		format,
		args...,
	)
}

func isOperator(s string) bool {
	switch s {
	case "+", "-", "*", "/", "<", ">", "<=", ">=", "and", "or":
		return true
	}
	return false
}
