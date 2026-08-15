package invariant

import (
	"fmt"
	"testing"

	"aiki/engine/semantics/evaluator"
	"aiki/engine/syntax"
	"aiki/engine/syntax/grammar"
)

// TestHandlerValidationPanicsOnMissingHandler verifies that missing production handlers panic at startup.
func TestHandlerValidationPanicsOnMissingHandler(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Error("expected panic for missing handler")
		}
		msg := fmt.Sprint(r)
		if msg != "grammar AST node has no evaluator handler: fake_stmt" {
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

	ev := evaluator.New(nil, nil)
	ev.SetGrammar(g) // should panic
}

// TestHandlerValidationPanicsOnMissingTokenHandler verifies that a TokenRef
// reachable from a production requires an evaluator handler. Merely defining a
// lexer token is not enough to make it an AST node.
func TestHandlerValidationPanicsOnMissingTokenHandler(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for missing token handler")
		}
		msg := fmt.Sprint(r)
		if msg != "grammar AST node has no evaluator handler: FAKE_TOKEN" {
			t.Errorf("unexpected panic message: %v", r)
		}
	}()

	g, err := grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
	if err != nil {
		t.Fatalf("load grammar: %v", err)
	}
	original := g.Productions["program"].Expr
	g.Productions["program"].Expr = &grammar.Sequence{Exprs: []grammar.Expression{
		original,
		&grammar.TokenRef{Name: "FAKE_TOKEN"},
	}}

	ev := evaluator.New(nil, nil)
	ev.SetGrammar(g) // should panic
}

// TestHandlerValidationPassesForRealGrammar verifies the real grammar has all handlers.
func TestHandlerValidationPassesForRealGrammar(t *testing.T) {
	// Should not panic
	g, err := grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
	if err != nil {
		t.Fatalf("load grammar: %v", err)
	}

	ev := evaluator.New(nil, nil)
	ev.SetGrammar(g) // should not panic
}
