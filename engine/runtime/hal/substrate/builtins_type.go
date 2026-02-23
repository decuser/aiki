package substrate

import (
	"aiki/engine/semantics/value"
)

func (g *GoRuntime) registerType() {
	g.register("type", func(args []value.Value) value.Value {
		if len(args) != 1 {
			return value.NewError("type: want 1 argument, got %d", len(args))
		}
		return &value.Symbol{Val: string(args[0].Type())}
	})

	g.register("inspect", func(args []value.Value) value.Value {
		if len(args) != 1 {
			return value.NewError("inspect: want 1 argument, got %d", len(args))
		}
		return &value.String{Val: args[0].Inspect()}
	})

	g.register("equal", func(args []value.Value) value.Value {
		if len(args) != 2 {
			return value.NewError("equal: want 2 arguments, got %d", len(args))
		}
		if valuesEqual(args[0], args[1]) {
			return value.TRUE
		}
		return value.FALSE
	})
}

func valuesEqual(a, b value.Value) bool {
	if a.Type() != b.Type() {
		return false
	}
	switch av := a.(type) {
	case *value.Number:
		bv := b.(*value.Number)
		return av.Val.Cmp(bv.Val) == 0
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
	case *value.Null:
		return true
	default:
		return false
	}
}
