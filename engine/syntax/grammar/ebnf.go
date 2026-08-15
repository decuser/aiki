package grammar

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"aiki/engine"
)

type ebnfParser struct {
	source string
	file   string
	pos    int
	line   int
	col    int
}

func NewParser(file, source string) *ebnfParser {
	return &ebnfParser{source: source, file: file, pos: 0, line: 1, col: 1}
}

func (p *ebnfParser) Parse() (*Grammar, error) {
	g := &Grammar{Productions: make(map[string]*Production)}
	p.skipWhitespaceAndComments()

	if p.peek() == '@' && p.lookingAt("@tokens") {
		tokens, err := p.parseTokensBlock()
		if err != nil {
			return nil, err
		}
		g.Tokens = tokens
		p.skipWhitespaceAndComments()
	}

	if p.peek() == '@' && p.lookingAt("@newline") {
		rule, err := p.parseNewlineBlock()
		if err != nil {
			return nil, err
		}
		g.Newline = rule
		p.skipWhitespaceAndComments()
	}
	if g.Newline == nil {
		return nil, p.error("missing @newline declaration")
	}
	if err := validateNewlineDeclaration(g); err != nil {
		return nil, err
	}

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

	for _, prod := range g.Productions {
		prod.Expr = resolveRefs(prod.Expr, g.Productions)
	}
	g.Reanalyze()
	return g, nil
}

func resolveRefs(expr Expression, prods map[string]*Production) Expression {
	switch e := expr.(type) {
	case *Sequence:
		for i, sub := range e.Exprs {
			e.Exprs[i] = resolveRefs(sub, prods)
		}
	case *Alternative:
		for i, sub := range e.Exprs {
			e.Exprs[i] = resolveRefs(sub, prods)
		}
	case *Repetition:
		e.Expr = resolveRefs(e.Expr, prods)
	case *Option:
		e.Expr = resolveRefs(e.Expr, prods)
	case *Group:
		e.Expr = resolveRefs(e.Expr, prods)
	case *TokenRef:
		if _, ok := prods[e.Name]; ok {
			return &Reference{Name: e.Name}
		}
	}
	return expr
}

func (p *ebnfParser) parseTokensBlock() ([]TokenDef, error) {
	if !p.consumeString("@tokens") {
		return nil, p.error("expected @tokens")
	}
	p.skipWhitespaceAndComments()
	if !p.consume('{') {
		return nil, p.error("expected '{'")
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
		return nil, p.error("expected '}'")
	}
	return tokens, nil
}

func (p *ebnfParser) parseNewlineBlock() (*NewlineRule, error) {
	startPos := p.position()
	if !p.consumeString("@newline") {
		return nil, p.error("expected @newline")
	}
	p.skipWhitespaceAndComments()
	if !p.consume('{') {
		return nil, p.error("expected '{'")
	}

	rule := &NewlineRule{Pos: startPos}
	p.skipWhitespaceAndComments()
	for p.peek() != '}' && p.pos < len(p.source) {
		if p.peek() == '@' {
			if !p.lookingAt("@help") {
				return nil, p.error("unknown @newline decorator")
			}
			name, value := p.parseDecorator()
			if name != "help" {
				return nil, p.error("expected @help")
			}
			rule.Meta.Help = value
			p.skipWhitespaceAndComments()
			continue
		}

		directivePos := p.position()
		name := p.parseIdentifier()
		if name == "" {
			return nil, p.error("expected @newline directive")
		}
		p.skipSpaces()
		args := strings.Fields(strings.TrimSpace(p.parseUntilNewlineOrAt()))
		switch name {
		case "token":
			if len(args) != 1 {
				return nil, fmt.Errorf("%s:%d:%d: @newline token expects exactly one token name", directivePos.File, directivePos.Line, directivePos.Col)
			}
			if rule.Token != "" {
				return nil, fmt.Errorf("%s:%d:%d: duplicate @newline token directive", directivePos.File, directivePos.Line, directivePos.Col)
			}
			rule.Token = args[0]
		case "after_token":
			if len(args) == 0 {
				return nil, fmt.Errorf("%s:%d:%d: @newline after_token requires at least one token name", directivePos.File, directivePos.Line, directivePos.Col)
			}
			rule.AfterToken = append(rule.AfterToken, args...)
		case "after_lexeme":
			if len(args) == 0 {
				return nil, fmt.Errorf("%s:%d:%d: @newline after_lexeme requires at least one lexeme", directivePos.File, directivePos.Line, directivePos.Col)
			}
			rule.AfterLexeme = append(rule.AfterLexeme, args...)
		case "suppress_in":
			if len(args) != 2 {
				return nil, fmt.Errorf("%s:%d:%d: @newline suppress_in expects an opener and closer", directivePos.File, directivePos.Line, directivePos.Col)
			}
			rule.SuppressIn = append(rule.SuppressIn, [2]string{args[0], args[1]})
		default:
			return nil, fmt.Errorf("%s:%d:%d: unknown @newline directive %q", directivePos.File, directivePos.Line, directivePos.Col, name)
		}
		p.skipWhitespaceAndComments()
	}
	if !p.consume('}') {
		return nil, p.error("expected '}'")
	}
	return rule, nil
}

