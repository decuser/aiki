package grammar

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"aiki/engine"
)

// Parser parses EBNFX grammar files.
type Parser struct {
	source string
	file   string
	pos    int
	line   int
	col    int
}

// NewParser creates a parser for the given source.
func NewParser(file, source string) *Parser {
	return &Parser{
		source: source,
		file:   file,
		pos:    0,
		line:   1,
		col:    1,
	}
}

// Parse parses the EBNFX source into a Grammar.
func (p *Parser) Parse() (*Grammar, error) {
	g := &Grammar{
		Productions: make(map[string]Production),
	}

	p.skipWhitespaceAndComments()

	// Parse @tokens block if present
	if p.peek() == '@' && p.lookingAt("@tokens") {
		tokens, err := p.parseTokensBlock()
		if err != nil {
			return nil, err
		}
		g.Tokens = tokens
		p.skipWhitespaceAndComments()
	}

	// Parse productions
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

	return g, nil
}

// parseTokensBlock parses @tokens { ... }
func (p *Parser) parseTokensBlock() ([]TokenDef, error) {
	p.expect("@tokens")
	p.skipWhitespaceAndComments()
	if err := p.expectChar('{'); err != nil {
		return nil, err
	}
	p.skipWhitespaceAndComments()

	var tokens []TokenDef

	for p.peek() != '}' && p.pos < len(p.source) {
		tok, err := p.parseTokenDef()
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, tok)
		p.skipWhitespaceAndComments()
	}

	if err := p.expectChar('}'); err != nil {
		return nil, err
	}

	return tokens, nil
}

// parseTokenDef parses a single token definition.
func (p *Parser) parseTokenDef() (TokenDef, error) {
	startPos := p.position()

	name := p.parseName()
	if name == "" {
		return TokenDef{}, p.error("expected token name")
	}

	p.skipInlineWhitespace()

	tok := TokenDef{
		Name: name,
		Pos:  startPos,
	}

	// Check for pattern or keyword list
	if p.peek() == '/' {
		pattern, err := p.parseRegex()
		if err != nil {
			return TokenDef{}, err
		}
		re, err := regexp.Compile("^" + pattern)
		if err != nil {
			return TokenDef{}, p.error("invalid regex: %s", err)
		}
		tok.Pattern = re
	} else {
		// Keyword/operator/delimiter list - no pattern, just literals
		// These are handled specially by the lexer
		p.skipToEndOfLine()
		return tok, nil
	}

	// Parse @skip and decorations
	p.skipInlineWhitespace()
	for p.peek() == '@' {
		decorator, value, err := p.parseDecorator()
		if err != nil {
			return TokenDef{}, err
		}
		switch decorator {
		case "skip":
			tok.Skip = true
		case "error":
			tok.Meta.Error = value
		case "template":
			tok.Meta.Template = value
		case "help":
			tok.Meta.Help = value
		}
		p.skipInlineWhitespace()
	}

	// Check for continued decorations on next lines
	for {
		savedPos, savedLine, savedCol := p.pos, p.line, p.col
		p.skipWhitespaceAndComments()
		if p.peek() == '@' {
			decorator, value, err := p.parseDecorator()
			if err != nil {
				return TokenDef{}, err
			}
			switch decorator {
			case "skip":
				tok.Skip = true
			case "error":
				tok.Meta.Error = value
			case "template":
				tok.Meta.Template = value
			case "help":
				tok.Meta.Help = value
			default:
				// Unknown decorator, restore position
				p.pos, p.line, p.col = savedPos, savedLine, savedCol
				break
			}
		} else if p.pos < len(p.source) && !unicode.IsUpper(rune(p.peek())) && p.peek() != '}' {
			// Next token definition or end
			p.pos, p.line, p.col = savedPos, savedLine, savedCol
			break
		} else {
			p.pos, p.line, p.col = savedPos, savedLine, savedCol
			break
		}
	}

	return tok, nil
}

