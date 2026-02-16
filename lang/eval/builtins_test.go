package tests

import (
	"testing"

	"aiki/lang/value"
)

func TestFirst(t *testing.T) {
	result := testEvalStrict(`first([10, 20, 30])`)
	testNumberValue(t, result, "10")
}

func TestRest(t *testing.T) {
	result := testEvalStrict(`rest([10, 20, 30])`)
	if result.Inspect() != "[20, 30]" {
		t.Errorf("got %s, want [20, 30]", result.Inspect())
	}
}

func TestLength(t *testing.T) {
	result := testEvalStrict(`length([1, 2, 3])`)
	testNumberValue(t, result, "3")
}

func TestType(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`type(42)`, ":number"},
		{`type("hello")`, ":string"},
		{`type(true)`, ":boolean"},
		{`type([1, 2])`, ":list"},
		{`type(:foo)`, ":symbol"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := testEvalStrict(tt.input)
			if result.Inspect() != tt.expected {
				t.Errorf("got %s, want %s", result.Inspect(), tt.expected)
			}
		})
	}
}

func TestShapeBuiltin(t *testing.T) {
	result := testEvalStrict(`
let @point [x, y]
let p = [@point, 10, 20]
shape(p)
`)
	if result.Inspect() != ":point" {
		t.Errorf("got %s, want :point", result.Inspect())
	}
}

func TestEqual(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{`equal(1, 1)`, true},
		{`equal(1, 2)`, false},
		{`equal("a", "a")`, true},
		{`equal([1, 2], [1, 2])`, true},
		{`equal([1, 2], [1, 3])`, false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := testEvalStrict(tt.input)
			testBooleanValue(t, result, tt.expected)
		})
	}
}

func TestToStr(t *testing.T) {
	result := testEvalStrict(`to_str(42)`)
	str, ok := result.(*value.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if str.Value != "42" {
		t.Errorf("got %s, want 42", str.Value)
	}
}

func TestToNumber(t *testing.T) {
	result := testEvalStrict(`to_number("42")`)
	testNumberValue(t, result, "42")
}

func TestToDecimal(t *testing.T) {
	result := testEvalStrict(`to_decimal(1/3, 4)`)
	str, ok := result.(*value.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if str.Value != "0.3333" {
		t.Errorf("got %s, want 0.3333", str.Value)
	}
}

func TestMath(t *testing.T) {
	result := testEvalStrict(`sqrt(4)`)
	testNumberValue(t, result, "2")
}

func TestAppendPrepend(t *testing.T) {
	result := testEvalStrict(`append([1, 2], 3)`)
	if result.Inspect() != "[1, 2, 3]" {
		t.Errorf("append: got %s, want [1, 2, 3]", result.Inspect())
	}

	result = testEvalStrict(`prepend([2, 3], 1)`)
	if result.Inspect() != "[1, 2, 3]" {
		t.Errorf("prepend: got %s, want [1, 2, 3]", result.Inspect())
	}
}

func TestApply(t *testing.T) {
	result := testEvalStrict(`
let add = (a, b) { return a + b }
apply(add, [3, 4])
`)
	testNumberValue(t, result, "7")
}