func validateNewlineDeclaration(g *Grammar) error {
	r := g.Newline
	if r == nil {
		return fmt.Errorf("missing @newline declaration")
	}
	if r.Token == "" {
		return fmt.Errorf("%s:%d:%d: @newline missing token directive", r.Pos.File, r.Pos.Line, r.Pos.Col)
	}

	tokenNames := make(map[string]bool, len(g.Tokens))
	literalLexemes := make(map[string]bool)
	for _, tok := range g.Tokens {
		tokenNames[tok.Name] = true
		for _, lexeme := range strings.Fields(tok.Literal) {
			literalLexemes[lexeme] = true
		}
	}
	if !tokenNames[r.Token] {
		return fmt.Errorf("%s:%d:%d: @newline token %q is not declared in @tokens", r.Pos.File, r.Pos.Line, r.Pos.Col, r.Token)
	}

	seenTokens := make(map[string]bool)
	for _, name := range r.AfterToken {
		if !tokenNames[name] {
			return fmt.Errorf("%s:%d:%d: @newline after_token %q is not declared in @tokens", r.Pos.File, r.Pos.Line, r.Pos.Col, name)
		}
		if seenTokens[name] {
			return fmt.Errorf("%s:%d:%d: duplicate @newline after_token %q", r.Pos.File, r.Pos.Line, r.Pos.Col, name)
		}
		seenTokens[name] = true
	}

	seenLexemes := make(map[string]bool)
	for _, lexeme := range r.AfterLexeme {
		if !literalLexemes[lexeme] {
			return fmt.Errorf("%s:%d:%d: @newline after_lexeme %q is not declared by a literal token", r.Pos.File, r.Pos.Line, r.Pos.Col, lexeme)
		}
		if seenLexemes[lexeme] {
			return fmt.Errorf("%s:%d:%d: duplicate @newline after_lexeme %q", r.Pos.File, r.Pos.Line, r.Pos.Col, lexeme)
		}
		seenLexemes[lexeme] = true
	}

	seenPairs := make(map[[2]string]bool)
	openers := make(map[string]string)
	closers := make(map[string]string)
	for _, pair := range r.SuppressIn {
		if !literalLexemes[pair[0]] || !literalLexemes[pair[1]] {
			return fmt.Errorf("%s:%d:%d: @newline suppress_in %q %q must name literal lexemes", r.Pos.File, r.Pos.Line, r.Pos.Col, pair[0], pair[1])
		}
		if pair[0] == pair[1] {
			return fmt.Errorf("%s:%d:%d: @newline suppress_in opener and closer must differ: %q", r.Pos.File, r.Pos.Line, r.Pos.Col, pair[0])
		}
		if seenPairs[pair] {
			return fmt.Errorf("%s:%d:%d: duplicate @newline suppress_in %q %q", r.Pos.File, r.Pos.Line, r.Pos.Col, pair[0], pair[1])
		}
		if prior, ok := openers[pair[0]]; ok && prior != pair[1] {
			return fmt.Errorf("%s:%d:%d: @newline opener %q has conflicting closers %q and %q", r.Pos.File, r.Pos.Line, r.Pos.Col, pair[0], prior, pair[1])
		}
		if prior, ok := closers[pair[1]]; ok && prior != pair[0] {
			return fmt.Errorf("%s:%d:%d: @newline closer %q has conflicting openers %q and %q", r.Pos.File, r.Pos.Line, r.Pos.Col, pair[1], prior, pair[0])
		}
		seenPairs[pair] = true
		openers[pair[0]] = pair[1]
		closers[pair[1]] = pair[0]
	}
	if len(r.AfterToken) == 0 && len(r.AfterLexeme) == 0 {
		return fmt.Errorf("%s:%d:%d: @newline requires after_token or after_lexeme", r.Pos.File, r.Pos.Line, r.Pos.Col)
	}
	return nil
}

