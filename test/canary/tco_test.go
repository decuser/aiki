package canary

import (
	"strings"
	"testing"

	"aiki/engine/runtime/hal/substrate"
	"aiki/engine/semantics/evaluator"
	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
	"aiki/engine/syntax/grammar"
)

// eval helper for canary tests - uses prelude scope for HAL access.
func eval(t *testing.T, source string) value.Value {
	g, _ := grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
	lexer := syntax.NewLexer(g, "test", source, nil)
	tokens, _ := lexer.Tokenize()
	parser := syntax.NewParser(g, tokens, source, nil)
	ast, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	rt := substrate.NewGoRuntime()
	ev := evaluator.New(rt, nil)
	// Use prelude scope to access HAL primitives directly
	return ev.Eval(ast, value.NewEnvWithScope(value.ScopePrelude))
}

// TestStackLimitNonTailRecursion verifies stack overflow is caught for non-tail recursion.
func TestStackLimitNonTailRecursion(t *testing.T) {
	src := `
	_stack_limit(50)
	let down = (n) {
		if (n <= 0) { return 0 }
		return (1 + down((n - 1)))
	}
	down(1000)
	`
	v := eval(t, src)
	if !value.IsFault(v) {
		t.Fatalf("expected fault, got %T %s", v, v.Inspect())
	}
	if !strings.Contains(v.Inspect(), "stack overflow") {
		t.Fatalf("expected stack overflow, got %s", v.Inspect())
	}
}

// TestProperTailCallExplicitReturn verifies TCO works with explicit return.
func TestProperTailCallExplicitReturn(t *testing.T) {
	src := `
	_stack_limit(50)
	let sum_tail = (n, acc) {
		if (n <= 0) { return acc }
		return sum_tail((n - 1), (acc + n))
	}
	sum_tail(5000, 0)
	`
	v := eval(t, src)
	if value.IsFault(v) {
		t.Fatalf("unexpected fault: %s", v.Inspect())
	}
	if v.Inspect() != "12502500" {
		t.Fatalf("got %s", v.Inspect())
	}
}

// TestProperTailCallImplicitIf verifies TCO works with implicit tail in if.
func TestProperTailCallImplicitIf(t *testing.T) {
	src := `
	_stack_limit(50)
	let sum_if = (n, acc) {
	    if (n <= 0) { return acc }
	    return sum_if((n - 1), (acc + n))
	}
	return sum_if(5000, 0)
	`
	v := eval(t, src)
	if value.IsFault(v) {
		t.Fatalf("unexpected fault: %s", v.Inspect())
	}
	if v.Inspect() != "12502500" {
		t.Fatalf("got %s", v.Inspect())
	}
}

// TestProperTailCallImplicitMatch verifies TCO works with implicit tail in match.
func TestProperTailCallImplicitMatch(t *testing.T) {
	src := `
	_stack_limit(50)
	let sum_match = (n, acc) {
	    match n {
		0 { return acc }
		_ { return sum_match((n - 1), (acc + n)) }
	    }
	    return [@error, "unreachable"]
	}
	return sum_match(5000, 0)
	`
	v := eval(t, src)
	if value.IsFault(v) {
		t.Fatalf("unexpected fault: %s", v.Inspect())
	}
	if v.Inspect() != "12502500" {
		t.Fatalf("got %s", v.Inspect())
	}
}