// parseProduction parses a production rule.
func (p *Parser) parseProduction() (Production, error) {
	startPos := p.position()

	name := p.parseName()
	if name == "" {
		return Production{}, p.error("expected production name")
	}

	p.skipWhitespaceAndComments()
	if err := p.expectChar('='); err != nil {
		return Production{}, err
	}
	p.skipWhitespaceAndComments()

	prod := Production{
		Name: name,
		Pos:  startPos,
	}

	// Parse expressions (alternatives)
	for {
		expr, err := p.parseExpression()
		if err != nil {
			return Production{}, err
		}
		prod.Expressions = append(prod.Expressions, expr)

		p.skipWhitespaceAndComments()
		if p.peek() == '|' && !p.lookingAt("|>") {
			p.advance() // consume |
			p.skipWhitespaceAndComments()
		} else {
			break
		}
	}

	// Parse decorations
	p.skipWhitespaceAndComments()
	for p.peek() == '@' {
		decorator, value, err := p.parseDecorator()
		if err != nil {
			return Production{}, err
		}
		switch decorator {
		case "error":
			prod.Meta.Error = value
		case "template":
			prod.Meta.Template = value
		case "help":
			prod.Meta.Help = value
		}
		p.skipWhitespaceAndComments()
	}

	return prod, nil
}

// parseExpression parses a sequence of terms.
func (p *Parser) parseExpression() (Expression, error) {
	var expr Expression

	for {
		p.skipInlineWhitespace()
		ch := p.peek()

		// End of expression
		if ch == 0 || ch == '|' || ch == '@' || ch == '\n' {
			break
		}
		// Also end on these unless in group
		if ch == ')' || ch == ']' || ch == '}' {
			break
		}

		term, err := p.parseTerm()
		if err != nil {
			return Expression{}, err
		}
		if term.Value != "" || term.Kind != TermLiteral {
			expr.Terms = append(expr.Terms, term)
		} else {
			break
		}
	}

	return expr, nil
}

// parseTerm parses a single term.
func (p *Parser) parseTerm() (Term, error) {
	startPos := p.position()
	ch := p.peek()

	var term Term
	term.Pos = startPos

	switch {
	case ch == '"':
		// Literal string
		lit, err := p.parseString()
		if err != nil {
			return Term{}, err
		}
		term.Value = lit
		term.Kind = TermLiteral

	case ch == '[':
		// Optional group - may contain alternatives
		p.advance()
		p.skipWhitespaceAndComments()
		
		var alts []Expression
		for {
			expr, err := p.parseExpression()
			if err != nil {
				return Term{}, err
			}
			alts = append(alts, expr)
			p.skipWhitespaceAndComments()
			if p.peek() == '|' && !p.lookingAt("|>") {
				p.advance() // consume |
				p.skipWhitespaceAndComments()
			} else {
				break
			}
		}
		
		p.skipWhitespaceAndComments()
		if err := p.expectChar(']'); err != nil {
			return Term{}, err
		}
		
		if len(alts) == 1 && len(alts[0].Terms) == 1 {
			term = alts[0].Terms[0]
			term.Optional = true
		} else {
			term.Value = "[group]"
			term.Kind = TermProduction
			term.Optional = true
		}

	case ch == '{':
		// Repeat group - may contain alternatives
		p.advance()
		p.skipWhitespaceAndComments()
		
		var alts []Expression
		for {
			expr, err := p.parseExpression()
			if err != nil {
				return Term{}, err
			}
			alts = append(alts, expr)
			p.skipWhitespaceAndComments()
			if p.peek() == '|' && !p.lookingAt("|>") {
				p.advance() // consume |
				p.skipWhitespaceAndComments()
			} else {
				break
			}
		}
		
		p.skipWhitespaceAndComments()
		if err := p.expectChar('}'); err != nil {
			return Term{}, err
		}
		
		if len(alts) == 1 && len(alts[0].Terms) == 1 {
			term = alts[0].Terms[0]
			term.Repeat = true
		} else {
			term.Value = "{group}"
			term.Kind = TermProduction
			term.Repeat = true
		}

	case ch == '(':
		// Grouping - may contain alternatives
		p.advance()
		p.skipWhitespaceAndComments()
		
		var alts []Expression
		for {
			expr, err := p.parseExpression()
			if err != nil {
				return Term{}, err
			}
			alts = append(alts, expr)
			p.skipWhitespaceAndComments()
			if p.peek() == '|' && !p.lookingAt("|>") {
				p.advance() // consume |
				p.skipWhitespaceAndComments()
			} else {
				break
			}
		}
		
		p.skipWhitespaceAndComments()
		if err := p.expectChar(')'); err != nil {
			return Term{}, err
		}
		
		if len(alts) == 1 && len(alts[0].Terms) == 1 {
			term = alts[0].Terms[0]
		} else {
			term.Value = "(group)"
			term.Kind = TermProduction
		}

	case unicode.IsUpper(rune(ch)):
		// Token reference
		name := p.parseName()
		term.Value = name
		term.Kind = TermToken

	case unicode.IsLower(rune(ch)) || ch == '_':
		// Production reference
		name := p.parseName()
		term.Value = name
		term.Kind = TermProduction

	default:
		return Term{}, nil
	}

	return term, nil
}

