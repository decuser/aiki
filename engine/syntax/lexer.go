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

type literalCandidate struct {
	tokName  string
	lit      string
	defIndex int
	litIndex int
}

// Lexer tokenizes source using only grammar token definitions.
//
// It emits every recognized token, including tokens marked @skip.
//
// Matching policy: maximal munch (longest match wins). If tied, earlier
// grammar token wins; if still tied, earlier literal in the token wins.
type Lexer struct {
	grammar  *grammar.Grammar
	source   string
	file     string
	pos      int
	line     int
	col      int
	observer engine.Observer
	literals []literalCandidate
}

func NewLexer(g *grammar.Grammar, file, source string, observer engine.Observer) *Lexer {
	if observer == nil {
		observer = engine.SilentObserver{}
	}

	var lits []literalCandidate
	for di, def := range g.Tokens {
		if def.Pattern != nil {
			continue
		}
		litStr := strings.TrimSpace(def.Literal)
		if litStr == "" {
			continue
		}
		parts := strings.Fields(litStr)
		for li, part := range parts {
			lits = append(lits, literalCandidate{
				tokName:  def.Name,
				lit:      part,
				defIndex: di,
				litIndex: li,
			})
		}
	}

	return &Lexer{
		grammar:  g,
		source:   source,
		file:     file,
		pos:      0,
		line:     1,
		col:      1,
		observer: observer,
		literals: lits,
	}
}

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

func (l *Lexer) Next() (Token, error) {
	if l.pos >= len(l.source) {
		return Token{Type: "EOF", Lexeme: "", Pos: l.position()}, nil
	}

	startPos := l.position()

	bestLen := 0
	bestType := ""
	bestLex := ""
	bestDefIndex := int(^uint(0) >> 1)
	bestLitIndex := int(^uint(0) >> 1)

	// Pattern tokens
	for di, def := range l.grammar.Tokens {
		if def.Pattern == nil {
			continue
		}
		match := def.Pattern.FindString(l.source[l.pos:])
		if match == "" {
			continue
		}
		mlen := len(match)
		if mlen > bestLen || (mlen == bestLen && di < bestDefIndex) {
			bestLen = mlen
			bestType = def.Name
			bestLex = match
			bestDefIndex = di
			bestLitIndex = 0
		}
	}

	// Literal tokens
	for _, c := range l.literals {
		if !strings.HasPrefix(l.source[l.pos:], c.lit) {
			continue
		}
		mlen := len(c.lit)
		if mlen > bestLen {
			bestLen = mlen
			bestType = c.tokName
			bestLex = c.lit
			bestDefIndex = c.defIndex
			bestLitIndex = c.litIndex
			continue
		}
		if mlen == bestLen {
			if c.defIndex < bestDefIndex || (c.defIndex == bestDefIndex && c.litIndex < bestLitIndex) {
				bestLen = mlen
				bestType = c.tokName
				bestLex = c.lit
				bestDefIndex = c.defIndex
				bestLitIndex = c.litIndex
			}
		}
	}

	if bestLen == 0 {
		ch := l.source[l.pos]
		message := fmt.Sprintf("unexpected character '%c'", ch)
		return Token{}, &SourceError{
			Kind:     "lex",
			Pos:      startPos,
			Message:  message,
			Rendered: engine.FormatWithCaret(startPos, engine.GetSourceLine(l.source, startPos.Line), message),
		}
	}

	tok := Token{Type: bestType, Lexeme: bestLex, Pos: startPos}
	l.advance(bestLen)
	l.observer.OnLex(tok.Type, tok.Lexeme, tok.Pos)
	return tok, nil
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
	return engine.Position{File: l.file, Line: l.line, Col: l.col}
}
