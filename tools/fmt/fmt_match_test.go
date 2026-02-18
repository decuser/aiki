package fmt
import "aiki/syntax"

import (
	"strings"
	"testing"

)

var testGrammar *syntax.Grammar

func init() {
	SetGrammar(syntax.Grammar())
}

func TestFmtMatchPreservesArms(t *testing.T) {
	input := `match x {
	1 { return "one" }
	2 { return "two" }
	_ { return "other" }
}`

	got, err := Format(input)
	if err != nil {
		t.Fatalf("format failed: %v", err)
	}

	// Must contain all arms
	if !strings.Contains(got, `"one"`) {
		t.Errorf("missing 'one' arm:\n%s", got)
	}
	if !strings.Contains(got, `"two"`) {
		t.Errorf("missing 'two' arm:\n%s", got)
	}
	if !strings.Contains(got, `"other"`) {
		t.Errorf("missing 'other' arm:\n%s", got)
	}
	if !strings.Contains(got, "_") {
		t.Errorf("missing wildcard arm:\n%s", got)
	}
}

func TestFmtMatchSymbolPatterns(t *testing.T) {
	input := `match type(val) {
	:number { return 1 }
	:string { return 2 }
	:symbol { return 3 }
	_ { return 0 }
}`

	got, err := Format(input)
	if err != nil {
		t.Fatalf("format failed: %v", err)
	}

	if !strings.Contains(got, ":number") {
		t.Errorf("missing :number pattern:\n%s", got)
	}
	if !strings.Contains(got, ":string") {
		t.Errorf("missing :string pattern:\n%s", got)
	}
	if !strings.Contains(got, ":symbol") {
		t.Errorf("missing :symbol pattern:\n%s", got)
	}
}

func TestFmtMatchNestedBlocks(t *testing.T) {
	input := `match x {
	:ok {
		if y {
			return 1
		}
		return 2
	}
	_ { return 0 }
}`

	got, err := Format(input)
	if err != nil {
		t.Fatalf("format failed: %v", err)
	}

	if !strings.Contains(got, "if y") {
		t.Errorf("missing nested if:\n%s", got)
	}
	if !strings.Contains(got, "return 1") {
		t.Errorf("missing nested return:\n%s", got)
	}
}

func TestFmtMatchIdempotent(t *testing.T) {
	input := `match type(val) {
	:number {
		return _hash_string(to_str(val))
	}
	:boolean {
		if val {
			return 1
		}
		return 0
	}
	_ {
		return 0
	}
}`

	first, err := Format(input)
	if err != nil {
		t.Fatalf("first format failed: %v", err)
	}

	second, err := Format(first)
	if err != nil {
		t.Fatalf("second format failed: %v", err)
	}

	if first != second {
		t.Errorf("not idempotent:\nfirst:\n%s\nsecond:\n%s", first, second)
	}
}

func TestFmtMatchBindingPattern(t *testing.T) {
	input := `match val {
	x { return x + 1 }
}`

	got, err := Format(input)
	if err != nil {
		t.Fatalf("format failed: %v", err)
	}

	if !strings.Contains(got, "x {") {
		t.Errorf("missing binding pattern:\n%s", got)
	}
	if !strings.Contains(got, "x + 1") {
		t.Errorf("missing body with bound variable:\n%s", got)
	}
}

// ============ PIPE EXPRESSION TESTS ============

func TestFmtPipeSingleLine(t *testing.T) {
	// Single pipe stays on one line
	input := `x |> f()`

	got, err := Format(input)
	if err != nil {
		t.Fatalf("format failed: %v", err)
	}

	// Should be on one line
	lines := strings.Split(strings.TrimSpace(got), "\n")
	if len(lines) != 1 {
		t.Errorf("single pipe should be one line, got %d:\n%s", len(lines), got)
	}
	if !strings.Contains(got, "|>") {
		t.Errorf("missing pipe operator:\n%s", got)
	}
}

func TestFmtPipeMultiLine(t *testing.T) {
	// Multiple pipes break across lines with |> at end
	input := `range(1, 10) |> filter(is_even) |> map(square) |> sum()`

	got, err := Format(input)
	if err != nil {
		t.Fatalf("format failed: %v", err)
	}

	// Should have multiple lines
	lines := strings.Split(strings.TrimSpace(got), "\n")
	if len(lines) < 2 {
		t.Errorf("multi-pipe should break across lines, got %d:\n%s", len(lines), got)
	}

	// Each |> should be at end of line (Go style)
	for i, line := range lines[:len(lines)-1] {
		if strings.Contains(line, "|>") && !strings.HasSuffix(strings.TrimSpace(line), "|>") {
			t.Errorf("line %d: pipe should be at end of line: %s", i+1, line)
		}
	}
}

func TestFmtPipePreservesExpressions(t *testing.T) {
	input := `range(1, 11) |> filter(is_even) |> map(square) |> sum() |> print()`

	got, err := Format(input)
	if err != nil {
		t.Fatalf("format failed: %v", err)
	}

	// All expressions must be preserved
	mustHave := []string{"range(1, 11)", "filter(is_even)", "map(square)", "sum()", "print()"}
	for _, expr := range mustHave {
		if !strings.Contains(got, expr) {
			t.Errorf("missing expression %q:\n%s", expr, got)
		}
	}
}

func TestFmtPipeIdempotent(t *testing.T) {
	input := `range(1, 10) |> filter(f) |> map(g) |> println()`

	first, err := Format(input)
	if err != nil {
		t.Fatalf("first format failed: %v", err)
	}

	second, err := Format(first)
	if err != nil {
		t.Fatalf("second format failed: %v", err)
	}

	if first != second {
		t.Errorf("not idempotent:\nfirst:\n%s\nsecond:\n%s", first, second)
	}
}

func TestFmtPipeInFunction(t *testing.T) {
	input := `let process = (data) {
	return data |> filter(valid) |> map(transform) |> collect()
}`

	got, err := Format(input)
	if err != nil {
		t.Fatalf("format failed: %v", err)
	}

	if !strings.Contains(got, "filter(valid)") {
		t.Errorf("missing filter:\n%s", got)
	}
	if !strings.Contains(got, "map(transform)") {
		t.Errorf("missing map:\n%s", got)
	}
	if !strings.Contains(got, "collect()") {
		t.Errorf("missing collect:\n%s", got)
	}
}
