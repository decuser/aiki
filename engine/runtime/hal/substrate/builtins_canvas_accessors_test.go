package substrate

import (
	"testing"

	"aiki/engine/semantics/value"
)

func testCanvasHandle(rt *GoRuntime, width, height int) *value.Canvas {
	cvs := &value.Canvas{ID: 1}
	rt.trackCanvas(cvs, &CanvasResource{
		Width: width, Height: height, BG: DefaultBG, FG: DefaultFG, PenSize: 1,
		Commands: make(chan CanvasCmd, 1), Done: make(chan struct{}), Ready: make(chan struct{}),
	})
	return cvs
}

func TestCanvasWidthHeight(t *testing.T) {
	rt := NewGoRuntime()
	cvs := testCanvasHandle(rt, 800, 600)

	w := rt.halCanvasWidth([]value.Value{cvs}, nil)
	wNum, ok := w.(*value.Number)
	if !ok {
		t.Fatalf("canvas_width: expected Number, got %T", w)
	}
	wVal, _ := wNum.Val.Float64()
	if int(wVal) != 800 {
		t.Errorf("canvas_width: expected 800, got %v", wVal)
	}

	h := rt.halCanvasHeight([]value.Value{cvs}, nil)
	hNum, ok := h.(*value.Number)
	if !ok {
		t.Fatalf("canvas_height: expected Number, got %T", h)
	}
	hVal, _ := hNum.Val.Float64()
	if int(hVal) != 600 {
		t.Errorf("canvas_height: expected 600, got %v", hVal)
	}
}

func TestCanvasWidthHeightErrors(t *testing.T) {
	rt := NewGoRuntime()
	v := rt.halCanvasWidth([]value.Value{}, nil)
	if _, ok := v.(*value.Fault); !ok {
		t.Errorf("expected Fault for no args, got %T", v)
	}

	v = rt.halCanvasHeight([]value.Value{value.NewNumber(1, 1), value.NewNumber(2, 1)}, nil)
	if _, ok := v.(*value.Fault); !ok {
		t.Errorf("expected Fault for too many args, got %T", v)
	}

	v = rt.halCanvasWidth([]value.Value{value.NewNumber(42, 1)}, nil)
	if _, ok := v.(*value.Fault); !ok {
		t.Errorf("expected Fault for non-canvas, got %T", v)
	}
}

func TestCanvasHandleIsRuntimeOwned(t *testing.T) {
	rt1 := NewGoRuntime()
	rt2 := NewGoRuntime()
	cvs := testCanvasHandle(rt1, 10, 20)
	if got := rt2.halCanvasAlive([]value.Value{cvs}, nil); got != value.FALSE {
		t.Fatalf("canvas handle from another runtime must not be alive, got %v", got)
	}
	if got := rt2.halCanvasWidth([]value.Value{cvs}, nil); !value.IsShapedError(got) {
		t.Fatalf("cross-runtime canvas handle must fail as resource error, got %v", got)
	}
}
