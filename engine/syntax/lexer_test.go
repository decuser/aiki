package syntax

import (
	"regexp"
	"strings"
	"testing"

	"aiki/engine"
	"aiki/engine/syntax/grammar"
)

// testGrammar creates a minimal grammar for testing.
func testGrammar() *grammar.Grammar {
	return &grammar.Grammar{
		Tokens: []grammar.TokenDef{
			{Name: "NUMBER", Pattern: mustCompile(`[0-9]+(\.[0-9]+)?(\/[0-9]+)?`)},
			{Name: "STRING", Pattern: mustCompile(`"[^"]*"`)},
			{Name: "RUNE", Pattern: mustCompile(`'([^'\\]|\\.)'`)},
			{Name: "SYMBOL", Pattern: mustCompile(`:[a-zA-Z_][a-zA-Z0-9_]*`)},
			{Name: "SHAPE", Pattern: mustCompile(`@[a-zA-Z_][a-zA-Z0-9_]*`)},
		},
		Productions: make(map[string]grammar.Production),
	}
}

func mustCompile(pattern string) *regexp.Regexp {
	return regexp.MustCompile("^" + pattern)
}

func TestLexerBasicTokens(t *testing.T) {
	g := testGrammar()
	source := `let x = 42`

	l := NewLexer(g, "test.ai", source, nil)
	tokens, err := l.Tokenize()
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
		{"EOF", ""},
	}

	if len(tokens) != len(expected) {
		t.Fatalf("expected %d tokens, got %d", len(expected), len(tokens))
	}

	for i, exp := range expected {
		if tokens[i].Type != exp.typ {
			t.Errorf("token %d: expected type %s, got %s", i, exp.typ, tokens[i].Type)
		}
		if tokens[i].Lexeme != exp.lexeme {
			t.Errorf("token %d: expected lexeme %q, got %q", i, exp.lexeme, tokens[i].Lexeme)
		}
	}
}

func TestLexerPosition(t *testing.T) {
	g := testGrammar()
	source := "let x = 1\nlet y = 2"

	l := NewLexer(g, "test.ai", source, nil)
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("tokenize error: %v", err)
	}

	// First "let" should be at line 1, col 1
	if tokens[0].Pos.Line != 1 || tokens[0].Pos.Col != 1 {
		t.Errorf("expected 1:1, got %d:%d", tokens[0].Pos.Line, tokens[0].Pos.Col)
	}

	// Second "let" should be at line 2, col 1
	// tokens: let x = 1 let y = 2 EOF
	// indices: 0   1 2 3 4   5 6 7 8
	if tokens[4].Pos.Line != 2 || tokens[4].Pos.Col != 1 {
		t.Errorf("expected 2:1, got %d:%d", tokens[4].Pos.Line, tokens[4].Pos.Col)
	}
}

func TestLexerOperators(t *testing.T) {
	g := testGrammar()
	source := `1 + 2 - 3 * 4 / 5 |> f`

	l := NewLexer(g, "test.ai", source, nil)
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("tokenize error: %v", err)
	}

	// Check operators
	ops := []string{"+", "-", "*", "/", "|>"}
	opIndex := 0
	for _, tok := range tokens {
		if tok.Type == "OPERATOR" {
			if opIndex >= len(ops) {
				t.Fatalf("too many operators")
			}
			if tok.Lexeme != ops[opIndex] {
				t.Errorf("expected operator %q, got %q", ops[opIndex], tok.Lexeme)
			}
			opIndex++
		}
	}
	if opIndex != len(ops) {
		t.Errorf("expected %d operators, got %d", len(ops), opIndex)
	}
}

func TestLexerDelimiters(t *testing.T) {
	g := testGrammar()
	source := `f(x, y) { [1, 2, 3] }`

	l := NewLexer(g, "test.ai", source, nil)
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("tokenize error: %v", err)
	}

	delims := []string{"(", ",", ")", "{", "[", ",", ",", "]", "}"}
	delimIndex := 0
	for _, tok := range tokens {
		if tok.Type == "DELIMITER" {
			if delimIndex >= len(delims) {
				t.Fatalf("too many delimiters")
			}
			if tok.Lexeme != delims[delimIndex] {
				t.Errorf("expected delimiter %q, got %q", delims[delimIndex], tok.Lexeme)
			}
			delimIndex++
		}
	}
}

func TestLexerKeywords(t *testing.T) {
	g := testGrammar()
	source := `if true { return false } else { while not x { let y = match z { } } }`

	l := NewLexer(g, "test.ai", source, nil)
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("tokenize error: %v", err)
	}

	keywords := []string{"if", "true", "return", "false", "else", "while", "not", "let", "match"}
	kwIndex := 0
	for _, tok := range tokens {
		if tok.Type == "KEYWORD" {
			if kwIndex >= len(keywords) {
				t.Fatalf("too many keywords")
			}
			if tok.Lexeme != keywords[kwIndex] {
				t.Errorf("expected keyword %q, got %q", keywords[kwIndex], tok.Lexeme)
			}
			kwIndex++
		}
	}
}

