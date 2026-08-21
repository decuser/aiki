package substrate

import (
	"math"
	"math/big"

	"aiki/engine/runtime/hal"
	"aiki/engine/semantics/value"
)

func halFloor(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("floor: want 1 argument, got %d", len(args))
	}
	n, ok := args[0].(*value.Number)
	if !ok {
		return value.NewFault("floor: expected number")
	}
	num := n.Numerator()
	denom := n.Denominator()
	q := new(big.Int)
	rem := new(big.Int)
	q.QuoRem(num, denom, rem)
	if num.Sign() < 0 && rem.Sign() != 0 {
		q.Sub(q, big.NewInt(1))
	}
	return value.NewNumberFromBigInt(q)
}

func halCeil(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("ceil: want 1 argument, got %d", len(args))
	}
	n, ok := args[0].(*value.Number)
	if !ok {
		return value.NewFault("ceil: expected number")
	}
	num := n.Numerator()
	denom := n.Denominator()
	q := new(big.Int)
	rem := new(big.Int)
	q.QuoRem(num, denom, rem)
	if num.Sign() > 0 && rem.Sign() != 0 {
		q.Add(q, big.NewInt(1))
	}
	return value.NewNumberFromBigInt(q)
}

func halTruncate(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("truncate: want 1 argument, got %d", len(args))
	}
	n, ok := args[0].(*value.Number)
	if !ok {
		return value.NewFault("truncate: expected number")
	}
	// Truncate toward zero: just integer division
	num := n.Numerator()
	denom := n.Denominator()
	q := new(big.Int).Quo(num, denom)
	return value.NewNumberFromBigInt(q)
}

func halModulo(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("modulo: want 2 arguments, got %d", len(args))
	}
	left, ok1 := args[0].(*value.Number)
	right, ok2 := args[1].(*value.Number)
	if !ok1 || !ok2 {
		return value.NewFault("modulo: expected numbers")
	}
	if !left.IsInt() || !right.IsInt() {
		return value.NewFault("modulo: requires integers")
	}
	if right.Sign() == 0 {
		return value.NewFault("modulo: division by zero")
	}
	l := left.Numerator()
	r := right.Numerator()
	result := new(big.Int).Mod(l, r)
	return value.NewNumberFromBigInt(result)
}

func halSqrt(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("sqrt: want 1 argument, got %d", len(args))
	}
	n, ok := args[0].(*value.Number)
	if !ok {
		return value.NewFault("sqrt: expected number")
	}
	if n.Sign() < 0 {
		return value.NewFault("sqrt: negative number")
	}
	f, exact := n.Float64()
	if !exact && (math.IsInf(f, 0) || math.IsNaN(f)) {
		return value.NewFault("sqrt: argument out of float64 range")
	}
	result := math.Sqrt(f)
	out, ok := value.NewNumberFromFloat64(result)
	if !ok {
		return value.NewFault("sqrt: result is not finite")
	}
	return out
}

func (g *GoRuntime) halSeed(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("seed: want 1 argument, got %d", len(args))
	}
	n, ok := args[0].(*value.Number)
	if !ok {
		return value.NewFault("seed: expected number")
	}
	if !n.IsInt() {
		return value.NewFault("seed: expected integer")
	}
	seed := n.Int64Value()
	g.mu.Lock()
	g.rng.Seed(seed)
	g.mu.Unlock()
	return value.EMPTY
}

func (g *GoRuntime) halRandom(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("random: want 1 argument, got %d", len(args))
	}
	n, ok := args[0].(*value.Number)
	if !ok {
		return value.NewFault("random: expected number")
	}
	if !n.IsInt() {
		return value.NewFault("random: expected integer")
	}
	max := n.Int64Value()
	if max <= 0 {
		return value.NewFault("random: max must be positive")
	}
	g.mu.Lock()
	result := g.rng.Int63n(max)
	g.mu.Unlock()
	return value.NewNumber(result, 1)
}
