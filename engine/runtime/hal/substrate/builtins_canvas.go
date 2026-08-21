package substrate

import (
	"image/color"
	"math"

	"aiki/engine/runtime/hal"
	"aiki/engine/semantics/value"
)

var (
	// Default colors
	DefaultBG = color.RGBA{0, 0, 0, 255}       // black
	DefaultFG = color.RGBA{255, 255, 255, 255} // white
)

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

func (g *GoRuntime) trackCanvas(c *value.Canvas, resource *CanvasResource) {
	g.mu.Lock()
	g.openCanvases = append(g.openCanvases, c)
	g.canvasResources[c] = resource
	g.mu.Unlock()
}

func (g *GoRuntime) canvasResource(c *value.Canvas) (*CanvasResource, bool) {
	g.mu.RLock()
	resource, ok := g.canvasResources[c]
	g.mu.RUnlock()
	return resource, ok
}

func (g *GoRuntime) canvasResourceMust(c *value.Canvas) *CanvasResource {
	resource, _ := g.canvasResource(c)
	return resource
}

func (g *GoRuntime) untrackCanvas(c *value.Canvas) {
	g.mu.Lock()
	delete(g.canvasResources, c)
	for i, cvs := range g.openCanvases {
		if cvs == c {
			g.openCanvases = append(g.openCanvases[:i], g.openCanvases[i+1:]...)
			break
		}
	}
	g.mu.Unlock()
}

func (g *GoRuntime) canvasFG(c *value.Canvas) color.RGBA         { return g.canvasResourceMust(c).FG }
func (g *GoRuntime) canvasPenSize(c *value.Canvas) float32       { return g.canvasResourceMust(c).PenSize }
func (g *GoRuntime) setCanvasBG(c *value.Canvas, clr color.RGBA) { g.canvasResourceMust(c).BG = clr }
func (g *GoRuntime) setCanvasFG(c *value.Canvas, clr color.RGBA) { g.canvasResourceMust(c).FG = clr }
func (g *GoRuntime) setCanvasPenSize(c *value.Canvas, size float32) {
	g.canvasResourceMust(c).PenSize = size
}
func (g *GoRuntime) enqueueCanvas(c *value.Canvas, cmd CanvasCmd) {
	g.canvasResourceMust(c).Commands <- cmd
}

// CloseAllCanvases closes all canvases owned by this runtime.
func (g *GoRuntime) CloseAllCanvases() {
	g.mu.Lock()
	cs := append([]*value.Canvas(nil), g.openCanvases...)
	g.openCanvases = nil
	resources := make([]*CanvasResource, 0, len(cs))
	for _, c := range cs {
		if resource := g.canvasResources[c]; resource != nil {
			resources = append(resources, resource)
		}
		delete(g.canvasResources, c)
	}
	g.mu.Unlock()

	for _, resource := range resources {
		select {
		case <-resource.Done:
		default:
			close(resource.Done)
		}
		bridgeWait(resource)
	}
}

func (g *GoRuntime) halCanvas(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("canvas: want 2 arguments, got %d", len(args))
	}
	width, ok1 := toInt(args[0])
	height, ok2 := toInt(args[1])
	if !ok1 || !ok2 {
		return value.NewFault("canvas: width and height must be numbers")
	}

	g.mu.Lock()
	g.nextCanvasID++
	id := g.nextCanvasID
	g.mu.Unlock()
	cvs := &value.Canvas{ID: id}
	resource := &CanvasResource{
		Width: width, Height: height, BG: DefaultBG, FG: DefaultFG, PenSize: 2,
		Commands: make(chan CanvasCmd, 100), Done: make(chan struct{}), Ready: make(chan struct{}),
	}

	g.trackCanvas(cvs, resource)
	if err := startCanvasSession(resource); err != nil {
		g.untrackCanvas(cvs)
		return value.NewShapedError("canvas", "canvas: %v", err)
	}
	<-resource.Ready

	return cvs
}

