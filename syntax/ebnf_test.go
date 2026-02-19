package syntax

import (
	"fmt"
	"strings"
	"testing"
)

func TestParseSimpleGrammar(t *testing.T) {
	source := `
@tokens {
    NUMBER  /[0-9]+/
    NAME    /[a-z]+/
    WS      /[ \t\n]+/  @skip
}

expr = NUMBER | NAME
`
	g, err := Parse(source)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if len(g.Tokens) != 3 {
		t.Errorf("expected 3 tokens, got %d", len(g.Tokens))
	}

	if g.Tokens[0].Name != "NUMBER" {
		t.Errorf("expected NUMBER, got %s", g.Tokens[0].Name)
	}

	if !g.Tokens[2].Skip {
		t.Errorf("expected WS to be skip")
	}

	if g.Start != "expr" {
		t.Errorf("expected start=expr, got %s", g.Start)
	}

	if _, ok := g.Productions["expr"]; !ok {
		t.Errorf("expected expr production")
	}
}

func TestTokenize(t *testing.T) {
	source := `
@tokens {
    NUMBER  /[0-9]+/
    PLUS    /\+/
    WS      /[ \t\n]+/  @skip
}

expr = NUMBER
`
	g, err := Parse(source)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	tokens, err := g.Tokenize("1 + 2")
	if err != nil {
		t.Fatalf("tokenize error: %v", err)
	}

	if len(tokens) != 3 {
		t.Errorf("expected 3 tokens, got %d", len(tokens))
		for _, tok := range tokens {
			t.Logf("  %s: %s", tok.Type, tok.Lexeme)
		}
	}

	if tokens[0].Type != "NUMBER" || tokens[0].Lexeme != "1" {
		t.Errorf("expected NUMBER:1, got %s:%s", tokens[0].Type, tokens[0].Lexeme)
	}
}

func TestParseSource(t *testing.T) {
	source := `
@tokens {
    NUMBER  /[0-9]+/
    WS      /[ \t\n]+/  @skip
}

expr = NUMBER
`
	g, err := Parse(source)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	ast, err := g.ParseSource("42")
	if err != nil {
		t.Fatalf("parse source error: %v", err)
	}

	if ast.Type != "expr" {
		t.Errorf("expected expr, got %s", ast.Type)
	}

	if len(ast.Children) != 1 {
		t.Errorf("expected 1 child, got %d", len(ast.Children))
	}

	if ast.Children[0].Value != "42" {
		t.Errorf("expected 42, got %s", ast.Children[0].Value)
	}
}

func TestParseAikiGrammar(t *testing.T) {
	g := LoadGrammarForTest(t, "grammar.ebnf")


	// Check tokens
	if len(g.Tokens) == 0 {
		t.Errorf("expected tokens")
	}

	// Check some productions exist
	prods := []string{"program", "statement", "let_stmt", "expr", "block"}
	for _, name := range prods {
		if _, ok := g.Productions[name]; !ok {
			t.Errorf("expected production: %s", name)
		}
	}

	// Check start
	if g.Start != "program" {
		t.Errorf("expected start=program, got %s", g.Start)
	}
}

func TestTokenizeAiki(t *testing.T) {
	g := LoadGrammarForTest(t, "grammar.ebnf")

	tokens, err := g.Tokenize("let x = 42")
	if err != nil {
		t.Fatalf("tokenize error: %v", err)
	}

	expected := []struct {
		typ    string
		lexeme string
	}{
		{"KEYWORD", "let"},
		{"NAME", "x"},
		{"OPERATOR", "="},
		{"NUMBER", "42"},
	}

	if len(tokens) != len(expected) {
		t.Errorf("expected %d tokens, got %d", len(expected), len(tokens))
		for _, tok := range tokens {
			t.Logf("  %s: %q", tok.Type, tok.Lexeme)
		}
		return
	}

	for i, exp := range expected {
		if tokens[i].Type != exp.typ || tokens[i].Lexeme != exp.lexeme {
			t.Errorf("token %d: expected %s:%q, got %s:%q",
				i, exp.typ, exp.lexeme, tokens[i].Type, tokens[i].Lexeme)
		}
	}
}

func TestParseSimpleAiki(t *testing.T) {
	g := LoadGrammarForTest(t, "grammar.ebnf")

	ast, err := g.ParseSource("42")
	if err != nil {
		t.Fatalf("parse source error: %v", err)
	}

	if ast.Type != "program" {
		t.Errorf("expected program, got %s", ast.Type)
	}
}

func TestParseLetStatement(t *testing.T) {
	g := LoadGrammarForTest(t, "grammar.ebnf")

	ast, err := g.ParseSource("let x = 42")
	if err != nil {
		t.Fatalf("parse source error: %v", err)
	}

	if ast.Type != "program" {
		t.Errorf("expected program, got %s", ast.Type)
	}
}

func TestParseFunctionLiteral(t *testing.T) {
	g := LoadGrammarForTest(t, "grammar.ebnf")

	ast, err := g.ParseSource("(a) { return a }")
	if err != nil {
		t.Fatalf("parse source error: %v", err)
	}

	if ast.Type != "program" {
		t.Errorf("expected program, got %s", ast.Type)
	}
}

