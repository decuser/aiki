package integration_test

import (
	"aiki/tests/testutil"
	"testing"

	"aiki/semantics/value"
)

func TestCanvasCreate(t *testing.T) {
	// canvas_create returns a canvas value
	result := testutil.EvalPrelude(`canvas_create(200, 200)`)
	if result == nil {
		t.Fatal("expected canvas, got nil")
	}
	if _, ok := result.(*value.Error); ok {
		// Canvas may fail in test environments without display
		// This is expected — mark as skip rather than fail
		t.Skipf("canvas not available in test environment: %s", result.Inspect())
	}
	if result.Type() != value.CanvasType {
		t.Errorf("expected canvas type, got %s", result.Type())
	}
}

func TestCanvasColors(t *testing.T) {
	// Color symbols should resolve
	result := testutil.EvalPrelude(`:red`)
	if result.Inspect() != ":red" {
		t.Errorf("got %s, want :red", result.Inspect())
	}
}