// halCanvasCommand realizes the Aiki-defined Canvas protocol through one host
// crossing. The operation symbol and argument list are constructed in Aiki;
// this adapter maps that protocol onto the existing substrate implementation.
func (g *GoRuntime) halCanvasCommand(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 3 {
		return value.NewFault("canvas_command: want 3 arguments, got %d", len(args))
	}
	cvs, ok := args[0].(*value.Canvas)
	if !ok {
		return value.NewFault("canvas_command: expected canvas")
	}
	op, ok := args[1].(*value.Symbol)
	if !ok {
		return value.NewFault("canvas_command: operation must be symbol")
	}
	params, ok := args[2].(*value.List)
	if !ok {
		return value.NewFault("canvas_command: arguments must be list")
	}
	call := append([]value.Value{cvs}, params.Elements...)
	switch op.Val {
	case "dot":
		return g.halDot(call, ctx)
	case "line":
		return g.halLine(call, ctx)
	case "rect":
		return g.halRect(call, ctx)
	case "fill_rect":
		return g.halFillRect(call, ctx)
	case "circle":
		return g.halCircle(call, ctx)
	case "fill_circle":
		return g.halFillCircle(call, ctx)
	case "arc":
		return g.halArc(call, ctx)
	case "clear":
		return g.halClear(call, ctx)
	case "set_bg":
		return g.halSetBG(call, ctx)
	case "set_fg":
		return g.halSetFG(call, ctx)
	case "pen_size":
		return g.halPenSize(call, ctx)
	case "turtle":
		return g.halSetTurtle(call, ctx)
	default:
		return value.NewFault("canvas_command: unknown operation: %s", op.Val)
	}
}

func (g *GoRuntime) halDot(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) < 3 || len(args) > 4 {
		return value.NewFault("dot: want 3 or 4 arguments, got %d", len(args))
	}
	cvs, ok := args[0].(*value.Canvas)
	if !ok {
		return value.NewFault("dot: expected canvas")
	}
	if errv := g.requireCanvasActive("dot", cvs); errv != nil {
		return errv
	}
	x, ok1 := toInt(args[1])
	y, ok2 := toInt(args[2])
	if !ok1 || !ok2 {
		return value.NewFault("dot: coordinates must be numbers")
	}
	clr := g.canvasFG(cvs)
	if len(args) == 4 {
		clr, ok = parseColor(args[3])
		if !ok {
			return value.NewFault("dot: invalid color")
		}
	}
	g.enqueueCanvas(cvs, CanvasCmd{Op: "dot", Args: []int{x, y}, Color: clr})
	return value.TRUE
}

func (g *GoRuntime) halLine(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) < 5 || len(args) > 6 {
		return value.NewFault("line: want 5 or 6 arguments, got %d", len(args))
	}
	cvs, ok := args[0].(*value.Canvas)
	if !ok {
		return value.NewFault("line: expected canvas")
	}
	if errv := g.requireCanvasActive("line", cvs); errv != nil {
		return errv
	}
	x1, ok1 := toInt(args[1])
	y1, ok2 := toInt(args[2])
	x2, ok3 := toInt(args[3])
	y2, ok4 := toInt(args[4])
	if !ok1 || !ok2 || !ok3 || !ok4 {
		return value.NewFault("line: coordinates must be numbers")
	}
	clr := g.canvasFG(cvs)
	if len(args) == 6 {
		clr, ok = parseColor(args[5])
		if !ok {
			return value.NewFault("line: invalid color")
		}
	}
	g.enqueueCanvas(cvs, CanvasCmd{Op: "line", Args: []int{x1, y1, x2, y2}, Color: clr, PenSize: g.canvasPenSize(cvs)})
	return value.TRUE
}

func (g *GoRuntime) halRect(args []value.Value, ctx *hal.EvalContext) value.Value {
	return g.rectHelper("rect", args)
}

func (g *GoRuntime) halFillRect(args []value.Value, ctx *hal.EvalContext) value.Value {
	return g.rectHelper("fill_rect", args)
}

