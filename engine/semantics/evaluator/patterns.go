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
		bindings[pattern.Value] = val
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
	var patternElems []*syntax.Node
	var shapeName string

	for _, child := range pattern.Children {
		if child.Type == "SHAPE" {
			shapeName = strings.TrimPrefix(child.Value, "@")
		} else if child.Type != "TERMINAL" {
			patternElems = append(patternElems, child)
		}
	}

	if shapeName != "" && list.Shape != shapeName {
		return false
	}

	if len(patternElems) != len(list.Elements) {
		return false
	}

	for i, pe := range patternElems {
		if !e.matchPattern(pe, list.Elements[i], bindings, env) {
			return false
		}
	}

	return true
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
