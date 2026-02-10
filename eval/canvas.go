package eval

import (
	"image/color"
	"sync"

	"aiki/canvas"
	"aiki/value"
)

// Track open canvases for cleanup on reset
var (
	openCanvases   []*value.Canvas
	openCanvasesMu sync.Mutex
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

func CloseAllCanvases() {
	openCanvasesMu.Lock()
	for _, c := range openCanvases {
		select {
		case <-c.Done:
			// already closed
		default:
			close(c.Done)
		}
	}
	openCanvases = nil
	openCanvasesMu.Unlock()
}

func init() {
	registerCanvasBuiltins()
}

func registerCanvasBuiltins() {
	builtins["canvas"] = &value.Builtin{Name: "canvas", Fn: builtinCanvas}
	builtins["get_bg"] = &value.Builtin{Name: "get_bg", Fn: builtinGetBG}
	builtins["set_bg"] = &value.Builtin{Name: "set_bg", Fn: builtinSetBG}
	builtins["get_fg"] = &value.Builtin{Name: "get_fg", Fn: builtinGetFG}
	builtins["set_fg"] = &value.Builtin{Name: "set_fg", Fn: builtinSetFG}
	builtins["pen_size"] = &value.Builtin{Name: "pen_size", Fn: builtinPenSize}
	builtins["get_pen_size"] = &value.Builtin{Name: "get_pen_size", Fn: builtinGetPenSize}
	builtins["line"] = &value.Builtin{Name: "line", Fn: builtinLine}
	builtins["rect"] = &value.Builtin{Name: "rect", Fn: builtinRect}
	builtins["fill_rect"] = &value.Builtin{Name: "fill_rect", Fn: builtinFillRect}
	builtins["circle"] = &value.Builtin{Name: "circle", Fn: builtinCircle}
	builtins["fill_circle"] = &value.Builtin{Name: "fill_circle", Fn: builtinFillCircle}
	builtins["oval"] = &value.Builtin{Name: "oval", Fn: builtinOval}
	builtins["fill_oval"] = &value.Builtin{Name: "fill_oval", Fn: builtinFillOval}
	builtins["dot"] = &value.Builtin{Name: "dot", Fn: builtinDot}
	builtins["text"] = &value.Builtin{Name: "text", Fn: builtinText}
	builtins["clear"] = &value.Builtin{Name: "clear", Fn: builtinClear}
	builtins["undo"] = &value.Builtin{Name: "undo", Fn: builtinUndo}
	builtins["redo"] = &value.Builtin{Name: "redo", Fn: builtinRedo}
	builtins["get_width"] = &value.Builtin{Name: "get_width", Fn: builtinGetWidth}
	builtins["get_height"] = &value.Builtin{Name: "get_height", Fn: builtinGetHeight}
	builtins["save"] = &value.Builtin{Name: "save", Fn: builtinSave}
	builtins["destroy"] = &value.Builtin{Name: "destroy", Fn: builtinDestroy}
}

func builtinCanvas(args ...value.Value) value.Value {
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
		BG:       canvas.DefaultBG,
		FG:       canvas.DefaultFG,
		PenSize:  2,
		Commands: make(chan value.CanvasCmd, 100),
		Done:     make(chan struct{}),
		Ready:    make(chan struct{}),
	}

	trackCanvas(cvs)
	go canvas.Run(cvs)
	<-cvs.Ready

	return cvs
}

func builtinGetBG(args ...value.Value) value.Value {
	if len(args) != 1 {
		return value.NewError("get_bg: want 1 argument, got %d", len(args))
	}
	cvs, ok := args[0].(*value.Canvas)
	if !ok {
		return value.NewError("get_bg: expected canvas")
	}
	return colorToList(cvs.BG)
}

func builtinSetBG(args ...value.Value) value.Value {
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
	return value.True
}

func builtinGetFG(args ...value.Value) value.Value {
	if len(args) != 1 {
		return value.NewError("get_fg: want 1 argument, got %d", len(args))
	}
	cvs, ok := args[0].(*value.Canvas)
	if !ok {
		return value.NewError("get_fg: expected canvas")
	}
	return colorToList(cvs.FG)
}

func builtinSetFG(args ...value.Value) value.Value {
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
	return value.True
}

func builtinPenSize(args ...value.Value) value.Value {
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
	return value.True
}

func builtinGetPenSize(args ...value.Value) value.Value {
	if len(args) != 1 {
		return value.NewError("get_pen_size: want 1 argument, got %d", len(args))
	}
	cvs, ok := args[0].(*value.Canvas)
	if !ok {
		return value.NewError("get_pen_size: expected canvas")
	}
	return value.NewNumber(int64(cvs.PenSize), 1)
}

func builtinLine(args ...value.Value) value.Value {
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
	return value.True
}

func builtinRect(args ...value.Value) value.Value {
	return rectHelper("rect", args)
}

func builtinFillRect(args ...value.Value) value.Value {
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
	return value.True
}

func builtinCircle(args ...value.Value) value.Value {
	return circleHelper("circle", args)
}

func builtinFillCircle(args ...value.Value) value.Value {
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
	return value.True
}

func builtinOval(args ...value.Value) value.Value {
	return ovalHelper("oval", args)
}

func builtinFillOval(args ...value.Value) value.Value {
	return ovalHelper("fill_oval", args)
}

