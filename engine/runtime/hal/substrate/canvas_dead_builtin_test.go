package substrate

import (
	"testing"

	"aiki/engine/semantics/value"
)

func TestCanvasAliveOrError(t *testing.T) {
	rt := NewGoRuntime()
	cvs := &value.Canvas{ID: 1}
	v := rt.canvasAliveOrError(cvs, "dot")
	if v == nil || !value.IsShapedError(v) {
		t.Fatalf("expected shaped error for closed canvas, got %v", v)
	}
}
