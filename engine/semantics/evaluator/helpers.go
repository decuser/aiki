package evaluator

import (
	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
)

// shouldHalt returns true if evaluation should halt on this value.
// During the shadow period, this checks both Fault (new) and Error (old).
func shouldHalt(v value.Value) bool {
	return value.IsError(v) || value.IsFault(v)
}

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
		env.CopyStack(),
		format,
		args...,
	)
}

func (e *Evaluator) makeFault(node *syntax.Node, env *value.Env, format string, args ...interface{}) *value.Fault {
	return value.NewFaultAt(
		env.GetFile(),
		node.Pos.Line,
		env.GetSourceLine(node.Pos.Line),
		env.CopyStack(),
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

// resolveTailCalls drains internal tail call sentinels so they never escape evaluator boundaries.
func (e *Evaluator) resolveTailCalls(v value.Value, env *value.Env) value.Value {
	for {
		tc, ok := v.(*tailCallValue)
		if !ok {
			return v
		}
		v = e.applyUserFunction(tc.Fn, tc.Args, tc.Node, env)
	}
}
