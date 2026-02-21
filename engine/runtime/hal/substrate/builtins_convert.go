package substrate

import (
	"fmt"

	"aiki/engine/semantics/value"
)

func (g *GoRuntime) registerConvert() {
	// Type - returns the type of a value as a symbol
	g.Register("type", func(args []value.Value) (value.Value, error) {
		if len(args) != 1 {
			return value.NullValue(), fmt.Errorf("type: want 1 argument")
		}
		return value.NewSymbol(value.TypeName(args[0].Type)), nil
	})

	// Inspect - returns string representation of a value
	g.Register("inspect", func(args []value.Value) (value.Value, error) {
		if len(args) != 1 {
			return value.NullValue(), fmt.Errorf("inspect: want 1 argument")
		}
		return value.NewString(args[0].Inspect()), nil
	})

	// Shape - get shape of a value
	g.Register("shape", func(args []value.Value) (value.Value, error) {
		if len(args) != 1 {
			return value.NullValue(), fmt.Errorf("shape: want 1 argument")
		}
		return value.NewSymbol(value.TypeName(args[0].Type)), nil
	})

	// Ord - get character code from rune/string
	g.Register("ord", func(args []value.Value) (value.Value, error) {
		if len(args) != 1 {
			return value.NullValue(), fmt.Errorf("ord: want 1 argument")
		}
		switch args[0].Type {
		case value.Rune:
			r := args[0].Data.(rune)
			return value.NewNumber(float64(r)), nil
		case value.String:
			s := args[0].Data.(string)
			if len(s) == 0 {
				return value.NewNumber(0), nil
			}
			runes := []rune(s)
			return value.NewNumber(float64(runes[0])), nil
		default:
			return value.NullValue(), fmt.Errorf("ord: expected rune or string")
		}
	})

	// to_str - convert to string
	g.Register("to_str", func(args []value.Value) (value.Value, error) {
		if len(args) != 1 {
			return value.NullValue(), fmt.Errorf("to_str: want 1 argument")
		}
		return value.NewString(args[0].Inspect()), nil
	})

	// to_number - convert string to number
	g.Register("to_number", func(args []value.Value) (value.Value, error) {
		if len(args) != 1 {
			return value.NullValue(), fmt.Errorf("to_number: want 1 argument")
		}
		if args[0].Type != value.String {
			return value.NullValue(), fmt.Errorf("to_number: expected string")
		}
		s := args[0].Data.(string)
		var f float64
		_, err := fmt.Sscanf(s, "%f", &f)
		if err != nil {
			return value.NullValue(), fmt.Errorf("to_number: invalid number %q", s)
		}
		return value.NewNumber(f), nil
	})

	// to_rune - convert number to rune
	g.Register("to_rune", func(args []value.Value) (value.Value, error) {
		if len(args) != 1 {
			return value.NullValue(), fmt.Errorf("to_rune: want 1 argument")
		}
		if args[0].Type != value.Number {
			return value.NullValue(), fmt.Errorf("to_rune: expected number")
		}
		n := int(args[0].Data.(float64))
		return value.Value{Type: value.Rune, Data: rune(n)}, nil
	})

	// to_bytes - convert string to byte list
	g.Register("to_bytes", func(args []value.Value) (value.Value, error) {
		if len(args) != 1 {
			return value.NullValue(), fmt.Errorf("to_bytes: want 1 argument")
		}
		if args[0].Type != value.String {
			return value.NullValue(), fmt.Errorf("to_bytes: expected string")
		}
		s := args[0].Data.(string)
		bytes := []byte(s)
		list := make([]value.Value, len(bytes))
		for i, b := range bytes {
			list[i] = value.NewNumber(float64(b))
		}
		return value.NewList(list), nil
	})

	// to_decimal - convert number to decimal string
	g.Register("to_decimal", func(args []value.Value) (value.Value, error) {
		if len(args) != 1 {
			return value.NullValue(), fmt.Errorf("to_decimal: want 1 argument")
		}
		if args[0].Type != value.Number {
			return value.NullValue(), fmt.Errorf("to_decimal: expected number")
		}
		f := args[0].Data.(float64)
		return value.NewString(fmt.Sprintf("%g", f)), nil
	})
}
