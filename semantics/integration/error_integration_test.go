package integration_test
import "aiki/syntax"

import (
	"aiki/semantics/eval"
	"aiki/semantics/testutil"
	"aiki/semantics/value"
	"strings"
	"testing"
)

// testEvalPreludeMultiline evaluates multiline input with source tracking.
func testEvalPreludeMultiline(input string) value.Value {
	env := testutil.SetupEnv()
	env.SetSource(input)
	env.SetFile("test.ai")
	return eval.RunNode(syntax.GetGrammar(), input, env)
}

func TestErrorHasPosition(t *testing.T) {
	// HAL error should get annotated with call site position
	result := testutil.EvalPrelude(`first([])`)

	err, ok := result.(*value.Error)
	if !ok {
		t.Fatalf("expected error, got %T", result)
	}

	if err.Line == 0 {
		t.Error("error should have line number")
	}

	if !strings.Contains(err.Message, "empty") {
		t.Errorf("expected 'empty' in message, got: %s", err.Message)
	}
}

func TestErrorHasSourceLine(t *testing.T) {
	// Error should capture the source line
	input := `let x = 42
first([])
let y = 10`

	result := testEvalPreludeMultiline(input)

	err, ok := result.(*value.Error)
	if !ok {
		t.Fatalf("expected error, got %T", result)
	}

	if err.Line != 2 {
		t.Errorf("expected line 2, got %d", err.Line)
	}

	if !strings.Contains(err.Source, "first([])") {
		t.Errorf("expected source to contain 'first([])', got: %s", err.Source)
	}
}

func TestErrorStackTrace(t *testing.T) {
	// Nested call should show stack trace
	input := `let f = () {
	first([])
}
f()`

	result := testEvalPreludeMultiline(input)

	err, ok := result.(*value.Error)
	if !ok {
		t.Fatalf("expected error, got %T", result)
	}

	inspect := err.Inspect()

	// Should mention 'f' in the error
	if !strings.Contains(inspect, "'f'") {
		t.Errorf("expected stack trace to mention 'f', got:\n%s", inspect)
	}

	// Should have "from" indicating stack trace
	if !strings.Contains(inspect, "from") {
		t.Errorf("expected 'from' in stack trace, got:\n%s", inspect)
	}
}

func TestErrorUndefinedHasPosition(t *testing.T) {
	// Language-level error (not HAL) should have position
	input := `let x = 1
y + 1`

	result := testEvalPreludeMultiline(input)

	err, ok := result.(*value.Error)
	if !ok {
		t.Fatalf("expected error, got %T", result)
	}

	if err.Line != 2 {
		t.Errorf("expected line 2, got %d", err.Line)
	}

	if !strings.Contains(err.Message, "undefined") {
		t.Errorf("expected 'undefined' in message, got: %s", err.Message)
	}
}

func TestErrorDivisionByZero(t *testing.T) {
	// Operator error should have position
	input := `let x = 10
let y = x / 0`

	result := testEvalPreludeMultiline(input)

	err, ok := result.(*value.Error)
	if !ok {
		t.Fatalf("expected error, got %T", result)
	}

	if err.Line != 2 {
		t.Errorf("expected line 2, got %d", err.Line)
	}

	if !strings.Contains(err.Message, "division by zero") {
		t.Errorf("expected 'division by zero' in message, got: %s", err.Message)
	}
}

func TestErrorIndexOutOfBounds(t *testing.T) {
	input := `let list = [1, 2, 3]
list[10]`

	result := testEvalPreludeMultiline(input)

	err, ok := result.(*value.Error)
	if !ok {
		t.Fatalf("expected error, got %T", result)
	}

	if err.Line != 2 {
		t.Errorf("expected line 2, got %d", err.Line)
	}

	if !strings.Contains(err.Message, "out of bounds") {
		t.Errorf("expected 'out of bounds' in message, got: %s", err.Message)
	}
}

func TestErrorCannotShadowBuiltin(t *testing.T) {
	result := testutil.EvalPrelude(`let first = 42`)

	err, ok := result.(*value.Error)
	if !ok {
		t.Fatalf("expected error, got %T", result)
	}

	if !strings.Contains(err.Message, "shadow") && !strings.Contains(err.Message, "builtin") {
		t.Errorf("expected shadow/builtin error, got: %s", err.Message)
	}
}

