package invariant

import (
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

	ev := evaluator.New(nil, nil)
	ev.SetGrammar(g) // should panic
}

// TestHandlerValidationPanicsOnMissingTokenHandler verifies that missing token handlers panic at startup.
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
