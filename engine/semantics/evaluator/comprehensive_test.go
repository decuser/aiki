package evaluator_test

import (
	"strings"
	"testing"

	"aiki/engine/internal"
	"aiki/engine/runtime/hal/substrate"
	"aiki/engine/semantics/evaluator"
	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
	"aiki/engine/syntax/definition"
)

func eval(source string) (value.Value, error) {
	grammar := definition.New()
	grammar.SetObserver(internal.SilentObserver{})
	runtime := substrate.NewGoRuntime()
	obs := internal.SilentObserver{}

	lexer := syntax.NewLexer("test", source, grammar)
	parser, err := syntax.NewParser(lexer, grammar)
	if err != nil {
		return value.NullValue(), err
	}

	ast, err := parser.Parse()
	if err != nil {
		return value.NullValue(), err
	}

	e := evaluator.New(runtime, obs)
	e.SetGrammar(grammar)

	scope := evaluator.NewScope(nil)
	return e.Eval(ast, scope)
}

func evalStr(source string) string {
	result, err := eval(source)
	if err != nil {
		return "ERROR: " + err.Error()
	}
	return result.Inspect()
}

func TestNumbers(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"5", "5"},
		{"42", "42"},
		// {"3/4", "0.75"}, // Rationals - grammar may not support
		{"-5", "-5"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := evalStr(tt.input)
			if got != tt.expected {
				t.Errorf("got %s, want %s", got, tt.expected)
			}
		})
	}
}

func TestBooleans(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"true", "true"},
		{"false", "false"},
		// {"not true", "false"},   // needs 'not' keyword
		// {"not false", "true"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := evalStr(tt.input)
			if got != tt.expected {
				t.Errorf("got %s, want %s", got, tt.expected)
			}
		})
	}
}

func TestArithmetic(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1 + 2", "3"},
		{"5 - 3", "2"},
		{"2 * 3", "6"},
		{"8 / 2", "4"},
		{"1 + 2 * 3", "9"},  // left-to-right: (1+2)*3
		{"2 * 3 + 1", "7"},  // left-to-right: (2*3)+1
		{"10 / 2 + 3", "8"}, // left-to-right: (10/2)+3
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := evalStr(tt.input)
			if got != tt.expected {
				t.Errorf("got %s, want %s", got, tt.expected)
			}
		})
	}
}

func TestComparison(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1 < 2", "true"},
		{"2 < 1", "false"},
		{"1 > 2", "false"},
		{"2 > 1", "true"},
		{"1 <= 1", "true"},
		{"1 <= 2", "true"},
		{"2 <= 1", "false"},
		{"1 >= 1", "true"},
		{"2 >= 1", "true"},
		{"1 >= 2", "false"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := evalStr(tt.input)
			if got != tt.expected {
				t.Errorf("got %s, want %s", got, tt.expected)
			}
		})
	}
}

func TestStrings(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"hello"`, "hello"},
		{`"hello" + " " + "world"`, "hello world"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := evalStr(tt.input)
			if got != tt.expected {
				t.Errorf("got %s, want %s", got, tt.expected)
			}
		})
	}
}

func TestLet(t *testing.T) {
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
			got := evalStr(tt.input)
			if got != tt.expected {
				t.Errorf("got %s, want %s", got, tt.expected)
			}
		})
	}
}

func TestUndefinedIdentifier(t *testing.T) {
	got := evalStr("x")
	if !strings.Contains(got, "undefined") {
		t.Errorf("expected undefined error, got %s", got)
	}
}

func TestIf(t *testing.T) {
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
			got := evalStr(tt.input)
			if got != tt.expected {
				t.Errorf("got %s, want %s", got, tt.expected)
			}
		})
	}
}

func TestWhile(t *testing.T) {
	input := `let x = 0
while x < 5 {
    x = x + 1
}
x`
	got := evalStr(input)
	if got != "5" {
		t.Errorf("got %s, want 5", got)
	}
}

func TestFunction(t *testing.T) {
	input := `let add = (a, b) {
    return a + b
}
add(2, 3)`
	got := evalStr(input)
	if got != "5" {
		t.Errorf("got %s, want 5", got)
	}
}

func TestFunctionImplicitReturn(t *testing.T) {
	input := `let f = (x) {
    x + 1
}
f(5)`
	got := evalStr(input)
	if got != "6" {
		t.Errorf("got %s, want 6", got)
	}
}

func TestRecursion(t *testing.T) {
	input := `let factorial = (n) {
    if n <= 1 {
        return 1
    }
    return n * factorial(n - 1)
}
factorial(5)`
	got := evalStr(input)
	if got != "120" {
		t.Errorf("got %s, want 120", got)
	}
}

func TestClosure(t *testing.T) {
	input := `let makeAdder = (x) {
    return (y) { return x + y }
}
let addFive = makeAdder(5)
addFive(3)`
	got := evalStr(input)
	if got != "8" {
		t.Errorf("got %s, want 8", got)
	}
}

func TestDivisionByZero(t *testing.T) {
	got := evalStr("1 / 0")
	if !strings.Contains(got, "division by zero") {
		t.Errorf("expected division by zero error, got %s", got)
	}
}

func TestAssignment(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"let x = 5\nx = 10\nx", "10"},
		{"let x = 1\nlet y = 2\nx = y\nx", "2"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := evalStr(tt.input)
			if got != tt.expected {
				t.Errorf("got %s, want %s", got, tt.expected)
			}
		})
	}
}

func TestPrint(t *testing.T) {
	// Just test it doesn't crash - output goes to stdout
	got := evalStr("print(42)")
	if strings.Contains(got, "ERROR") {
		t.Errorf("print failed: %s", got)
	}
}

func TestLength(t *testing.T) {
	got := evalStr(`length("hello")`)
	if got != "5" {
		t.Errorf("got %s, want 5", got)
	}
}

func TestEqual(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"equal(5, 5)", "true"},
		{"equal(5, 3)", "false"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := evalStr(tt.input)
			if got != tt.expected {
				t.Errorf("got %s, want %s", got, tt.expected)
			}
		})
	}
}
