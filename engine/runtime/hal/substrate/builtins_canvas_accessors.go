package substrate

import (
	"aiki/engine/runtime/hal"
	"aiki/engine/semantics/value"
)

func halCanvasWidth(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("canvas_width: want 1 argument, got %d", len(args))
	}
	cvs, ok := args[0].(*value.Canvas)
	if !ok {
		return value.NewFault("canvas_width: expected canvas")
	}
	return value.NewNumber(int64(cvs.Width), 1)
}

func halCanvasHeight(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("canvas_height: want 1 argument, got %d", len(args))
	}
	cvs, ok := args[0].(*value.Canvas)
	if !ok {
		return value.NewFault("canvas_height: expected canvas")
	}
	return value.NewNumber(int64(cvs.Height), 1)
}
