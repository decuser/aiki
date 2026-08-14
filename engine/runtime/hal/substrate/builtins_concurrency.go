package substrate

import (
	"aiki/engine"
	"aiki/engine/runtime/hal"
	"aiki/engine/semantics/value"
)

// halChannel creates a new channel.
func halChannel(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 0 {
		return value.NewFault("channel: want 0 arguments, got %d", len(args))
	}
	return value.NewChannel()
}

// halSend sends a value on a channel. Blocks until received.
func halSend(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("send: want 2 arguments, got %d", len(args))
	}
	ch, ok := args[0].(*value.Channel)
	if !ok {
		return value.NewFault("send: first argument must be channel, got %s", args[0].Type())
	}
	if !ch.CanSend() {
		return value.NewFault("send: channel is receive-only")
	}
	semanticHit(ctx, engine.SemanticSend)
	if ctx != nil && ctx.AsyncFault != nil {
		select {
		case ch.C <- args[1]:
			return value.TRUE
		case fault := <-ctx.AsyncFault:
			return fault
		}
	}
	ch.C <- args[1]
	return value.TRUE
}

// halRecv receives a value from a channel. Blocks until sent.
func halRecv(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("recv: want 1 argument, got %d", len(args))
	}
	ch, ok := args[0].(*value.Channel)
	if !ok {
		return value.NewFault("recv: argument must be channel, got %s", args[0].Type())
	}
	if ctx != nil && ctx.AsyncFault != nil {
		select {
		case v := <-ch.C:
			semanticHit(ctx, engine.SemanticReceive)
			return v
		case fault := <-ctx.AsyncFault:
			return fault
		}
	}
	v := <-ch.C
	semanticHit(ctx, engine.SemanticReceive)
	return v
}
