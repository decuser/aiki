package ebnf

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Comment represents a source comment with its line number.
type Comment struct {
	Line  int
	Text  string
	IsEOL bool // true if comment follows code on same line
}

// Lexer tokenizes source using token definitions from Grammar
type Lexer struct {
	defs        []TokenDef
	source      string
	pos         int
	line        int
	col         int
	Comments    []Comment // collected comments
	lastTokLine int       // line of last emitted token
}

// NewLexer creates a lexer from token definitions
func NewLexer(defs []TokenDef, source string) *Lexer {
	return &Lexer{
		defs:   defs,
		source: source,
		pos:    0,
		line:   1,
		col:    1,
	}
}

// Tokenize returns all tokens from source
func (g *Grammar) Tokenize(source string) ([]Token, error) {
	l := NewLexer(g.Tokens, source)
	return l.Tokenize()
}

// TokenizeWithComments returns all tokens and collected comments from source.
func (g *Grammar) TokenizeWithComments(source string) ([]Token, []Comment, error) {
	l := NewLexer(g.Tokens, source)
	tokens, err := l.Tokenize()
	return tokens, l.Comments, err
}

// Tokenize returns all tokens
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

// Next returns the next token
func (l *Lexer) Next() (Token, error) {
	for l.pos < len(l.source) {
		startLine := l.line
		startCol := l.col
		startPos := l.pos

		// Try literal tokens first (keywords, operators), then patterns
		// This ensures "let" matches KEYWORD before NAME's regex
		var bestMatch struct {
			def    *TokenDef
			lexeme string
		}

		for i := range l.defs {
			def := &l.defs[i]
			var lexeme string
			var ok bool

			if def.Literal != "" {
				lexeme, ok = l.matchLiterals(def.Literal)
				if ok && len(lexeme) > len(bestMatch.lexeme) {
					bestMatch.def = def
					bestMatch.lexeme = lexeme
				}
			}
		}

		// If no literal matched, try patterns
		if bestMatch.def == nil {
			for i := range l.defs {
				def := &l.defs[i]
				if def.Pattern != nil {
					lexeme, ok := l.matchPattern(def.Pattern)
					if ok && len(lexeme) > len(bestMatch.lexeme) {
						bestMatch.def = def
						bestMatch.lexeme = lexeme
					}
				}
			}
		}

		if bestMatch.def != nil {
			l.advance(bestMatch.lexeme)

			if bestMatch.def.Skip {
				// Collect comments instead of silently discarding
				if strings.HasPrefix(bestMatch.lexeme, "#") {
					text := strings.TrimRight(bestMatch.lexeme, "\n")
					isEOL := l.lastTokLine == startLine
					l.Comments = append(l.Comments, Comment{
						Line:  startLine,
						Text:  text,
						IsEOL: isEOL,
					})
				}

				if l.pos == startPos {
					return Token{}, fmt.Errorf("line %d col %d: lexer stuck on skip", l.line, l.col)
				}
				continue
			}

			l.lastTokLine = startLine

			return Token{
				Type:   bestMatch.def.Name,
				Lexeme: bestMatch.lexeme,
				Line:   startLine,
				Column: startCol,
			}, nil
		}

		// No token matched
		ch, _ := utf8.DecodeRuneInString(l.source[l.pos:])
		return Token{}, fmt.Errorf("line %d col %d: unexpected character '%c'", l.line, l.col, ch)
	}

	return Token{Type: "EOF", Line: l.line, Column: l.col}, nil
}

func (l *Lexer) matchPattern(pattern interface{ FindString(string) string }) (string, bool) {
	remaining := l.source[l.pos:]
	match := pattern.FindString(remaining)
	if match != "" && strings.HasPrefix(remaining, match) {
		return match, true
	}
	return "", false
}

func (l *Lexer) matchLiterals(literals string) (string, bool) {
	remaining := l.source[l.pos:]

	// Literals are space-separated keywords/operators
	// Try longest match first
	parts := strings.Fields(literals)

	// Sort by length descending for longest match
	for i := 0; i < len(parts); i++ {
		for j := i + 1; j < len(parts); j++ {
			if len(parts[j]) > len(parts[i]) {
				parts[i], parts[j] = parts[j], parts[i]
			}
		}
	}

	for _, lit := range parts {
		if strings.HasPrefix(remaining, lit) {
			// For keywords, ensure not followed by identifier char
			if isKeyword(lit) {
				after := remaining[len(lit):]
				if len(after) > 0 {
					r, _ := utf8.DecodeRuneInString(after)
					if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
						continue // not a keyword match
					}
				}
			}
			return lit, true
		}
	}

	return "", false
}

func isKeyword(s string) bool {
	if len(s) == 0 {
		return false
	}
	r, _ := utf8.DecodeRuneInString(s)
	return unicode.IsLetter(r)
}

func (l *Lexer) advance(lexeme string) {
	for _, ch := range lexeme {
		if ch == '\n' {
			l.line++
			l.col = 1
		} else {
			l.col++
		}
	}
	l.pos += len(lexeme)
}
