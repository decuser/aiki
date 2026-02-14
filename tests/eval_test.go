package tests

import (
	"testing"

	"aiki/lang/value"
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
		{"10 / 2 + 3", "8"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := testEval(tt.input)
			testNumberValue(t, result, tt.expected)
		})
	}
}

func TestEvalComparison(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"1 < 2", true},
		{"2 < 1", false},
		{"1 > 2", false},
		{"2 > 1", true},
		{"1 <= 1", true},
		{"1 <= 2", true},
		{"2 <= 1", false},
		{"1 >= 1", true},
		{"2 >= 1", true},
		{"1 >= 2", false},
		{"5 == 5", true},
		{"5 == 3", false},
		{"5 != 3", true},
		{"5 != 5", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := testEval(tt.input)
			testBooleanValue(t, result, tt.expected)
		})
	}
}

func TestEvalLogical(t *testing.T) {
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

func TestEvalStrings(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"hello"`, "hello"},
		{`"hello" + " " + "world"`, "hello world"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := testEval(tt.input)
			str, ok := result.(*value.String)
			if !ok {
				t.Fatalf("expected String, got %T", result)
			}
			if str.Value != tt.expected {
				t.Errorf("got %s, want %s", str.Value, tt.expected)
			}
		})
	}
}

func TestEvalLet(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"let x = 5\nx", "5"},
		{"let x = 5\nlet y = 10\nx + y", "15"},
		{"let x = 5\nlet y = x\ny", "5"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := testEval(tt.input)
			testNumberValue(t, result, tt.expected)
		})
	}
}

func TestUndefinedIdentifier(t *testing.T) {
	result := testEval("x")
	err, ok := result.(*value.Error)
	if !ok {
		t.Fatalf("expected Error, got %T", result)
	}
	if err.Message != "undefined: x" {
		t.Errorf("got %s, want 'undefined: x'", err.Message)
	}
}

func TestEvalIf(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"if true { 1 }", "1"},
		{"if false { 1 }", "null"},
		{"if true { 1 } else { 2 }", "1"},
		{"if false { 1 } else { 2 }", "2"},
		{"if 1 < 2 { 10 } else { 20 }", "10"},
		{"if 1 > 2 { 10 } else { 20 }", "20"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := testEval(tt.input)
			if tt.expected == "null" {
				if result.Type() != value.NullType {
					t.Errorf("expected null, got %s", result.Inspect())
				}
			} else {
				testNumberValue(t, result, tt.expected)
			}
		})
	}
}

func TestEvalWhile(t *testing.T) {
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

func TestEvalList(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"[1, 2, 3]", "[1, 2, 3]"},
		{"[]", "[]"},
		{"[1 + 1, 2 + 2]", "[2, 4]"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := testEval(tt.input)
			if result.Inspect() != tt.expected {
				t.Errorf("got %s, want %s", result.Inspect(), tt.expected)
			}
		})
	}
}

func TestEvalListIndex(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"[1, 2, 3][0]", "1"},
		{"[1, 2, 3][1]", "2"},
		{"[1, 2, 3][2]", "3"},
		{"let a = [10, 20, 30]\na[1]", "20"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := testEval(tt.input)
			testNumberValue(t, result, tt.expected)
		})
	}
}

func TestEvalFunction(t *testing.T) {
	input := `
let add = (a, b) {
    return a + b
}
add(2, 3)
`
	result := testEval(input)
	testNumberValue(t, result, "5")
}

func TestEvalRecursion(t *testing.T) {
	input := `
let factorial = (n) {
    if n <= 1 {
        return 1
    }
    return n * factorial(n - 1)
}
factorial(5)
`
	result := testEval(input)
	testNumberValue(t, result, "120")
}

func TestEvalClosure(t *testing.T) {
	input := `
let makeAdder = (x) {
    return (y) { return x + y }
}
let addFive = makeAdder(5)
addFive(3)
`
	result := testEval(input)
	testNumberValue(t, result, "8")
}

func TestEvalRestParam(t *testing.T) {
	input := `
let sum = (...args) {
    let total = 0
    let i = 0
    while i < length(args) {
        total = total + args[i]
        i = i + 1
    }
    return total
}
sum(1, 2, 3, 4, 5)
`
	result := testEval(input)
	testNumberValue(t, result, "15")
}

func TestEvalShapedList(t *testing.T) {
	input := `
let @point [x, y]
let p = [@point, 10, 20]
p.x + p.y
`
	result := testEval(input)
	testNumberValue(t, result, "30")
}

func TestEvalDivisionByZero(t *testing.T) {
	result := testEval("1 / 0")
	err, ok := result.(*value.Error)
	if !ok {
		t.Fatalf("expected Error, got %T", result)
	}
	if err.Message != "division by zero" {
		t.Errorf("got %s, want 'division by zero'", err.Message)
	}
}

func TestEvalAssignment(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"let x = 5\nx = 10\nx", "10"},
		{"let x = 1\nlet y = 2\nx = y\nx", "2"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := testEval(tt.input)
			testNumberValue(t, result, tt.expected)
		})
	}
}

func TestEvalFunctionReturn(t *testing.T) {
	input := `
let earlyReturn = (n) {
    if n < 0 {
        return -1
    }
    return n * 2
}
earlyReturn(-5)
`
	result := testEval(input)
	testNumberValue(t, result, "-1")
}

func TestEvalNestedFunction(t *testing.T) {
	input := `
let outer = (x) {
    let inner = (y) {
        return x + y
    }
    return inner(10)
}
outer(5)
`
	result := testEval(input)
	testNumberValue(t, result, "15")
}

func TestEvalImplicitReturn(t *testing.T) {
	input := `
let f = (x) {
    x + 1
}
f(5)
`
	result := testEval(input)
	testNumberValue(t, result, "6")
}
