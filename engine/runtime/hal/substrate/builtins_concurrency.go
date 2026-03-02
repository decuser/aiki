package substrate

import (
	"aiki/engine/runtime/hal"
	"aiki/engine/semantics/value"
)

// halChannel creates a new channel.
func halChannel(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 0 {
		return value.NewError("channel: want 0 arguments, got %d", len(args))
	}
	return value.NewChannel()
}

// halSend sends a value on a channel. Blocks until received.
func halSend(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewError("send: want 2 arguments, got %d", len(args))
	}
	ch, ok := args[0].(*value.Channel)
	if !ok {
		return value.NewError("send: first argument must be channel, got %s", args[0].Type())
	}
	ch.C <- args[1]
	return value.TRUE
}

// halRecv receives a value from a channel. Blocks until sent.
func halRecv(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewError("recv: want 1 argument, got %d", len(args))
	}
	ch, ok := args[0].(*value.Channel)
	if !ok {
		return value.NewError("recv: argument must be channel, got %s", args[0].Type())
	}
	return <-ch.C
}
