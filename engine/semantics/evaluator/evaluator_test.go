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
	env := value.NewEnvWithScope(value.ScopePrelude)
	env.SetAuthority(rt.AuthorityForSource("engine/runtime/prelude/prelude.ai"))
	return ev.Eval(ast, env)
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

func TestHandlerCoverageMatchesGrammarAST(t *testing.T) {
	g, err := grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
	if err != nil {
		t.Fatalf("load grammar: %v", err)
	}
	if err := validateHandlerCoverage(g, handlers); err != nil {
		t.Fatal(err)
	}

	refs := g.Analysis().TokenRefs()
	for _, name := range []string{"NAME", "NUMBER", "STRING", "RUNE", "SYMBOL", "SHAPE"} {
		if _, ok := refs[name]; !ok {
			t.Errorf("expected production TokenRef %s", name)
		}
	}
	if _, ok := g.Productions["BINOP"]; !ok {
		t.Error("expected BINOP to be a named production")
	}
	if _, ok := refs["BINOP"]; ok {
		t.Error("BINOP is a production, not a TokenRef")
	}
	for _, name := range []string{"KEYWORD", "OPERATOR", "DELIMITER", "NEWLINE"} {
		if _, ok := refs[name]; ok {
			t.Errorf("lexical-only token unexpectedly production-referenced: %s", name)
		}
		if _, ok := handlers[name]; ok {
			t.Errorf("lexical-only token unexpectedly has evaluator handler: %s", name)
		}
	}
}

func TestHandlerCoverageRejectsBothDirections(t *testing.T) {
	g, err := grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
	if err != nil {
		t.Fatalf("load grammar: %v", err)
	}

	missing := make(map[string]handlerFunc, len(handlers))
	for name, h := range handlers {
		missing[name] = h
	}
	delete(missing, "NAME")
	if err := validateHandlerCoverage(g, missing); err == nil {
		t.Fatal("expected missing grammar node handler to fail")
	}

	extra := make(map[string]handlerFunc, len(handlers)+1)
	for name, h := range handlers {
		extra[name] = h
	}
	extra["KEYWORD"] = (*Evaluator).evalTerminal
	if err := validateHandlerCoverage(g, extra); err == nil {
		t.Fatal("expected handler without grammar AST node to fail")
	}
}

func TestBinaryOperatorCoverageMatchesGrammar(t *testing.T) {
	g, err := grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
	if err != nil {
		t.Fatalf("load grammar: %v", err)
	}

	ops := g.Analysis().TerminalAlternatives("BINOP")
	if err := validateBinaryOperatorCoverage(ops, binaryOperatorSemantics); err != nil {
		t.Fatal(err)
	}
	if len(ops) != 10 {
		t.Fatalf("BINOP alternatives = %d, want 10", len(ops))
	}
	for _, op := range []string{"+", "-", "*", "/", "<", ">", "<=", ">=", "and", "or"} {
		if _, ok := ops[op]; !ok {
			t.Errorf("grammar BINOP missing %q", op)
		}
	}
}

func TestBinaryOperatorCoverageRejectsBothDirections(t *testing.T) {
	g, err := grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
	if err != nil {
		t.Fatalf("load grammar: %v", err)
	}
	ops := g.Analysis().TerminalAlternatives("BINOP")

	missing := make(map[string]binaryOperatorKind, len(binaryOperatorSemantics)-1)
	for op, kind := range binaryOperatorSemantics {
		if op != "+" {
			missing[op] = kind
		}
	}
	if err := validateBinaryOperatorCoverage(ops, missing); err == nil {
		t.Fatal("expected grammar operator without evaluator semantics to fail")
	}

	extra := make(map[string]binaryOperatorKind, len(binaryOperatorSemantics)+1)
	for op, kind := range binaryOperatorSemantics {
		extra[op] = kind
	}
	extra["fake-op"] = operatorAdd
	if err := validateBinaryOperatorCoverage(ops, extra); err == nil {
		t.Fatal("expected evaluator operator without grammar BINOP to fail")
	}
}

func TestBinaryOperatorMembershipComesFromGrammar(t *testing.T) {
	g, err := grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
	if err != nil {
		t.Fatalf("load grammar: %v", err)
	}
	ev := New(substrate.NewGoRuntime(), nil)
	ev.SetGrammar(g)
	if !ev.isBinaryOperator("+") {
		t.Fatal("expected grammar-declared + to be an operator")
	}
	if ev.isBinaryOperator("fake-op") {
		t.Fatal("unexpected operator not declared by grammar")
	}
}