func (p *ebnfParser) parseTokenDef() (TokenDef, error) {
	startPos := p.position()
	name := p.parseIdentifier()
	if name == "" {
		return TokenDef{}, p.error("expected token name")
	}
	p.skipSpaces()

	def := TokenDef{Name: name, Pos: startPos}

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
		def.Literal = strings.TrimSpace(p.parseUntilNewlineOrAt())
	}

	// Parse decorations (may span multiple lines)
	for {
		p.skipWhitespaceAndComments()
		if p.peek() != '@' {
			break
		}
		// Check if this is a new token or production (uppercase after @)
		if p.lookingAt("@tokens") || p.lookingAt("@skip") || p.lookingAt("@error") || p.lookingAt("@template") || p.lookingAt("@help") {
			if p.lookingAt("@tokens") {
				break // new block
			}
			dname, value := p.parseDecorator()
			switch dname {
			case "skip":
				def.Skip = true
			case "error":
				def.Meta.Error = value
			case "template":
				def.Meta.Template = value
			case "help":
				def.Meta.Help = value
			}
		} else {
			break
		}
	}
	return def, nil
}

func (p *ebnfParser) parseProduction() (*Production, error) {
	startPos := p.position()
	name := p.parseIdentifier()
	if name == "" {
		return nil, p.error("expected production name")
	}
	p.skipWhitespaceAndComments()
	if !p.consume('=') {
		return nil, p.error("expected '='")
	}
	p.skipWhitespaceAndComments()

	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	prod := &Production{Name: name, Expr: expr, Pos: startPos}

	// Parse decorations (may span multiple lines)
	for {
		p.skipWhitespaceAndComments()
		if p.peek() != '@' {
			break
		}
		if !p.lookingAt("@error") && !p.lookingAt("@template") && !p.lookingAt("@help") {
			break
		}
		dname, value := p.parseDecorator()
		switch dname {
		case "error":
			prod.Meta.Error = value
		case "template":
			prod.Meta.Template = value
		case "help":
			prod.Meta.Help = value
		}
	}
	return prod, nil
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
		if p.peek() != '|' || p.lookingAt("|>") {
			break
		}
		p.consume('|')
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

	if ch == 0 || ch == '|' || ch == ')' || ch == ']' || ch == '}' || ch == '@' {
		return nil, nil
	}

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

	if ch == '"' {
		p.consume('"')
		start := p.pos
		for p.pos < len(p.source) && p.source[p.pos] != '"' {
			p.advance()
		}
		value := p.source[start:p.pos]
		if !p.consume('"') {
			return nil, p.error("unterminated string")
		}
		return &Terminal{Value: value}, nil
	}

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
	for _, r := range name {
		if !unicode.IsUpper(r) && r != '_' {
			return false
		}
	}
	return len(name) > 0
}

func (p *ebnfParser) parseDecorator() (string, string) {
	if !p.consume('@') {
		return "", ""
	}
	name := p.parseIdentifier()
	p.skipSpaces()
	var value string
	if p.peek() == '"' {
		p.consume('"')
		start := p.pos
		for p.pos < len(p.source) && p.source[p.pos] != '"' {
			if p.source[p.pos] == '\\' && p.pos+1 < len(p.source) {
				p.pos += 2 // skip escape sequence
			} else {
				p.pos++
			}
		}
		value = p.source[start:p.pos]
		p.consume('"')
	}
	return name, value
}

func (p *ebnfParser) parseRegex() (string, error) {
	if !p.consume('/') {
		return "", p.error("expected '/'")
	}
	start := p.pos
	for p.pos < len(p.source) && p.source[p.pos] != '/' {
		if p.source[p.pos] == '\\' && p.pos+1 < len(p.source) {
			p.pos += 2
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

func (p *ebnfParser) parseUntilNewlineOrAt() string {
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

func (p *ebnfParser) advance() {
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

func (p *ebnfParser) consume(ch byte) bool {
	if p.peek() == ch {
		p.advance()
		return true
	}
	return false
}

func (p *ebnfParser) consumeString(s string) bool {
	if p.lookingAt(s) {
		for range s {
			p.advance()
		}
		return true
	}
	return false
}

func (p *ebnfParser) lookingAt(s string) bool {
	return strings.HasPrefix(p.source[p.pos:], s)
}

func (p *ebnfParser) skipSpaces() {
	for p.pos < len(p.source) {
		ch := p.source[p.pos]
		if ch == ' ' || ch == '\t' {
			p.advance()
		} else {
			break
		}
	}
}

func (p *ebnfParser) skipWhitespaceAndComments() {
	for p.pos < len(p.source) {
		ch := p.source[p.pos]
		if ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n' {
			p.advance()
		} else if ch == '#' {
			for p.pos < len(p.source) && p.source[p.pos] != '\n' {
				p.advance()
			}
		} else {
			break
		}
	}
}

func (p *ebnfParser) position() engine.Position {
	return engine.Position{File: p.file, Line: p.line, Col: p.col}
}

func (p *ebnfParser) error(msg string) error {
	return fmt.Errorf("%s:%d:%d: %s", p.file, p.line, p.col, msg)
}