// parseDecorator parses @name or @name "value"
func (p *Parser) parseDecorator() (string, string, error) {
	if err := p.expectChar('@'); err != nil {
		return "", "", err
	}

	name := p.parseName()
	if name == "" {
		return "", "", p.error("expected decorator name")
	}

	p.skipInlineWhitespace()

	var value string
	if p.peek() == '"' {
		var err error
		value, err = p.parseString()
		if err != nil {
			return "", "", err
		}
	}

	return name, value, nil
}

// parseString parses a quoted string.
func (p *Parser) parseString() (string, error) {
	if err := p.expectChar('"'); err != nil {
		return "", err
	}

	var sb strings.Builder
	for p.pos < len(p.source) {
		ch := p.source[p.pos]
		if ch == '"' {
			p.advance()
			return sb.String(), nil
		}
		if ch == '\\' && p.pos+1 < len(p.source) {
			p.advance()
			next := p.source[p.pos]
			switch next {
			case 'n':
				sb.WriteByte('\n')
			case 't':
				sb.WriteByte('\t')
			case '"':
				sb.WriteByte('"')
			case '\\':
				sb.WriteByte('\\')
			default:
				sb.WriteByte(next)
			}
			p.advance()
		} else {
			sb.WriteByte(ch)
			p.advance()
		}
	}

	return "", p.error("unterminated string")
}

// parseRegex parses /pattern/
func (p *Parser) parseRegex() (string, error) {
	if err := p.expectChar('/'); err != nil {
		return "", err
	}

	var sb strings.Builder
	for p.pos < len(p.source) {
		ch := p.source[p.pos]
		if ch == '/' {
			p.advance()
			return sb.String(), nil
		}
		if ch == '\\' && p.pos+1 < len(p.source) {
			sb.WriteByte(ch)
			p.advance()
			sb.WriteByte(p.source[p.pos])
			p.advance()
		} else {
			sb.WriteByte(ch)
			p.advance()
		}
	}

	return "", p.error("unterminated regex")
}

// parseName parses an identifier.
func (p *Parser) parseName() string {
	start := p.pos
	for p.pos < len(p.source) {
		ch := p.source[p.pos]
		if unicode.IsLetter(rune(ch)) || unicode.IsDigit(rune(ch)) || ch == '_' {
			p.advance()
		} else {
			break
		}
	}
	return p.source[start:p.pos]
}

// Helper methods

func (p *Parser) peek() byte {
	if p.pos >= len(p.source) {
		return 0
	}
	return p.source[p.pos]
}

func (p *Parser) advance() {
	if p.pos < len(p.source) {
		if p.source[p.pos] == '\n' {
			p.line++
			p.col = 1
		} else {
			p.col++
		}
		p.pos++
	}
}

func (p *Parser) lookingAt(s string) bool {
	return strings.HasPrefix(p.source[p.pos:], s)
}

func (p *Parser) expect(s string) error {
	if !p.lookingAt(s) {
		return p.error("expected '%s'", s)
	}
	for range s {
		p.advance()
	}
	return nil
}

func (p *Parser) expectChar(ch byte) error {
	if p.peek() != ch {
		return p.error("expected '%c'", ch)
	}
	p.advance()
	return nil
}

func (p *Parser) skipWhitespaceAndComments() {
	for p.pos < len(p.source) {
		ch := p.source[p.pos]
		if ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n' {
			p.advance()
		} else if ch == '#' {
			// Skip comment to end of line
			for p.pos < len(p.source) && p.source[p.pos] != '\n' {
				p.advance()
			}
		} else {
			break
		}
	}
}

func (p *Parser) skipInlineWhitespace() {
	for p.pos < len(p.source) {
		ch := p.source[p.pos]
		if ch == ' ' || ch == '\t' {
			p.advance()
		} else {
			break
		}
	}
}

func (p *Parser) skipToEndOfLine() {
	for p.pos < len(p.source) && p.source[p.pos] != '\n' {
		p.advance()
	}
}

func (p *Parser) position() engine.Position {
	return engine.Position{
		File: p.file,
		Line: p.line,
		Col:  p.col,
	}
}

func (p *Parser) error(format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	return fmt.Errorf("%s:%d:%d: %s", p.file, p.line, p.col, msg)
}
