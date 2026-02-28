package refactor

import (
	"testing"

	"aiki/engine/runtime/hal"
	"aiki/engine/runtime/hal/substrate"
	"aiki/engine/semantics/evaluator"
	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
	"aiki/engine/syntax/grammar"
)

// TestBoundaryUserCannotSeeHAL verifies user scope cannot resolve _prefixed names.
func TestBoundaryUserCannotSeeHAL(t *testing.T) {
	code := `_print("should fail")`

	result, err := evalWithScope(code, hal.ScopeUser)
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

// TestBoundaryUserCanSeePrint verifies user scope can resolve non-prefixed names.
func TestBoundaryUserCanSeePrint(t *testing.T) {
	code := `print`

	result, err := evalWithScope(code, hal.ScopeUser)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}

	// Should resolve to builtin, not error
	if value.IsError(result) {
		t.Errorf("expected print to resolve in user scope, got error: %s", result.Inspect())
	}
}

// TestBoundaryPreludeCanSeeHAL verifies prelude scope can resolve _prefixed names.
func TestBoundaryPreludeCanSeeHAL(t *testing.T) {
	code := `_print`

	result, err := evalWithScope(code, hal.ScopePrelude)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}

	// Should resolve to builtin, not error
	if value.IsError(result) {
		t.Errorf("expected _print to resolve in prelude scope, got error: %s", result.Inspect())
	}
}

// TestBoundaryPreludeCanSeePrint verifies prelude scope can also see non-prefixed.
func TestBoundaryPreludeCanSeePrint(t *testing.T) {
	code := `print`

	result, err := evalWithScope(code, hal.ScopePrelude)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}

	if value.IsError(result) {
		t.Errorf("expected print to resolve in prelude scope, got error: %s", result.Inspect())
	}
}

// evalWithScope evaluates code with the specified scope.
func evalWithScope(code string, scope hal.Scope) (value.Value, error) {
	g, err := grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
	if err != nil {
		return nil, err
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

	rt := substrate.NewGoRuntime()
	ev := evaluator.NewWithScope(rt, nil, scope)
	env := value.NewEnv()

	return ev.Eval(ast, env), nil
}
