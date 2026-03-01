package substrate

import (
	"math"
	"math/big"
	"math/rand"
	"time"

	"aiki/engine/runtime/hal"
	"aiki/engine/semantics/value"
)

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

func halFloor(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewError("floor: want 1 argument, got %d", len(args))
	}
	n, ok := args[0].(*value.Number)
	if !ok {
		return value.NewError("floor: expected number")
	}
	num := n.Val.Num()
	denom := n.Val.Denom()
	q := new(big.Int).Div(num, denom)
	rem := new(big.Int).Mod(num, denom)
	if num.Sign() < 0 && rem.Sign() != 0 {
		q.Sub(q, big.NewInt(1))
	}
	return &value.Number{Val: new(big.Rat).SetInt(q)}
}

func halCeil(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewError("ceil: want 1 argument, got %d", len(args))
	}
	n, ok := args[0].(*value.Number)
	if !ok {
		return value.NewError("ceil: expected number")
	}
	num := n.Val.Num()
	denom := n.Val.Denom()
	q := new(big.Int).Div(num, denom)
	rem := new(big.Int).Mod(num, denom)
	if num.Sign() > 0 && rem.Sign() != 0 {
		q.Add(q, big.NewInt(1))
	}
	return &value.Number{Val: new(big.Rat).SetInt(q)}
}

func halModulo(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewError("modulo: want 2 arguments, got %d", len(args))
	}
	left, ok1 := args[0].(*value.Number)
	right, ok2 := args[1].(*value.Number)
	if !ok1 || !ok2 {
		return value.NewError("modulo: expected numbers")
	}
	if !left.Val.IsInt() || !right.Val.IsInt() {
		return value.NewError("modulo: requires integers")
	}
	if right.Val.Sign() == 0 {
		return value.NewError("modulo: division by zero")
	}
	l := left.Val.Num()
	r := right.Val.Num()
	result := new(big.Int).Mod(l, r)
	return &value.Number{Val: new(big.Rat).SetInt(result)}
}

func halSqrt(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewError("sqrt: want 1 argument, got %d", len(args))
	}
	n, ok := args[0].(*value.Number)
	if !ok {
		return value.NewError("sqrt: expected number")
	}
	f, _ := n.Val.Float64()
	if f < 0 {
		return value.NewError("sqrt: negative number")
	}
	r := new(big.Rat).SetFloat64(math.Sqrt(f))
	return &value.Number{Val: r}
}

func halSeed(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewError("seed: want 1 argument, got %d", len(args))
	}
	n, ok := args[0].(*value.Number)
	if !ok {
		return value.NewError("seed: expected number")
	}
	if !n.Val.IsInt() {
		return value.NewError("seed: expected integer")
	}
	seed := n.Val.Num().Int64()
	rng = rand.New(rand.NewSource(seed))
	return value.EMPTY
}

func halRandom(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewError("random: want 1 argument, got %d", len(args))
	}
	n, ok := args[0].(*value.Number)
	if !ok {
		return value.NewError("random: expected number")
	}
	if !n.Val.IsInt() {
		return value.NewError("random: expected integer")
	}
	max := n.Val.Num().Int64()
	if max <= 0 {
		return value.NewError("random: max must be positive")
	}
	result := rng.Int63n(max)
	return &value.Number{Val: big.NewRat(result, 1)}
}
