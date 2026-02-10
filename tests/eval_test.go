package tests

import (
	"testing"

	"aiki/eval"
	"aiki/value"
)

func TestEvalNumbers(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"5", "5"},
		{"42", "42"},
		{"3/4", "3/4"},
		{"-5", "-5"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := testEval(tt.input)
			testNumberValue(t, result, tt.expected)
		})
	}
}

func TestEvalBooleans(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
		{"not true", false},
		{"not false", true},
		{"not not true", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := testEval(tt.input)
			testBooleanValue(t, result, tt.expected)
		})
	}
}

func TestEvalArithmetic(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1 + 2", "3"},
		{"5 - 3", "2"},
		{"2 * 3", "6"},
		{"8 / 2", "4"},
		{"5 % 3", "2"},
		{"1 + 2 * 3", "9"},
		{"2 * 3 + 1", "7"},
		{"(1 + 2) * 3", "9"},
		{"1 + (2 * 3)", "7"},
		{"10 - 5 - 2", "3"},
		{"1 / 3 * 3", "1"},
		{"(1/3) + (1/3) + (1/3)", "1"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := testEval(tt.input)
			testNumberValue(t, result, tt.expected)
		})
	}
}

func TestEvalComparisons(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"1 < 2", true},
		{"2 < 1", false},
		{"1 > 2", false},
		{"2 > 1", true},
		{"1 == 1", true},
		{"1 == 2", false},
		{"1 != 2", true},
		{"1 != 1", false},
		{"1 <= 1", true},
		{"1 <= 2", true},
		{"2 <= 1", false},
		{"1 >= 1", true},
		{"2 >= 1", true},
		{"1 >= 2", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := testEval(tt.input)
			testBooleanValue(t, result, tt.expected)
		})
	}
}

func TestEvalLogicalOperators(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true and true", true},
		{"true and false", false},
		{"false and true", false},
		{"false and false", false},
		{"true or true", true},
		{"true or false", true},
		{"false or true", true},
		{"false or false", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := testEval(tt.input)
			testBooleanValue(t, result, tt.expected)
		})
	}
}

func TestEvalLetStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"let x = 5\nx", "5"},
		{"let x = 5 * 5\nx", "25"},
		{"let a = 5\nlet b = a\nb", "5"},
		{"let a = 5\nlet b = a\nlet c = a + b + 5\nc", "15"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := testEval(tt.input)
			testNumberValue(t, result, tt.expected)
		})
	}
}

func TestEvalAssignment(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"let x = 5\nx = 10\nx", "10"},
		{"let x = 1\nx = x + 1\nx", "2"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := testEval(tt.input)
			testNumberValue(t, result, tt.expected)
		})
	}
}

func TestEvalUndefinedVariable(t *testing.T) {
	result := testEval("x")

	err, ok := result.(*value.Error)
	if !ok {
		t.Fatalf("expected Error, got %T", result)
	}
	if err.Message != "undefined: x" {
		t.Errorf("wrong error: %s", err.Message)
	}
}

func TestEvalFunctions(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"let f = (n) { return n }\nf(5)", "5"},
		{"let f = (n) { return n * 2 }\nf(5)", "10"},
		{"let add = (a, b) { return a + b }\nadd(2, 3)", "5"},
		{"let f = (n) { return n * 2 }\nf(f(5))", "20"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := testEval(tt.input)
			testNumberValue(t, result, tt.expected)
		})
	}
}

func TestEvalClosures(t *testing.T) {
	input := `
let newAdder = (x) {
	return (n) { return x + n }
}
let addTwo = newAdder(2)
addTwo(3)
`
	result := testEval(input)
	testNumberValue(t, result, "5")
}

func TestEvalIfStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		isNull   bool
	}{
		{"if true { return 10 }", "10", false},
		{"if false { return 10 }", "", true},
		{"if 1 < 2 { return 10 }", "10", false},
		{"if 1 > 2 { return 10 }", "", true},
		{"if 1 > 2 { return 10 } else { return 20 }", "20", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := testEval(tt.input)
			if tt.isNull {
				if result != value.NULL {
					t.Errorf("expected NULL, got %v", result)
				}
			} else {
				testNumberValue(t, result, tt.expected)
			}
		})
	}
}

func TestEvalWhileStatements(t *testing.T) {
	input := `
let x = 0
while x < 5 {
	x = x + 1
}
x
`
	result := testEval(input)
	testNumberValue(t, result, "5")
}

func TestEvalLists(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"[1, 2, 3].0", "1"},
		{"[1, 2, 3].1", "2"},
		{"[1, 2, 3].2", "3"},
		{"first([1, 2, 3])", "1"},
		{"len([1, 2, 3])", "3"},
		{"len([])", "0"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := testEval(tt.input)
			testNumberValue(t, result, tt.expected)
		})
	}
}

