package tests

import (
	"testing"

	"aiki/lexer"
	"aiki/token"
)

func TestLexerBasicTokens(t *testing.T) {
	tests := []struct {
		input    string
		expected []token.Type
	}{
		{"1 + 2", []token.Type{token.Number, token.Plus, token.Number, token.EOF}},
		{"42", []token.Type{token.Number, token.EOF}},
		{"3.14", []token.Type{token.Number, token.EOF}},
		{"1/3", []token.Type{token.Number, token.Slash, token.Number, token.EOF}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := lexer.New(tt.input)
			tokens := l.Tokenize()

			if len(tokens) != len(tt.expected) {
				t.Fatalf("token count: got %d, want %d\ntokens: %v", len(tokens), len(tt.expected), tokens)
			}

			for i, tok := range tokens {
				if tok.Type != tt.expected[i] {
					t.Errorf("token[%d]: got %v, want %v", i, tok.Type, tt.expected[i])
				}
			}
		})
	}
}

func TestLexerStringsAndRunes(t *testing.T) {
	tests := []struct {
		input    string
		expected []token.Type
	}{
		{`"hello"`, []token.Type{token.String, token.EOF}},
		{`'a'`, []token.Type{token.Rune, token.EOF}},
		{`'€'`, []token.Type{token.Rune, token.EOF}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := lexer.New(tt.input)
			tokens := l.Tokenize()

			if len(tokens) != len(tt.expected) {
				t.Fatalf("token count: got %d, want %d", len(tokens), len(tt.expected))
			}

			for i, tok := range tokens {
				if tok.Type != tt.expected[i] {
					t.Errorf("token[%d]: got %v, want %v", i, tok.Type, tt.expected[i])
				}
			}
		})
	}
}

func TestLexerSymbolsAndShapes(t *testing.T) {
	tests := []struct {
		input    string
		expected []token.Type
	}{
		{":active", []token.Type{token.Symbol, token.EOF}},
		{":ok", []token.Type{token.Symbol, token.EOF}},
		{"@point", []token.Type{token.Shape, token.EOF}},
		{"@user_data", []token.Type{token.Shape, token.EOF}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := lexer.New(tt.input)
			tokens := l.Tokenize()

			for i, tok := range tokens {
				if tok.Type != tt.expected[i] {
					t.Errorf("token[%d]: got %v, want %v", i, tok.Type, tt.expected[i])
				}
			}
		})
	}
}

func TestLexerKeywords(t *testing.T) {
	tests := []struct {
		input    string
		expected token.Type
	}{
		{"let", token.Let},
		{"if", token.If},
		{"else", token.Else},
		{"while", token.While},
		{"match", token.Match},
		{"return", token.Return},
		{"true", token.True},
		{"false", token.False},
		{"and", token.And},
		{"or", token.Or},
		{"not", token.Not},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := lexer.New(tt.input)
			tokens := l.Tokenize()

			if tokens[0].Type != tt.expected {
				t.Errorf("got %v, want %v", tokens[0].Type, tt.expected)
			}
		})
	}
}

func TestLexerOperators(t *testing.T) {
	tests := []struct {
		input    string
		expected []token.Type
	}{
		{"+ - * / %", []token.Type{token.Plus, token.Minus, token.Star, token.Slash, token.Percent, token.EOF}},
		{"== != < > <= >=", []token.Type{token.Eq, token.NotEq, token.Lt, token.Gt, token.LtEq, token.GtEq, token.EOF}},
		{"=", []token.Type{token.Assign, token.EOF}},
		{"|>", []token.Type{token.Pipe, token.EOF}},
		{".", []token.Type{token.Dot, token.EOF}},
		{",", []token.Type{token.Comma, token.EOF}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := lexer.New(tt.input)
			tokens := l.Tokenize()

			for i, tok := range tokens {
				if tok.Type != tt.expected[i] {
					t.Errorf("token[%d]: got %v, want %v", i, tok.Type, tt.expected[i])
				}
			}
		})
	}
}

func TestLexerDelimiters(t *testing.T) {
	input := "( ) [ ] { } ,"
	expected := []token.Type{token.LParen, token.RParen, token.LBracket, token.RBracket, token.LBrace, token.RBrace, token.Comma, token.EOF}

	l := lexer.New(input)
	tokens := l.Tokenize()

	for i, tok := range tokens {
		if tok.Type != expected[i] {
			t.Errorf("token[%d]: got %v, want %v", i, tok.Type, expected[i])
		}
	}
}

func TestLexerCombined(t *testing.T) {
	tests := []struct {
		input    string
		expected []token.Type
	}{
		{"1 + 2 * 3", []token.Type{token.Number, token.Plus, token.Number, token.Star, token.Number, token.EOF}},
		{"(1 + 2)", []token.Type{token.LParen, token.Number, token.Plus, token.Number, token.RParen, token.EOF}},
		{"[1, 2, 3]", []token.Type{token.LBracket, token.Number, token.Comma, token.Number, token.Comma, token.Number, token.RBracket, token.EOF}},
		{"[@point, 10, 20]", []token.Type{token.LBracket, token.Shape, token.Comma, token.Number, token.Comma, token.Number, token.RBracket, token.EOF}},
		{"(n) { return n }", []token.Type{token.LParen, token.Name, token.RParen, token.LBrace, token.Return, token.Name, token.RBrace, token.EOF}},
		{"x |> f() |> g()", []token.Type{token.Name, token.Pipe, token.Name, token.LParen, token.RParen, token.Pipe, token.Name, token.LParen, token.RParen, token.EOF}},
		{"let x = 5", []token.Type{token.Let, token.Name, token.Assign, token.Number, token.EOF}},
		{"let @point [x, y]", []token.Type{token.Let, token.Shape, token.LBracket, token.Name, token.Comma, token.Name, token.RBracket, token.EOF}},
		{"f(a, b, c)", []token.Type{token.Name, token.LParen, token.Name, token.Comma, token.Name, token.Comma, token.Name, token.RParen, token.EOF}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := lexer.New(tt.input)
			tokens := l.Tokenize()

			if len(tokens) != len(tt.expected) {
				t.Fatalf("token count: got %d, want %d\ntokens: %v", len(tokens), len(tt.expected), tokens)
			}

			for i, tok := range tokens {
				if tok.Type != tt.expected[i] {
					t.Errorf("token[%d]: got %v, want %v", i, tok.Type, tt.expected[i])
				}
			}
		})
	}
}

func TestLexerComments(t *testing.T) {
	tests := []struct {
		input    string
		expected []token.Type
	}{
		{"1 + 2 # comment", []token.Type{token.Number, token.Plus, token.Number, token.EOF}},
		{"# full line comment\n42", []token.Type{token.Number, token.EOF}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := lexer.New(tt.input)
			tokens := l.Tokenize()

			if len(tokens) != len(tt.expected) {
				t.Fatalf("token count: got %d, want %d", len(tokens), len(tt.expected))
			}

			for i, tok := range tokens {
				if tok.Type != tt.expected[i] {
					t.Errorf("token[%d]: got %v, want %v", i, tok.Type, tt.expected[i])
				}
			}
		})
	}
}

func TestLexerIllegalTokens(t *testing.T) {
	tests := []string{"|", "!", ":", "@"}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			l := lexer.New(input)
			tokens := l.Tokenize()

			hasIllegal := false
			for _, tok := range tokens {
				if tok.Type == token.Illegal {
					hasIllegal = true
					break
				}
			}
			if !hasIllegal {
				t.Errorf("expected Illegal token for input %q", input)
			}
		})
	}
}
