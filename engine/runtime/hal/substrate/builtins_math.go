package substrate

import (
	"math/big"

	"aiki/engine/semantics/value"
)

func halFloor(args []value.Value) value.Value {
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

func halCeil(args []value.Value) value.Value {
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
