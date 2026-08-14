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
	onParse func(production string, depth int, pos engine.Position)
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
