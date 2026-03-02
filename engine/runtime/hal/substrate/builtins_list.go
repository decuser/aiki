package substrate

import (
	"aiki/engine/runtime/hal"
	"aiki/engine/semantics/value"
)

func halFirst(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewError("first: want 1 argument, got %d", len(args))
	}
	switch a := args[0].(type) {
	case *value.List:
		if len(a.Elements) == 0 {
			return value.NewError("first: empty list")
		}
		return a.Elements[0]
	case *value.String:
		runes := []rune(a.Val)
		if len(runes) == 0 {
			return value.NewError("first: empty string")
		}
		return &value.Rune{Val: runes[0]}
	default:
		return value.NewError("first: expected list or string")
	}
}

func halRest(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewError("rest: want 1 argument, got %d", len(args))
	}
	switch a := args[0].(type) {
	case *value.List:
		if len(a.Elements) == 0 {
			return &value.List{Elements: []value.Value{}}
		}
		return &value.List{Elements: a.Elements[1:]}
	case *value.String:
		runes := []rune(a.Val)
		if len(runes) == 0 {
			return &value.String{Val: ""}
		}
		return &value.String{Val: string(runes[1:])}
	default:
		return value.NewError("rest: expected list or string")
	}
}

func halLength(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewError("length: want 1 argument, got %d", len(args))
	}
	switch a := args[0].(type) {
	case *value.List:
		return value.NewNumber(int64(len(a.Elements)), 1)
	case *value.String:
		return value.NewNumber(int64(len([]rune(a.Val))), 1)
	default:
		return value.NewError("length: expected list or string")
	}
}

func halPrepend(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewError("prepend: want 2 arguments, got %d", len(args))
	}
	list, ok := args[0].(*value.List)
	if !ok {
		return value.NewError("prepend: first argument must be list")
	}
	newElems := make([]value.Value, len(list.Elements)+1)
	newElems[0] = args[1]
	copy(newElems[1:], list.Elements)
	return &value.List{Elements: newElems, Shape: list.Shape}
}

func halAppend(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewError("append: want 2 arguments, got %d", len(args))
	}
	list, ok := args[0].(*value.List)
	if !ok {
		return value.NewError("append: first argument must be list")
	}
	newElems := make([]value.Value, len(list.Elements)+1)
	copy(newElems, list.Elements)
	newElems[len(list.Elements)] = args[1]
	return &value.List{Elements: newElems, Shape: list.Shape}
}

func halEmpty(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewError("empty: want 1 argument, got %d", len(args))
	}
	switch a := args[0].(type) {
	case *value.List:
		if len(a.Elements) == 0 {
			return value.TRUE
		}
		return value.FALSE
	case *value.String:
		if len(a.Val) == 0 {
			return value.TRUE
		}
		return value.FALSE
	default:
		return value.NewError("empty: expected list or string")
	}
}

func halRange(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) < 1 || len(args) > 2 {
		return value.NewError("range: want 1 or 2 arguments")
	}
	var start, end int64
	if len(args) == 1 {
		n, ok := args[0].(*value.Number)
		if !ok || !n.Val.IsInt() {
			return value.NewError("range: expected integer")
		}
		start = 0
		end = n.Val.Num().Int64()
	} else {
		s, ok := args[0].(*value.Number)
		if !ok || !s.Val.IsInt() {
			return value.NewError("range: expected integer")
		}
		e, ok := args[1].(*value.Number)
		if !ok || !e.Val.IsInt() {
			return value.NewError("range: expected integer")
		}
		start = s.Val.Num().Int64()
		end = e.Val.Num().Int64()
	}
	var elems []value.Value
	for i := start; i < end; i++ {
		elems = append(elems, value.NewNumber(i, 1))
	}
	return &value.List{Elements: elems}
}

// halMakeShapedList creates a shaped list from a shape name and elements list.
func halMakeShapedList(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewError("make_shaped_list: want 2 arguments, got %d", len(args))
	}
	sym, ok := args[0].(*value.Symbol)
	if !ok {
		return value.NewError("make_shaped_list: expected symbol for shape, got %s", args[0].Type())
	}
	list, ok := args[1].(*value.List)
	if !ok {
		return value.NewError("make_shaped_list: expected list for elements, got %s", args[1].Type())
	}
	return &value.List{Elements: list.Elements, Shape: sym.Val}
}
