package lint

import (
	"testing"

	"aiki/ebnf"
)

var testGrammar *ebnf.Grammar

func init() {
	// From cmd/lint/, grammar.ebnf is at ../grammar.ebnf
	g, err := ebnf.ParseFile("../grammar.ebnf")
	if err != nil {
		panic("failed to load grammar: " + err.Error())
	}
	testGrammar = g
	SetGrammar(g)
}

func TestLintCleanCode(t *testing.T) {
	clean := []string{
		`let x = 5`,
		`let add = (a, b) { return a + b }`,
		`let result = add(1, 2)`,
		`if true { let y = 1 }`,
		`while true { let z = 1 }`,
		`let f = (n) { return n + 1 }`,
		`let items = [1, 2, 3]`,
		`print("hello")`,
		`let MAX_SIZE = 100`,
		`let _internal = 0`,
	}

	for _, src := range clean {
		t.Run(src, func(t *testing.T) {
			diags, err := LintSource(grammar, src)
			if err != nil {
				t.Fatalf("lint error: %v", err)
			}
			errors := filterErrors(diags)
			if len(errors) > 0 {
				t.Errorf("expected clean, got errors: %v", errors)
			}
		})
	}
}

func TestLintUndefined(t *testing.T) {
	tests := []struct {
		name   string
		source string
		expect string
	}{
		{
			name:   "undefined variable",
			source: `let x = y`,
			expect: "undefined: 'y'",
		},
		{
			name:   "undefined function",
			source: `foo(1)`,
			expect: "undefined: 'foo'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diags, err := LintSource(grammar, tt.source)
			if err != nil {
				t.Fatalf("lint error: %v", err)
			}
			if !hasDiagContaining(diags, tt.expect) {
				t.Errorf("expected diagnostic containing %q, got: %v", tt.expect, diags)
			}
		})
	}
}

func TestLintNaming(t *testing.T) {
	tests := []struct {
		name   string
		source string
		expect string
	}{
		{
			name:   "camelCase rejected",
			source: `let myVar = 5`,
			expect: "naming:",
		},
		{
			name:   "PascalCase rejected",
			source: `let MyVar = 5`,
			expect: "naming:",
		},
		{
			name:   "MixedSCREAM rejected",
			source: `let MAX_size = 5`,
			expect: "naming:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diags, err := LintSource(grammar, tt.source)
			if err != nil {
				t.Fatalf("lint error: %v", err)
			}
			if !hasDiagContaining(diags, tt.expect) {
				t.Errorf("expected diagnostic containing %q, got: %v", tt.expect, diags)
			}
		})
	}
}

func TestLintValidNaming(t *testing.T) {
	valid := []string{
		`let x = 1`,
		`let my_var = 1`,
		`let MAX_SIZE = 1`,
		`let _private = 1`,
		`let _PRIVATE = 1`,
	}

	for _, src := range valid {
		t.Run(src, func(t *testing.T) {
			diags, err := LintSource(grammar, src)
			if err != nil {
				t.Fatalf("lint error: %v", err)
			}
			for _, d := range diags {
				if d.Level == "error" && contains(d.Message, "naming:") {
					t.Errorf("expected valid naming, got: %s", d.Message)
				}
			}
		})
	}
}

func TestLintShadowWarning(t *testing.T) {
	source := `let x = 1
let f = () {
	let x = 2
}`
	diags, err := LintSource(grammar, source)
	if err != nil {
		t.Fatalf("lint error: %v", err)
	}
	if !hasDiagContaining(diags, "shadow:") {
		t.Errorf("expected shadow warning, got: %v", diags)
	}
	// Should be warning, not error
	for _, d := range diags {
		if contains(d.Message, "shadow:") && d.Level != "warning" {
			t.Errorf("shadow should be warning, got %s", d.Level)
		}
	}
}

func TestLintExportPrefixWarning(t *testing.T) {
	source := `let _private = 1
export [_private]`
	diags, err := LintSource(grammar, source)
	if err != nil {
		t.Fatalf("lint error: %v", err)
	}
	if !hasDiagContaining(diags, "_prefix") {
		t.Errorf("expected _prefix warning, got: %v", diags)
	}
}

func TestLintBuiltinsAvailable(t *testing.T) {
	// HAL builtins should not trigger undefined
	sources := []string{
		`print("hello")`,
		`let n = length([1, 2, 3])`,
		`let t = type(42)`,
		`let f = first([1, 2])`,
		`let r = rest([1, 2])`,
	}

	for _, src := range sources {
		t.Run(src, func(t *testing.T) {
			diags, err := LintSource(grammar, src)
			if err != nil {
				t.Fatalf("lint error: %v", err)
			}
			errors := filterErrors(diags)
			if len(errors) > 0 {
				t.Errorf("expected clean (builtins available), got: %v", errors)
			}
		})
	}
}

func TestLintStrictExportsAvailable(t *testing.T) {
	// Strict exports should not trigger undefined
	sources := []string{
		`let r = range(1, 10)`,
		`let s = sum([1, 2, 3])`,
		`let m = map([1, 2], (x) { return x })`,
		`let h = hash_new()`,
	}

	for _, src := range sources {
		t.Run(src, func(t *testing.T) {
			diags, err := LintSource(grammar, src)
			if err != nil {
				t.Fatalf("lint error: %v", err)
			}
			errors := filterErrors(diags)
			if len(errors) > 0 {
				t.Errorf("expected clean (strict exports available), got: %v", errors)
			}
		})
	}
}

func TestLintFuncParams(t *testing.T) {
	// Parameters should be defined within function body
	source := `let f = (x, y) { return x + y }`
	diags, err := LintSource(grammar, source)
	if err != nil {
		t.Fatalf("lint error: %v", err)
	}
	errors := filterErrors(diags)
	if len(errors) > 0 {
		t.Errorf("expected clean, got: %v", errors)
	}
}

func TestLintMatchPatternBindings(t *testing.T) {
	// Pattern variables should be available in match arm body
	source := `let x = 1
match x {
	n { return n }
}`
	diags, err := LintSource(grammar, source)
	if err != nil {
		t.Fatalf("lint error: %v", err)
	}
	errors := filterErrors(diags)
	if len(errors) > 0 {
		t.Errorf("expected clean (pattern bindings), got: %v", errors)
	}
}

// --- helpers ---

func filterErrors(diags []Diagnostic) []Diagnostic {
	var errs []Diagnostic
	for _, d := range diags {
		if d.Level == "error" {
			errs = append(errs, d)
		}
	}
	return errs
}

func hasDiagContaining(diags []Diagnostic, substr string) bool {
	for _, d := range diags {
		if contains(d.Message, substr) {
			return true
		}
	}
	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchStr(s, substr)
}

func searchStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
