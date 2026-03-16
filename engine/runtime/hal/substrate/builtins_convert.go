package substrate

import (
	"fmt"

	"aiki/engine/runtime/hal"
	"aiki/engine/semantics/value"
)

// halShape returns the shape of a list, or :list if unshaped.
func halShape(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("shape: want 1 argument, got %d", len(args))
	}
	list, ok := args[0].(*value.List)
	if !ok {
		return &value.Symbol{Val: "list"}
	}
	if list.Shape == "" {
		return &value.Symbol{Val: "list"}
	}
	return &value.Symbol{Val: list.Shape}
}

// halToStr converts any value to a string.
func halToStr(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("to_str: want 1 argument, got %d", len(args))
	}
	return &value.String{Val: args[0].Inspect()}
}

// halToDecimal formats a number with specified decimal places.
func halToDecimal(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("to_decimal: want 2 arguments, got %d", len(args))
	}
	n, ok := args[0].(*value.Number)
	if !ok {
		return value.NewFault("to_decimal: first argument must be number")
	}
	places, ok := args[1].(*value.Number)
	if !ok || !places.Val.IsInt() {
		return value.NewFault("to_decimal: second argument must be integer")
	}
	p := int(places.Val.Num().Int64())
	f, _ := n.Val.Float64()
	format := fmt.Sprintf("%%.%df", p)
	return &value.String{Val: fmt.Sprintf(format, f)}
}

// halToNumber parses a string into a number.
func halToNumber(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("to_number: want 1 argument, got %d", len(args))
	}
	s, ok := args[0].(*value.String)
	if !ok {
		return value.NewFault("to_number: expected string")
	}
	n, err := value.NewNumberFromString(s.Val)
	if err != nil {
		return value.NewShapedError("parse", "to_number: invalid number: %s", s.Val)
	}
	return n
}

// halChr converts an integer code point to a rune.
func halChr(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("chr: want 1 argument, got %d", len(args))
	}
	n, ok := args[0].(*value.Number)
	if !ok || !n.Val.IsInt() {
		return value.NewFault("chr: expected integer")
	}
	code := n.Val.Num().Int64()
	if code < 0 || code > 0x10FFFF {
		return value.NewShapedError("range", "chr: code point out of range: %d", code)
	}
	return &value.Rune{Val: rune(code)}
}
