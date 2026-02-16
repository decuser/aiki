package tests

import (
	"testing"

	"aiki/internal/ebnf"
	"aiki/lang/eval"
	"aiki/lang/value"
	"aiki/runtime/prelude"
)

var testGrammar *ebnf.Grammar

func init() {
	var err error
	testGrammar, err = ebnf.ParseFile("../cmd/aiki/grammar.ebnf")
	if err != nil {
		panic("failed to load grammar: " + err.Error())
	}
	eval.SetNodeGrammar(testGrammar)
}

// setupEnv creates an env with strict loaded.
func setupEnv() *value.Env {
	env := value.NewEnv(nil)
	result := eval.RunNode(testGrammar, prelude.Source, env)
	if _, ok := result.(*value.Error); ok {
		panic("failed to load strict: " + result.Inspect())
	}
	env.SnapshotStrict()
	return env
}

// testEval evaluates without strict (bare env).
func testEval(input string) value.Value {
	env := value.NewEnv(nil)
	return eval.RunNode(testGrammar, input, env)
}

// testEvalStrict evaluates with strict loaded.
func EvalStrict(input string) value.Value {
	env := setupEnv()
	return eval.RunNode(testGrammar, input, env)
}

func testNumberValue(t *testing.T, v value.Value, expected string) {
	t.Helper()
	num, ok := v.(*value.Number)
	if !ok {
		t.Fatalf("expected Number, got %T (%v)", v, v)
	}
	if num.Inspect() != expected {
		t.Errorf("got %s, want %s", num.Inspect(), expected)
	}
}

func testBooleanValue(t *testing.T, v value.Value, expected bool) {
	t.Helper()
	b, ok := v.(*value.Boolean)
	if !ok {
		t.Fatalf("expected Boolean, got %T (%v)", v, v)
	}
	if b.Value != expected {
		t.Errorf("got %v, want %v", b.Value, expected)
	}
}
