// Package grammar provides the EBNF parser and grammar types.
package grammar

import (
	"regexp"

	"aiki/engine"
)

// Grammar holds token definitions, productions, and metadata.
type Grammar struct {
	Tokens      []TokenDef
	Productions map[string]Production
	Start       string // first production name
}

// TokenDef defines a lexical token.
type TokenDef struct {
	Name    string
	Pattern *regexp.Regexp
	Skip    bool
	Pos     engine.Position
	Meta    Meta
}

// Production defines a grammar rule.
type Production struct {
	Name        string
	Expressions []Expression // alternatives separated by |
	Pos         engine.Position
	Meta        Meta
}

// Expression is a sequence of terms (one alternative).
type Expression struct {
	Terms []Term
}

// Term is a single element in an expression.
type Term struct {
	Value    string // literal, token name, or production name
	Kind     TermKind
	Optional bool // [ ... ]
	Repeat   bool // { ... }
	Pos      engine.Position
}

// TermKind distinguishes term types.
type TermKind int

const (
	TermLiteral    TermKind = iota // "let", "|>"
	TermToken                      // NUMBER, NAME
	TermProduction                 // expr, statement
)

// Meta holds decorations for a token or production.
type Meta struct {
	Error    string // parse failure message
	Template string // readable syntax hint
	Help     string // short description
	Doc      string // full documentation with examples
}

// GetProduction returns a production by name.
func (g *Grammar) GetProduction(name string) (Production, bool) {
	p, ok := g.Productions[name]
	return p, ok
}

// GetToken returns a token definition by name.
func (g *Grammar) GetToken(name string) (TokenDef, bool) {
	for _, t := range g.Tokens {
		if t.Name == name {
			return t, true
		}
	}
	return TokenDef{}, false
}
