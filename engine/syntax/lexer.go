package syntax

import (
	"fmt"
	"strings"

	"aiki/engine"
	"aiki/engine/syntax/grammar"
)

// Token represents a lexical token.
type Token struct {
	Type   string
	Lexeme string
	Pos    engine.Position
}

// Lexer tokenizes source using grammar definitions.
type Lexer struct {
	grammar  *grammar.Grammar
	source   string
	file     string
	pos      int
	line     int
	col      int
	observer engine.Observer
}

// NewLexer creates a lexer for the given source.
func NewLexer(g *grammar.Grammar, file, source string, observer engine.Observer) *Lexer {
	if observer == nil {
		observer = engine.SilentObserver{}
	}
	return &Lexer{
		grammar:  g,
		source:   source,
		file:     file,
		pos:      0,
		line:     1,
		col:      1,
		observer: observer,
	}
}

// Tokenize returns all tokens from the source.
func (l *Lexer) Tokenize() ([]Token, error) {
	var tokens []Token

	for {
		tok, err := l.Next()
		if err != nil {
			return nil, err
		}
		if tok.Type == "EOF" {
			break
		}
		tokens = append(tokens, tok)
	}

	return tokens, nil
}

// Next returns the next token.
func (l *Lexer) Next() (Token, error) {
	// Skip whitespace and comments
	l.skipIgnored()

	if l.pos >= len(l.source) {
		tok := Token{
			Type:   "EOF",
			Lexeme: "",
			Pos:    l.position(),
		}
		l.observer.OnLex(tok.Type, tok.Lexeme, tok.Pos)
		return tok, nil
	}

	startPos := l.position()

	// Try identifier/keyword first (before patterns)
	if l.isIdentStart(l.peek()) {
	    ident := l.readIdent()
	    tokType := "NAME"
	    if l.isKeyword(ident) {
		tokType = "KEYWORD"
	    }
	    tok := Token{
		Type:   tokType,
		Lexeme: ident,
		Pos:    startPos,
	    }
	    l.observer.OnLex(tok.Type, tok.Lexeme, tok.Pos)
	    return tok, nil
	}

	// Try each token definition
	for _, def := range l.grammar.Tokens {
		if def.Skip {
			continue // Already handled in skipIgnored
		}
		if def.Pattern == nil {
			continue // Keywords/operators handled separately
		}

		match := def.Pattern.FindString(l.source[l.pos:])
		if match != "" {
			tok := Token{
				Type:   def.Name,
				Lexeme: match,
				Pos:    startPos,
			}
			l.advance(len(match))
			l.observer.OnLex(tok.Type, tok.Lexeme, tok.Pos)
			return tok, nil
		}
	}


	// Try delimiters before operators (... must match before .)
	if delim := l.matchDelimiter(); delim != "" {
		tok := Token{
			Type:   "DELIMITER",
			Lexeme: delim,
			Pos:    startPos,
		}
		l.advance(len(delim))
		l.observer.OnLex(tok.Type, tok.Lexeme, tok.Pos)
		return tok, nil
	}

	if op := l.matchOperator(); op != "" {
		tok := Token{
			Type:   "OPERATOR",
			Lexeme: op,
			Pos:    startPos,
		}
		l.advance(len(op))
		l.observer.OnLex(tok.Type, tok.Lexeme, tok.Pos)
		return tok, nil
	}

	// Unknown character
	ch := l.source[l.pos]
	return Token{}, fmt.Errorf("%s", engine.FormatWithCaret(
		startPos,
		engine.GetSourceLine(l.source, startPos.Line),
		fmt.Sprintf("unexpected character '%c'", ch),
	))
}

// skipIgnored skips whitespace and comments.
func (l *Lexer) skipIgnored() {
	for l.pos < len(l.source) {
		ch := l.source[l.pos]

		// Whitespace
		if ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n' {
			l.advanceOne()
			continue
		}

		// Comment
		if ch == '#' {
			for l.pos < len(l.source) && l.source[l.pos] != '\n' {
				l.advanceOne()
			}
			continue
		}

		break
	}
}

// readIdent reads an identifier.
func (l *Lexer) readIdent() string {
	start := l.pos
	for l.pos < len(l.source) && l.isIdentChar(l.source[l.pos]) {
		l.advanceOne()
	}
	return l.source[start:l.pos]
}

// isIdentStart returns true if ch can start an identifier.
func (l *Lexer) isIdentStart(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

// isIdentChar returns true if ch can be in an identifier.
func (l *Lexer) isIdentChar(ch byte) bool {
	return l.isIdentStart(ch) || (ch >= '0' && ch <= '9')
}

// isKeyword returns true if ident is a keyword.
func (l *Lexer) isKeyword(ident string) bool {
	keywords := []string{
		"let", "if", "else", "while", "match", "return",
		"true", "false", "and", "or", "not",
	}
	for _, kw := range keywords {
		if ident == kw {
			return true
		}
	}
	return false
}

// matchOperator tries to match an operator at current position.
func (l *Lexer) matchOperator() string {
	// Longer operators first
	operators := []string{"|>", "<=", ">=", "+", "-", "*", "/", "<", ">", ".", "="}
	for _, op := range operators {
		if strings.HasPrefix(l.source[l.pos:], op) {
			return op
		}
	}
	return ""
}

// matchDelimiter tries to match a delimiter at current position.
func (l *Lexer) matchDelimiter() string {
	// Longer delimiters first
	delimiters := []string{"...", "(", ")", "[", "]", "{", "}", ",", ";"}
	for _, d := range delimiters {
		if strings.HasPrefix(l.source[l.pos:], d) {
			return d
		}
	}
	return ""
}

// Helper methods

func (l *Lexer) peek() byte {
	if l.pos >= len(l.source) {
		return 0
	}
	return l.source[l.pos]
}

func (l *Lexer) advanceOne() {
	if l.pos < len(l.source) {
		if l.source[l.pos] == '\n' {
			l.line++
			l.col = 1
		} else {
			l.col++
		}
		l.pos++
	}
}

func (l *Lexer) advance(n int) {
	for i := 0; i < n; i++ {
		l.advanceOne()
	}
}

func (l *Lexer) position() engine.Position {
	return engine.Position{
		File: l.file,
		Line: l.line,
		Col:  l.col,
	}
}
