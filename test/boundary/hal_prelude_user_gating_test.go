package boundary

import (
	"fmt"
	"testing"

	"aiki/engine/runtime/hal/substrate"
	"aiki/engine/runtime/prelude"
	"aiki/engine/semantics/evaluator"
	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
	"aiki/engine/syntax/grammar"
)

// TestUserCannotSeeHAL verifies user scope cannot resolve _prefixed names.
func TestUserCannotSeeHAL(t *testing.T) {
	code := `_print("should fail")`

	result, err := evalWithScope(code, value.ScopeUser, false)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}

	// Should be a fault about undefined _print
	if !value.IsFault(result) {
		t.Errorf("expected fault for _print without authority, got: %s", result.Inspect())
	}
	faultVal := result.(*value.Fault)
	if faultVal.Message != "undefined: _print" {
		t.Errorf("expected 'undefined: _print', got: %s", faultVal.Message)
	}
}

// TestUserCanSeePrintAfterPrelude verifies user scope can see print after prelude loads.
func TestUserCanSeePrintAfterPrelude(t *testing.T) {
	code := `print`

	result, err := evalWithScope(code, value.ScopeUser, true)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}

	// Should resolve to function from prelude, not fault
	if value.IsFault(result) {
		t.Errorf("expected print to resolve without authority after prelude, got fault: %s", result.Inspect())
	}
}

// TestUserCannotSeePrintWithoutPrelude verifies user scope cannot see print without prelude.
func TestUserCannotSeePrintWithoutPrelude(t *testing.T) {
	code := `print`

	result, err := evalWithScope(code, value.ScopeUser, false)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}

	// Should be undefined without prelude
	if !value.IsFault(result) {
		t.Errorf("expected fault for print without prelude, got: %s", result.Inspect())
	}
}

// TestPreludeCanSeeHAL verifies prelude scope can resolve _prefixed names.
func TestPreludeCanSeeHAL(t *testing.T) {
	code := `_print`

	result, err := evalWithScope(code, value.ScopePrelude, false)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}

	// Should resolve to builtin, not fault
	if value.IsFault(result) {
		t.Errorf("expected _print to resolve with canonical host authority, got fault: %s", result.Inspect())
	}
}

// TestPreludeCannotSeeNonPrefixed verifies prelude scope cannot see non-prefixed without loading prelude.
func TestPreludeCannotSeeNonPrefixed(t *testing.T) {
	code := `print`

	result, err := evalWithScope(code, value.ScopePrelude, false)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}

	// Without prelude loading, print is not defined
	if !value.IsFault(result) {
		t.Errorf("expected print to be undefined with canonical host authority without prelude, got: %s", result.Inspect())
	}
}

// TestRuntimeRegistryOnlyHasUnderscored verifies runtime binding names remain _prefixed.
func TestRuntimeRegistryOnlyHasUnderscored(t *testing.T) {
	rt := substrate.NewGoRuntime()

	// Canonical host authority does not change the runtime binding namespace.
	if _, ok := rt.GetBuiltin("print", value.NewAuthority("HAL.io.print", "_length")); ok {
		t.Error("HAL registry should not contain non-prefixed 'print'")
	}
	if _, ok := rt.GetBuiltin("length", value.NewAuthority("HAL.io.print", "_length")); ok {
		t.Error("HAL registry should not contain non-prefixed 'length'")
	}

	// Raw binding lookup succeeds only when the corresponding authority is present.
	if _, ok := rt.GetBuiltin("_print", value.NewAuthority("HAL.io.print", "_length")); !ok {
		t.Error("HAL registry should contain '_print'")
	}
	if _, ok := rt.GetBuiltin("_length", value.NewAuthority("HAL.io.print", "_length")); !ok {
		t.Error("HAL registry should contain '_length'")
	}
}

// TestUserScopeGetsNothingFromRuntime verifies user scope gets nothing from runtime directly.
func TestUserScopeGetsNothingFromRuntime(t *testing.T) {
	rt := substrate.NewGoRuntime()

	// User scope should get nothing, even for _prefixed
	if rt.HasBuiltin("_print", value.NoAuthority()) {
		t.Error("user scope should not see _print")
	}
	if rt.HasBuiltin("print", value.NoAuthority()) {
		t.Error("user scope should not see print from runtime")
	}
}

// TestUserCannotSeeHALChr verifies user scope cannot see _chr.
func TestUserCannotSeeHALChr(t *testing.T) {
	code := `_chr(65)`

	result, err := evalWithScope(code, value.ScopeUser, false)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}

	if !value.IsFault(result) {
		t.Errorf("expected fault for _chr without authority, got: %s", result.Inspect())
	}
}

