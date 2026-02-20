// definition/parser.go - parses EBNF grammar files
package definition

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// Grammar holds the complete parsed EBNF result
type Grammar struct {
	Tokens      []TokenDef
	Productions map[string]*Production
	Start       string
}

// TokenDef defines a token type for the lexer
type TokenDef struct {
	Name    string
	Pattern *regexp.Regexp
	Literal string
	Skip    bool
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

// Parse parses EBNF source and returns a Grammar
func Parse(source string) (*Grammar, error) {
	p := &ebnfParser{
		source: source,
		pos:    0,
		line:   1,
		col:    1,
	}
	return p.parse()
}

type ebnfParser struct {
	source string
	pos    int
	line   int
	col    int
}

func (p *ebnfParser) parse() (*Grammar, error) {
	g := &Grammar{
		Productions: make(map[string]*Production),
	}

	p.skipWhitespaceAndComments()

	// Parse @tokens block if present
	if p.peek() == '@' {
		tokens, err := p.parseTokensBlock()
		if err != nil {
			return nil, err
		}
		g.Tokens = tokens
		p.skipWhitespaceAndComments()
	}

	// Pass 1: Parse productions
	for p.pos < len(p.source) {
		p.skipWhitespaceAndComments()
		if p.pos >= len(p.source) {
			break
		}

		prod, err := p.parseProduction()
		if err != nil {
			return nil, err
		}

		if g.Start == "" {
			g.Start = prod.Name
		}
		g.Productions[prod.Name] = prod

		p.skipWhitespaceAndComments()
	}

	// Pass 2: Resolve references
	// Convert TokenRef to Reference if name exists in productions
	for _, prod := range g.Productions {
		prod.Expr = resolveRefs(prod.Expr, g.Productions)
	}

	return g, nil
}

// resolveRefs converts TokenRef to Reference if the name is a production
func resolveRefs(expr Expression, prods map[string]*Production) Expression {
	switch e := expr.(type) {
	case *Sequence:
		for i, sub := range e.Exprs {
			e.Exprs[i] = resolveRefs(sub, prods)
		}
		return e
	case *Alternative:
		for i, sub := range e.Exprs {
			e.Exprs[i] = resolveRefs(sub, prods)
		}
		return e
	case *Repetition:
		e.Expr = resolveRefs(e.Expr, prods)
		return e
	case *Option:
		e.Expr = resolveRefs(e.Expr, prods)
		return e
	case *Group:
		e.Expr = resolveRefs(e.Expr, prods)
		return e
	case *TokenRef:
		// If this name is a production, convert to Reference
		if _, ok := prods[e.Name]; ok {
			return &Reference{Name: e.Name}
		}
		return e
	default:
		return expr
	}
}

func (p *ebnfParser) parseTokensBlock() ([]TokenDef, error) {
	if !p.consumeString("@tokens") {
		return nil, p.error("expected @tokens")
	}

	p.skipWhitespaceAndComments()

	if !p.consume('{') {
		return nil, p.error("expected '{' after @tokens")
	}

	var tokens []TokenDef

	p.skipWhitespaceAndComments()
	for p.peek() != '}' && p.pos < len(p.source) {
		tok, err := p.parseTokenDef()
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, tok)
		p.skipWhitespaceAndComments()
	}

	if !p.consume('}') {
		return nil, p.error("expected '}' after tokens block")
	}

	return tokens, nil
}

func (p *ebnfParser) parseTokenDef() (TokenDef, error) {
	name := p.parseIdentifier()
	if name == "" {
		return TokenDef{}, p.error("expected token name")
	}

	p.skipSpaces()

	var def TokenDef
	def.Name = name

	// Check for regex pattern /.../ or literal keywords
	if p.peek() == '/' {
		pattern, err := p.parseRegex()
		if err != nil {
			return TokenDef{}, err
		}
		re, err := regexp.Compile("^" + pattern)
		if err != nil {
			return TokenDef{}, p.error("invalid regex: " + err.Error())
		}
		def.Pattern = re
	} else {
		// Literal keywords/operators (space separated until newline)
		literals := p.parseUntilNewline()
		def.Literal = strings.TrimSpace(literals)
	}

	// Check for @skip
	p.skipSpaces()
	if p.consumeString("@skip") {
		def.Skip = true
	}

	return def, nil
}

func (p *ebnfParser) parseRegex() (string, error) {
	if !p.consume('/') {
		return "", p.error("expected '/'")
	}

	start := p.pos
	for p.pos < len(p.source) && p.source[p.pos] != '/' {
		if p.source[p.pos] == '\\' && p.pos+1 < len(p.source) {
			p.pos += 2 // skip escaped char
		} else {
			p.pos++
		}
	}
	pattern := p.source[start:p.pos]

	if !p.consume('/') {
		return "", p.error("unterminated regex")
	}

	return pattern, nil
}

func (p *ebnfParser) parseProduction() (*Production, error) {
	name := p.parseIdentifier()
	if name == "" {
		return nil, p.error("expected production name")
	}

	p.skipWhitespaceAndComments()

	if !p.consume('=') {
		return nil, p.error("expected '=' after production name")
	}

	p.skipWhitespaceAndComments()

	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	// Optional trailing .
	p.skipWhitespaceAndComments()
	p.consume('.')

	return &Production{Name: name, Expr: expr}, nil
}

