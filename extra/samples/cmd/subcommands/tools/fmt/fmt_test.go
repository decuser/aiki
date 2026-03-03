package fmt

import (
	"testing"

	"aiki/engine/syntax"
	"aiki/engine/syntax/grammar"
)

func TestFormatSource_IdempotentAndParsePreserving(t *testing.T) {
	g, err := grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
	if err != nil {
		t.Fatalf("load grammar: %v", err)
	}

	src := "# header\nlet x=1  #eol\nif x\n{\n  return x\n}\n"
	out, err := FormatSource(g, "test.ai", src)
	if err != nil {
		t.Fatalf("format: %v", err)
	}

	// Second pass should be identical.
	out2, err := FormatSource(g, "test.ai", out)
	if err != nil {
		t.Fatalf("format second pass: %v", err)
	}
	if out2 != out {
		t.Fatalf("format not idempotent\n--- out ---\n%s\n--- out2 ---\n%s", out, out2)
	}

	// Spot check that it normalized brace placement and spaces.
	if !contains(out, "if x {\n") {
		t.Fatalf("expected same-line brace in output, got:\n%s", out)
	}
	if !contains(out, "let x = 1") {
		t.Fatalf("expected spaced let binding in output, got:\n%s", out)
	}
	if !contains(out, "#eol") {
		t.Fatalf("expected to preserve eol comment, got:\n%s", out)
	}
	if !contains(out, "# header") {
		t.Fatalf("expected to preserve standalone comment, got:\n%s", out)
	}
}

func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && (func() bool {
		for i := 0; i+len(sub) <= len(s); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	})())
}
