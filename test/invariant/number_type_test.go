package invariant

import (
	"math/big"
	"testing"

	"aiki/engine/runtime/hal/substrate"
	"aiki/engine/semantics/evaluator"
	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
	"aiki/engine/syntax/grammar"
)

// eval helper for invariant tests
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
	return ev.Eval(ast, value.NewEnvWithScope(value.ScopePrelude))
}

// TestNumbersAreBigRat verifies all number operations return *big.Rat, not float64.
func TestNumbersAreBigRat(t *testing.T) {
	tests := []struct {
		name   string
		expr   string
		verify func(t *testing.T, v value.Value)
	}{
		{
			name: "integer literal",
			expr: "42",
			verify: func(t *testing.T, v value.Value) {
				assertBigRat(t, v)
			},
		},
		{
			name: "rational literal",
			expr: "1/3",
			verify: func(t *testing.T, v value.Value) {
				n := assertBigRat(t, v)
				if n != nil && n.Val.Cmp(big.NewRat(1, 3)) != 0 {
					t.Errorf("expected 1/3, got %s", n.Val.RatString())
				}
			},
		},
		{
			name: "addition preserves rational",
			expr: "1/3 + 1/3",
			verify: func(t *testing.T, v value.Value) {
				n := assertBigRat(t, v)
				if n != nil && n.Val.Cmp(big.NewRat(2, 3)) != 0 {
					t.Errorf("expected 2/3, got %s", n.Val.RatString())
				}
			},
		},
		{
			name: "division creates rational",
			expr: "1 / 3",
			verify: func(t *testing.T, v value.Value) {
				n := assertBigRat(t, v)
				if n != nil && n.Val.Cmp(big.NewRat(1, 3)) != 0 {
					t.Errorf("expected 1/3, got %s", n.Val.RatString())
				}
			},
		},
		{
			name: "multiplication preserves rational",
			expr: "(1/3) * 3",
			verify: func(t *testing.T, v value.Value) {
				n := assertBigRat(t, v)
				if n != nil && n.Val.Cmp(big.NewRat(1, 1)) != 0 {
					t.Errorf("expected 1, got %s", n.Val.RatString())
				}
			},
		},
		{
			name: "sqrt returns big.Rat (via SetFloat64)",
			expr: "_sqrt_inexact(4)",
			verify: func(t *testing.T, v value.Value) {
				assertBigRat(t, v)
			},
		},
		{
			name: "trig returns big.Rat (via SetFloat64)",
			expr: "_cos_inexact(0)",
			verify: func(t *testing.T, v value.Value) {
				assertBigRat(t, v)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := eval(t, tt.expr)
			tt.verify(t, v)
		})
	}
}

// assertBigRat verifies v is a *value.Number with *big.Rat inside.
func assertBigRat(t *testing.T, v value.Value) *value.Number {
	t.Helper()
	if v == nil {
		t.Error("value is nil")
		return nil
	}
	n, ok := v.(*value.Number)
	if !ok {
		t.Errorf("expected *value.Number, got %T", v)
		return nil
	}
	if n.Val == nil {
		t.Error("Number.Val is nil, expected *big.Rat")
		return nil
	}
	return n
}
