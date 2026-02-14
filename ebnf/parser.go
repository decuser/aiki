package ebnf

import (
	"fmt"
)

// Parser parses tokens into AST using grammar
type Parser struct {
	grammar *Grammar
	tokens  []Token
	pos     int
}

// NewParser creates a parser for the given grammar and tokens
func NewParser(grammar *Grammar, tokens []Token) *Parser {
	return &Parser{
		grammar: grammar,
		tokens:  tokens,
		pos:     0,
	}
}

// Parse parses tokens into AST
func (g *Grammar) Parse(tokens []Token) (*Node, error) {
	p := NewParser(g, tokens)
	return p.Parse()
}

// ParseSource tokenizes and parses source
func (g *Grammar) ParseSource(source string) (*Node, error) {
	tokens, err := g.Tokenize(source)
	if err != nil {
		return nil, err
	}
	return g.Parse(tokens)
}

// Parse parses from the start production
func (p *Parser) Parse() (*Node, error) {
	node, err := p.parseProduction(p.grammar.Start)
	if err != nil {
		return nil, err
	}

	// Should consume all tokens
	if p.pos < len(p.tokens) {
		tok := p.tokens[p.pos]
		return nil, fmt.Errorf("line %d col %d: unexpected token '%s'", tok.Line, tok.Column, tok.Lexeme)
	}

	return node, nil
}

func (p *Parser) parseProduction(name string) (*Node, error) {
	prod, ok := p.grammar.Productions[name]
	if !ok {
		return nil, fmt.Errorf("undefined production: %s", name)
	}

	startPos := p.pos
	children, err := p.parseExpr(prod.Expr)
	if err != nil {
		p.pos = startPos // backtrack
		return nil, err
	}

	node := &Node{
		Type:     name,
		Children: children,
	}
	if len(children) > 0 {
		node.Line = children[0].Line
		node.Column = children[0].Column
	}

	return node, nil
}

func (p *Parser) parseExpr(expr Expression) ([]*Node, error) {
	switch e := expr.(type) {
	case *Sequence:
		return p.parseSequence(e)
	case *Alternative:
		return p.parseAlternative(e)
	case *Repetition:
		return p.parseRepetition(e)
	case *Option:
		return p.parseOption(e)
	case *Group:
		return p.parseExpr(e.Expr)
	case *Terminal:
		return p.parseTerminal(e)
	case *Reference:
		return p.parseReference(e)
	case *TokenRef:
		return p.parseTokenRef(e)
	default:
		return nil, fmt.Errorf("unknown expression type: %T", expr)
	}
}

func (p *Parser) parseSequence(seq *Sequence) ([]*Node, error) {
	var all []*Node
	for _, expr := range seq.Exprs {
		nodes, err := p.parseExpr(expr)
		if err != nil {
			return nil, err
		}
		all = append(all, nodes...)
	}
	return all, nil
}

func (p *Parser) parseAlternative(alt *Alternative) ([]*Node, error) {
	startPos := p.pos
	for _, expr := range alt.Exprs {
		p.pos = startPos
		nodes, err := p.parseExpr(expr)
		if err == nil {
			return nodes, nil
		}
	}
	return nil, p.error("no alternative matched")
}

func (p *Parser) parseRepetition(rep *Repetition) ([]*Node, error) {
	var all []*Node
	for {
		startPos := p.pos
		nodes, err := p.parseExpr(rep.Expr)
		if err != nil {
			p.pos = startPos // backtrack
			break
		}
		all = append(all, nodes...)
	}
	return all, nil // zero or more, always succeeds
}

func (p *Parser) parseOption(opt *Option) ([]*Node, error) {
	startPos := p.pos
	nodes, err := p.parseExpr(opt.Expr)
	if err != nil {
		p.pos = startPos // backtrack
		return nil, nil  // optional, return empty
	}
	return nodes, nil
}

func (p *Parser) parseTerminal(term *Terminal) ([]*Node, error) {
	if p.pos >= len(p.tokens) {
		return nil, p.error("unexpected end of input, expected '%s'", term.Value)
	}

	tok := p.tokens[p.pos]
	if tok.Lexeme != term.Value {
		return nil, p.error("expected '%s', got '%s'", term.Value, tok.Lexeme)
	}

	p.pos++
	return []*Node{{
		Type:   "TERMINAL",
		Value:  tok.Lexeme,
		Line:   tok.Line,
		Column: tok.Column,
	}}, nil
}

func (p *Parser) parseReference(ref *Reference) ([]*Node, error) {
	node, err := p.parseProduction(ref.Name)
	if err != nil {
		return nil, err
	}
	return []*Node{node}, nil
}

func (p *Parser) parseTokenRef(ref *TokenRef) ([]*Node, error) {
	if p.pos >= len(p.tokens) {
		return nil, p.error("unexpected end of input, expected %s", ref.Name)
	}

	tok := p.tokens[p.pos]
	if tok.Type != ref.Name {
		return nil, p.error("expected %s, got %s", ref.Name, tok.Type)
	}

	p.pos++
	return []*Node{{
		Type:   ref.Name,
		Value:  tok.Lexeme,
		Line:   tok.Line,
		Column: tok.Column,
	}}, nil
}

func (p *Parser) error(format string, args ...interface{}) error {
	if p.pos < len(p.tokens) {
		tok := p.tokens[p.pos]
		prefix := fmt.Sprintf("line %d col %d: ", tok.Line, tok.Column)
		return fmt.Errorf(prefix+format, args...)
	}
	return fmt.Errorf("end of input: "+format, args...)
}
