package substrate

import (
	"testing"

	"aiki/engine/semantics/value"
)

func TestCanvasWidthHeight(t *testing.T) {
	cvs := &value.Canvas{
		Width:    800,
		Height:   600,
		Commands: make(chan value.CanvasCmd, 1),
		Done:     make(chan struct{}),
		Ready:    make(chan struct{}),
	}

	// Test width
	w := halCanvasWidth([]value.Value{cvs}, nil)
	wNum, ok := w.(*value.Number)
	if !ok {
		t.Fatalf("canvas_width: expected Number, got %T", w)
	}
	wVal, _ := wNum.Val.Float64()
	if int(wVal) != 800 {
		t.Errorf("canvas_width: expected 800, got %v", wVal)
	}

	// Test height
	h := halCanvasHeight([]value.Value{cvs}, nil)
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
	// Wrong arg count
	v := halCanvasWidth([]value.Value{}, nil)
	if _, ok := v.(*value.Fault); !ok {
		t.Errorf("expected Fault for no args, got %T", v)
	}

	v = halCanvasHeight([]value.Value{value.NewNumber(1, 1), value.NewNumber(2, 1)}, nil)
	if _, ok := v.(*value.Fault); !ok {
		t.Errorf("expected Fault for too many args, got %T", v)
	}

	// Wrong type
	v = halCanvasWidth([]value.Value{value.NewNumber(42, 1)}, nil)
	if _, ok := v.(*value.Fault); !ok {
		t.Errorf("expected Fault for non-canvas, got %T", v)
	}
}
