package substrate

import (
	"math/big"
	"strings"
	"unicode"

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

// halToStr converts a value to its string representation.
// Unlike Inspect(), this returns the value content, not display format.
func halToStr(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("to_str: want 1 argument, got %d", len(args))
	}
	switch v := args[0].(type) {
	case *value.Rune:
		return &value.String{Val: string(v.Val)}
	case *value.String:
		return v
	case *value.Number:
		return &value.String{Val: v.Val.RatString()}
	case *value.Boolean:
		if v.Val {
			return &value.String{Val: "true"}
		}
		return &value.String{Val: "false"}
	case *value.Symbol:
		return &value.String{Val: ":" + v.Val}
	default:
		return &value.String{Val: args[0].Inspect()}
	}
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
	if places.Val.Sign() < 0 {
		return value.NewFault("to_decimal: places must not be negative")
	}
	if !places.Val.Num().IsInt64() {
		return value.NewFault("to_decimal: places out of range")
	}
	p := int(places.Val.Num().Int64())

	// The exact rational is formatted by long division. No floating-point
	// value appears in this path: a float64 round-trip would misreport
	// digits for magnitudes, place counts, or denominators outside its range.
	neg := n.Val.Sign() < 0
	num := new(big.Int).Abs(n.Val.Num())
	den := new(big.Int).Set(n.Val.Denom())

	scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(p)), nil)
	scaled := new(big.Int).Mul(num, scale)
	quo := new(big.Int)
	rem := new(big.Int)
	quo.QuoRem(scaled, den, rem)

	// Round half away from zero, decided on the exact remainder.
	twice := new(big.Int).Lsh(rem, 1)
	if twice.Cmp(den) >= 0 {
		quo.Add(quo, big.NewInt(1))
	}

	digits := quo.String()
	out := digits
	if p > 0 {
		if len(digits) <= p {
			digits = strings.Repeat("0", p+1-len(digits)) + digits
		}
		out = digits[:len(digits)-p] + "." + digits[len(digits)-p:]
	}
	if neg {
		out = "-" + out
	}
	return &value.String{Val: out}
}

// halToSymbol constructs a symbol from its value-level name.
func halToSymbol(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("to_symbol: want 1 argument, got %d", len(args))
	}
	s, ok := args[0].(*value.String)
	if !ok {
		return value.NewFault("to_symbol: expected string")
	}
	return &value.Symbol{Val: s.Val}
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

// halChars decomposes a string into its Unicode rune values in one pass.
// The public surface remains lib/string.chars; this is its substrate primitive.
func halChars(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("chars: want 1 argument, got %d", len(args))
	}
	s, ok := args[0].(*value.String)
	if !ok {
		return value.NewFault("chars: expected string")
	}
	runes := []rune(s.Val)
	elems := make([]value.Value, len(runes))
	for i, r := range runes {
		elems[i] = &value.Rune{Val: r}
	}
	return &value.List{Elements: elems}
}

// halUpper converts a string to uppercase using Unicode rules.
func halUpper(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("upper: want 1 argument, got %d", len(args))
	}
	s, ok := args[0].(*value.String)
	if !ok {
		return value.NewFault("upper: expected string")
	}
	return &value.String{Val: strings.ToUpper(s.Val)}
}

// halLower converts a string to lowercase using Unicode rules.
func halLower(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("lower: want 1 argument, got %d", len(args))
	}
	s, ok := args[0].(*value.String)
	if !ok {
		return value.NewFault("lower: expected string")
	}
	return &value.String{Val: strings.ToLower(s.Val)}
}

// halUpperRune converts a rune to uppercase using Unicode rules.
func halUpperRune(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("upper_rune: want 1 argument, got %d", len(args))
	}
	r, ok := args[0].(*value.Rune)
	if !ok {
		return value.NewFault("upper_rune: expected rune")
	}
	return &value.Rune{Val: unicode.ToUpper(r.Val)}
}

// halLowerRune converts a rune to lowercase using Unicode rules.
func halLowerRune(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("lower_rune: want 1 argument, got %d", len(args))
	}
	r, ok := args[0].(*value.Rune)
	if !ok {
		return value.NewFault("lower_rune: expected rune")
	}
	return &value.Rune{Val: unicode.ToLower(r.Val)}
}