func ovalHelper(op string, args []value.Value) value.Value {
	if len(args) < 5 || len(args) > 6 {
		return value.NewError("%s: want 5 or 6 arguments, got %d", op, len(args))
	}
	cvs, ok := args[0].(*value.Canvas)
	if !ok {
		return value.NewError("%s: expected canvas", op)
	}
	x, ok1 := toInt(args[1])
	y, ok2 := toInt(args[2])
	rx, ok3 := toInt(args[3])
	ry, ok4 := toInt(args[4])
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
	cvs.Commands <- value.CanvasCmd{Op: op, Args: []int{x, y, rx, ry}, Color: clr, PenSize: cvs.PenSize}
	return value.True
}

func builtinDot(args ...value.Value) value.Value {
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
	return value.True
}

func builtinText(args ...value.Value) value.Value {
	if len(args) < 4 || len(args) > 5 {
		return value.NewError("text: want 4 or 5 arguments, got %d", len(args))
	}
	cvs, ok := args[0].(*value.Canvas)
	if !ok {
		return value.NewError("text: expected canvas")
	}
	x, ok1 := toInt(args[1])
	y, ok2 := toInt(args[2])
	if !ok1 || !ok2 {
		return value.NewError("text: coordinates must be numbers")
	}
	str, ok := args[3].(*value.String)
	if !ok {
		return value.NewError("text: expected string")
	}
	clr := cvs.FG
	if len(args) == 5 {
		clr, ok = parseColor(args[4])
		if !ok {
			return value.NewError("text: invalid color")
		}
	}
	cvs.Commands <- value.CanvasCmd{Op: "text", Args: []int{x, y}, Color: clr, Text: str.Value}
	return value.True
}

func builtinClear(args ...value.Value) value.Value {
	if len(args) != 1 {
		return value.NewError("clear: want 1 argument, got %d", len(args))
	}
	cvs, ok := args[0].(*value.Canvas)
	if !ok {
		return value.NewError("clear: expected canvas")
	}
	cvs.Commands <- value.CanvasCmd{Op: "clear"}
	return value.True
}

func builtinUndo(args ...value.Value) value.Value {
	if len(args) != 1 {
		return value.NewError("undo: want 1 argument, got %d", len(args))
	}
	cvs, ok := args[0].(*value.Canvas)
	if !ok {
		return value.NewError("undo: expected canvas")
	}
	cvs.Commands <- value.CanvasCmd{Op: "undo"}
	return value.True
}

func builtinRedo(args ...value.Value) value.Value {
	if len(args) != 1 {
		return value.NewError("redo: want 1 argument, got %d", len(args))
	}
	cvs, ok := args[0].(*value.Canvas)
	if !ok {
		return value.NewError("redo: expected canvas")
	}
	cvs.Commands <- value.CanvasCmd{Op: "redo"}
	return value.True
}

func builtinGetWidth(args ...value.Value) value.Value {
	if len(args) != 1 {
		return value.NewError("get_width: want 1 argument, got %d", len(args))
	}
	cvs, ok := args[0].(*value.Canvas)
	if !ok {
		return value.NewError("get_width: expected canvas")
	}
	return value.NewNumber(int64(cvs.Width), 1)
}

func builtinGetHeight(args ...value.Value) value.Value {
	if len(args) != 1 {
		return value.NewError("get_height: want 1 argument, got %d", len(args))
	}
	cvs, ok := args[0].(*value.Canvas)
	if !ok {
		return value.NewError("get_height: expected canvas")
	}
	return value.NewNumber(int64(cvs.Height), 1)
}

func builtinSave(args ...value.Value) value.Value {
	if len(args) != 2 {
		return value.NewError("save: want 2 arguments, got %d", len(args))
	}
	cvs, ok := args[0].(*value.Canvas)
	if !ok {
		return value.NewError("save: expected canvas")
	}
	path, ok := args[1].(*value.String)
	if !ok {
		return value.NewError("save: expected string path")
	}
	result := make(chan error, 1)
	cvs.Commands <- value.CanvasCmd{Op: "save", Path: path.Value, Result: result}
	if err := <-result; err != nil {
		return value.NewError("save: %s", err)
	}
	return value.True
}

func builtinDestroy(args ...value.Value) value.Value {
	if len(args) != 1 {
		return value.NewError("destroy: want 1 argument, got %d", len(args))
	}
	cvs, ok := args[0].(*value.Canvas)
	if !ok {
		return value.NewError("destroy: expected canvas")
	}
	untrackCanvas(cvs)
	close(cvs.Done)
	return value.True
}

// Helpers

func toInt(v value.Value) (int, bool) {
	n, ok := v.(*value.Number)
	if !ok {
		return 0, false
	}
	f, _ := n.Value.Float64()
	return int(f), true
}

func parseColor(v value.Value) (color.RGBA, bool) {
	switch c := v.(type) {
	case *value.Symbol:
		if clr, ok := canvas.Colors[c.Value]; ok {
			return clr, true
		}
		return color.RGBA{}, false
	case *value.List:
		if len(c.Elements) != 3 {
			return color.RGBA{}, false
		}
		r, ok1 := toInt(c.Elements[0])
		g, ok2 := toInt(c.Elements[1])
		b, ok3 := toInt(c.Elements[2])
		if !ok1 || !ok2 || !ok3 {
			return color.RGBA{}, false
		}
		return color.RGBA{uint8(r), uint8(g), uint8(b), 255}, true
	}
	return color.RGBA{}, false
}

func colorToList(c color.RGBA) *value.List {
	return &value.List{
		Elements: []value.Value{
			value.NewNumber(int64(c.R), 1),
			value.NewNumber(int64(c.G), 1),
			value.NewNumber(int64(c.B), 1),
		},
	}
}
