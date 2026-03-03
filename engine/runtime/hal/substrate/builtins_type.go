package substrate

import (
	"aiki/engine/runtime/hal"
	"aiki/engine/semantics/value"
)

func halType(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewError("type: want 1 argument, got %d", len(args))
	}
	return &value.Symbol{Val: string(args[0].Type())}
}

func halInspect(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewError("inspect: want 1 argument, got %d", len(args))
	}
	return &value.String{Val: args[0].Inspect()}
}

func halEqual(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewError("equal: want 2 arguments, got %d", len(args))
	}
	if valuesEqual(args[0], args[1]) {
		return value.TRUE
	}
	return value.FALSE
}

func halOrd(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewError("ord: want 1 argument")
	}
	switch v := args[0].(type) {
	case *value.Rune:
		return value.NewNumber(int64(v.Val), 1)
	case *value.String:
		if len(v.Val) == 0 {
			return value.NewNumber(0, 1)
		}
		return value.NewNumber(int64([]rune(v.Val)[0]), 1)
	default:
		return value.NewError("ord: expected rune or string")
	}
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
	default:
		return false
	}
}

// halStackLimit sets the evaluator stack limit (non tail call frames).
// n must be an integer >= 1.
func halStackLimit(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewError("stack_limit: want 1 argument, got %d", len(args))
	}
	if ctx == nil || ctx.Env == nil {
		return value.NewError("stack_limit: environment not available")
	}
	num, ok := args[0].(*value.Number)
	if !ok || !num.Val.IsInt() {
		return value.NewError("stack_limit: n must be integer >= 1")
	}
	n := num.Val.Num().Int64()
	if n < 1 {
		return value.NewError("stack_limit: n must be integer >= 1")
	}
	ctx.Env.SetStackLimit(int(n))
	return value.EMPTY
}
