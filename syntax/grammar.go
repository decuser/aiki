package syntax

import "regexp"

// Grammar holds the complete parsed EBNF result
type Grammar struct {
	Tokens      []TokenDef
	Productions map[string]*Production
	Start       string // first production name
}

// TokenDef defines a token type for the lexer
type TokenDef struct {
	Name    string
	Pattern *regexp.Regexp // for regex tokens
	Literal string         // for keyword/operator literals
	Skip    bool           // true for whitespace, comments
}

// Production is a named grammar rule
type Production struct {
	Name string
	Expr Expression
}

// Expression is the interface for grammar expressions
type Expression interface {
	exprNode()
}

// Concrete expression types

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

// Reference: expr, block (production reference)
type Reference struct {
	Name string
}

func (e *Reference) exprNode() {}

// TokenRef: NAME, NUMBER (token class reference)
type TokenRef struct {
	Name string
}

func (e *TokenRef) exprNode() {}
