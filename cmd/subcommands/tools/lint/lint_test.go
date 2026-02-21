package lint

import "aiki/syntax"

import (
	"strings"
	"testing"
)

func init() {
	// Initialize strictExports by calling SetGrammar
	SetGrammar(syntax.GetGrammar())
}

func lintSource(t *testing.T, source string) []Diagnostic {
	t.Helper()
	diags, err := LintSource(syntax.GetGrammar(), source)
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

func TestLintPreludeExportsOK(t *testing.T) {
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
	// export takes symbols - :does_not_exist is valid syntax
	// The linter doesn't validate that exported symbols are defined
	// because that's a runtime/evaluator concern, not a lint concern.
	// This test verifies no lint errors for valid syntax.
	diags := lintSource(t, `
export(:does_not_exist)
`)
	// Should have no lint errors for valid syntax
	// (undefined export is a runtime error, not lint)
	for _, d := range diags {
		if d.Level == "error" {
			t.Errorf("expected no lint errors for export syntax, got: %s", d.Message)
		}
	}
}

func TestLintExportPrivate(t *testing.T) {
	// export takes symbols - the _prefix convention is about naming,
	// but exporting a _prefixed name isn't currently a lint warning.
	// This test verifies the syntax is valid.
	diags := lintSource(t, `
let _internal = 5
export(:_internal)
`)
	// No errors expected - _prefix export warning is future work
	for _, d := range diags {
		if d.Level == "error" {
			t.Errorf("expected no errors, got: %s", d.Message)
		}
	}
}

// ============ NEW TESTS FOR BLOCK-SCOPED VARIABLES ============

func TestLintWhileBlockVariable(t *testing.T) {
	// Variable defined inside while should be visible within the loop
	diags := lintSource(t, `
let i = 0
while i < 10 {
	let x = i * 2
	println(x)
	i = i + 1
}
`)
	for _, d := range diags {
		if strings.Contains(d.Message, "undefined") && strings.Contains(d.Message, "'x'") {
			t.Errorf("'x' defined in while block should be visible, got: %s", d.Message)
		}
	}
}

func TestLintIfBlockVariable(t *testing.T) {
	// Variable defined inside if should be visible within the block
	diags := lintSource(t, `
let cond = true
if cond {
	let result = 42
	println(result)
}
`)
	for _, d := range diags {
		if strings.Contains(d.Message, "undefined") && strings.Contains(d.Message, "'result'") {
			t.Errorf("'result' defined in if block should be visible, got: %s", d.Message)
		}
	}
}

func TestLintNestedBlockVariables(t *testing.T) {
	// Variables in nested blocks
	diags := lintSource(t, `
let i = 0
while i < 10 {
	let outer = i
	if outer > 5 {
		let inner = outer * 2
		println(inner)
	}
	i = i + 1
}
`)
	for _, d := range diags {
		if strings.Contains(d.Message, "undefined") {
			t.Errorf("nested block variables should be visible, got: %s", d.Message)
		}
	}
}

func TestLintBlockVariableNotVisibleOutside(t *testing.T) {
	// Variable defined inside block should NOT be visible outside
	diags := lintSource(t, `
if true {
	let inside = 42
}
println(inside)
`)
	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "undefined") && strings.Contains(d.Message, "'inside'") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected undefined 'inside' (out of scope), got %v", diags)
	}
}

func TestLintFunctionBlockVariable(t *testing.T) {
	// Variable defined inside function body
	diags := lintSource(t, `
let f = () {
	let local = 42
	return local
}
`)
	for _, d := range diags {
		if strings.Contains(d.Message, "undefined") && strings.Contains(d.Message, "'local'") {
			t.Errorf("'local' defined in function should be visible, got: %s", d.Message)
		}
	}
}

func TestLintNestedFunctionVariables(t *testing.T) {
	// Outer function's variables visible to inner
	diags := lintSource(t, `
let outer = () {
	let x = 10
	let inner = () {
		return x + 1
	}
	return inner()
}
`)
	for _, d := range diags {
		if strings.Contains(d.Message, "undefined") {
			t.Errorf("closure variable 'x' should be visible, got: %s", d.Message)
		}
	}
}

func TestLintMatchPatternBinding(t *testing.T) {
	// Variables bound in match patterns
	diags := lintSource(t, `
let val = 42
match val {
	x {
		println(x)
	}
}
`)
	for _, d := range diags {
		if strings.Contains(d.Message, "undefined") && strings.Contains(d.Message, "'x'") {
			t.Errorf("pattern binding 'x' should be visible in arm, got: %s", d.Message)
		}
	}
}

func TestLintRestParam(t *testing.T) {
	// Rest parameter should be visible in function body
	diags := lintSource(t, `
let f = (...args) {
	return length(args)
}
`)
	for _, d := range diags {
		if strings.Contains(d.Message, "undefined") && strings.Contains(d.Message, "'args'") {
			t.Errorf("rest param 'args' should be visible, got: %s", d.Message)
		}
	}
}

func TestLintMixedParams(t *testing.T) {
	// Regular params plus rest param
	diags := lintSource(t, `
let f = (a, b, ...rest) {
	return a + b + length(rest)
}
`)
	for _, d := range diags {
		if strings.Contains(d.Message, "undefined") {
			t.Errorf("all params should be visible, got: %s", d.Message)
		}
	}
}

func TestLintComplexWhileLoop(t *testing.T) {
	diags := lintSource(t, `
let _HASH_SIZE = 64
let make_buckets = () {
	let buckets = []
	let i = 0
	while i < _HASH_SIZE {
		buckets = append(buckets, [])
		i = i + 1
	}
	return buckets
}
`)
	for _, d := range diags {
		if d.Level == "error" {
			t.Errorf("expected no errors, got: %s", d.Message)
		}
	}
}