func TestParseFunctionWithParams(t *testing.T) {
	g := LoadGrammarForTest(t, "grammar.ebnf")

	ast, err := g.ParseSource("let add = (a, b) { return a }")
	if err != nil {
		t.Fatalf("parse source error: %v", err)
	}

	if ast.Type != "program" {
		t.Errorf("expected program, got %s", ast.Type)
	}
}

func TestParseInfixExpression(t *testing.T) {
	g := LoadGrammarForTest(t, "grammar.ebnf")

	ast, err := g.ParseSource("a + b")
	if err != nil {
		t.Fatalf("parse source error: %v", err)
	}

	if ast.Type != "program" {
		t.Errorf("expected program, got %s", ast.Type)
	}
}

func TestParseFunctionWithInfix(t *testing.T) {
	g := LoadGrammarForTest(t, "grammar.ebnf")

	ast, err := g.ParseSource("let add = (a, b) { return a + b }")
	if err != nil {
		t.Fatalf("parse source error: %v", err)
	}

	if ast.Type != "program" {
		t.Errorf("expected program, got %s", ast.Type)
	}
}

func TestParseIfStatement(t *testing.T) {
	g := LoadGrammarForTest(t, "grammar.ebnf")

	ast, err := g.ParseSource("if x { return 1 }")
	if err != nil {
		t.Fatalf("parse source error: %v", err)
	}

	if ast.Type != "program" {
		t.Errorf("expected program, got %s", ast.Type)
	}
}

func TestParseIfElse(t *testing.T) {
	g := LoadGrammarForTest(t, "grammar.ebnf")

	ast, err := g.ParseSource("if x { return 1 } else { return 2 }")
	if err != nil {
		t.Fatalf("parse source error: %v", err)
	}

	if ast.Type != "program" {
		t.Errorf("expected program, got %s", ast.Type)
	}
}

func TestParseWhile(t *testing.T) {
	g := LoadGrammarForTest(t, "grammar.ebnf")

	ast, err := g.ParseSource("while x { x = x - 1 }")
	if err != nil {
		t.Fatalf("parse source error: %v", err)
	}

	if ast.Type != "program" {
		t.Errorf("expected program, got %s", ast.Type)
	}
}

func TestParsePipe(t *testing.T) {
	g := LoadGrammarForTest(t, "grammar.ebnf")

	ast, err := g.ParseSource("x |> f()")
	if err != nil {
		t.Fatalf("parse source error: %v", err)
	}

	if ast.Type != "program" {
		t.Errorf("expected program, got %s", ast.Type)
	}
}

func TestParseList(t *testing.T) {
	g := LoadGrammarForTest(t, "grammar.ebnf")

	ast, err := g.ParseSource("[1, 2, 3]")
	if err != nil {
		t.Fatalf("parse source error: %v", err)
	}

	if ast.Type != "program" {
		t.Errorf("expected program, got %s", ast.Type)
	}
}

func TestParseShapedList(t *testing.T) {
	g := LoadGrammarForTest(t, "grammar.ebnf")

	ast, err := g.ParseSource("[@point, 10, 20]")
	if err != nil {
		t.Fatalf("parse source error: %v", err)
	}

	if ast.Type != "program" {
		t.Errorf("expected program, got %s", ast.Type)
	}
}

func TestParseCall(t *testing.T) {
	g := LoadGrammarForTest(t, "grammar.ebnf")

	ast, err := g.ParseSource("f(1, 2)")
	if err != nil {
		t.Fatalf("parse source error: %v", err)
	}

	if ast.Type != "program" {
		t.Errorf("expected program, got %s", ast.Type)
	}
}

func TestParseIndex(t *testing.T) {
	g := LoadGrammarForTest(t, "grammar.ebnf")

	ast, err := g.ParseSource("list[0]")
	if err != nil {
		t.Fatalf("parse source error: %v", err)
	}

	if ast.Type != "program" {
		t.Errorf("expected program, got %s", ast.Type)
	}
}

func TestParseAccess(t *testing.T) {
	g := LoadGrammarForTest(t, "grammar.ebnf")

	ast, err := g.ParseSource("point.x")
	if err != nil {
		t.Fatalf("parse source error: %v", err)
	}

	if ast.Type != "program" {
		t.Errorf("expected program, got %s", ast.Type)
	}
}

func TestParseMultipleStatements(t *testing.T) {
	g := LoadGrammarForTest(t, "grammar.ebnf")

	ast, err := g.ParseSource("let x = 1\nlet y = 2\nx + y")
	if err != nil {
		t.Fatalf("parse source error: %v", err)
	}

	if ast.Type != "program" {
		t.Errorf("expected program, got %s", ast.Type)
	}

	// Should have 3 statements
	stmts := ast.ChildrenByType("statement")
	if len(stmts) != 3 {
		t.Errorf("expected 3 statements, got %d", len(stmts))
	}
}

func printAST(n *Node, indent int) {
	prefix := strings.Repeat("  ", indent)
	if n.IsTerminal() {
		fmt.Printf("%s%s: %q\n", prefix, n.Type, n.Value)
	} else {
		fmt.Printf("%s%s\n", prefix, n.Type)
		for _, c := range n.Children {
			printAST(c, indent+1)
		}
	}
}
