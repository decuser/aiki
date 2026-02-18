package fmt

import (
	"strings"
	"testing"

	"aiki/internal/ebnf"
	"aiki/lang"
)

var commentTestGrammar *ebnf.Grammar

func init() {
	SetGrammar(lang.Grammar())
}

func TestFmtCommentOrder(t *testing.T) {
	input := "# first comment\n# second comment\nlet x = 1\n"

	got, err := Format(input)
	if err != nil {
		t.Fatalf("format failed: %v", err)
	}

	if !strings.Contains(got, "# first comment") {
		t.Errorf("missing first comment:\n%s", got)
	}
	if !strings.Contains(got, "# second comment") {
		t.Errorf("missing second comment:\n%s", got)
	}

	// First should come before second
	i1 := strings.Index(got, "# first")
	i2 := strings.Index(got, "# second")
	if i1 > i2 {
		t.Errorf("comments out of order:\n%s", got)
	}
}

func TestFmtCommentOrderMultipleBlocks(t *testing.T) {
	input := "# block one\n# still block one\nlet x = 1\n\n# block two\n# still block two\nlet y = 2\n"

	got, err := Format(input)
	if err != nil {
		t.Fatalf("format failed: %v", err)
	}

	if !strings.Contains(got, "# block one") {
		t.Errorf("missing 'block one' comment:\n%s", got)
	}
	if !strings.Contains(got, "# still block one") {
		t.Errorf("missing 'still block one' comment:\n%s", got)
	}
	if !strings.Contains(got, "# block two") {
		t.Errorf("missing 'block two' comment:\n%s", got)
	}
	if !strings.Contains(got, "# still block two") {
		t.Errorf("missing 'still block two' comment:\n%s", got)
	}
}

func TestFmtEOLComment(t *testing.T) {
	input := "let x = 1  # end of line comment\n"

	got, err := Format(input)
	if err != nil {
		t.Fatalf("format failed: %v", err)
	}

	if !strings.Contains(got, "# end of line comment") {
		t.Errorf("missing EOL comment:\n%s", got)
	}

	// EOL comment should be on same line as statement
	lines := strings.Split(got, "\n")
	found := false
	for _, line := range lines {
		if strings.Contains(line, "let x") && strings.Contains(line, "# end of line") {
			found = true
		}
	}
	if !found {
		t.Errorf("EOL comment not on same line as statement:\n%s", got)
	}
}

func TestFmtBlankLinePreservation(t *testing.T) {
	input := "let x = 1\n\nlet y = 2\n"

	got, err := Format(input)
	if err != nil {
		t.Fatalf("format failed: %v", err)
	}

	// Should have a blank line between the two lets
	if !strings.Contains(got, "let x = 1\n\nlet y = 2") {
		t.Errorf("blank line not preserved:\n%q", got)
	}
}

func TestFmtIdempotentWithComments(t *testing.T) {
	input := "# Hash map implementation\n# A hash is a list of 64 buckets\n\nlet _HASH_SIZE = 64\n"

	first, err := Format(input)
	if err != nil {
		t.Fatalf("first format failed: %v", err)
	}

	second, err := Format(first)
	if err != nil {
		t.Fatalf("second format failed: %v", err)
	}

	if first != second {
		t.Errorf("not idempotent with comments:\nfirst:\n%s\nsecond:\n%s", first, second)
	}
}
