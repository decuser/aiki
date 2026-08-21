package syntax

import (
	"strings"
	"testing"

	"aiki/engine"
	"aiki/engine/syntax/grammar"
)

func loadTestGrammar(t *testing.T) *grammar.Grammar {
	g, err := grammar.Load("grammar.ebnfx", EbnfxSource, "grammar.help", HelpSource)
	if err != nil {
		t.Fatalf("loading grammar: %v", err)
	}
	return g
}

func parseSource(t *testing.T, source string) *Node {
	g := loadTestGrammar(t)
	lexer := NewLexer(g, "test.ai", source, nil)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}
	parser := NewParser(g, tokens, source, nil)
	ast, err := parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}
	return ast
}

func TestParserLet(t *testing.T) {
	ast := parseSource(t, `let x = 42`)

	if ast.Type != "program" {
		t.Errorf("expected program, got %s", ast.Type)
	}
	if len(ast.Children) == 0 {
		t.Fatal("expected children")
	}

	// Find let_stmt
	found := findNode(ast, "let_stmt")
	if found == nil {
		t.Fatal("expected let_stmt node")
	}
}

func TestParserIf(t *testing.T) {
	ast := parseSource(t, `if true { 1 }`)

	found := findNode(ast, "if_stmt")
	if found == nil {
		t.Fatal("expected if_stmt node")
	}
}

func TestParserIfElse(t *testing.T) {
	ast := parseSource(t, `if true { 1 } else { 2 }`)

	found := findNode(ast, "if_stmt")
	if found == nil {
		t.Fatal("expected if_stmt node")
	}

	// Should have block for both branches
	blocks := findAllNodes(found, "block")
	if len(blocks) < 2 {
		t.Errorf("expected 2 blocks, got %d", len(blocks))
	}
}

func TestParserWhile(t *testing.T) {
	ast := parseSource(t, `while true { 1 }`)

	found := findNode(ast, "while_stmt")
	if found == nil {
		t.Fatal("expected while_stmt node")
	}
}

func TestParserFunction(t *testing.T) {
	ast := parseSource(t, `let f = (x) { x + 1 }`)

	found := findNode(ast, "func_literal")
	if found == nil {
		t.Fatal("expected func_literal node")
	}
}

func TestParserFunctionCall(t *testing.T) {
	ast := parseSource(t, `f(1, 2)`)

	found := findNode(ast, "call")
	if found == nil {
		t.Fatal("expected call node")
	}
}

func TestParserList(t *testing.T) {
	ast := parseSource(t, `[1, 2, 3]`)

	found := findNode(ast, "list_literal")
	if found == nil {
		t.Fatal("expected list_literal node")
	}
}

func TestParserShapedList(t *testing.T) {
	ast := parseSource(t, `[@point, 1, 2]`)

	found := findNode(ast, "list_literal")
	if found == nil {
		t.Fatal("expected list_literal node")
	}

	// Should have SHAPE child
	shape := findNode(found, "SHAPE")
	if shape == nil {
		t.Fatal("expected SHAPE node")
	}
}

func TestParserPipe(t *testing.T) {
	ast := parseSource(t, `1 |> f`)

	// The pipe operator should be present
	found := findNode(ast, "pipe_expr")
	if found == nil {
		t.Fatal("expected pipe_expr node")
	}
}

func TestParserMatch(t *testing.T) {
	ast := parseSource(t, `match x { 1 { :one } 2 { :two } }`)

	found := findNode(ast, "match_stmt")
	if found == nil {
		t.Fatal("expected match_stmt node")
	}
}

func TestParserPosition(t *testing.T) {
	source := "let x = 1\nlet y = 2"
	ast := parseSource(t, source)

	if ast.Pos.Line != 1 {
		t.Errorf("expected line 1, got %d", ast.Pos.Line)
	}

	// Find second let
	stmts := findAllNodes(ast, "let_stmt")
	if len(stmts) < 2 {
		t.Fatal("expected 2 let statements")
	}

	if stmts[1].Pos.Line != 2 {
		t.Errorf("expected second let at line 2, got %d", stmts[1].Pos.Line)
	}
}

func TestParserObserver(t *testing.T) {
	g := loadTestGrammar(t)
	source := `let x = 1`

	lexer := NewLexer(g, "test.ai", source, nil)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	var observed []string
	obs := &testParserObserver{
		onParse: func(production string, depth int, pos engine.Position) {
			observed = append(observed, production)
		},
	}

	parser := NewParser(g, tokens, source, obs)
	_, err = parser.Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}

	if len(observed) == 0 {
		t.Error("expected observer to be called")
	}

	// Should include program
	hasProgram := false
	for _, p := range observed {
		if p == "program" {
			hasProgram = true
			break
		}
	}
	if !hasProgram {
		t.Error("expected 'program' in observed productions")
	}
}

