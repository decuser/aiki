package testutil

import (
	"testing"

	"aiki/internal/ebnf"
	"aiki/lang"
	"aiki/lang/eval"
	"aiki/lang/value"
	"aiki/runtime/prelude"
)

var testGrammar *ebnf.Grammar

func init() {
	testGrammar = lang.Grammar()
	eval.SetNodeGrammar(testGrammar)
}

// setupEnv creates an env with strict loaded.
func SetupEnv() *value.Env {
	env := value.NewEnv(nil)
	result := eval.RunNode(testGrammar, prelude.Source, env)
	if _, ok := result.(*value.Error); ok {
		panic("failed to load strict: " + result.Inspect())
	}
	env.SnapshotStrict()
	return env
}

// testEval evaluates without strict (bare env).
func TestEval(input string) value.Value {
	env := value.NewEnv(nil)
	return eval.RunNode(testGrammar, input, env)
}

// testEvalStrict evaluates with strict loaded.
func EvalStrict(input string) value.Value {
	env := SetupEnv()
	return eval.RunNode(testGrammar, input, env)
}

func TestNumberValue(t *testing.T, v value.Value, expected string) {
	t.Helper()
	num, ok := v.(*value.Number)
	if !ok {
		t.Fatalf("expected Number, got %T (%v)", v, v)
	}
	if num.Inspect() != expected {
		t.Errorf("got %s, want %s", num.Inspect(), expected)
	}
}

func TestBooleanValue(t *testing.T, v value.Value, expected bool) {
	t.Helper()
	b, ok := v.(*value.Boolean)
	if !ok {
		t.Fatalf("expected Boolean, got %T (%v)", v, v)
	}
	if b.Value != expected {
		t.Errorf("got %v, want %v", b.Value, expected)
	}
}
