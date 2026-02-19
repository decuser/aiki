package testutil

import (
	"testing"

	"aiki/runtime/prelude"
	"aiki/semantics/eval"
	"aiki/semantics/value"
	"aiki/syntax"
)

var testGrammar *syntax.Grammar

func init() {
	testGrammar = syntax.GetGrammar()
	eval.SetNodeGrammar(testGrammar)
}

// SetupEnv creates an env with prelude loaded.
func SetupEnv() *value.Env {
	env := value.NewEnv(nil)
	// Ensure host services are available for intrinsics like spawn.
	env.Host = NewTestHost()
	result := eval.RunNode(testGrammar, prelude.Source, env)
	if _, ok := result.(*value.Error); ok {
		panic("failed to load prelude: " + result.Inspect())
	}
	env.SnapshotPrelude()
	return env
}

// TestEval evaluates without prelude (bare env).
func TestEval(input string) value.Value {
	env := value.NewEnv(nil)
	env.Host = NewTestHost()
	return eval.RunNode(testGrammar, input, env)
}

// EvalPrelude evaluates with prelude loaded.
func EvalPrelude(input string) value.Value {
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
