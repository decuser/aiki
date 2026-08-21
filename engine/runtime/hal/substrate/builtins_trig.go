package substrate

import (
	"math"
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
	f, exact := n.Float64()
	if !exact && (math.IsInf(f, 0) || math.IsNaN(f)) {
		return value.NewFault("cos: argument out of float64 range")
	}
	result := math.Cos(f)
	out, ok := value.NewNumberFromFloat64(result)
	if !ok {
		return value.NewFault("cos: result is not finite")
	}
	return out
}

func halSin(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("sin: want 1 argument, got %d", len(args))
	}
	n, ok := args[0].(*value.Number)
	if !ok {
		return value.NewFault("sin: expected number")
	}
	f, exact := n.Float64()
	if !exact && (math.IsInf(f, 0) || math.IsNaN(f)) {
		return value.NewFault("sin: argument out of float64 range")
	}
	result := math.Sin(f)
	out, ok := value.NewNumberFromFloat64(result)
	if !ok {
		return value.NewFault("sin: result is not finite")
	}
	return out
}

func halSleep(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("sleep: want 1 argument, got %d", len(args))
	}
	n, ok := args[0].(*value.Number)
	if !ok {
		return value.NewFault("sleep: expected number (milliseconds)")
	}
	ms, _ := n.Float64()
	time.Sleep(time.Duration(ms) * time.Millisecond)
	return value.TRUE
}
