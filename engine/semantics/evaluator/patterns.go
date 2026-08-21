package evaluator

import (
	"strings"

	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
)

func (e *Evaluator) matchPattern(pattern *syntax.Node, val value.Value, bindings map[string]value.Value, env *value.Env) bool {
	// Unwrap single-child pattern
	if pattern.Type == "pattern" && len(pattern.Children) == 1 {
		return e.matchPattern(pattern.Children[0], val, bindings, env)
	}

	// List pattern: pattern with "[" ... "]" structure
	if pattern.Type == "pattern" && len(pattern.Children) > 0 {
		if pattern.Children[0].Type == "TERMINAL" && pattern.Children[0].Value == "[" {
			list, ok := val.(*value.List)
			if !ok {
				return false
			}
			return e.matchListPattern(pattern, list, bindings, env)
		}
	}

	// Unwrap literal node
	if pattern.Type == "literal" && len(pattern.Children) > 0 {
		return e.matchPattern(pattern.Children[0], val, bindings, env)
	}

	if (pattern.Type == "NAME" || pattern.Type == "TERMINAL") && pattern.Value == "_" {
		return true
	}

	if pattern.Type == "NAME" {
		if bindings != nil {
			bindings[pattern.Value] = val
		}
		return true
	}

	if pattern.Type == "NUMBER" || pattern.Type == "STRING" ||
		pattern.Type == "RUNE" || pattern.Type == "SYMBOL" {
		patternVal := e.Eval(pattern, env)
		return e.valuesEqual(patternVal, val)
	}

	// Boolean literal patterns are TERMINAL nodes carrying "true" or "false".
	// The wildcard "_" is handled above, so only these two remain.
	if pattern.Type == "TERMINAL" && (pattern.Value == "true" || pattern.Value == "false") {
		patternVal := e.Eval(pattern, env)
		return e.valuesEqual(patternVal, val)
	}

	if pattern.Type == "list_literal" {
		list, ok := val.(*value.List)
		if !ok {
			return false
		}
		return e.matchListPattern(pattern, list, bindings, env)
	}

	return false
}

func (e *Evaluator) matchListPattern(pattern *syntax.Node, list *value.List, bindings map[string]value.Value, env *value.Env) bool {
	var shapeName string
	elemCount := 0
	for _, child := range pattern.Children {
		if child.Type == "SHAPE" {
			shapeName = strings.TrimPrefix(child.Value, "@")
		} else if child.Type != "TERMINAL" {
			elemCount++
		}
	}

	if shapeName != "" && list.Shape != shapeName {
		return false
	}
	if elemCount != len(list.Elements) {
		return false
	}

	i := 0
	for _, child := range pattern.Children {
		if child.Type == "TERMINAL" || child.Type == "SHAPE" {
			continue
		}
		if !e.matchPattern(child, list.Elements[i], bindings, env) {
			return false
		}
		i++
	}
	return true
}

// patternBinds reports whether matching the pattern can introduce a lexical
// binding. It traverses the immutable AST directly and allocates nothing.
func patternBinds(pattern *syntax.Node) bool {
	if pattern == nil {
		return false
	}
	if pattern.Type == "NAME" && pattern.Value != "_" {
		return true
	}
	for _, child := range pattern.Children {
		if patternBinds(child) {
			return true
		}
	}
	return false
}

func patternBindingMap(pattern *syntax.Node) map[string]value.Value {
	if !patternBinds(pattern) {
		return nil
	}
	return make(map[string]value.Value)
}

func (e *Evaluator) valuesEqual(a, b value.Value) bool {
	if a.Type() != b.Type() {
		return false
	}
	switch av := a.(type) {
	case *value.Number:
		bv := b.(*value.Number)
		return av.Equal(bv)
	case *value.String:
		bv := b.(*value.String)
		return av.Val == bv.Val
	case *value.Symbol:
		bv := b.(*value.Symbol)
		return av.Val == bv.Val
	case *value.Boolean:
		bv := b.(*value.Boolean)
		return av.Val == bv.Val
	case *value.Rune:
		bv := b.(*value.Rune)
		return av.Val == bv.Val
	default:
		return false
	}
}