func TestLexerLiterals(t *testing.T) {
	g := testGrammar()
	source := `42 3.14 1/3 "hello" 'a' :ok @point`

	l := NewLexer(g, "test.ai", source, nil)
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("tokenize error: %v", err)
	}

	expected := []struct {
		typ    string
		lexeme string
	}{
		{"NUMBER", "42"},
		{"NUMBER", "3.14"},
		{"NUMBER", "1/3"},
		{"STRING", `"hello"`},
		{"RUNE", "'a'"},
		{"SYMBOL", ":ok"},
		{"SHAPE", "@point"},
		{"EOF", ""},
	}

	if len(tokens) != len(expected) {
		t.Fatalf("expected %d tokens, got %d", len(expected), len(tokens))
	}

	for i, exp := range expected {
		if tokens[i].Type != exp.typ {
			t.Errorf("token %d: expected type %s, got %s", i, exp.typ, tokens[i].Type)
		}
		if tokens[i].Lexeme != exp.lexeme {
			t.Errorf("token %d: expected lexeme %q, got %q", i, exp.lexeme, tokens[i].Lexeme)
		}
	}
}

func TestLexerComments(t *testing.T) {
	g := testGrammar()
	source := `let x = 1 # this is a comment
let y = 2`

	l := NewLexer(g, "test.ai", source, nil)
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("tokenize error: %v", err)
	}

	// Should not include comment
	for _, tok := range tokens {
		if strings.Contains(tok.Lexeme, "#") || strings.Contains(tok.Lexeme, "comment") {
			t.Errorf("comment leaked into tokens: %v", tok)
		}
	}

	// Should have 8 tokens: let x = 1 let y = 2 EOF
	if len(tokens) != 9 {
		t.Errorf("expected 9 tokens, got %d", len(tokens))
	}
}

func TestLexerObserver(t *testing.T) {
	g := testGrammar()
	source := `let x = 1`

	var observed []string
	obs := &testObserver{
		onLex: func(token, lexeme string, pos engine.Position) {
			observed = append(observed, token+":"+lexeme)
		},
	}

	l := NewLexer(g, "test.ai", source, obs)
	_, err := l.Tokenize()
	if err != nil {
		t.Fatalf("tokenize error: %v", err)
	}

	expected := []string{"KEYWORD:let", "NAME:x", "OPERATOR:=", "NUMBER:1", "EOF:"}
	if len(observed) != len(expected) {
		t.Fatalf("expected %d observations, got %d", len(expected), len(observed))
	}

	for i, exp := range expected {
		if observed[i] != exp {
			t.Errorf("observation %d: expected %q, got %q", i, exp, observed[i])
		}
	}
}

func TestLexerError(t *testing.T) {
	g := testGrammar()
	source := "let x = `invalid`"

	l := NewLexer(g, "test.ai", source, nil)
	_, err := l.Tokenize()
	if err == nil {
		t.Fatal("expected error for invalid character")
	}

	// Should have caret in error
	if !strings.Contains(err.Error(), "^") {
		t.Errorf("expected caret in error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "unexpected character") {
		t.Errorf("expected 'unexpected character' in error, got: %v", err)
	}
}

func TestLexerRestParam(t *testing.T) {
	g := testGrammar()
	source := `(...args)`

	l := NewLexer(g, "test.ai", source, nil)
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("tokenize error: %v", err)
	}

	// Should have: ( ... args ) EOF
	expected := []struct {
		typ    string
		lexeme string
	}{
		{"DELIMITER", "("},
		{"DELIMITER", "..."},
		{"NAME", "args"},
		{"DELIMITER", ")"},
		{"EOF", ""},
	}

	if len(tokens) != len(expected) {
		t.Fatalf("expected %d tokens, got %d: %v", len(expected), len(tokens), tokens)
	}

	for i, exp := range expected {
		if tokens[i].Type != exp.typ || tokens[i].Lexeme != exp.lexeme {
			t.Errorf("token %d: expected %s:%q, got %s:%q",
				i, exp.typ, exp.lexeme, tokens[i].Type, tokens[i].Lexeme)
		}
	}
}

// testObserver implements Observer for testing.
type testObserver struct {
	onLex func(token, lexeme string, pos engine.Position)
}

func (o *testObserver) OnLex(token, lexeme string, pos engine.Position) {
	if o.onLex != nil {
		o.onLex(token, lexeme, pos)
	}
}

func (o *testObserver) OnParse(production string, depth int, pos engine.Position) {}
func (o *testObserver) OnEval(node, result string, scope int, pos engine.Position) {}
func (o *testObserver) OnEffect(action, target string, pos engine.Position)        {}
