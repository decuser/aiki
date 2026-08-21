package substrate

import (
	"math"
	"time"

	"aiki/engine/runtime/hal"
	"aiki/engine/semantics/value"
)

// halAfter creates a receive-only event channel that produces true once after
// the requested number of milliseconds. The event channel has internal
// capacity one so an abandoned timer never leaves a blocked sender behind.
func halAfter(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("after: want 1 argument, got %d", len(args))
	}
	n, ok := args[0].(*value.Number)
	if !ok {
		return value.NewFault("after: expected number (milliseconds)")
	}
	ms, _ := n.Float64()
	if math.IsInf(ms, 0) || math.IsNaN(ms) {
		return value.NewFault("after: milliseconds out of range")
	}
	if ms < 0 {
		return value.NewFault("after: milliseconds must be non-negative")
	}

	// Convert after scaling so fractional milliseconds retain their duration
	// rather than being truncated to an integer millisecond.
	d := time.Duration(ms * float64(time.Millisecond))
	if d < 0 {
		return value.NewFault("after: milliseconds out of range")
	}

	ch := value.NewEventChannel()
	time.AfterFunc(d, func() {
		ch.C <- value.TRUE
	})
	return ch
}

// halTimeNow returns Unix time in integer milliseconds.
func halTimeNow(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 0 {
		return value.NewFault("_time_now: want 0 arguments, got %d", len(args))
	}
	return value.NewNumber(time.Now().UnixMilli(), 1)
}
