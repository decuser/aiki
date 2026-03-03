package substrate

import (
	"image/color"
	"sync"

	"aiki/engine/runtime/hal"
	"aiki/engine/semantics/value"
)

var (
	openCanvases   []*value.Canvas
	openCanvasesMu sync.Mutex

	// Default colors
	DefaultBG = color.RGBA{0, 0, 0, 255}       // black
	DefaultFG = color.RGBA{255, 255, 255, 255} // white
)

func trackCanvas(c *value.Canvas) {
	openCanvasesMu.Lock()
	openCanvases = append(openCanvases, c)
	openCanvasesMu.Unlock()
}

func untrackCanvas(c *value.Canvas) {
	openCanvasesMu.Lock()
	for i, cvs := range openCanvases {
		if cvs == c {
			openCanvases = append(openCanvases[:i], openCanvases[i+1:]...)
			break
		}
	}
	openCanvasesMu.Unlock()
}

// CloseAllCanvases closes all open canvases.
func CloseAllCanvases() {
	openCanvasesMu.Lock()
	for _, c := range openCanvases {
		select {
		case <-c.Done:
		default:
			close(c.Done)
		}
	}
	openCanvases = nil
	openCanvasesMu.Unlock()
}

func halCanvas(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewError("canvas: want 2 arguments, got %d", len(args))
	}
	width, ok1 := toInt(args[0])
	height, ok2 := toInt(args[1])
	if !ok1 || !ok2 {
		return value.NewError("canvas: width and height must be numbers")
	}

	cvs := &value.Canvas{
		Width:    width,
		Height:   height,
		BG:       DefaultBG,
		FG:       DefaultFG,
		PenSize:  2,
		Commands: make(chan value.CanvasCmd, 100),
		Done:     make(chan struct{}),
		Ready:    make(chan struct{}),
	}

	trackCanvas(cvs)
	go RunEbiten(cvs)
	<-cvs.Ready

	return cvs
}

func halDot(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) < 3 || len(args) > 4 {
		return value.NewError("dot: want 3 or 4 arguments, got %d", len(args))
	}
	cvs, ok := args[0].(*value.Canvas)
	if !ok {
		return value.NewError("dot: expected canvas")
	}
	x, ok1 := toInt(args[1])
	y, ok2 := toInt(args[2])
	if !ok1 || !ok2 {
		return value.NewError("dot: coordinates must be numbers")
	}
	clr := cvs.FG
	if len(args) == 4 {
		clr, ok = parseColor(args[3])
		if !ok {
			return value.NewError("dot: invalid color")
		}
	}
	cvs.Commands <- value.CanvasCmd{Op: "dot", Args: []int{x, y}, Color: clr}
	return value.TRUE
}

func halLine(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) < 5 || len(args) > 6 {
		return value.NewError("line: want 5 or 6 arguments, got %d", len(args))
	}
	cvs, ok := args[0].(*value.Canvas)
	if !ok {
		return value.NewError("line: expected canvas")
	}
	x1, ok1 := toInt(args[1])
	y1, ok2 := toInt(args[2])
	x2, ok3 := toInt(args[3])
	y2, ok4 := toInt(args[4])
	if !ok1 || !ok2 || !ok3 || !ok4 {
		return value.NewError("line: coordinates must be numbers")
	}
	clr := cvs.FG
	if len(args) == 6 {
		clr, ok = parseColor(args[5])
		if !ok {
			return value.NewError("line: invalid color")
		}
	}
	cvs.Commands <- value.CanvasCmd{Op: "line", Args: []int{x1, y1, x2, y2}, Color: clr, PenSize: cvs.PenSize}
	return value.TRUE
}

func halRect(args []value.Value, ctx *hal.EvalContext) value.Value {
	return rectHelper("rect", args)
}

func halFillRect(args []value.Value, ctx *hal.EvalContext) value.Value {
	return rectHelper("fill_rect", args)
}

func rectHelper(op string, args []value.Value) value.Value {
	if len(args) < 5 || len(args) > 6 {
		return value.NewError("%s: want 5 or 6 arguments, got %d", op, len(args))
	}
	cvs, ok := args[0].(*value.Canvas)
	if !ok {
		return value.NewError("%s: expected canvas", op)
	}
	x, ok1 := toInt(args[1])
	y, ok2 := toInt(args[2])
	w, ok3 := toInt(args[3])
	h, ok4 := toInt(args[4])
	if !ok1 || !ok2 || !ok3 || !ok4 {
		return value.NewError("%s: dimensions must be numbers", op)
	}
	clr := cvs.FG
	if len(args) == 6 {
		clr, ok = parseColor(args[5])
		if !ok {
			return value.NewError("%s: invalid color", op)
		}
	}
	cvs.Commands <- value.CanvasCmd{Op: op, Args: []int{x, y, w, h}, Color: clr, PenSize: cvs.PenSize}
	return value.TRUE
}

