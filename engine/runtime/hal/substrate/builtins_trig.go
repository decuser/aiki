package substrate

import (
	"math"
	"math/big"
	"time"

	"aiki/engine/runtime/hal"
	"aiki/engine/semantics/value"
)

func halCos(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("cos: want 1 argument, got %d", len(args))
	}
	n, ok := args[0].(*value.Number)
	if !ok {
		return value.NewFault("cos: expected number")
	}
	f, _ := n.Val.Float64()
	r := new(big.Rat).SetFloat64(math.Cos(f))
	return &value.Number{Val: r}
}

func halSin(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("sin: want 1 argument, got %d", len(args))
	}
	n, ok := args[0].(*value.Number)
	if !ok {
		return value.NewFault("sin: expected number")
	}
	f, _ := n.Val.Float64()
	r := new(big.Rat).SetFloat64(math.Sin(f))
	return &value.Number{Val: r}
}

func halSleep(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("sleep: want 1 argument, got %d", len(args))
	}
	n, ok := args[0].(*value.Number)
	if !ok {
		return value.NewFault("sleep: expected number (milliseconds)")
	}
	ms, _ := n.Val.Float64()
	time.Sleep(time.Duration(ms) * time.Millisecond)
	return value.TRUE
}