// TestUserCanSeeChrAfterPrelude verifies user scope can see chr after prelude.
func TestUserCanSeeChrAfterPrelude(t *testing.T) {
	code := `chr(65)`

	result, err := evalWithScope(code, value.ScopeUser, true)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}

	if value.IsFault(result) {
		t.Errorf("expected chr to work without authority after prelude, got fault: %s", result.Inspect())
	}
	r, ok := result.(*value.Rune)
	if !ok || r.Val != 'A' {
		t.Errorf("expected 'A', got: %s", result.Inspect())
	}
}

// TestUserCanSeeAppendAfterPrelude verifies user scope can see append after prelude.
func TestUserCanSeeAppendAfterPrelude(t *testing.T) {
	code := `append([1, 2], 3)`

	result, err := evalWithScope(code, value.ScopeUser, true)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}

	if value.IsFault(result) {
		t.Errorf("expected append to work without authority after prelude, got fault: %s", result.Inspect())
	}
	list, ok := result.(*value.List)
	if !ok || len(list.Elements) != 3 {
		t.Errorf("expected [1, 2, 3], got: %s", result.Inspect())
	}
}

// TestChrOrdRoundTrip verifies chr and ord are inverses.
func TestChrOrdRoundTrip(t *testing.T) {
	code := `ord(chr(8364))`

	result, err := evalWithScope(code, value.ScopeUser, true)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}

	if value.IsFault(result) {
		t.Errorf("expected round trip to work, got fault: %s", result.Inspect())
	}
	num, ok := result.(*value.Number)
	if !ok {
		t.Fatalf("expected number, got: %s", result.Inspect())
	}
	if !num.Equal(value.NewNumber(8364, 1)) {
		t.Errorf("expected 8364, got: %s", result.Inspect())
	}
}

// TestUserCannotSeeHALUpper verifies user scope cannot see _upper.
func TestUserCannotSeeHALUpper(t *testing.T) {
	code := `_upper("hello")`

	result, err := evalWithScope(code, value.ScopeUser, false)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}

	if !value.IsFault(result) {
		t.Errorf("expected fault for _upper without authority, got: %s", result.Inspect())
	}
}

// evalWithScope evaluates code with the specified scope.
func evalWithScope(code string, scope value.Scope, loadPrel bool) (value.Value, error) {
	g, err := grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
	if err != nil {
		return nil, err
	}

	rt := substrate.NewGoRuntime()

	// Create env with specified scope
	env := value.NewEnvWithScope(scope)
	if scope == value.ScopePrelude {
		env.SetAuthority(rt.AuthorityForSource("engine/runtime/prelude/prelude.ai"))
	}

	// Optionally load prelude first (into a prelude-scope env, then enclose)
	if loadPrel {
		preludeEnv := value.NewEnvWithScope(value.ScopePrelude)
		preludeEnv.SetAuthority(rt.AuthorityForSource("engine/runtime/prelude/prelude.ai"))
		if err := loadPreludeInto(g, rt, preludeEnv); err != nil {
			return nil, err
		}
		// Create user env enclosed by prelude
		env = value.NewEnclosedEnv(preludeEnv)
		env.SetAuthority(value.NoAuthority())
	}

	lexer := syntax.NewLexer(g, "<test>", code, nil)
	tokens, err := lexer.Tokenize()
	if err != nil {
		return nil, err
	}

	parser := syntax.NewParser(g, tokens, code, nil)
	ast, err := parser.Parse()
	if err != nil {
		return nil, err
	}

	ev := evaluator.New(rt, nil)

	return ev.Eval(ast, env), nil
}

// loadPreludeInto loads prelude.ai into the environment.
func loadPreludeInto(g *grammar.Grammar, rt *substrate.GoRuntime, env *value.Env) error {
	lexer := syntax.NewLexer(g, "<prelude>", prelude.Source, nil)
	tokens, err := lexer.Tokenize()
	if err != nil {
		return err
	}

	parser := syntax.NewParser(g, tokens, prelude.Source, nil)
	ast, err := parser.Parse()
	if err != nil {
		return err
	}

	ev := evaluator.New(rt, nil)
	result := ev.Eval(ast, env)
	if faultVal, ok := result.(*value.Fault); ok {
		return fmt.Errorf("%s", faultVal.Message)
	}

	return nil
}
