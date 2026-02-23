package substrate

import (
	"aiki/engine/semantics/value"
)

func (g *GoRuntime) registerList() {
	g.register("first", func(args []value.Value) value.Value {
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
	})

	g.register("rest", func(args []value.Value) value.Value {
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
	})

	g.register("length", func(args []value.Value) value.Value {
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
	})

	g.register("prepend", func(args []value.Value) value.Value {
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
		return &value.List{Elements: newElems}
	})

	g.register("append", func(args []value.Value) value.Value {
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
		return &value.List{Elements: newElems}
	})

	g.register("empty", func(args []value.Value) value.Value {
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
	})
}
