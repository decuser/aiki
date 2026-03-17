package evaluator

import "testing"

func TestMatchLiteral(t *testing.T) {
	tests := []struct{ in, out string }{
		{`let x = 1; match x { 1 { :one } _ { :other } }`, ":one"},
		{`let x = 2; match x { 1 { :one } _ { :other } }`, ":other"},
		{`let x = "hi"; match x { "hi" { :yes } _ { :no } }`, ":yes"},
		{`let x = :foo; match x { :foo { :yes } _ { :no } }`, ":yes"},
	}
	for _, tt := range tests {
		v := eval(t, tt.in)
		if v.Inspect() != tt.out {
			t.Errorf("%s: got %s, want %s", tt.in, v.Inspect(), tt.out)
		}
	}
}

func TestMatchWildcard(t *testing.T) {
	v := eval(t, `let x = 42; match x { _ { :matched } }`)
	if v.Inspect() != ":matched" {
		t.Errorf("got %s", v.Inspect())
	}
}

func TestMatchBinding(t *testing.T) {
	v := eval(t, `let x = 42; match x { n { n + 1 } }`)
	if v.Inspect() != "43" {
		t.Errorf("got %s", v.Inspect())
	}
}

func TestMatchList(t *testing.T) {
	tests := []struct{ in, out string }{
		{`let x = [1, 2]; match x { [a, b] { a + b } _ { 0 } }`, "3"},
		{`let x = [1]; match x { [a, b] { a + b } [a] { a } _ { 0 } }`, "1"},
		{`let x = []; match x { [] { :empty } _ { :other } }`, ":empty"},
		{`let x = [1, 2, 3]; match x { [a, b, c] { a + b + c } _ { 0 } }`, "6"},
	}
	for _, tt := range tests {
		v := eval(t, tt.in)
		if v.Inspect() != tt.out {
			t.Errorf("%s: got %s, want %s", tt.in, v.Inspect(), tt.out)
		}
	}
}

func TestMatchShapedList(t *testing.T) {
	tests := []struct{ in, out string }{
		{`let x = [@ok, 42]; match x { [@ok, v] { v } _ { 0 } }`, "42"},
		{`let x = [@error, :io, "fail"]; match x { [@error, k, m] { k } _ { :none } }`, ":io"},
		{`let x = [@end]; match x { [@end] { :done } _ { :other } }`, ":done"},
		{`let x = [@point, 1, 2]; match x { [@point, x, y] { x + y } _ { 0 } }`, "3"},
	}
	for _, tt := range tests {
		v := eval(t, tt.in)
		if v.Inspect() != tt.out {
			t.Errorf("%s: got %s, want %s", tt.in, v.Inspect(), tt.out)
		}
	}
}

func TestMatchShapeMismatch(t *testing.T) {
	tests := []struct{ in, out string }{
		{`let x = [@ok, 1]; match x { [@error, v] { v } _ { :fallback } }`, ":fallback"},
		{`let x = [1, 2]; match x { [@ok, a, b] { a } _ { :plain } }`, ":plain"},
		{`let x = [@end]; match x { [@ok, v] { v } _ { :nope } }`, ":nope"},
	}
	for _, tt := range tests {
		v := eval(t, tt.in)
		if v.Inspect() != tt.out {
			t.Errorf("%s: got %s, want %s", tt.in, v.Inspect(), tt.out)
		}
	}
}

func TestMatchNestedList(t *testing.T) {
	tests := []struct{ in, out string }{
		{`let x = [[1, 2], 3]; match x { [[a, b], c] { a + b + c } _ { 0 } }`, "6"},
		{`let x = [@wrap, [1, 2]]; match x { [@wrap, [a, b]] { a + b } _ { 0 } }`, "3"},
	}
	for _, tt := range tests {
		v := eval(t, tt.in)
		if v.Inspect() != tt.out {
			t.Errorf("%s: got %s, want %s", tt.in, v.Inspect(), tt.out)
		}
	}
}

func TestMatchNoMatch(t *testing.T) {
	v := eval(t, `let x = 5; match x { 1 { :one } 2 { :two } }`)
	if v.Inspect() != "[]" {
		t.Errorf("expected empty list for no match, got %s", v.Inspect())
	}
}
