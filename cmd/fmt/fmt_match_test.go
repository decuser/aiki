package fmt

import (
	"strings"
	"testing"

	"aiki/ebnf"
)

var testGrammar *ebnf.Grammar

func init() {
	var err error
	testGrammar, err = ebnf.ParseFile("../../cmd/grammar.ebnf")
	if err != nil {
		panic("failed to load grammar: " + err.Error())
	}
	SetGrammar(testGrammar)
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
