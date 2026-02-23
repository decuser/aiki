package grammar

import (
	"strings"
	"testing"
)

func TestParseTokensBlock(t *testing.T) {
	source := `@tokens {
    NUMBER      /[0-9]+/
        @error "expected number"
        @help "integer literal"
    STRING      /"[^"]*"/  @skip
    NAME        /[a-z_]+/
}`

	p := NewParser("test.ebnfx", source)
	g, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if len(g.Tokens) != 3 {
		t.Fatalf("expected 3 tokens, got %d", len(g.Tokens))
	}

	// Check NUMBER
	num := g.Tokens[0]
	if num.Name != "NUMBER" {
		t.Errorf("expected NUMBER, got %s", num.Name)
	}
	if num.Meta.Error != "expected number" {
		t.Errorf("expected error 'expected number', got '%s'", num.Meta.Error)
	}
	if num.Meta.Help != "integer literal" {
		t.Errorf("expected help 'integer literal', got '%s'", num.Meta.Help)
	}

	// Check STRING has skip
	str := g.Tokens[1]
	if !str.Skip {
		t.Error("expected STRING to have skip")
	}
}

func TestParseProduction(t *testing.T) {
	source := `program = { statement }
    @error "expected statement"
    @template "statement ..."
    @help "sequence of statements"`

	p := NewParser("test.ebnfx", source)
	g, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	prod, ok := g.Productions["program"]
	if !ok {
		t.Fatal("expected program production")
	}

	if prod.Meta.Error != "expected statement" {
		t.Errorf("expected error decorator, got '%s'", prod.Meta.Error)
	}
	if prod.Meta.Template != "statement ..." {
		t.Errorf("expected template decorator, got '%s'", prod.Meta.Template)
	}
	if prod.Meta.Help != "sequence of statements" {
		t.Errorf("expected help decorator, got '%s'", prod.Meta.Help)
	}
}

func TestParseAlternatives(t *testing.T) {
	source := `statement = let_stmt | assign_stmt | expr_stmt`

	p := NewParser("test.ebnfx", source)
	g, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	prod := g.Productions["statement"]
	if len(prod.Expressions) != 3 {
		t.Fatalf("expected 3 alternatives, got %d", len(prod.Expressions))
	}

	names := []string{"let_stmt", "assign_stmt", "expr_stmt"}
	for i, expr := range prod.Expressions {
		if len(expr.Terms) != 1 {
			t.Errorf("expression %d: expected 1 term, got %d", i, len(expr.Terms))
			continue
		}
		if expr.Terms[0].Value != names[i] {
			t.Errorf("expected %s, got %s", names[i], expr.Terms[0].Value)
		}
	}
}

func TestParseLiterals(t *testing.T) {
	source := `let_stmt = "let" NAME "=" expr`

	p := NewParser("test.ebnfx", source)
	g, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	prod := g.Productions["let_stmt"]
	if len(prod.Expressions) != 1 {
		t.Fatalf("expected 1 expression, got %d", len(prod.Expressions))
	}

	terms := prod.Expressions[0].Terms
	if len(terms) != 4 {
		t.Fatalf("expected 4 terms, got %d", len(terms))
	}

	// "let"
	if terms[0].Kind != TermLiteral || terms[0].Value != "let" {
		t.Errorf("expected literal 'let', got %v", terms[0])
	}
	// NAME
	if terms[1].Kind != TermToken || terms[1].Value != "NAME" {
		t.Errorf("expected token NAME, got %v", terms[1])
	}
	// "="
	if terms[2].Kind != TermLiteral || terms[2].Value != "=" {
		t.Errorf("expected literal '=', got %v", terms[2])
	}
	// expr
	if terms[3].Kind != TermProduction || terms[3].Value != "expr" {
		t.Errorf("expected production expr, got %v", terms[3])
	}
}

