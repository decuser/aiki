package substrate

import (
	"math/big"

	"aiki/engine/runtime/hal"
	"aiki/engine/semantics/value"
)

func bitsNonNegativeInteger(v value.Value, what string) (*big.Int, *value.Fault) {
	n, ok := v.(*value.Number)
	if !ok || !n.Val.IsInt() {
		return nil, value.NewFault("%s must be a non-negative integer", what)
	}
	z := new(big.Int).Set(n.Val.Num())
	if z.Sign() < 0 {
		return nil, value.NewFault("%s must be a non-negative integer", what)
	}
	return z, nil
}

func bitsShift(v value.Value, what string) (uint, *value.Fault) {
	z, fault := bitsNonNegativeInteger(v, what)
	if fault != nil {
		return 0, fault
	}
	if !z.IsUint64() {
		return 0, value.NewFault("%s is too large", what)
	}
	n := z.Uint64()
	if uint64(uint(n)) != n {
		return 0, value.NewFault("%s is too large", what)
	}
	return uint(n), nil
}

func bitsNumber(z *big.Int) value.Value {
	return &value.Number{Val: new(big.Rat).SetInt(z)}
}

func halBitsAnd(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("bits.bit_and: want 2 arguments, got %d", len(args))
	}
	a, fault := bitsNonNegativeInteger(args[0], "bits.bit_and: a")
	if fault != nil {
		return fault
	}
	b, fault := bitsNonNegativeInteger(args[1], "bits.bit_and: b")
	if fault != nil {
		return fault
	}
	return bitsNumber(new(big.Int).And(a, b))
}

func halBitsOr(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("bits.bit_or: want 2 arguments, got %d", len(args))
	}
	a, fault := bitsNonNegativeInteger(args[0], "bits.bit_or: a")
	if fault != nil {
		return fault
	}
	b, fault := bitsNonNegativeInteger(args[1], "bits.bit_or: b")
	if fault != nil {
		return fault
	}
	return bitsNumber(new(big.Int).Or(a, b))
}

func halBitsXor(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("bits.bit_xor: want 2 arguments, got %d", len(args))
	}
	a, fault := bitsNonNegativeInteger(args[0], "bits.bit_xor: a")
	if fault != nil {
		return fault
	}
	b, fault := bitsNonNegativeInteger(args[1], "bits.bit_xor: b")
	if fault != nil {
		return fault
	}
	return bitsNumber(new(big.Int).Xor(a, b))
}

func halBitsNot(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("bits.bit_not: want 2 arguments, got %d", len(args))
	}
	a, fault := bitsNonNegativeInteger(args[0], "bits.bit_not: a")
	if fault != nil {
		return fault
	}
	width, fault := bitsShift(args[1], "bits.bit_not: width")
	if fault != nil {
		return fault
	}
	if width == 0 {
		return value.NewNumber(0, 1)
	}
	mask := new(big.Int).Lsh(big.NewInt(1), width)
	mask.Sub(mask, big.NewInt(1))
	low := new(big.Int).And(a, mask)
	return bitsNumber(new(big.Int).Xor(mask, low))
}

func halBitsShl(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("bits.shl: want 2 arguments, got %d", len(args))
	}
	a, fault := bitsNonNegativeInteger(args[0], "bits.shl: a")
	if fault != nil {
		return fault
	}
	n, fault := bitsShift(args[1], "bits.shl: count")
	if fault != nil {
		return fault
	}
	return bitsNumber(new(big.Int).Lsh(a, n))
}

func halBitsShr(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("bits.shr: want 2 arguments, got %d", len(args))
	}
	a, fault := bitsNonNegativeInteger(args[0], "bits.shr: a")
	if fault != nil {
		return fault
	}
	n, fault := bitsShift(args[1], "bits.shr: count")
	if fault != nil {
		return fault
	}
	return bitsNumber(new(big.Int).Rsh(a, n))
}