func TestEvalShapedLists(t *testing.T) {
	input := `
let @point [x, y]
let p = [@point, 10, 20]
p.x + p.y
`
	result := testEval(input)
	testNumberValue(t, result, "30")
}

func TestEvalShapedListPositionalAccess(t *testing.T) {
	input := `
let @point [x, y]
let p = [@point, 10, 20]
p.0 + p.1
`
	result := testEval(input)
	testNumberValue(t, result, "30")
}

func TestEvalStrings(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"hello" + " world"`, "hello world"},
		{`len("hello")`, "5"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := testEval(tt.input)
			switch r := result.(type) {
			case *value.String:
				if r.Value != tt.expected {
					t.Errorf("got %q, want %q", r.Value, tt.expected)
				}
			case *value.Number:
				testNumberValue(t, result, tt.expected)
			default:
				t.Fatalf("unexpected type: %T", result)
			}
		})
	}
}

func TestEvalSymbols(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{":ok == :ok", true},
		{":ok == :error", false},
		{":ok != :error", true},
		{":foo == :foo", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := testEval(tt.input)
			testBooleanValue(t, result, tt.expected)
		})
	}
}

func TestEvalPipe(t *testing.T) {
	input := `
let double = (n) { return n * 2 }
let add1 = (n) { return n + 1 }
5 |> double() |> add1()
`
	result := testEval(input)
	testNumberValue(t, result, "11")
}

func TestEvalPipeWithBuiltins(t *testing.T) {
	input := "[1, 2, 3] |> first()"
	result := testEval(input)
	testNumberValue(t, result, "1")
}

func TestEvalMatchStatement(t *testing.T) {
	input := `
let @ok [value]
let result = [@ok, 42]
match result {
	[@ok, val] { val }
	_ { 0 }
}
`
	result := testEval(input)
	testNumberValue(t, result, "42")
}

func TestEvalMatchWildcard(t *testing.T) {
	input := `
match 999 {
	_ { "other" }
}
`
	result := testEval(input)
	s, ok := result.(*value.String)
	if !ok {
		t.Fatalf("expected String, got %T", result)
	}
	if s.Value != "other" {
		t.Errorf("got %q, want %q", s.Value, "other")
	}
}

func TestEvalDivisionByZero(t *testing.T) {
	result := testEval("1 / 0")

	list, ok := result.(*value.List)
	if !ok {
		t.Fatalf("expected List (error shape), got %T", result)
	}
	if list.Shape != "error" {
		t.Errorf("expected error shape, got %s", list.Shape)
	}
}

func TestEvalBuiltinType(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"type(5)", ":number"},
		{"type(true)", ":boolean"},
		{`type("hello")`, ":string"},
		{"type('a')", ":rune"},
		{"type(:ok)", ":symbol"},
		{"type([1, 2])", ":list"},
		{"type((n) { return n })", ":function"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := testEval(tt.input)
			sym, ok := result.(*value.Symbol)
			if !ok {
				t.Fatalf("expected Symbol, got %T", result)
			}
			if sym.Inspect() != tt.expected {
				t.Errorf("got %s, want %s", sym.Inspect(), tt.expected)
			}
		})
	}
}

func TestEvalBuiltinShape(t *testing.T) {
	input := `
let @point [x, y]
let p = [@point, 10, 20]
shape(p)
`
	result := testEval(input)
	sym, ok := result.(*value.Symbol)
	if !ok {
		t.Fatalf("expected Symbol, got %T", result)
	}
	if sym.Value != "point" {
		t.Errorf("got %s, want point", sym.Value)
	}
}

func TestEvalRecursion(t *testing.T) {
	input := `
let factorial = (n) {
	if n <= 1 { return 1 }
	return n * factorial(n - 1)
}
factorial(5)
`
	result := testEval(input)
	testNumberValue(t, result, "120")
}

// Helpers

func testEval(input string) value.Value {
	env := value.NewEnv(nil)
	return eval.Run(input, env)
}

func testNumberValue(t *testing.T, v value.Value, expected string) {
	t.Helper()
	num, ok := v.(*value.Number)
	if !ok {
		t.Fatalf("expected Number, got %T (%v)", v, v)
	}
	if num.Inspect() != expected {
		t.Errorf("got %s, want %s", num.Inspect(), expected)
	}
}

func testBooleanValue(t *testing.T, v value.Value, expected bool) {
	t.Helper()
	b, ok := v.(*value.Boolean)
	if !ok {
		t.Fatalf("expected Boolean, got %T (%v)", v, v)
	}
	if b.Value != expected {
		t.Errorf("got %v, want %v", b.Value, expected)
	}
}