func TestParserError(t *testing.T) {
	g := loadTestGrammar(t)
	source := `let = 1` // missing name

	lexer := NewLexer(g, "test.ai", source, nil)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	parser := NewParser(g, tokens, source, nil)
	_, err = parser.Parse()
	if err == nil {
		t.Fatal("expected parse error")
	}

	// Should have caret in error
	if !strings.Contains(err.Error(), "^") {
		t.Errorf("expected caret in error, got: %v", err)
	}
}

func TestParserSuppressionIgnoresUnmatchedCloser(t *testing.T) {
	g := loadTestGrammar(t)
	tokens := []Token{
		{Type: "DELIMITER", Lexeme: "]", Pos: engine.Position{File: "test.ai", Line: 1, Col: 1}},
		{Type: "DELIMITER", Lexeme: "(", Pos: engine.Position{File: "test.ai", Line: 1, Col: 2}},
		{Type: "NUMBER", Lexeme: "1", Pos: engine.Position{File: "test.ai", Line: 1, Col: 3}},
		{Type: "NEWLINE", Lexeme: "\n", Pos: engine.Position{File: "test.ai", Line: 1, Col: 4}},
		{Type: "DELIMITER", Lexeme: ")", Pos: engine.Position{File: "test.ai", Line: 2, Col: 1}},
	}

	p := NewParser(g, tokens, "]\n(1\n)", nil)
	for _, tok := range p.tokens {
		if tok.Lexeme == ";" {
			t.Fatalf("unmatched closer corrupted later suppression; inserted terminator at %v", tok.Pos)
		}
	}
}

func TestParserReportsUnavailableNewlineAnalysisToDiagnosticObserver(t *testing.T) {
	g := loadTestGrammar(t)
	delete(g.Productions, "expr")
	g.Reanalyze()

	var kind, message string
	obs := &testParserObserver{
		onDiagnostic: func(gotKind, gotMessage string, pos engine.Position) {
			kind = gotKind
			message = gotMessage
		},
	}

	_ = NewParser(g, nil, "", obs)
	if kind != "grammar-newline-analysis" {
		t.Fatalf("diagnostic kind = %q, want grammar-newline-analysis", kind)
	}
	if !strings.Contains(message, "no expr production") {
		t.Fatalf("diagnostic message = %q, want missing expr reason", message)
	}
}

// Helper functions

func findNode(node *Node, nodeType string) *Node {
	if node.Type == nodeType {
		return node
	}
	for _, child := range node.Children {
		if found := findNode(child, nodeType); found != nil {
			return found
		}
	}
	return nil
}

func findAllNodes(node *Node, nodeType string) []*Node {
	var result []*Node
	if node.Type == nodeType {
		result = append(result, node)
	}
	for _, child := range node.Children {
		result = append(result, findAllNodes(child, nodeType)...)
	}
	return result
}

type testParserObserver struct {
	onParse      func(production string, depth int, pos engine.Position)
	onDiagnostic func(kind string, message string, pos engine.Position)
}

func (o *testParserObserver) OnLex(token, lexeme string, pos engine.Position) {}
func (o *testParserObserver) OnParse(production string, depth int, pos engine.Position) {
	if o.onParse != nil {
		o.onParse(production, depth, pos)
	}
}
func (o *testParserObserver) OnEval(node, result string, scope int, pos engine.Position) {}
func (o *testParserObserver) OnEffect(action, target string, pos engine.Position)        {}
func (o *testParserObserver) OnFormat(method, output, node string, depth int)            {}
func (o *testParserObserver) OnDiagnostic(kind, message string, pos engine.Position) {
	if o.onDiagnostic != nil {
		o.onDiagnostic(kind, message, pos)
	}
}

func TestParserSelect(t *testing.T) {
	ast := parseSource(t, `select {
		let msg = recv(commands) { println(msg) }
		recv(interrupts) { println(:interrupt) }
		default { println(:idle) }
	}`)

	found := findNode(ast, "select_stmt")
	if found == nil {
		t.Fatal("expected select_stmt node")
	}
	cases := findAllNodes(found, "select_case")
	if len(cases) != 2 {
		t.Fatalf("expected 2 select cases, got %d", len(cases))
	}
	if findNode(found, "select_default") == nil {
		t.Fatal("expected select_default node")
	}
}

