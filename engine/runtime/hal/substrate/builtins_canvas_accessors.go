package substrate

import (
	"image/color"

	"aiki/engine/runtime/hal"
	"aiki/engine/semantics/value"
)

func (g *GoRuntime) halCanvasWidth(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("canvas_width: want 1 argument, got %d", len(args))
	}
	cvs, ok := args[0].(*value.Canvas)
	if !ok {
		return value.NewFault("canvas_width: expected canvas")
	}
	resource, ok := g.canvasResource(cvs)
	if !ok {
		return value.NewShapedError("canvas", "canvas_width: canvas closed")
	}
	return value.NewNumber(int64(resource.Width), 1)
}

func (g *GoRuntime) halCanvasHeight(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("canvas_height: want 1 argument, got %d", len(args))
	}
	cvs, ok := args[0].(*value.Canvas)
	if !ok {
		return value.NewFault("canvas_height: expected canvas")
	}
	resource, ok := g.canvasResource(cvs)
	if !ok {
		return value.NewShapedError("canvas", "canvas_height: canvas closed")
	}
	return value.NewNumber(int64(resource.Height), 1)
}

func (g *GoRuntime) halCanvasAlive(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("canvas_alive: want 1 argument, got %d", len(args))
	}
	cvs, ok := args[0].(*value.Canvas)
	if !ok {
		return value.FALSE
	}
	resource, ok := g.canvasResource(cvs)
	if !ok {
		return value.FALSE
	}
	select {
	case <-resource.Done:
		return value.FALSE
	default:
		return value.TRUE
	}
}

// halSetTurtle sets the turtle overlay state through the owning Canvas resource.
func (g *GoRuntime) halSetTurtle(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 6 {
		return value.NewFault("set_turtle: want 6 arguments, got %d", len(args))
	}
	cvs, ok := args[0].(*value.Canvas)
	if !ok {
		return value.NewFault("set_turtle: expected canvas")
	}
	resource, ok := g.canvasResource(cvs)
	if !ok || !canvasSessionAlive(resource) {
		return value.NewShapedError("canvas", "set_turtle: canvas closed")
	}
	x, ok := args[1].(*value.Number)
	if !ok {
		return value.NewFault("set_turtle: x must be number")
	}
	y, ok := args[2].(*value.Number)
	if !ok {
		return value.NewFault("set_turtle: y must be number")
	}
	heading, ok := args[3].(*value.Number)
	if !ok {
		return value.NewFault("set_turtle: heading must be number")
	}
	if args[4] != value.TRUE && args[4] != value.FALSE {
		return value.NewFault("set_turtle: visible must be boolean")
	}
	visible := args[4] == value.TRUE
	sym, ok := args[5].(*value.Symbol)
	if !ok {
		return value.NewFault("set_turtle: color must be symbol")
	}
	clr := symbolToColor(sym.Val)

	xf, _ := x.Val.Float64()
	yf, _ := y.Val.Float64()
	hf, _ := heading.Val.Float64()

	SendCanvasTurtle(resource, xf, yf, hf, visible, clr)
	return value.TRUE
}

func symbolToColor(name string) color.RGBA {
	switch name {
	case "white":
		return color.RGBA{255, 255, 255, 255}
	case "black":
		return color.RGBA{0, 0, 0, 255}
	case "red":
		return color.RGBA{255, 0, 0, 255}
	case "green":
		return color.RGBA{0, 255, 0, 255}
	case "blue":
		return color.RGBA{0, 0, 255, 255}
	case "yellow":
		return color.RGBA{255, 255, 0, 255}
	case "cyan":
		return color.RGBA{0, 255, 255, 255}
	case "magenta":
		return color.RGBA{255, 0, 255, 255}
	case "orange":
		return color.RGBA{255, 165, 0, 255}
	case "purple":
		return color.RGBA{128, 0, 128, 255}
	case "pink":
		return color.RGBA{255, 192, 203, 255}
	case "gray", "grey":
		return color.RGBA{128, 128, 128, 255}
	default:
		return color.RGBA{255, 255, 255, 255}
	}
}
