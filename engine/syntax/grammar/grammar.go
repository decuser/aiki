// Package grammar provides the EBNF parser and grammar types.
package grammar

import (
	"regexp"

	"aiki/engine"
)

// Grammar holds token definitions, surface policy, and productions.
type Grammar struct {
	Tokens      []TokenDef
	Productions map[string]*Production
	Start       string
	Newline     *NewlineRule
	analysis    *Analysis
}

// NewlineRule declares how physical newline tokens become statement boundaries.
type NewlineRule struct {
	Token       string
	AfterToken  []string
	AfterLexeme []string
	SuppressIn  [][2]string
	Meta        Meta
	Pos         engine.Position
}

// TokenDef defines a lexical token.
type TokenDef struct {
	Name    string
	Pattern *regexp.Regexp
	Literal string // for keywords/operators
	Skip    bool
	Pos     engine.Position
	Meta    Meta
}

// Production is a named grammar rule.
type Production struct {
	Name string
	Expr Expression
	Pos  engine.Position
	Meta Meta
}

// Meta holds decorations for a token or production.
type Meta struct {
	Error    string
	Template string
	Help     string
	Doc      string
}

// Expression is the interface for grammar expressions.
type Expression interface {
	exprNode()
}

// Sequence: a b c
type Sequence struct {
	Exprs []Expression
}

func (e *Sequence) exprNode() {}

// Alternative: a | b | c
type Alternative struct {
	Exprs []Expression
}

func (e *Alternative) exprNode() {}

// Repetition: { a }
type Repetition struct {
	Expr Expression
}

func (e *Repetition) exprNode() {}

// Option: [ a ]
type Option struct {
	Expr Expression
}

func (e *Option) exprNode() {}

// Group: ( a )
type Group struct {
	Expr Expression
}

func (e *Group) exprNode() {}

// Terminal: "let"
type Terminal struct {
	Value string
}

func (e *Terminal) exprNode() {}

// Reference: production reference
type Reference struct {
	Name string
}

func (e *Reference) exprNode() {}

// TokenRef: token class reference (NAME, NUMBER, etc.)
type TokenRef struct {
	Name string
}

func (e *TokenRef) exprNode() {}

// GetProduction returns a production by name.
func (g *Grammar) GetProduction(name string) (*Production, bool) {
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