func TestParserConsumesGrammarNewlineCompletionRule(t *testing.T) {
	g := loadTestGrammar(t)
	source := "let x = 1\n\t+ 2\n"
	lexer := NewLexer(g, "test.ai", source, nil)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("lexer error: %v", err)
	}

	// The declared rule terminates after NUMBER, so the leading binary
	// continuation is rejected under the baseline policy.
	if _, err := NewParser(g, tokens, source, nil).Parse(); err == nil {
		t.Fatal("expected baseline newline rule to reject leading '+'")
	}

	// Remove NUMBER from the grammar declaration only. If the parser consumes
	// the declaration rather than a private completion list, the newline is now
	// dropped and the same token stream parses as `let x = 1 + 2`.
	kept := g.Newline.AfterToken[:0]
	for _, name := range g.Newline.AfterToken {
		if name != "NUMBER" {
			kept = append(kept, name)
		}
	}
	g.Newline.AfterToken = kept

	if _, err := NewParser(g, tokens, source, nil).Parse(); err != nil {
		t.Fatalf("parser did not consume modified grammar newline rule: %v", err)
	}
}

func TestParserNewlineContinuationDiagnosticUsesGrammarHelp(t *testing.T) {
	g := loadTestGrammar(t)
	g.Newline.Meta.Help = "TEST NEWLINE POLICY"
	source := "let x = 1\n\t+ 2\n"

	lexer := NewLexer(g, "test.ai", source, nil)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatal(err)
	}
	parser := NewParser(g, tokens, source, nil)
	_, err = parser.Parse()
	if err == nil {
		t.Fatal("expected parse error")
	}

	got := err.Error()
	for _, want := range []string{
		"the previous newline ended the statement",
		"newline: TEST NEWLINE POLICY",
		"'+' continues an expression",
		"Place '+' before the newline it continues.",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("diagnostic missing %q:\n%s", want, got)
		}
	}
}

func TestRecordFailureStoresRelevantProductionWithoutStackMaterialization(t *testing.T) {
	p := &Parser{errorProduction: "expression"}
	pos := engine.Position{File: "test.ai", Line: 3, Col: 7}
	p.pos = 11
	p.recordFailure(pos, "+", "term")

	if !p.hasFailure {
		t.Fatal("recordFailure did not retain failure")
	}
	if got := p.failure.Production; got != "expression" {
		t.Fatalf("failure production = %q, want expression", got)
	}
	if p.failure.Stack != nil {
		t.Fatalf("parser materialized legacy failure stack: %#v", p.failure.Stack)
	}
	if p.failure.Pos != pos || p.failure.Got != "+" || p.failure.Expected != "term" {
		t.Fatalf("failure payload changed: %#v", p.failure)
	}
}

func TestRecordFailureSpeculativeReplacementAllocatesZero(t *testing.T) {
	p := &Parser{errorProduction: "expression"}
	pos := engine.Position{File: "test.ai", Line: 1, Col: 1}
	p.pos = 1

	allocs := testing.AllocsPerRun(1000, func() {
		p.recordFailure(pos, "x", "term")
	})
	if allocs != 0 {
		t.Fatalf("recordFailure allocations = %v, want 0", allocs)
	}
}

func TestTerminalMismatchUsesZeroAllocationControlFlow(t *testing.T) {
	p := &Parser{
		tokens: []Token{{
			Type:   "IDENT",
			Lexeme: "actual",
			Pos:    engine.Position{File: "test.ai", Line: 1, Col: 1},
		}},
		observer: engine.SilentObserver{},
	}
	term := &grammar.Terminal{Value: "expected"}

	allocs := testing.AllocsPerRun(1000, func() {
		p.pos = 0
		p.furthest = 0
		p.hasFailure = false
		if _, err := p.parseTerminal(term); err != errParserNoMatch {
			t.Fatalf("terminal mismatch error = %v, want internal no-match", err)
		}
	})
	if allocs != 0 {
		t.Fatalf("terminal mismatch allocations = %v, want 0", allocs)
	}

	if got := p.failure.expectedText(); got != "'expected'" {
		t.Fatalf("recorded terminal expectation = %q, want %q", got, "'expected'")
	}
	if p.failure.Expected != "" {
		t.Fatalf("speculative terminal mismatch materialized Expected = %q", p.failure.Expected)
	}
}

func TestTerminalMismatchRenderingPreservesQuotedExpectation(t *testing.T) {
	g := loadTestGrammar(t)
	source := "actual"
	parser := &Parser{
		grammar: g,
		tokens: []Token{{
			Type:   "IDENT",
			Lexeme: "actual",
			Pos:    engine.Position{File: "test.ai", Line: 1, Col: 1},
		}},
		source:   source,
		observer: engine.SilentObserver{},
	}

	parser.recordTerminalFailure(
		engine.Position{File: "test.ai", Line: 1, Col: 1},
		"actual",
		"=",
	)

	err := parser.renderFailure()
	if !strings.Contains(err.Error(), "expected '='") {
		t.Fatalf("terminal diagnostic lost quoted expectation:\n%s", err)
	}
}