func (p *ebnfParser) parseExpression() (Expression, error) {
	return p.parseAlternative()
}

func (p *ebnfParser) parseAlternative() (Expression, error) {
	first, err := p.parseSequenceExpr()
	if err != nil {
		return nil, err
	}

	var alts []Expression
	alts = append(alts, first)

	for {
		p.skipWhitespaceAndComments()
		if !p.consume('|') {
			break
		}
		p.skipWhitespaceAndComments()

		next, err := p.parseSequenceExpr()
		if err != nil {
			return nil, err
		}
		alts = append(alts, next)
	}

	if len(alts) == 1 {
		return alts[0], nil
	}
	return &Alternative{Exprs: alts}, nil
}

func (p *ebnfParser) parseSequenceExpr() (Expression, error) {
	var terms []Expression

	for {
		p.skipSpaces()
		term, err := p.parseTerm()
		if err != nil {
			return nil, err
		}
		if term == nil {
			break
		}
		terms = append(terms, term)
	}

	if len(terms) == 0 {
		return nil, p.error("expected expression")
	}
	if len(terms) == 1 {
		return terms[0], nil
	}
	return &Sequence{Exprs: terms}, nil
}

func (p *ebnfParser) parseTerm() (Expression, error) {
	p.skipSpaces()

	ch := p.peek()

	// End of term indicators
	if ch == 0 || ch == '|' || ch == ')' || ch == ']' || ch == '}' || ch == '.' || ch == '\n' || ch == '@' {
		return nil, nil
	}

	// Group: ( ... )
	if ch == '(' {
		p.consume('(')
		p.skipWhitespaceAndComments()
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		p.skipWhitespaceAndComments()
		if !p.consume(')') {
			return nil, p.error("expected ')'")
		}
		return &Group{Expr: expr}, nil
	}

	// Option: [ ... ]
	if ch == '[' {
		p.consume('[')
		p.skipWhitespaceAndComments()
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		p.skipWhitespaceAndComments()
		if !p.consume(']') {
			return nil, p.error("expected ']'")
		}
		return &Option{Expr: expr}, nil
	}

	// Repetition: { ... }
	if ch == '{' {
		p.consume('{')
		p.skipWhitespaceAndComments()
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		p.skipWhitespaceAndComments()
		if !p.consume('}') {
			return nil, p.error("expected '}'")
		}
		return &Repetition{Expr: expr}, nil
	}

	// Terminal: "..."
	if ch == '"' {
		p.consume('"')
		start := p.pos
		for p.pos < len(p.source) && p.source[p.pos] != '"' {
			p.pos++
		}
		value := p.source[start:p.pos]
		if !p.consume('"') {
			return nil, p.error("unterminated string")
		}
		return &Terminal{Value: value}, nil
	}

	// Identifier: production reference or token class
	if unicode.IsLetter(rune(ch)) || ch == '_' {
		name := p.parseIdentifier()
		if isTokenClass(name) {
			return &TokenRef{Name: name}, nil
		}
		return &Reference{Name: name}, nil
	}

	return nil, nil
}

func isTokenClass(name string) bool {
	// Token classes are all uppercase
	for _, r := range name {
		if !unicode.IsUpper(r) && r != '_' {
			return false
		}
	}
	return len(name) > 0
}

func (p *ebnfParser) parseIdentifier() string {
	start := p.pos
	for p.pos < len(p.source) {
		ch := p.source[p.pos]
		if unicode.IsLetter(rune(ch)) || unicode.IsDigit(rune(ch)) || ch == '_' {
			p.pos++
		} else {
			break
		}
	}
	return p.source[start:p.pos]
}

func (p *ebnfParser) parseUntilNewline() string {
	start := p.pos
	for p.pos < len(p.source) && p.source[p.pos] != '\n' && p.source[p.pos] != '@' {
		p.pos++
	}
	return p.source[start:p.pos]
}

func (p *ebnfParser) peek() byte {
	if p.pos >= len(p.source) {
		return 0
	}
	return p.source[p.pos]
}

func (p *ebnfParser) consume(ch byte) bool {
	if p.peek() == ch {
		p.pos++
		if ch == '\n' {
			p.line++
			p.col = 1
		} else {
			p.col++
		}
		return true
	}
	return false
}

func (p *ebnfParser) consumeString(s string) bool {
	if strings.HasPrefix(p.source[p.pos:], s) {
		p.pos += len(s)
		p.col += len(s)
		return true
	}
	return false
}

func (p *ebnfParser) skipSpaces() {
	for p.pos < len(p.source) {
		ch := p.source[p.pos]
		if ch == ' ' || ch == '\t' {
			p.pos++
			p.col++
		} else {
			break
		}
	}
}

func (p *ebnfParser) skipWhitespaceAndComments() {
	for p.pos < len(p.source) {
		ch := p.source[p.pos]
		if ch == ' ' || ch == '\t' {
			p.pos++
			p.col++
		} else if ch == '\n' {
			p.pos++
			p.line++
			p.col = 1
		} else if ch == '\r' {
			p.pos++
		} else if ch == '#' {
			// Skip comment to end of line
			for p.pos < len(p.source) && p.source[p.pos] != '\n' {
				p.pos++
			}
		} else {
			break
		}
	}
}

func (p *ebnfParser) error(msg string) error {
	return fmt.Errorf("line %d col %d: %s", p.line, p.col, msg)
}
