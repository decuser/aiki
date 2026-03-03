package evaluator

import (
	"testing"

	"aiki/engine/runtime/hal/substrate"
	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
	"aiki/engine/syntax/grammar"
)

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
	ev := New(rt, nil)
	// Use prelude scope to access HAL primitives directly
	return ev.Eval(ast, value.NewEnvWithScope(value.ScopePrelude))
}

func TestEvalNumber(t *testing.T) {
	v := eval(t, "42")
	if v.Inspect() != "42" {
		t.Errorf("got %s", v.Inspect())
	}
}

func TestEvalArithmetic(t *testing.T) {
	tests := []struct{ in, out string }{
		{"1 + 2", "3"},
		{"5 - 3", "2"},
		{"2 * 3", "6"},
		{"8 / 4", "2"},
		{"1 / 3", "1/3"},
	}
	for _, tt := range tests {
		v := eval(t, tt.in)
		if v.Inspect() != tt.out {
			t.Errorf("%s: got %s, want %s", tt.in, v.Inspect(), tt.out)
		}
	}
}

func TestEvalLet(t *testing.T) {
	v := eval(t, "let x = 5\nx * 2")
	if v.Inspect() != "10" {
		t.Errorf("got %s", v.Inspect())
	}
}

func TestEvalIf(t *testing.T) {
	v := eval(t, "if true { 1 } else { 2 }")
	if v.Inspect() != "1" {
		t.Errorf("got %s", v.Inspect())
	}
	v = eval(t, "if false { 1 } else { 2 }")
	if v.Inspect() != "2" {
		t.Errorf("got %s", v.Inspect())
	}
}

func TestEvalWhile(t *testing.T) {
	v := eval(t, "let x = 0\nwhile x < 3 { x = x + 1 }\nx")
	if v.Inspect() != "3" {
		t.Errorf("got %s", v.Inspect())
	}
}

func TestEvalFunction(t *testing.T) {
	v := eval(t, "let f = (x) { x + 1 }\nf(5)")
	if v.Inspect() != "6" {
		t.Errorf("got %s", v.Inspect())
	}
}

func TestEvalList(t *testing.T) {
	v := eval(t, "[1, 2, 3]")
	if v.Inspect() != "[1, 2, 3]" {
		t.Errorf("got %s", v.Inspect())
	}
}

func TestEvalBuiltin(t *testing.T) {
	v := eval(t, "_length([1, 2, 3])")
	if v.Inspect() != "3" {
		t.Errorf("got %s", v.Inspect())
	}
}

func TestEvalPipe(t *testing.T) {
	v := eval(t, "[1, 2, 3] |> _first")
	if v.Inspect() != "1" {
		t.Errorf("got %s", v.Inspect())
	}
}

func TestHandlerValidationPanicsOnMissingHandler(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Error("expected panic for missing handler")
		}
		msg, ok := r.(string)
		if !ok || msg != "grammar production has no handler: fake_stmt" {
			t.Errorf("unexpected panic message: %v", r)
		}
	}()

	g := &grammar.Grammar{
		Productions: map[string]*grammar.Production{
			"program":   {},
			"fake_stmt": {}, // no handler for this
		},
		Tokens: []grammar.TokenDef{},
	}

	ev := New(nil, nil)
	ev.SetGrammar(g) // should panic
}

func TestHandlerValidationPanicsOnMissingTokenHandler(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Error("expected panic for missing token handler")
		}
	}()

	g := &grammar.Grammar{
		Productions: map[string]*grammar.Production{
			"program": {},
		},
		Tokens: []grammar.TokenDef{
			{Name: "FAKE_TOKEN", Skip: false}, // no handler for this
		},
	}

	ev := New(nil, nil)
	ev.SetGrammar(g) // should panic
}

func TestHandlerValidationPassesForRealGrammar(t *testing.T) {
	// Should not panic
	g, err := grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
	if err != nil {
		t.Fatalf("load grammar: %v", err)
	}

	ev := New(nil, nil)
	ev.SetGrammar(g) // should not panic
}
