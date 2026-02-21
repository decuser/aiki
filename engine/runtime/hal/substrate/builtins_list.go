package substrate

import (
	"fmt"

	"aiki/engine/semantics/value"
)

func (g *GoRuntime) registerList() {
	// Length - returns length of list or string
	g.Register("length", func(args []value.Value) (value.Value, error) {
		if len(args) != 1 {
			return value.NullValue(), fmt.Errorf("length: want 1 argument")
		}
		switch args[0].Type {
		case value.List:
			if list, ok := args[0].Data.([]value.Value); ok {
				return value.NewNumber(float64(len(list))), nil
			}
		case value.String:
			if s, ok := args[0].Data.(string); ok {
				return value.NewNumber(float64(len([]rune(s)))), nil
			}
		}
		return value.NullValue(), fmt.Errorf("length: invalid type %s", value.TypeName(args[0].Type))
	})

	// First - returns first element of list OR first character of string
	g.Register("first", func(args []value.Value) (value.Value, error) {
	    if len(args) != 1 {
		return value.NullValue(), fmt.Errorf("first: want 1 argument")
	    }
	    switch args[0].Type {
	    case value.List:
		list, ok := args[0].Data.([]value.Value)
		if !ok || len(list) == 0 {
		    return value.NullValue(), fmt.Errorf("first: empty list")
		}
		return list[0], nil
	    case value.String:
		s := args[0].Data.(string)
		if len(s) == 0 {
		    return value.NullValue(), fmt.Errorf("first: empty string")
		}
		runes := []rune(s)
		return value.Value{Type: value.Rune, Data: runes[0]}, nil
	    default:
		return value.NullValue(), fmt.Errorf("first: expected list or string")
	    }
	})

	// Rest - returns all but first element of list OR all but first character of string
	g.Register("rest", func(args []value.Value) (value.Value, error) {
		if len(args) != 1 {
			return value.NullValue(), fmt.Errorf("rest: want 1 argument")
		}

		switch args[0].Type {
		case value.List:
			list, ok := args[0].Data.([]value.Value)
			if !ok || len(list) == 0 {
				return value.NewList(nil), nil
			}
			if len(list) == 1 {
				return value.NewList(nil), nil
			}
			// copy to avoid aliasing caller backing array
			out := make([]value.Value, len(list)-1)
			copy(out, list[1:])
			return value.NewList(out), nil

		case value.String:
			s, ok := args[0].Data.(string)
			if !ok || len(s) == 0 {
				return value.NewString(""), nil
			}
			r := []rune(s)
			if len(r) <= 1 {
				return value.NewString(""), nil
			}
			return value.NewString(string(r[1:])), nil

		default:
			return value.NullValue(), fmt.Errorf("rest: expected list or string")
		}
	})

	// Cons - prepends element to list
	g.Register("cons", func(args []value.Value) (value.Value, error) {
		if len(args) != 2 {
			return value.NullValue(), fmt.Errorf("cons: want 2 arguments")
		}
		if args[1].Type != value.List {
			return value.NullValue(), fmt.Errorf("cons: second argument must be list")
		}
		list, _ := args[1].Data.([]value.Value)
		newList := make([]value.Value, 0, len(list)+1)
		newList = append(newList, args[0])
		newList = append(newList, list...)
		return value.NewList(newList), nil
	})

	// Prepend - prepends element to list
	g.Register("prepend", func(args []value.Value) (value.Value, error) {
		if len(args) != 2 {
			return value.NullValue(), fmt.Errorf("prepend: want 2 arguments")
		}
		if args[0].Type != value.List {
			return value.NullValue(), fmt.Errorf("prepend: first argument must be list")
		}
		list, _ := args[0].Data.([]value.Value)
		newList := make([]value.Value, 0, len(list)+1)
		newList = append(newList, args[1])
		newList = append(newList, list...)
		return value.NewList(newList), nil
	})

	// Append - appends element to list
	g.Register("append", func(args []value.Value) (value.Value, error) {
		if len(args) != 2 {
			return value.NullValue(), fmt.Errorf("append: want 2 arguments")
		}
		if args[0].Type != value.List {
			return value.NullValue(), fmt.Errorf("append: first argument must be list")
		}
		list, _ := args[0].Data.([]value.Value)
		newList := make([]value.Value, len(list), len(list)+1)
		copy(newList, list)
		newList = append(newList, args[1])
		return value.NewList(newList), nil
	})

	// Equal - equality comparison
	g.Register("equal", func(args []value.Value) (value.Value, error) {
		if len(args) != 2 {
			return value.NullValue(), fmt.Errorf("equal: want 2 arguments")
		}
		return value.Value{Type: value.Boolean, Data: valuesEqual(args[0], args[1])}, nil
	})

	// Channel primitives
	g.Register("channel", func(args []value.Value) (value.Value, error) {
		ch := make(chan value.Value)
		return value.Value{Type: value.Channel, Data: ch}, nil
	})

	g.Register("send", func(args []value.Value) (value.Value, error) {
		if len(args) != 2 {
			return value.NullValue(), fmt.Errorf("send: want 2 arguments")
		}
		if args[0].Type != value.Channel {
			return value.NullValue(), fmt.Errorf("send: first argument must be channel")
		}
		ch, _ := args[0].Data.(chan value.Value)
		ch <- args[1]
		return value.True(), nil
	})

	g.Register("recv", func(args []value.Value) (value.Value, error) {
		if len(args) != 1 {
			return value.NullValue(), fmt.Errorf("recv: want 1 argument")
		}
		if args[0].Type != value.Channel {
			return value.NullValue(), fmt.Errorf("recv: argument must be channel")
		}
		ch, _ := args[0].Data.(chan value.Value)
		return <-ch, nil
	})
}

// valuesEqual compares two values for equality.
func valuesEqual(a, b value.Value) bool {
	if a.Type != b.Type {
		return false
	}

	switch a.Type {
	case value.Null:
		return true
	case value.Number:
		av, _ := a.AsNumber()
		bv, _ := b.AsNumber()
		return av == bv
	case value.Boolean:
		av, _ := a.AsBool()
		bv, _ := b.AsBool()
		return av == bv
	case value.String:
		av, _ := a.AsString()
		bv, _ := b.AsString()
		return av == bv
	case value.Symbol:
		av, _ := a.Data.(string)
		bv, _ := b.Data.(string)
		return av == bv
	case value.Rune:
		av, _ := a.Data.(rune)
		bv, _ := b.Data.(rune)
		return av == bv
	case value.List:
		al, _ := a.AsList()
		bl, _ := b.AsList()
		if len(al) != len(bl) {
			return false
		}
		for i := range al {
			if !valuesEqual(al[i], bl[i]) {
				return false
			}
		}
		return true
	default:
		return false
	}
}
