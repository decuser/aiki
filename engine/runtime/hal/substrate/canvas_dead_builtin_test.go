package substrate

import (
	"testing"

	"aiki/engine/semantics/value"
)

func TestCanvasAliveOrError(t *testing.T) {
	cvs := &value.Canvas{
		Width:    10,
		Height:   10,
		Commands: make(chan value.CanvasCmd, 1),
		Done:     make(chan struct{}),
		Ready:    make(chan struct{}),
	}

	// Not started, so no session should exist.
	v := canvasAliveOrError(cvs, "dot")
	if v == nil || !value.IsShapedError(v) {
		t.Fatalf("expected shaped error for closed canvas, got %v", v)
	}
}
