package syntax

import (
	"aiki/engine"
	"fmt"
)

type Token struct {
	Type   string
	Lexeme string
	Pos    engine.Position
}

type Lexer struct {
	source   string
	contract GrammarContract
	file     string // Added field
	pos      int
	line     int
	col      int
}

func NewLexer(file string, source string, contract GrammarContract) *Lexer {
	return &Lexer{
		file:     file,
		source:   source,
		contract: contract,
		line:     1,
		col:      1,
	}
}

func (l *Lexer) NextToken() (Token, error) {
	for l.pos < len(l.source) {
		remaining := l.source[l.pos:]
		matched := false

		for _, def := range l.contract.GetTokens() {
			loc := def.Pattern.FindStringIndex(remaining)
			if loc != nil && loc[0] == 0 {
				lexeme := remaining[loc[0]:loc[1]]
				t := Token{
					Type:   def.Name,
					Lexeme: lexeme,
					Pos:    engine.Position{File: l.file, Line: l.line, Col: l.col},
				}
				
				l.advance(len(lexeme))
				matched = true
				if def.Skip { break }
				
				l.contract.Observe().OnLex(t.Type, t.Lexeme, t.Pos)
				return t, nil
			}
		}
		if !matched {
			return Token{}, fmt.Errorf("%s:%d:%d: lexical error", l.file, l.line, l.col)
		}
	}
	return Token{Type: "EOF", Pos: engine.Position{File: l.file, Line: l.line, Col: l.col}}, nil
}

func (l *Lexer) advance(n int) {
	for i := 0; i < n; i++ {
		if l.source[l.pos] == '\n' {
			l.line++
			l.col = 1
		} else {
			l.col++
		}
		l.pos++
	}
}
