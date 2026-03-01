package syntax

import (
	"fmt"

	"aiki/engine"
	"aiki/engine/syntax/grammar"
)

type Parser struct {
	grammar  *grammar.Grammar
	tokens   []Token
	pos      int
	furthest int // furthest position reached - for better error reporting
	source   string
	observer engine.Observer
}

func NewParser(g *grammar.Grammar, tokens []Token, source string, observer engine.Observer) *Parser {
	if observer == nil {
		observer = engine.SilentObserver{}
	}
	return &Parser{grammar: g, tokens: tokens, pos: 0, furthest: 0, source: source, observer: observer}
}

func (p *Parser) Parse() (*Node, error) {
	node, err := p.parseProduction(p.grammar.Start)
	if err != nil {
		return nil, err
	}
	if p.pos < len(p.tokens) {
		// Use furthest position for better error location
		errPos := p.furthest
		if errPos >= len(p.tokens) {
			errPos = len(p.tokens) - 1
		}
		tok := p.tokens[errPos]
		return nil, fmt.Errorf("%s", engine.FormatWithCaret(tok.Pos, engine.GetSourceLine(p.source, tok.Pos.Line),
			fmt.Sprintf("unexpected token '%s'", tok.Lexeme)))
	}
	return node, nil
}

func (p *Parser) parseProduction(name string) (*Node, error) {
	prod, ok := p.grammar.GetProduction(name)
	if !ok {
		return nil, fmt.Errorf("undefined production: %s", name)
	}

	startPos := p.pos
	startTok := p.peek()
	p.observer.OnParse(name, 0, startTok.Pos)

	children, err := p.parseExpr(prod.Expr)
	if err != nil {
		p.pos = startPos
		return nil, err
	}

	node := &Node{Type: name, Children: children, Pos: startTok.Pos}
	return node, nil
}

func (p *Parser) parseExpr(expr grammar.Expression) ([]*Node, error) {
	switch e := expr.(type) {
	case *grammar.Sequence:
		return p.parseSequence(e)
	case *grammar.Alternative:
		return p.parseAlternative(e)
	case *grammar.Repetition:
		return p.parseRepetition(e)
	case *grammar.Option:
		return p.parseOption(e)
	case *grammar.Group:
		return p.parseExpr(e.Expr)
	case *grammar.Terminal:
		return p.parseTerminal(e)
	case *grammar.Reference:
		return p.parseReference(e)
	case *grammar.TokenRef:
		return p.parseTokenRef(e)
	default:
		return nil, fmt.Errorf("unknown expression type: %T", expr)
	}
}

func (p *Parser) parseSequence(seq *grammar.Sequence) ([]*Node, error) {
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

func (p *Parser) parseAlternative(alt *grammar.Alternative) ([]*Node, error) {
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

func (p *Parser) parseRepetition(rep *grammar.Repetition) ([]*Node, error) {
	var all []*Node
	for {
		startPos := p.pos
		nodes, err := p.parseExpr(rep.Expr)
		if err != nil {
			p.pos = startPos
			break
		}
		all = append(all, nodes...)
	}
	return all, nil
}

func (p *Parser) parseOption(opt *grammar.Option) ([]*Node, error) {
	startPos := p.pos
	nodes, err := p.parseExpr(opt.Expr)
	if err != nil {
		p.pos = startPos
		return nil, nil
	}
	return nodes, nil
}

func (p *Parser) parseTerminal(term *grammar.Terminal) ([]*Node, error) {
	if p.pos >= len(p.tokens) {
		return nil, p.error("unexpected end of input, expected '%s'", term.Value)
	}
	tok := p.tokens[p.pos]
	if tok.Lexeme != term.Value {
		return nil, p.error("expected '%s', got '%s'", term.Value, tok.Lexeme)
	}
	p.advance()
	return []*Node{{Type: "TERMINAL", Value: tok.Lexeme, Pos: tok.Pos}}, nil
}

func (p *Parser) parseReference(ref *grammar.Reference) ([]*Node, error) {
	node, err := p.parseProduction(ref.Name)
	if err != nil {
		return nil, err
	}
	return []*Node{node}, nil
}

func (p *Parser) parseTokenRef(ref *grammar.TokenRef) ([]*Node, error) {
	if p.pos >= len(p.tokens) {
		return nil, p.error("unexpected end of input, expected %s", ref.Name)
	}
	tok := p.tokens[p.pos]
	if !p.matchesToken(tok, ref.Name) {
		return nil, p.error("expected %s, got %s", ref.Name, tok.Type)
	}
	p.advance()
	return []*Node{{Type: ref.Name, Value: tok.Lexeme, Pos: tok.Pos}}, nil
}

// advance moves to next token and tracks furthest position reached.
func (p *Parser) advance() {
	p.pos++
	if p.pos > p.furthest {
		p.furthest = p.pos
	}
}

func (p *Parser) matchesToken(tok Token, refName string) bool {
	if tok.Type == refName {
		return true
	}
	// KEYWORD matches any keyword token
	if refName == "KEYWORD" && tok.Type == "KEYWORD" {
		return true
	}
	return false
}

func (p *Parser) peek() Token {
	if p.pos >= len(p.tokens) {
		return Token{Type: "EOF"}
	}
	return p.tokens[p.pos]
}

func (p *Parser) error(format string, args ...interface{}) error {
	tok := p.peek()
	msg := fmt.Sprintf(format, args...)
	return fmt.Errorf("%s", engine.FormatWithCaret(tok.Pos, engine.GetSourceLine(p.source, tok.Pos.Line), msg))
}
