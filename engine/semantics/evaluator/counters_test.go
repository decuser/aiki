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

func TestCounterTailCalls(t *testing.T) {
	ev, g := makeEvalWithCounters(t)
	evalSource(t, ev, g, `
let countdown = (n) {
	if n <= 0 {
		return 0
	}
	countdown(n - 1)
}
countdown(5)
`)
	if ev.Counters.Call != 6 {
		t.Errorf("call: expected 6 including 5 proper tail calls, got %d", ev.Counters.Call)
	}
}

func TestNumberRealizationUsesEnvironmentProbe(t *testing.T) {
	g, err := grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
	if err != nil {
		t.Fatalf("loading grammar: %v", err)
	}
	ev := New(nil, engine.SilentObserver{})
	ev.SetGrammar(g)
	counters := NewCounters()
	source := "1 + 2 * 3"
	lexer := syntax.NewLexer(g, "test.ai", source, engine.SilentObserver{})
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lex error: %v", err)
	}
	parser := syntax.NewParser(g, tokens, source, engine.SilentObserver{})
	tree, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	env := value.NewEnvWithScope(value.ScopeUser)
	env.SetSemanticProbe(counters)
	env.SetFile("test.ai")
	env.SetSource(source)
	result := ev.Eval(tree, env)
	if got := result.Inspect(); got != "9" {
		t.Fatalf("result: got %s want 9", got)
	}
	n := counters.NumberSnapshot()
	if n.ResultSmallInteger != 2 {
		t.Fatalf("small integer results through environment probe: got %d want 2", n.ResultSmallInteger)
	}
}

func TestNumberRealizationCounters(t *testing.T) {
	c := NewCounters()
	a := value.NewNumber(2, 1)
	b := value.NewNumber(3, 1)
	c.NumberArithmeticResult(a, b, a.Add(b))

	f1, ok := value.NewNumberFromFloat64(0.5)
	if !ok {
		t.Fatal("0.5 carrier rejected")
	}
	f2, ok := value.NewNumberFromFloat64(0.25)
	if !ok {
		t.Fatal("0.25 carrier rejected")
	}
	c.NumberArithmeticResult(f1, f2, f1.Add(f2))

	n := c.NumberSnapshot()
	if n.ResultSmallInteger != 1 {
		t.Fatalf("small integer results: got %d want 1", n.ResultSmallInteger)
	}
	if n.ResultBinaryCarrier != 1 || n.BinaryCertified != 1 || n.BinaryFallback != 0 {
		t.Fatalf("binary realization: %+v", n)
	}
}

func TestNumberCallRealizationUsesEnvironmentProbe(t *testing.T) {
	g, err := grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
	if err != nil {
		t.Fatalf("loading grammar: %v", err)
	}
	ev := New(nil, engine.SilentObserver{})
	ev.SetGrammar(g)
	counters := NewCounters()
	source := "let f = () { 42 }\nf()\n"
	lexer := syntax.NewLexer(g, "test.ai", source, engine.SilentObserver{})
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lex error: %v", err)
	}
	parser := syntax.NewParser(g, tokens, source, engine.SilentObserver{})
	tree, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	env := value.NewEnvWithScope(value.ScopeUser)
	env.SetSemanticProbe(counters)
	env.SetFile("test.ai")
	env.SetSource(source)
	result := ev.Eval(tree, env)
	if got := result.Inspect(); got != "42" {
		t.Fatalf("result: got %s want 42", got)
	}
	n := counters.NumberCallSnapshot()
	if n.ResultSmallInteger != 1 {
		t.Fatalf("small integer call returns through environment probe: got %d want 1", n.ResultSmallInteger)
	}
}

func TestCallRealizationSnapshot(t *testing.T) {
	c := NewCounters()
	c.UserCallEntry()
	c.UserCallEntry()
	c.SubstrateCall()
	c.TailCallReuse()

	got := c.CallSnapshot()
	if got.UserEntry != 2 || got.Substrate != 1 || got.TailReuse != 1 {
		t.Fatalf("call realization = %+v, want user=2 substrate=1 tail=1", got)
	}
}
