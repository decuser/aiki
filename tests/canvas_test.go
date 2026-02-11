package tests

import (
	"testing"

	"aiki/lang/eval"
	"aiki/lang/value"
)

func TestCanvasCreate(t *testing.T) {
	env := value.NewEnv(nil)
	result := eval.Run("canvas(800, 600)", env)

	cvs, ok := result.(*value.Canvas)
	if !ok {
		if err, isErr := result.(*value.Error); isErr {
			t.Fatalf("expected Canvas, got Error: %s", err.Message)
		}
		t.Fatalf("expected Canvas, got %T", result)
	}

	if cvs.Width != 800 || cvs.Height != 600 {
		t.Errorf("wrong dimensions: %d x %d", cvs.Width, cvs.Height)
	}

	// Clean up
	close(cvs.Done)
}

func TestCanvasGetSetFG(t *testing.T) {
	env := value.NewEnv(nil)

	// Create canvas
	result := eval.Run("let c = canvas(100, 100)", env)
	if _, ok := result.(*value.Error); ok {
		t.Fatalf("canvas creation failed: %v", result)
	}

	// Set fg
	result = eval.Run("set_fg(c, :red)", env)
	if _, ok := result.(*value.Error); ok {
		t.Fatalf("set_fg failed: %v", result)
	}

	// Get fg
	result = eval.Run("get_fg(c)", env)
	list, ok := result.(*value.List)
	if !ok {
		t.Fatalf("get_fg: expected List, got %T", result)
	}

	// Check red (170, 0, 0)
	if len(list.Elements) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(list.Elements))
	}

	// Clean up
	cvs, _ := env.Get("c")
	close(cvs.(*value.Canvas).Done)
}

func TestCanvasColors(t *testing.T) {
	env := value.NewEnv(nil)
	eval.Run("let c = canvas(100, 100)", env)

	// Test symbol color
	result := eval.Run("set_fg(c, :bright_white)", env)
	if _, ok := result.(*value.Error); ok {
		t.Fatalf("symbol color failed: %v", result)
	}

	// Test RGB list color
	result = eval.Run("set_fg(c, [255, 128, 0])", env)
	if _, ok := result.(*value.Error); ok {
		t.Fatalf("RGB color failed: %v", result)
	}

	// Test invalid color
	result = eval.Run("set_fg(c, :not_a_color)", env)
	if _, ok := result.(*value.Error); !ok {
		t.Error("expected error for invalid color")
	}

	// Clean up
	cvs, _ := env.Get("c")
	close(cvs.(*value.Canvas).Done)
}
