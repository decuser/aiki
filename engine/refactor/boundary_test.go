package refactor

import (
	"fmt"
	"testing"

	"aiki/engine/runtime/hal"
	"aiki/engine/runtime/hal/substrate"
	"aiki/engine/runtime/prelude"
	"aiki/engine/semantics/evaluator"
	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
	"aiki/engine/syntax/grammar"
)

// TestBoundaryUserCannotSeeHAL verifies user scope cannot resolve _prefixed names.
func TestBoundaryUserCannotSeeHAL(t *testing.T) {
	code := `_print("should fail")`

	result, err := evalWithScope(code, hal.ScopeUser, false)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}

	// Should be an error about undefined _print
	if !value.IsError(result) {
		t.Errorf("expected error for _print in user scope, got: %s", result.Inspect())
	}
	errVal := result.(*value.Error)
	if errVal.Message != "undefined: _print" {
		t.Errorf("expected 'undefined: _print', got: %s", errVal.Message)
	}
}

// TestBoundaryUserCanSeePrintAfterPrelude verifies user scope can see print after prelude loads.
func TestBoundaryUserCanSeePrintAfterPrelude(t *testing.T) {
	code := `print`

	result, err := evalWithScope(code, hal.ScopeUser, true)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}

	// Should resolve to function from prelude, not error
	if value.IsError(result) {
		t.Errorf("expected print to resolve in user scope after prelude, got error: %s", result.Inspect())
	}
}

// TestBoundaryUserCannotSeePrintWithoutPrelude verifies user scope cannot see print without prelude.
func TestBoundaryUserCannotSeePrintWithoutPrelude(t *testing.T) {
	code := `print`

	result, err := evalWithScope(code, hal.ScopeUser, false)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}

	// Should be undefined without prelude
	if !value.IsError(result) {
		t.Errorf("expected error for print without prelude, got: %s", result.Inspect())
	}
}

// TestBoundaryPreludeCanSeeHAL verifies prelude scope can resolve _prefixed names.
func TestBoundaryPreludeCanSeeHAL(t *testing.T) {
	code := `_print`

	result, err := evalWithScope(code, hal.ScopePrelude, false)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}

	// Should resolve to builtin, not error
	if value.IsError(result) {
		t.Errorf("expected _print to resolve in prelude scope, got error: %s", result.Inspect())
	}
}

// TestBoundaryPreludeCannotSeeNonPrefixed verifies prelude scope cannot see non-prefixed without loading prelude.
func TestBoundaryPreludeCannotSeeNonPrefixed(t *testing.T) {
	code := `print`

	result, err := evalWithScope(code, hal.ScopePrelude, false)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}

	// Without prelude loading, print is not defined
	if !value.IsError(result) {
		t.Errorf("expected print to be undefined in prelude scope without prelude, got: %s", result.Inspect())
	}
}

// TestBoundaryHALRegistryOnlyHasUnderscored verifies all HAL registry entries are _prefixed.
func TestBoundaryHALRegistryOnlyHasUnderscored(t *testing.T) {
	rt := substrate.NewGoRuntime()

	// Try to get non-prefixed from prelude scope - should fail
	if _, ok := rt.GetBuiltin("print", hal.ScopePrelude); ok {
		t.Error("HAL registry should not contain non-prefixed 'print'")
	}
	if _, ok := rt.GetBuiltin("length", hal.ScopePrelude); ok {
		t.Error("HAL registry should not contain non-prefixed 'length'")
	}

	// _prefixed should work
	if _, ok := rt.GetBuiltin("_print", hal.ScopePrelude); !ok {
		t.Error("HAL registry should contain '_print'")
	}
	if _, ok := rt.GetBuiltin("_length", hal.ScopePrelude); !ok {
		t.Error("HAL registry should contain '_length'")
	}
}

// TestBoundaryUserScopeGetsNothingFromRuntime verifies user scope gets nothing from runtime directly.
func TestBoundaryUserScopeGetsNothingFromRuntime(t *testing.T) {
	rt := substrate.NewGoRuntime()

	// User scope should get nothing, even for _prefixed
	if rt.HasBuiltin("_print", hal.ScopeUser) {
		t.Error("user scope should not see _print")
	}
	if rt.HasBuiltin("print", hal.ScopeUser) {
		t.Error("user scope should not see print from runtime")
	}
}

// evalWithScope evaluates code with the specified scope.
func evalWithScope(code string, scope hal.Scope, loadPrelude bool) (value.Value, error) {
	g, err := grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
	if err != nil {
		return nil, err
	}

	rt := substrate.NewGoRuntime()
	env := value.NewEnv()

	// Optionally load prelude first
	if loadPrelude {
		if err := loadPreludeInto(g, rt, env); err != nil {
			return nil, err
		}
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

	ev := evaluator.NewWithScope(rt, nil, scope)

	return ev.Eval(ast, env), nil
}

// loadPreludeInto loads prelude.ai into the environment with ScopePrelude.
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

	ev := evaluator.NewWithScope(rt, nil, hal.ScopePrelude)
	result := ev.Eval(ast, env)
	if errVal, ok := result.(*value.Error); ok {
		return fmt.Errorf("%s", errVal.Message)
	}

	return nil
}
