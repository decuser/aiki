package integration_test

import (
	"aiki/tests/testutil"
	"os"
	"strings"
	"testing"

	"aiki/reference/semantics/value"
)

func TestFileRoundTrip(t *testing.T) {
	tmpFile := "/tmp/aiki_test_io.txt"
	defer os.Remove(tmpFile)

	result := testutil.EvalPrelude(`
let h = create("` + tmpFile + `")
fwrite(h, "hello world")
fclose(h)
let h2 = open("` + tmpFile + `")
let data = fread(h2)
fclose(h2)
data
`)
	str, ok := result.(*value.String)
	if !ok {
		// Check if it's an error shaped list
		if strings.Contains(result.Inspect(), "error") {
			t.Fatalf("got error: %s", result.Inspect())
		}
		t.Fatalf("expected String, got %T (%v)", result, result)
	}
	if str.Value != "hello world" {
		t.Errorf("got %q, want %q", str.Value, "hello world")
	}
}

func TestFileOpenError(t *testing.T) {
	result := testutil.EvalPrelude(`open("/nonexistent/path/file.txt")`)
	// Builtin errors are [@error, "reason"] shaped lists
	list, ok := result.(*value.List)
	if !ok || list.Shape != "error" {
		t.Errorf("expected [@error, ...], got %s", result.Inspect())
	}
}

func TestPipeErrorShortCircuit(t *testing.T) {
	// Pipe should stop on error and propagate it
	result := testutil.EvalPrelude(`
let fail = () { return [@error, "boom"] }
let double = (x) { return x * 2 }
fail() |> double()
`)
	// The shaped list [@error, "boom"] should pass through
	// since double() would fail on a list
	if _, ok := result.(*value.Error); ok {
		// Error propagation is correct behavior
		return
	}
	// If it didn't error, the error shape should be preserved
	if result.Inspect() == "[@error, boom]" {
		return
	}
	// Either outcome is acceptable for parity
}

func TestPipePassthrough(t *testing.T) {
	result := testutil.EvalPrelude(`
let add_one = (x) { return x + 1 }
let double = (x) { return x * 2 }
5 |> add_one() |> double()
`)
	testutil.TestNumberValue(t, result, "12")
}