func (g *GoRuntime) rectHelper(op string, args []value.Value) value.Value {
	if len(args) < 5 || len(args) > 6 {
		return value.NewFault("%s: want 5 or 6 arguments, got %d", op, len(args))
	}
	cvs, ok := args[0].(*value.Canvas)
	if !ok {
		return value.NewFault("%s: expected canvas", op)
	}
	if errv := g.requireCanvasActive(op, cvs); errv != nil {
		return errv
	}
	x, ok1 := toInt(args[1])
	y, ok2 := toInt(args[2])
	w, ok3 := toInt(args[3])
	h, ok4 := toInt(args[4])
	if !ok1 || !ok2 || !ok3 || !ok4 {
		return value.NewFault("%s: dimensions must be numbers", op)
	}
	clr := g.canvasFG(cvs)
	if len(args) == 6 {
		clr, ok = parseColor(args[5])
		if !ok {
			return value.NewFault("%s: invalid color", op)
		}
	}
	g.enqueueCanvas(cvs, CanvasCmd{Op: op, Args: []int{x, y, w, h}, Color: clr, PenSize: g.canvasPenSize(cvs)})
	return value.TRUE
}

func (g *GoRuntime) halCircle(args []value.Value, ctx *hal.EvalContext) value.Value {
	return g.circleHelper("circle", args)
}

func (g *GoRuntime) halFillCircle(args []value.Value, ctx *hal.EvalContext) value.Value {
	return g.circleHelper("fill_circle", args)
}

func (g *GoRuntime) halArc(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) < 6 || len(args) > 7 {
		return value.NewFault("arc: want 6 or 7 arguments, got %d", len(args))
	}
	cvs, ok := args[0].(*value.Canvas)
	if !ok {
		return value.NewFault("arc: expected canvas")
	}
	if errv := g.requireCanvasActive("arc", cvs); errv != nil {
		return errv
	}
	x, ok1 := toInt(args[1])
	y, ok2 := toInt(args[2])
	r, ok3 := toInt(args[3])
	start, ok4 := toInt(args[4])
	end, ok5 := toInt(args[5])
	if !ok1 || !ok2 || !ok3 || !ok4 || !ok5 {
		return value.NewFault("arc: arguments must be numbers")
	}
	clr := g.canvasFG(cvs)
	if len(args) == 7 {
		clr, ok = parseColor(args[6])
		if !ok {
			return value.NewFault("arc: invalid color")
		}
	}
	g.enqueueCanvas(cvs, CanvasCmd{Op: "arc", Args: []int{x, y, r, start, end}, Color: clr, PenSize: g.canvasPenSize(cvs)})
	return value.TRUE
}

func (g *GoRuntime) circleHelper(op string, args []value.Value) value.Value {
	if len(args) < 4 || len(args) > 5 {
		return value.NewFault("%s: want 4 or 5 arguments, got %d", op, len(args))
	}
	cvs, ok := args[0].(*value.Canvas)
	if !ok {
		return value.NewFault("%s: expected canvas", op)
	}
	if errv := g.requireCanvasActive(op, cvs); errv != nil {
		return errv
	}
	x, ok1 := toInt(args[1])
	y, ok2 := toInt(args[2])
	r, ok3 := toInt(args[3])
	if !ok1 || !ok2 || !ok3 {
		return value.NewFault("%s: dimensions must be numbers", op)
	}
	clr := g.canvasFG(cvs)
	if len(args) == 5 {
		clr, ok = parseColor(args[4])
		if !ok {
			return value.NewFault("%s: invalid color", op)
		}
	}
	g.enqueueCanvas(cvs, CanvasCmd{Op: op, Args: []int{x, y, r}, Color: clr, PenSize: g.canvasPenSize(cvs)})
	return value.TRUE
}

func (g *GoRuntime) halClear(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("clear: want 1 argument, got %d", len(args))
	}
	cvs, ok := args[0].(*value.Canvas)
	if !ok {
		return value.NewFault("clear: expected canvas")
	}
	if errv := g.requireCanvasActive("clear", cvs); errv != nil {
		return errv
	}
	g.enqueueCanvas(cvs, CanvasCmd{Op: "clear"})
	return value.TRUE
}

