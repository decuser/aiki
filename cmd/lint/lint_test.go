package lint

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
}

func lintSource(t *testing.T, source string) []Diagnostic {
	t.Helper()
	diags, err := LintSource(testGrammar, source)
	if err != nil {
		t.Fatalf("lint error: %v", err)
	}
	return diags
}

func TestLintCleanCode(t *testing.T) {
	diags := lintSource(t, `
let x = 5
let y = x + 1
`)
	if len(diags) > 0 {
		t.Errorf("expected no diagnostics, got %d: %v", len(diags), diags)
	}
}

func TestLintUndefined(t *testing.T) {
	diags := lintSource(t, `
let x = y + 1
`)
	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "undefined") && strings.Contains(d.Message, "'y'") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected undefined 'y' diagnostic, got %v", diags)
	}
}

func TestLintNamingViolation(t *testing.T) {
	diags := lintSource(t, `
let badName = 5
`)
	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "naming") && strings.Contains(d.Message, "badName") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected naming violation for 'badName', got %v", diags)
	}
}

func TestLintScreamingCase(t *testing.T) {
	diags := lintSource(t, `
let MAX_SIZE = 100
`)
	for _, d := range diags {
		if strings.Contains(d.Message, "MAX_SIZE") {
			t.Errorf("SCREAMING_SNAKE_CASE should be valid, got: %s", d.Message)
		}
	}
}

func TestLintForwardReference(t *testing.T) {
	// Two-pass: b uses a which is defined later at top level
	diags := lintSource(t, `
let b = a + 1
let a = 5
`)
	for _, d := range diags {
		if strings.Contains(d.Message, "undefined") && strings.Contains(d.Message, "'a'") {
			t.Errorf("forward reference 'a' should be allowed at top level, got: %s", d.Message)
		}
	}
}

func TestLintShadowWarning(t *testing.T) {
	diags := lintSource(t, `
let x = 5
let f = () {
	let x = 10
	x
}
`)
	found := false
	for _, d := range diags {
		if d.Level == "warning" && strings.Contains(d.Message, "shadow") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected shadow warning, got %v", diags)
	}
}

func TestLintBuiltinsOK(t *testing.T) {
	diags := lintSource(t, `
let x = length([1, 2, 3])
`)
	for _, d := range diags {
		if strings.Contains(d.Message, "undefined") && strings.Contains(d.Message, "length") {
			t.Errorf("builtin 'length' should be recognized, got: %s", d.Message)
		}
	}
}

func TestLintStrictExportsOK(t *testing.T) {
	diags := lintSource(t, `
let x = map([1, 2], (n) { return n + 1 })
`)
	for _, d := range diags {
		if strings.Contains(d.Message, "undefined") && strings.Contains(d.Message, "map") {
			t.Errorf("strict export 'map' should be recognized, got: %s", d.Message)
		}
	}
}

func TestLintParams(t *testing.T) {
	diags := lintSource(t, `
let f = (good_name) { return good_name }
`)
	if len(diags) > 0 {
		t.Errorf("expected no diagnostics for valid params, got %v", diags)
	}
}

func TestLintBadParam(t *testing.T) {
	diags := lintSource(t, `
let f = (badParam) { return badParam }
`)
	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "naming") && strings.Contains(d.Message, "badParam") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected naming violation for 'badParam', got %v", diags)
	}
}

func TestLintExportUndefined(t *testing.T) {
	diags := lintSource(t, `
export [does_not_exist]
`)
	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "export") && strings.Contains(d.Message, "does_not_exist") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected export warning for undefined name, got %v", diags)
	}
}

func TestLintExportPrivate(t *testing.T) {
	diags := lintSource(t, `
let _internal = 5
export [_internal]
`)
	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "_prefix") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected _prefix export warning, got %v", diags)
	}
}

func TestLintMatchPattern(t *testing.T) {
	diags := lintSource(t, `
let x = 5
match x {
	n { n + 1 }
}
`)
	// 'n' should be defined in the match arm scope
	for _, d := range diags {
		if strings.Contains(d.Message, "undefined") && strings.Contains(d.Message, "'n'") {
			t.Errorf("match pattern name 'n' should be defined in arm scope, got: %s", d.Message)
		}
	}
}