func TestParseOptionalAndRepeat(t *testing.T) {
	source := `program = { statement }
if_stmt = "if" expr block [ "else" block ]`

	p := NewParser("test.ebnfx", source)
	g, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	// Check repeat in program
	prog := g.Productions["program"]
	if len(prog.Expressions[0].Terms) != 1 {
		t.Fatalf("expected 1 term in program")
	}
	if !prog.Expressions[0].Terms[0].Repeat {
		t.Error("expected statement to be repeat")
	}

	// Check optional in if_stmt
	ifStmt := g.Productions["if_stmt"]
	terms := ifStmt.Expressions[0].Terms
	// Last term should be optional
	found := false
	for _, term := range terms {
		if term.Optional {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected optional term in if_stmt")
	}
}

func TestParsePipeOperator(t *testing.T) {
	// |> should not be confused with |
	source := `pipe_expr = infix_expr { "|>" postfix_expr }`

	p := NewParser("test.ebnfx", source)
	g, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	prod := g.Productions["pipe_expr"]
	if len(prod.Expressions) != 1 {
		t.Fatalf("expected 1 expression (not split by |>), got %d", len(prod.Expressions))
	}
}

func TestParseHelp(t *testing.T) {
	source := `program

A program is a sequence of statements.

let x = 1
print(x)
===
let_stmt

Binds a value to a name.

let x = 42
===`

	entries, err := ParseHelp("test.help", source)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	prog := entries["program"]
	if !strings.Contains(prog.Doc, "sequence of statements") {
		t.Errorf("expected doc content, got '%s'", prog.Doc)
	}
	if !strings.Contains(prog.Doc, "let x = 1") {
		t.Errorf("expected example in doc, got '%s'", prog.Doc)
	}
}

func TestLoadValidates(t *testing.T) {
	ebnfx := `@tokens {
    NUMBER /[0-9]+/
}
program = { statement }`

	// Missing help for NUMBER and program
	help := `statement

A statement.
===`

	_, err := Load("test.ebnfx", ebnfx, "test.help", help)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "missing help") {
		t.Errorf("expected 'missing help' error, got: %v", err)
	}
}

func TestLoadMergesDoc(t *testing.T) {
	ebnfx := `@tokens {
    NUMBER /[0-9]+/
        @help "short help"
}
program = { statement }
    @help "short help"
statement = expr`

	help := `NUMBER

Full documentation for NUMBER with examples.

42
123
===
program

Full documentation for program.

let x = 1
===
statement

Full doc for statement.
===`

	g, err := Load("test.ebnfx", ebnfx, "test.help", help)
	if err != nil {
		t.Fatalf("load error: %v", err)
	}

	// Check token got doc merged
	tok, _ := g.GetToken("NUMBER")
	if tok.Meta.Help != "short help" {
		t.Errorf("expected short help preserved, got '%s'", tok.Meta.Help)
	}
	if !strings.Contains(tok.Meta.Doc, "Full documentation for NUMBER") {
		t.Errorf("expected doc merged, got '%s'", tok.Meta.Doc)
	}

	// Check production got doc merged
	prod, _ := g.GetProduction("program")
	if !strings.Contains(prod.Meta.Doc, "Full documentation for program") {
		t.Errorf("expected doc merged, got '%s'", prod.Meta.Doc)
	}
}

func TestPositionTracking(t *testing.T) {
	source := `@tokens {
    NUMBER /[0-9]+/
}

program = { statement }`

	p := NewParser("test.ebnfx", source)
	g, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	// NUMBER should be on line 2
	tok, _ := g.GetToken("NUMBER")
	if tok.Pos.Line != 2 {
		t.Errorf("expected NUMBER on line 2, got %d", tok.Pos.Line)
	}

	// program should be on line 5
	prod, _ := g.GetProduction("program")
	if prod.Pos.Line != 5 {
		t.Errorf("expected program on line 5, got %d", prod.Pos.Line)
	}
}
