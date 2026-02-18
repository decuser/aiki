package integration_test

import (
	"aiki/semantics/testutil"
	"testing"

	"aiki/semantics/value"
)

func TestEvalPipe(t *testing.T) {
	result := testutil.EvalPrelude(`
let add_one = (x) { return x + 1 }
let double = (x) { return x * 2 }
5 |> add_one() |> double()
`)
	testutil.TestNumberValue(t, result, "12")
}

func TestEvalPipeWithBuiltins(t *testing.T) {
	result := testutil.EvalPrelude(`
range(1, 6) |> sum()
`)
	testutil.TestNumberValue(t, result, "15")
}

func TestEvalMatchStatement(t *testing.T) {
	result := testutil.EvalPrelude(`
let x = 2
match x {
	1 { "one" }
	2 { "two" }
	_ { "other" }
}
`)
	str, ok := result.(*value.String)
	if !ok {
		t.Fatalf("expected String, got %T (%v)", result, result)
	}
	if str.Value != "two" {
		t.Errorf("got %s, want two", str.Value)
	}
}

func TestEvalMatchWildcard(t *testing.T) {
	result := testutil.EvalPrelude(`
let x = 99
match x {
	1 { "one" }
	_ { "other" }
}
`)
	str, ok := result.(*value.String)
	if !ok {
		t.Fatalf("expected String, got %T (%v)", result, result)
	}
	if str.Value != "other" {
		t.Errorf("got %s, want other", str.Value)
	}
}

func TestEvalSymbols(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`:foo`, ":foo"},
		{`:bar`, ":bar"},
		{`let x = :hello
x`, ":hello"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := testutil.TestEval(tt.input)
			if result.Inspect() != tt.expected {
				t.Errorf("got %s, want %s", result.Inspect(), tt.expected)
			}
		})
	}
}

func TestEvalBuiltinType(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`type(42)`, ":number"},
		{`type("hi")`, ":string"},
		{`type(true)`, ":boolean"},
		{`type([])`, ":list"},
		{`type(:x)`, ":symbol"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := testutil.EvalPrelude(tt.input)
			if result.Inspect() != tt.expected {
				t.Errorf("got %s, want %s", result.Inspect(), tt.expected)
			}
		})
	}
}

func TestEvalBuiltinShape(t *testing.T) {
	result := testutil.EvalPrelude(`
let @point [x, y]
let p = [@point, 1, 2]
shape(p)
`)
	if result.Inspect() != ":point" {
		t.Errorf("got %s, want :point", result.Inspect())
	}
}

func TestEvalShapedListPositionalAccess(t *testing.T) {
	result := testutil.EvalPrelude(`
let @point [x, y]
let p = [@point, 10, 20]
p[0]
`)
	testutil.TestNumberValue(t, result, "10")
}

func TestEvalRune(t *testing.T) {
	result := testutil.TestEval(`'a'`)
	r, ok := result.(*value.Rune)
	if !ok {
		t.Fatalf("expected Rune, got %T", result)
	}
	if r.Value != 'a' {
		t.Errorf("got %c, want a", r.Value)
	}
}

func TestEvalGroupExpression(t *testing.T) {
	result := testutil.TestEval(`(1 + 2) * 3`)
	testutil.TestNumberValue(t, result, "9")
}
