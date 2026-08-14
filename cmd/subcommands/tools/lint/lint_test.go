package lint

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
	"aiki/engine/syntax/grammar"
)

func TestLint_CheckFormatting(t *testing.T) {
	dir := t.TempDir()
	p1 := filepath.Join(dir, "a.ai")
	// Unformatted: missing spaces and brace on next line.
	if err := os.WriteFile(p1, []byte("let x=1\nif x {\nreturn x\n}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	p2 := filepath.Join(dir, "b.ai")
	if err := os.WriteFile(p2, []byte("let y = 2\n"), 0644); err != nil {
		t.Fatal(err)
	}

	bad, err := CheckFormatting([]string{dir + "/..."}, false)
	if err != nil {
		t.Fatalf("CheckFormatting: %v", err)
	}
	if len(bad) != 1 {
		t.Fatalf("expected 1 bad file, got %d: %v", len(bad), bad)
	}
	if bad[0] != p1 {
		t.Fatalf("expected bad file %s, got %s", p1, bad[0])
	}
}

func lintSource(t *testing.T, src string) []Diagnostic {
	t.Helper()
	g, err := grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
	if err != nil {
		t.Fatal(err)
	}
	diags, err := LintSource(g, "test.ai", src, value.ScopeUser)
	if err != nil {
		t.Fatalf("LintSource error: %v", err)
	}
	return diags
}

func TestLintCleanCode(t *testing.T) {
	diags := lintSource(t, "let x = 5\nlet y = x + 1\n")
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
}

func TestLintUndefined(t *testing.T) {
	diags := lintSource(t, "let x = y + 1\n")
	found := false
	for _, d := range diags {
		if d.Level == "error" && strings.Contains(d.Message, "undefined") && strings.Contains(d.Message, "y") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected undefined y, got %v", diags)
	}
}

func TestLintNamingViolation(t *testing.T) {
	diags := lintSource(t, "let badName = 5\n")
	found := false
	for _, d := range diags {
		if d.Level == "warning" && strings.Contains(d.Message, "naming") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected naming warning, got %v", diags)
	}
}

func TestLintForwardReferenceTopLevel(t *testing.T) {
	diags := lintSource(t, "let b = a + 1\nlet a = 5\n")
	for _, d := range diags {
		if d.Level == "error" && strings.Contains(d.Message, "undefined") && strings.Contains(d.Message, "a") {
			t.Fatalf("forward reference should be allowed at top level, got %v", diags)
		}
	}
}

func TestLintShadowWarning(t *testing.T) {
	diags := lintSource(t, "let x = 5\nlet f = () {\n let x = 10\n x\n}\n")
	found := false
	for _, d := range diags {
		if d.Level == "warning" && strings.Contains(d.Message, "shadow") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected shadow warning, got %v", diags)
	}
}

func TestLintPreludeBuiltinsOK(t *testing.T) {
	diags := lintSource(t, "let x = length([1, 2, 3])\n")
	for _, d := range diags {
		if d.Level == "error" && strings.Contains(d.Message, "undefined") && strings.Contains(d.Message, "length") {
			t.Fatalf("prelude builtin length should be defined, got %v", diags)
		}
	}
}

func TestLintParamsNaming(t *testing.T) {
	diags := lintSource(t, "let f = (badParam) { return badParam }\n")
	found := false
	for _, d := range diags {
		if d.Level == "warning" && strings.Contains(d.Message, "badParam") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected param naming warning, got %v", diags)
	}
}

func TestLintMatchPatternBindsName(t *testing.T) {
	diags := lintSource(t, "match 1 { x { println(x) } }\n")
	for _, d := range diags {
		if d.Level == "error" && strings.Contains(d.Message, "undefined") && strings.Contains(d.Message, "x") {
			t.Fatalf("pattern binder x should be defined in arm, got %v", diags)
		}
	}
}

func TestLintIfBlockDoesNotScopeBindings(t *testing.T) {
	// Bindings inside if blocks are visible outside — the evaluator
	// does not create a new environment for if/while blocks.
	src := "if true {\n\tlet y = 42\n}\nlet z = y\n"
	diags := lintSource(t, src)
	for _, d := range diags {
		if d.Level == "error" && strings.Contains(d.Message, "undefined") && strings.Contains(d.Message, "y") {
			t.Fatalf("linter falsely reports y as undefined inside if block: %v", diags)
		}
	}
}

func TestLintWhileBlockDoesNotScopeBindings(t *testing.T) {
	src := "let i = 0\nwhile i < 1 {\n\tlet found = true\n\ti = i + 1\n}\nlet x = found\n"
	diags := lintSource(t, src)
	for _, d := range diags {
		if d.Level == "error" && strings.Contains(d.Message, "undefined") && strings.Contains(d.Message, "found") {
			t.Fatalf("linter falsely reports found as undefined after while block: %v", diags)
		}
	}
}

func TestLintMatchArmScopesBindings(t *testing.T) {
	// Match arms DO create enclosed environments — pattern bindings
	// should not leak outside.
	src := "let val = 42\nmatch val {\n\tx { let inner = x }\n}\nlet z = inner\n"
	diags := lintSource(t, src)
	found := false
	for _, d := range diags {
		if d.Level == "error" && strings.Contains(d.Message, "undefined") && strings.Contains(d.Message, "inner") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected inner to be undefined outside match arm, got %v", diags)
	}
}

func TestLintFunctionBodyScopesBindings(t *testing.T) {
	// Function bodies create their own scope.
	src := "let f = () {\n\tlet local = 1\n}\nlet x = local\n"
	diags := lintSource(t, src)
	found := false
	for _, d := range diags {
		if d.Level == "error" && strings.Contains(d.Message, "undefined") && strings.Contains(d.Message, "local") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected local to be undefined outside function, got %v", diags)
	}
}