func TestInspectAtLayerUser(t *testing.T) {
	// Create error with mixed layer stack
	// Stack order: oldest first, newest last
	stack := []value.StackFrame{
		{Name: "main", File: "test.ai", Line: 1, Layer: value.LayerUser},
		{Name: "helper", File: "test.ai", Line: 5, Layer: value.LayerUser},
		{Name: "hash_get", File: "prelude.ai", Line: 100, Layer: value.LayerPrelude},
		{Name: "first", File: "", Line: 0, Layer: value.LayerHal},
	}

	err := &value.Error{
		Message: "empty list",
		File:    "test.ai",
		Line:    5,
		Source:  "hash_get(table, key)",
		Stack:   stack,
	}

	// User layer should show user frames in "from" lines, not strict/hal
	userView := err.InspectAtLayer(value.LayerUser)

	// Should show "helper" (user layer, second newest)
	if !strings.Contains(userView, "'helper'") {
		t.Errorf("user view should show helper, got:\n%s", userView)
	}
	// Should show "main" (user layer, oldest)
	if !strings.Contains(userView, "'main'") {
		t.Errorf("user view should show main, got:\n%s", userView)
	}
	// "from" lines should NOT include strict-layer "hash_get"
	// (Note: the header may show it as innermost visible, but "from" shouldn't)
	fromLines := strings.Split(userView, "\n")
	for _, line := range fromLines {
		if strings.Contains(line, "from") && strings.Contains(line, "hash_get") {
			t.Errorf("user view 'from' should not show strict layer hash_get, got:\n%s", userView)
		}
	}
}

func TestInspectAtLayerPrelude(t *testing.T) {
	stack := []value.StackFrame{
		{Name: "main", File: "test.ai", Line: 1, Layer: value.LayerUser},
		{Name: "helper", File: "test.ai", Line: 5, Layer: value.LayerUser},
		{Name: "hash_get", File: "prelude.ai", Line: 100, Layer: value.LayerPrelude},
		{Name: "first", File: "", Line: 0, Layer: value.LayerHal},
	}

	err := &value.Error{
		Message: "empty list",
		File:    "prelude.ai",
		Line:    100,
		Source:  "first(bucket)",
		Stack:   stack,
	}

	// Prelude layer should show strict and user, not hal
	strictView := err.InspectAtLayer(value.LayerPrelude)

	// Should show user frames
	if !strings.Contains(strictView, "'helper'") {
		t.Errorf("strict view should show user layer helper, got:\n%s", strictView)
	}
	// Should show strict frame in "from" (hash_get is second newest after hal)
	if !strings.Contains(strictView, "'hash_get'") {
		t.Errorf("strict view should show strict layer hash_get, got:\n%s", strictView)
	}
}

func TestInspectAtLayerHal(t *testing.T) {
	stack := []value.StackFrame{
		{Name: "main", File: "test.ai", Line: 1, Layer: value.LayerUser},
		{Name: "hash_get", File: "prelude.ai", Line: 100, Layer: value.LayerPrelude},
		{Name: "first", File: "hal", Line: 0, Layer: value.LayerHal},
	}

	err := &value.Error{
		Message: "empty list",
		File:    "hal",
		Line:    0,
		Stack:   stack,
	}

	// HAL layer should show everything
	halView := err.InspectAtLayer(value.LayerHal)
	if !strings.Contains(halView, "'main'") {
		t.Errorf("hal view should show user layer, got:\n%s", halView)
	}
	if !strings.Contains(halView, "'hash_get'") {
		t.Errorf("hal view should show strict layer, got:\n%s", halView)
	}
}

func TestAnnotateError(t *testing.T) {
	// Bare error gets annotated
	bare := value.NewError("something went wrong")
	if bare.Line != 0 {
		t.Fatal("bare error should have no line")
	}

	stack := []value.StackFrame{{Name: "test", File: "test.ai", Line: 5}}
	annotated := value.AnnotateError(bare, "test.ai", 10, "some_call()", stack)

	if annotated.Line != 10 {
		t.Errorf("expected line 10, got %d", annotated.Line)
	}
	if annotated.File != "test.ai" {
		t.Errorf("expected file test.ai, got %s", annotated.File)
	}
	if annotated.Source != "some_call()" {
		t.Errorf("expected source 'some_call()', got %s", annotated.Source)
	}
}

func TestAnnotateErrorPreservesExisting(t *testing.T) {
	// Error with position should not be re-annotated
	existing := value.NewErrorAt("original.ai", 5, "original()", nil, "original error")

	stack := []value.StackFrame{{Name: "test", File: "test.ai", Line: 10}}
	result := value.AnnotateError(existing, "other.ai", 99, "other()", stack)

	// Should preserve original position
	if result.Line != 5 {
		t.Errorf("expected original line 5, got %d", result.Line)
	}
	if result.File != "original.ai" {
		t.Errorf("expected original file, got %s", result.File)
	}
}
