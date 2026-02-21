package substrate

import (
	"fmt"
	"math"
	"math/rand"

	"aiki/engine/semantics/value"
)

func (g *GoRuntime) registerMath() {
	// Modulo - remainder after division
	g.Register("modulo", func(args []value.Value) (value.Value, error) {
		if len(args) != 2 {
			return value.NullValue(), fmt.Errorf("modulo: want 2 arguments")
		}
		if args[0].Type != value.Number || args[1].Type != value.Number {
			return value.NullValue(), fmt.Errorf("modulo: arguments must be numbers")
		}
		a := int(args[0].Data.(float64))
		b := int(args[1].Data.(float64))
		if b == 0 {
			return value.NullValue(), fmt.Errorf("modulo: division by zero")
		}
		return value.NewNumber(float64(a % b)), nil
	})

	// Sqrt - square root
	g.Register("sqrt", func(args []value.Value) (value.Value, error) {
		if len(args) != 1 {
			return value.NullValue(), fmt.Errorf("sqrt: want 1 argument")
		}
		if args[0].Type != value.Number {
			return value.NullValue(), fmt.Errorf("sqrt: expected number")
		}
		f := args[0].Data.(float64)
		if f < 0 {
			return value.NullValue(), fmt.Errorf("sqrt: negative number")
		}
		return value.NewNumber(math.Sqrt(f)), nil
	})

	// Sin - sine
	g.Register("sin", func(args []value.Value) (value.Value, error) {
		if len(args) != 1 {
			return value.NullValue(), fmt.Errorf("sin: want 1 argument")
		}
		if args[0].Type != value.Number {
			return value.NullValue(), fmt.Errorf("sin: expected number")
		}
		f := args[0].Data.(float64)
		return value.NewNumber(math.Sin(f)), nil
	})

	// Cos - cosine
	g.Register("cos", func(args []value.Value) (value.Value, error) {
		if len(args) != 1 {
			return value.NullValue(), fmt.Errorf("cos: want 1 argument")
		}
		if args[0].Type != value.Number {
			return value.NullValue(), fmt.Errorf("cos: expected number")
		}
		f := args[0].Data.(float64)
		return value.NewNumber(math.Cos(f)), nil
	})

	// Abs - absolute value
	g.Register("abs", func(args []value.Value) (value.Value, error) {
		if len(args) != 1 {
			return value.NullValue(), fmt.Errorf("abs: want 1 argument")
		}
		if args[0].Type != value.Number {
			return value.NullValue(), fmt.Errorf("abs: expected number")
		}
		f := args[0].Data.(float64)
		return value.NewNumber(math.Abs(f)), nil
	})

	// Random - random integer in [0, max)
	g.Register("random", func(args []value.Value) (value.Value, error) {
		if len(args) != 1 {
			return value.NullValue(), fmt.Errorf("random: want 1 argument")
		}
		if args[0].Type != value.Number {
			return value.NullValue(), fmt.Errorf("random: expected number")
		}
		max := int64(args[0].Data.(float64))
		if max <= 0 {
			return value.NullValue(), fmt.Errorf("random: max must be positive")
		}
		return value.NewNumber(float64(rand.Int63n(max))), nil
	})
}
