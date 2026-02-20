// engine/syntax/parser.go
package syntax

import "fmt"

type Parser struct {
	lexer    *Lexer
	contract GrammarContract
	look     Token
}

type marker struct {
	pos  int
	line int
	col  int
	look Token
}

func NewParser(l *Lexer, c GrammarContract) (*Parser, error) {
	p := &Parser{lexer: l, contract: c}
	if err := p.advance(); err != nil {
		return nil, err
	}
	return p, nil
}

func (p *Parser) Parse() (*Node, error) {
	node, err := p.matchProduction(p.contract.GetStart(), 0)
	if err != nil {
		return nil, err
	}
	if p.look.Type != "EOF" {
		return nil, fmt.Errorf("%s:%d:%d: unexpected %s '%s' after %s",
			p.look.Pos.File, p.look.Pos.Line, p.look.Pos.Col, p.look.Type, p.look.Lexeme, p.contract.GetStart())
	}
	return node, nil
}

func (p *Parser) matchProduction(name string, depth int) (*Node, error) {
	prod, ok := p.contract.GetProduction(name)
	if !ok {
		return nil, fmt.Errorf("unknown production: %s", name)
	}

	p.contract.Observe().OnParse(name, depth, p.look.Pos)

	for _, expr := range prod.Expressions {
		m := p.mark()
		node := &Node{Type: name, Pos: p.look.Pos}
		success := true

		for _, term := range expr {
			if !p.matchTerm(term, node, depth) {
				success = false
				break
			}
		}

		if success {
			return node, nil
		}
		p.reset(m)
	}

	return nil, fmt.Errorf("rule '%s' failed", name)
}

func (p *Parser) matchTerm(term Term, parent *Node, depth int) bool {
	matchOnce := func() bool {
		if term.IsSymbol {
			child, err := p.matchProduction(term.Value, depth+1)
			if err != nil {
				return false
			}
			parent.Children = append(parent.Children, child)
			return true
		}
		if p.look.Lexeme == term.Value || p.look.Type == term.Value {
			parent.Children = append(parent.Children, &Node{
				Type:  p.look.Type,
				Value: p.look.Lexeme,
				Pos:   p.look.Pos,
			})
			p.advance()
			return true
		}
		return false
	}

	if term.IsRepeat {
		for matchOnce() {
		}
		return true
	}
	if matchOnce() {
		return true
	}
	return term.IsOption
}

func (p *Parser) advance() error {
	tok, err := p.lexer.NextToken()
	if err != nil {
		return err
	}
	p.look = tok
	return nil
}

func (p *Parser) mark() marker {
	return marker{p.lexer.pos, p.lexer.line, p.lexer.col, p.look}
}

func (p *Parser) reset(m marker) {
	p.lexer.pos, p.lexer.line, p.lexer.col, p.look = m.pos, m.line, m.col, m.look
}