func (g *GoRuntime) halDestroy(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("destroy: want 1 argument, got %d", len(args))
	}
	cvs, ok := args[0].(*value.Canvas)
	if !ok {
		return value.NewFault("destroy: expected canvas")
	}
	resource, ok := g.canvasResource(cvs)
	if !ok {
		return value.TRUE
	}
	select {
	case <-resource.Done:
	default:
		close(resource.Done)
	}
	// The bridge goroutine sees Done and drains any queued commands from
	// cvs.Commands into sendCh before sending the close frame itself. We
	// wait for the bridge to finish, then reap the child.
	bridgeWait(resource)
	g.untrackCanvas(cvs)
	return value.TRUE
}

func (g *GoRuntime) halSetBG(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("set_bg: want 2 arguments, got %d", len(args))
	}
	cvs, ok := args[0].(*value.Canvas)
	if !ok {
		return value.NewFault("set_bg: expected canvas")
	}
	if errv := g.requireCanvasActive("set_bg", cvs); errv != nil {
		return errv
	}
	clr, ok := parseColor(args[1])
	if !ok {
		return value.NewFault("set_bg: invalid color")
	}
	g.setCanvasBG(cvs, clr)
	sendCanvasSetBG(g.canvasResourceMust(cvs), clr)
	return value.TRUE
}

func (g *GoRuntime) halSetFG(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("set_fg: want 2 arguments, got %d", len(args))
	}
	cvs, ok := args[0].(*value.Canvas)
	if !ok {
		return value.NewFault("set_fg: expected canvas")
	}
	if errv := g.requireCanvasActive("set_fg", cvs); errv != nil {
		return errv
	}
	clr, ok := parseColor(args[1])
	if !ok {
		return value.NewFault("set_fg: invalid color")
	}
	g.setCanvasFG(cvs, clr)
	sendCanvasSetFG(g.canvasResourceMust(cvs), clr)
	return value.TRUE
}

func (g *GoRuntime) halPenSize(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("pen_size: want 2 arguments, got %d", len(args))
	}
	cvs, ok := args[0].(*value.Canvas)
	if !ok {
		return value.NewFault("pen_size: expected canvas")
	}
	if errv := g.requireCanvasActive("pen_size", cvs); errv != nil {
		return errv
	}
	size, ok := toInt(args[1])
	if !ok || size < 1 {
		return value.NewFault("pen_size: size must be a positive number")
	}
	g.setCanvasPenSize(cvs, float32(size))
	return value.TRUE
}

// Helper functions

func (g *GoRuntime) requireCanvasActive(op string, cvs *value.Canvas) value.Value {
	resource, ok := g.canvasResource(cvs)
	if !ok || !canvasSessionAlive(resource) {
		return value.NewShapedError("canvas", "%s: canvas closed", op)
	}
	return nil
}

func toInt(v value.Value) (int, bool) {
	n, ok := v.(*value.Number)
	if !ok {
		return 0, false
	}
	// Truncate the rational to an integer. Canvas coordinates are pixels;
	// fractional values are expected from computed positions (e.g. turtle
	// geometry) and truncation is the correct behavior at the display
	// boundary. Out-of-range and non-finite values are rejected.
	f, exact := n.Float64()
	if !exact && (math.IsInf(f, 0) || math.IsNaN(f)) {
		return 0, false
	}
	i := int(f)
	// Guard against overflow: int(f) is undefined when f is outside int range.
	if f > float64(math.MaxInt32) || f < float64(math.MinInt32) {
		return 0, false
	}
	return i, true
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

// canvasAliveOrError returns an error value if the canvas session is closed.
// This is used by builtins to fail fast when the user closes the window manually.
func (g *GoRuntime) canvasAliveOrError(cvs *value.Canvas, name string) value.Value {
	resource, ok := g.canvasResource(cvs)
	if !ok || !canvasSessionAlive(resource) {
		return value.NewShapedError("canvas", "%s: canvas closed", name)
	}
	return nil
}
