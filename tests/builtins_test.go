package tests

import (
	"os"
	"testing"

	"aiki/eval"
	"aiki/value"
)

// TestBuiltinsFollowTheWay verifies that builtins return raw values on success,
// [@error, reason] on failure. Success is not wrapped in [@ok, ...].

func TestTonum(t *testing.T) {
	env := setupIOEnv()

	// Success returns raw number
	result := eval.Run(`tonum("42")`, env)
	num, ok := result.(*value.Number)
	if !ok {
		t.Fatalf("tonum success: expected Number, got %T: %v", result, result)
	}
	if num.Inspect() != "42" {
		t.Errorf("tonum success: got %s, want 42", num.Inspect())
	}

	// Failure returns [@error, reason]
	result = eval.Run(`tonum("not a number")`, env)
	list, ok := result.(*value.List)
	if !ok {
		t.Fatalf("tonum failure: expected List, got %T: %v", result, result)
	}
	if list.Shape != "error" {
		t.Errorf("tonum failure: expected error shape, got %s", list.Shape)
	}
}

func TestOpen(t *testing.T) {
	env := setupIOEnv()

	// Failure returns [@error, reason]
	result := eval.Run(`open("nonexistent_file_xyz.txt")`, env)
	list, ok := result.(*value.List)
	if !ok {
		t.Fatalf("open failure: expected List, got %T: %v", result, result)
	}
	if list.Shape != "error" {
		t.Errorf("open failure: expected error shape, got %s", list.Shape)
	}
}

func TestCreateBuiltin(t *testing.T) {
	env := setupIOEnv()
	defer os.Remove("test_builtin_create.txt")

	// Success returns raw handle
	result := eval.Run(`create("test_builtin_create.txt")`, env)
	h, ok := result.(*value.Handle)
	if !ok {
		if list, isList := result.(*value.List); isList && list.Shape == "ok" {
			t.Fatalf("create success: returned [@ok, handle], should return raw handle")
		}
		t.Fatalf("create success: expected Handle, got %T: %v", result, result)
	}
	h.File.Close()
}

func TestFirst(t *testing.T) {
	env := setupIOEnv()

	// Success returns raw value
	result := eval.Run(`first([1, 2, 3])`, env)
	num, ok := result.(*value.Number)
	if !ok {
		t.Fatalf("first success: expected Number, got %T: %v", result, result)
	}
	if num.Inspect() != "1" {
		t.Errorf("first success: got %s, want 1", num.Inspect())
	}
}

func TestNth(t *testing.T) {
	env := setupIOEnv()

	// Success returns raw value
	result := eval.Run(`nth([10, 20, 30], 1)`, env)
	num, ok := result.(*value.Number)
	if !ok {
		t.Fatalf("nth success: expected Number, got %T: %v", result, result)
	}
	if num.Inspect() != "20" {
		t.Errorf("nth success: got %s, want 20", num.Inspect())
	}
}

func TestLen(t *testing.T) {
	env := setupIOEnv()

	result := eval.Run(`len([1, 2, 3])`, env)
	num, ok := result.(*value.Number)
	if !ok {
		t.Fatalf("len: expected Number, got %T: %v", result, result)
	}
	if num.Inspect() != "3" {
		t.Errorf("len: got %s, want 3", num.Inspect())
	}
}

func TestType(t *testing.T) {
	env := setupIOEnv()

	result := eval.Run(`type(42)`, env)
	sym, ok := result.(*value.Symbol)
	if !ok {
		t.Fatalf("type: expected Symbol, got %T: %v", result, result)
	}
	if sym.Value != "number" {
		t.Errorf("type: got %s, want number", sym.Value)
	}
}

func TestShapeBuiltin(t *testing.T) {
	env := setupIOEnv()

	// Raw list
	result := eval.Run(`shape([1, 2, 3])`, env)
	sym, ok := result.(*value.Symbol)
	if !ok {
		t.Fatalf("shape raw: expected Symbol, got %T: %v", result, result)
	}
	if sym.Value != "list" {
		t.Errorf("shape raw: got %s, want list", sym.Value)
	}

	// Shaped list
	result = eval.Run(`shape([@error, "test"])`, env)
	sym, ok = result.(*value.Symbol)
	if !ok {
		t.Fatalf("shape error: expected Symbol, got %T: %v", result, result)
	}
	if sym.Value != "error" {
		t.Errorf("shape error: got %s, want error", sym.Value)
	}
}

func TestEqual(t *testing.T) {
	env := setupIOEnv()

	result := eval.Run(`equal([1, 2], [1, 2])`, env)
	b, ok := result.(*value.Boolean)
	if !ok {
		t.Fatalf("equal: expected Boolean, got %T: %v", result, result)
	}
	if !b.Value {
		t.Errorf("equal: got false, want true")
	}
}

func TestTostr(t *testing.T) {
	env := setupIOEnv()

	result := eval.Run(`tostr(42)`, env)
	str, ok := result.(*value.String)
	if !ok {
		t.Fatalf("tostr: expected String, got %T: %v", result, result)
	}
	if str.Value != "42" {
		t.Errorf("tostr: got %s, want 42", str.Value)
	}
}

func TestTodecimal(t *testing.T) {
	env := setupIOEnv()

	result := eval.Run(`todecimal(3.14159, 2)`, env)
	str, ok := result.(*value.String)
	if !ok {
		t.Fatalf("todecimal: expected String, got %T: %v", result, result)
	}
	if str.Value != "3.14" {
		t.Errorf("todecimal: got %s, want 3.14", str.Value)
	}
}

func TestMath(t *testing.T) {
	env := setupIOEnv()

	// sqrt
	result := eval.Run(`sqrt(4)`, env)
	num, ok := result.(*value.Number)
	if !ok {
		t.Fatalf("sqrt: expected Number, got %T: %v", result, result)
	}
	if num.Inspect() != "2" {
		t.Errorf("sqrt: got %s, want 2", num.Inspect())
	}

	// random returns number
	result = eval.Run(`type(random(100))`, env)
	sym, ok := result.(*value.Symbol)
	if !ok {
		t.Fatalf("random type: expected Symbol, got %T: %v", result, result)
	}
	if sym.Value != "number" {
		t.Errorf("random type: got %s, want number", sym.Value)
	}
}
