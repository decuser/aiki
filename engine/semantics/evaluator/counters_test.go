package evaluator

import (
	"testing"

	"aiki/engine"
	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
	"aiki/engine/syntax/grammar"
)

func makeEvalWithCounters(t *testing.T) (*Evaluator, *grammar.Grammar) {
	t.Helper()
	g, err := grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
	if err != nil {
		t.Fatalf("loading grammar: %v", err)
	}
	ev := New(nil, engine.SilentObserver{})
	ev.SetGrammar(g)
	ev.Counters = NewCounters()
	return ev, g
}

func evalSource(t *testing.T, ev *Evaluator, g *grammar.Grammar, source string) value.Value {
	t.Helper()
	lexer := syntax.NewLexer(g, "test.ai", source, engine.SilentObserver{})
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lex error: %v", err)
	}
	parser := syntax.NewParser(g, tokens, "test.ai", engine.SilentObserver{})
	tree, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	env := value.NewEnvWithScope(value.ScopeUser)
	env.SetFile("test.ai")
	env.SetSource(source)
	return ev.Eval(tree, env)
}

func TestCounterArithmetic(t *testing.T) {
	ev, g := makeEvalWithCounters(t)
	evalSource(t, ev, g, "1 + 2 + 3")
	// Two additions
	if ev.Counters.Arithmetic != 2 {
		t.Errorf("arithmetic: expected 2, got %d", ev.Counters.Arithmetic)
	}
}

func TestCounterComparison(t *testing.T) {
	ev, g := makeEvalWithCounters(t)
	evalSource(t, ev, g, "1 < 2")
	if ev.Counters.Comparison != 1 {
		t.Errorf("comparison: expected 1, got %d", ev.Counters.Comparison)
	}
}

func TestCounterCall(t *testing.T) {
	ev, g := makeEvalWithCounters(t)
	evalSource(t, ev, g, `
let f = () { 1 }
f()
f()
f()
`)
	if ev.Counters.Call != 3 {
		t.Errorf("call: expected 3, got %d", ev.Counters.Call)
	}
}

func TestCounterIteration(t *testing.T) {
	ev, g := makeEvalWithCounters(t)
	evalSource(t, ev, g, `
let x = 0
while x < 10 {
    x = x + 1
}
`)
	if ev.Counters.Iteration != 10 {
		t.Errorf("iteration: expected 10, got %d", ev.Counters.Iteration)
	}
	// Also: 10 comparisons in the condition, 10 additions in the body
	if ev.Counters.Comparison != 11 {
		// 10 true + 1 false exit check
		t.Errorf("comparison: expected 11, got %d", ev.Counters.Comparison)
	}
	if ev.Counters.Arithmetic != 10 {
		t.Errorf("arithmetic: expected 10, got %d", ev.Counters.Arithmetic)
	}
}

func TestCounterIndex(t *testing.T) {
	ev, g := makeEvalWithCounters(t)
	evalSource(t, ev, g, `
let xs = [10, 20, 30]
xs[0]
xs[1]
xs[2]
`)
	if ev.Counters.Index != 3 {
		t.Errorf("index: expected 3, got %d", ev.Counters.Index)
	}
}

func TestCounterNegation(t *testing.T) {
	ev, g := makeEvalWithCounters(t)
	evalSource(t, ev, g, "-5")
	if ev.Counters.Arithmetic != 1 {
		t.Errorf("arithmetic (negation): expected 1, got %d", ev.Counters.Arithmetic)
	}
}

func TestCountersDisabledByDefault(t *testing.T) {
	g, err := grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
	if err != nil {
		t.Fatalf("loading grammar: %v", err)
	}
	ev := New(nil, engine.SilentObserver{})
	ev.SetGrammar(g)
	// Counters is nil — no probes fire, no panic
	evalSource(t, ev, g, "1 + 2 * 3")
	if ev.Counters != nil {
		t.Error("counters should be nil by default")
	}
}

func TestCoverageHits(t *testing.T) {
	g, err := grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
	if err != nil {
		t.Fatalf("loading grammar: %v", err)
	}
	ev := New(nil, engine.SilentObserver{})
	ev.SetGrammar(g)
	ev.Counters = NewCoverageCounters()

	evalSource(t, ev, g, "1 + 2\n3 + 4\n")

	cov := ev.Counters.Coverage()
	if cov == nil {
		t.Fatal("coverage map should not be nil")
	}
	if cov["test.ai:1"] == 0 {
		t.Error("line 1 should have coverage hits")
	}
	if cov["test.ai:2"] == 0 {
		t.Error("line 2 should have coverage hits")
	}
}
