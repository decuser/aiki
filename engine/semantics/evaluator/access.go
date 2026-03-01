package evaluator

import (
	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
)

func (e *Evaluator) evalIndex(val value.Value, node *syntax.Node, env *value.Env) value.Value {
	list, ok := val.(*value.List)
	if !ok {
		if s, ok := val.(*value.String); ok {
			return e.evalStringIndex(s, node, env)
		}
		return e.makeError(node, env, "cannot index %s", val.Type())
	}

	for _, child := range node.Children {
		if child.Type != "TERMINAL" {
			idx := e.Eval(child, env)
			if value.IsError(idx) {
				return idx
			}
			num, ok := idx.(*value.Number)
			if !ok {
				return e.makeError(node, env, "index must be a number")
			}
			if !num.Val.IsInt() {
				return e.makeError(node, env, "index must be an integer")
			}
			i := int(num.Val.Num().Int64())
			if i < 0 || i >= len(list.Elements) {
				return e.makeError(node, env, "index out of bounds: %d", i)
			}
			return list.Elements[i]
		}
	}

	return value.EMPTY
}

func (e *Evaluator) evalStringIndex(s *value.String, node *syntax.Node, env *value.Env) value.Value {
	for _, child := range node.Children {
		if child.Type != "TERMINAL" {
			idx := e.Eval(child, env)
			if value.IsError(idx) {
				return idx
			}
			num, ok := idx.(*value.Number)
			if !ok {
				return e.makeError(node, env, "index must be a number")
			}
			if !num.Val.IsInt() {
				return e.makeError(node, env, "index must be an integer")
			}
			i := int(num.Val.Num().Int64())
			runes := []rune(s.Val)
			if i < 0 || i >= len(runes) {
				return e.makeError(node, env, "index out of bounds: %d", i)
			}
			return &value.Rune{Val: runes[i]}
		}
	}
	return value.EMPTY
}

func (e *Evaluator) evalAccess(val value.Value, node *syntax.Node, env *value.Env) value.Value {
	list, ok := val.(*value.List)
	if !ok || list.Shape == "" {
		return e.makeError(node, env, "cannot access field on %s", val.Type())
	}

	for _, child := range node.Children {
		if child.Type == "NAME" {
			fieldName := child.Value
			shapeDef, ok := env.GetShape(list.Shape)
			if !ok {
				return e.makeError(node, env, "unknown shape: %s", list.Shape)
			}
			for i, f := range shapeDef.Fields {
				if f == fieldName && i < len(list.Elements) {
					return list.Elements[i]
				}
			}
			return e.makeError(node, env, "no field %s in @%s", fieldName, list.Shape)
		}
	}

	return value.EMPTY
}