func halCircle(args []value.Value, ctx *hal.EvalContext) value.Value {
	return circleHelper("circle", args)
}

func halFillCircle(args []value.Value, ctx *hal.EvalContext) value.Value {
	return circleHelper("fill_circle", args)
}

func circleHelper(op string, args []value.Value) value.Value {
	if len(args) < 4 || len(args) > 5 {
		return value.NewError("%s: want 4 or 5 arguments, got %d", op, len(args))
	}
	cvs, ok := args[0].(*value.Canvas)
	if !ok {
		return value.NewError("%s: expected canvas", op)
	}
	x, ok1 := toInt(args[1])
	y, ok2 := toInt(args[2])
	r, ok3 := toInt(args[3])
	if !ok1 || !ok2 || !ok3 {
		return value.NewError("%s: dimensions must be numbers", op)
	}
	clr := cvs.FG
	if len(args) == 5 {
		clr, ok = parseColor(args[4])
		if !ok {
			return value.NewError("%s: invalid color", op)
		}
	}
	cvs.Commands <- value.CanvasCmd{Op: op, Args: []int{x, y, r}, Color: clr, PenSize: cvs.PenSize}
	return value.TRUE
}

func halClear(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewError("clear: want 1 argument, got %d", len(args))
	}
	cvs, ok := args[0].(*value.Canvas)
	if !ok {
		return value.NewError("clear: expected canvas")
	}
	cvs.Commands <- value.CanvasCmd{Op: "clear"}
	return value.TRUE
}

func halDestroy(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewError("destroy: want 1 argument, got %d", len(args))
	}
	cvs, ok := args[0].(*value.Canvas)
	if !ok {
		return value.NewError("destroy: expected canvas")
	}
	select {
	case <-cvs.Done:
	default:
		close(cvs.Done)
	}
	untrackCanvas(cvs)
	return value.TRUE
}

func halSetBG(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewError("set_bg: want 2 arguments, got %d", len(args))
	}
	cvs, ok := args[0].(*value.Canvas)
	if !ok {
		return value.NewError("set_bg: expected canvas")
	}
	clr, ok := parseColor(args[1])
	if !ok {
		return value.NewError("set_bg: invalid color")
	}
	cvs.BG = clr
	return value.TRUE
}

func halSetFG(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewError("set_fg: want 2 arguments, got %d", len(args))
	}
	cvs, ok := args[0].(*value.Canvas)
	if !ok {
		return value.NewError("set_fg: expected canvas")
	}
	clr, ok := parseColor(args[1])
	if !ok {
		return value.NewError("set_fg: invalid color")
	}
	cvs.FG = clr
	return value.TRUE
}

func halPenSize(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewError("pen_size: want 2 arguments, got %d", len(args))
	}
	cvs, ok := args[0].(*value.Canvas)
	if !ok {
		return value.NewError("pen_size: expected canvas")
	}
	size, ok := toInt(args[1])
	if !ok || size < 1 {
		return value.NewError("pen_size: size must be a positive number")
	}
	cvs.PenSize = float32(size)
	return value.TRUE
}

// Helper functions

func toInt(v value.Value) (int, bool) {
	n, ok := v.(*value.Number)
	if !ok {
		return 0, false
	}
	f, _ := n.Val.Float64()
	return int(f), true
}

func parseColor(v value.Value) (color.RGBA, bool) {
	switch c := v.(type) {
	case *value.Symbol:
		return colorFromName(c.Val)
	case *value.List:
		if len(c.Elements) >= 3 {
			r, ok1 := toInt(c.Elements[0])
			g, ok2 := toInt(c.Elements[1])
			b, ok3 := toInt(c.Elements[2])
			a := 255
			if len(c.Elements) >= 4 {
				a, _ = toInt(c.Elements[3])
			}
			if ok1 && ok2 && ok3 {
				return color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)}, true
			}
		}
	}
	return color.RGBA{}, false
}

func colorFromName(name string) (color.RGBA, bool) {
	colors := map[string]color.RGBA{
		"black":   {0, 0, 0, 255},
		"white":   {255, 255, 255, 255},
		"red":     {255, 0, 0, 255},
		"green":   {0, 255, 0, 255},
		"blue":    {0, 0, 255, 255},
		"yellow":  {255, 255, 0, 255},
		"cyan":    {0, 255, 255, 255},
		"magenta": {255, 0, 255, 255},
		"orange":  {255, 165, 0, 255},
		"purple":  {128, 0, 128, 255},
		"pink":    {255, 192, 203, 255},
		"gray":    {128, 128, 128, 255},
		"grey":    {128, 128, 128, 255},
	}
	c, ok := colors[name]
	return c, ok
}
