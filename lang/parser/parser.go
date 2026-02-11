package parser

import (
	"fmt"

	"aiki/lang/ast"
	"aiki/lang/lexer"
	"aiki/lang/token"
)

type Parser struct {
	tokens   []token.Token
	pos      int
	current  token.Token
	peek     token.Token
	errors   []string
	comments []lexer.Comment
}

func New(input string) *Parser {
	l := lexer.New(input)
	tokens := l.Tokenize()

	p := &Parser{
		tokens:   tokens,
		comments: l.Comments,
	}
	p.advance()
	p.advance()
	return p
}

func (p *Parser) advance() {
	p.current = p.peek
	if p.pos < len(p.tokens) {
		p.peek = p.tokens[p.pos]
		p.pos++
	}
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) Comments() []lexer.Comment {
	return p.comments
}

func (p *Parser) error(msg string) {
	p.errors = append(p.errors, fmt.Sprintf("line %d: %s", p.current.Line, msg))
}

func (p *Parser) Parse() *ast.Program {
	program := &ast.Program{}

	for p.current.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
	}

	return program
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.current.Type {
	case token.Let:
		return p.parseLetStatement()
	case token.Return:
		return p.parseReturnStatement()
	case token.If:
		return p.parseIfStatement()
	case token.While:
		return p.parseWhileStatement()
	case token.Match:
		return p.parseMatchStatement()
	case token.Export:
		return p.parseExportStatement()
	case token.From:
		return p.parseImportStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseLetStatement() ast.Statement {
	tok := p.current
	p.advance()

	// Shape definition: let @name [fields]
	if p.current.Type == token.Shape {
		return p.parseShapeStatement(tok)
	}

	// Normal binding: let name = expr
	if p.current.Type != token.Name {
		p.error("expected name after let")
		return nil
	}

	name := &ast.Identifier{Token: p.current, Value: p.current.Lexeme}
	p.advance()

	if p.current.Type != token.Assign {
		p.error("expected = after name")
		return nil
	}
	p.advance()

	value := p.parseExpression()
	if value == nil {
		return nil
	}

	return &ast.LetStatement{Token: tok, Name: name, Value: value}
}

func (p *Parser) parseShapeStatement(tok token.Token) ast.Statement {
	shapeName := p.current.Lexeme[1:] // strip @
	p.advance()

	if p.current.Type != token.LBracket {
		p.error("expected [ after shape name")
		return nil
	}
	p.advance()

	var fields []string
	var embeds []string

	for p.current.Type != token.RBracket && p.current.Type != token.EOF {
		if p.current.Type == token.Shape {
			embeds = append(embeds, p.current.Lexeme[1:])
		} else if p.current.Type == token.Name {
			fields = append(fields, p.current.Lexeme)
		} else {
			p.error("expected field name or embedded shape")
			return nil
		}
		p.advance()

		// Optional comma
		if p.current.Type == token.Comma {
			p.advance()
		}
	}

	if p.current.Type != token.RBracket {
		p.error("expected ]")
		return nil
	}
	p.advance()
	return &ast.ShapeStatement{
		Token:  tok,
		Name:   shapeName,
		Fields: fields,
		Embeds: embeds,
	}
}

func (p *Parser) parseReturnStatement() ast.Statement {
	tok := p.current
	p.advance()

	value := p.parseExpression()
	return &ast.ReturnStatement{Token: tok, Value: value}
}

func (p *Parser) parseIfStatement() ast.Statement {
	tok := p.current
	p.advance()

	condition := p.parseExpression()
	if condition == nil {
		return nil
	}

	if p.current.Type != token.LBrace {
		p.error("expected { after if condition")
		return nil
	}
	consequence := p.parseBlockStatement()

	var alternative *ast.BlockStatement
	if p.current.Type == token.Else {
		p.advance() // consume 'else'
		if p.current.Type != token.LBrace {
			p.error("expected { after else")
			return nil
		}
		alternative = p.parseBlockStatement()
	}

	return &ast.IfStatement{
		Token:       tok,
		Condition:   condition,
		Consequence: consequence,
		Alternative: alternative,
	}
}

func (p *Parser) parseWhileStatement() ast.Statement {
	tok := p.current
	p.advance()

	condition := p.parseExpression()
	if condition == nil {
		return nil
	}

	if p.current.Type != token.LBrace {
		p.error("expected { after while condition")
		return nil
	}
	body := p.parseBlockStatement()

	return &ast.WhileStatement{
		Token:     tok,
		Condition: condition,
		Body:      body,
	}
}

func (p *Parser) parseMatchStatement() ast.Statement {
	tok := p.current
	p.advance()

	value := p.parseExpression()
	if value == nil {
		return nil
	}

	if p.current.Type != token.LBrace {
		p.error("expected { after match value")
		return nil
	}
	p.advance()

	var arms []*ast.MatchArm
	for p.current.Type != token.RBrace && p.current.Type != token.EOF {
		arm := p.parseMatchArm()
		if arm != nil {
			arms = append(arms, arm)
		}
	}

	p.advance()

	return &ast.MatchStatement{Token: tok, Value: value, Arms: arms}
}

func (p *Parser) parseMatchArm() *ast.MatchArm {
	pattern := p.parsePattern()
	if pattern == nil {
		return nil
	}

	if p.current.Type != token.LBrace {
		p.error("expected { after pattern")
		return nil
	}
	body := p.parseBlockStatement()

	return &ast.MatchArm{Pattern: pattern, Body: body}
}

func (p *Parser) parsePattern() ast.Pattern {
	switch p.current.Type {
	case token.Name:
		if p.current.Lexeme == "_" {
			pat := &ast.WildcardPattern{Token: p.current}
			p.advance()
			return pat
		}
		pat := &ast.NamePattern{Token: p.current, Name: p.current.Lexeme}
		p.advance()
		return pat
	case token.LBracket:
		return p.parseListPattern()
	default:
		// Literal pattern
		expr := p.parseExpression()
		if expr == nil {
			return nil
		}
		return &ast.LiteralPattern{Token: p.current, Value: expr}
	}
}

func (p *Parser) parseListPattern() ast.Pattern {
	tok := p.current
	p.advance()

	// Check for shaped list pattern: [@name ...]
	if p.current.Type == token.Shape {
		shapeName := p.current.Lexeme[1:]
		p.advance()

		var elements []ast.Pattern
		for p.current.Type != token.RBracket && p.current.Type != token.EOF {
			// Skip leading comma after shape name
			if p.current.Type == token.Comma {
				p.advance()
				continue
			}
			pat := p.parsePattern()
			if pat != nil {
				elements = append(elements, pat)
			}
			// Optional comma between elements
			if p.current.Type == token.Comma {
				p.advance()
			}
		}

		if p.current.Type != token.RBracket {
			p.error("expected ]")
			return nil
		}
		p.advance()

		return &ast.ShapedListPattern{Token: tok, Shape: shapeName, Elements: elements}
	}

	// Raw list pattern
	var elements []ast.Pattern
	for p.current.Type != token.RBracket && p.current.Type != token.EOF {
		pat := p.parsePattern()
		if pat != nil {
			elements = append(elements, pat)
		}
		// Optional comma between elements
		if p.current.Type == token.Comma {
			p.advance()
		}
	}

	if p.current.Type != token.RBracket {
		p.error("expected ]")
		return nil
	}
	p.advance()

	return &ast.ListPattern{Token: tok, Elements: elements}
}

func (p *Parser) parseExportStatement() ast.Statement {
	tok := p.current
	p.advance()
	if p.current.Type != token.LBracket {
		p.error("expected [ after export")
		return nil
	}
	p.advance()
	// fmt.Printf("DEBUG: type=%v lexeme=%q\n", p.current.Type, p.current.Lexeme)
	var names []string
	for p.current.Type == token.Name {
		// fmt.Printf("DEBUG name: %q\n", p.current.Lexeme)
		names = append(names, p.current.Lexeme)
		p.advance()
		// fmt.Printf("DEBUG after advance: type=%v lexeme=%q\n", p.current.Type, p.current.Lexeme)
		if p.current.Type == token.Comma {
			p.advance()
		}
	}
	// fmt.Printf("DEBUG exited loop: type=%v lexeme=%q\n", p.current.Type, p.current.Lexeme)

	if p.current.Type != token.RBracket {
		p.error("expected ]")
		return nil
	}
	p.advance()
	return &ast.ExportStatement{Token: tok, Names: names}
}

func (p *Parser) parseImportStatement() ast.Statement {
	tok := p.current
	p.advance()

	if p.current.Type != token.Name {
		p.error("expected module name after from")
		return nil
	}
	moduleName := p.current.Lexeme
	p.advance()

	if p.current.Type != token.Use {
		p.error("expected use after module name")
		return nil
	}
	p.advance()

	if p.current.Type != token.LBracket {
		p.error("expected [ after use")
		return nil
	}
	p.advance()

	var names []string
	for p.current.Type == token.Name {
		names = append(names, p.current.Lexeme)
		p.advance()
		if p.current.Type == token.Comma {
			p.advance()
		}
	}

	if p.current.Type != token.RBracket {
		p.error("expected ]")
		return nil
	}

	return &ast.ImportStatement{Token: tok, Module: moduleName, Names: names}
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.current}
	p.advance() // consume '{'

	for p.current.Type != token.RBrace && p.current.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
	}

	p.advance()
	return block
}

func (p *Parser) parseExpressionStatement() ast.Statement {
	tok := p.current

	// Check for assignment: name = expr
	if p.current.Type == token.Name && p.peek.Type == token.Assign {
		name := &ast.Identifier{Token: p.current, Value: p.current.Lexeme}
		p.advance()
		p.advance()
		value := p.parseExpression()
		return &ast.AssignStatement{Token: tok, Name: name, Value: value}
	}

	expr := p.parseExpression()
	if expr == nil {
		return nil
	}
	return &ast.ExpressionStatement{Token: tok, Expression: expr}
}

func (p *Parser) parseExpression() ast.Expression {
	return p.parsePipe()
}

func (p *Parser) parsePipe() ast.Expression {
	left := p.parseInfix()

	for p.current.Type == token.Pipe {
		tok := p.current
		p.advance()
		right := p.parseCall()
		if call, ok := right.(*ast.CallExpression); ok {
			left = &ast.PipeExpression{Token: tok, Left: left, Right: call}
		} else {
			p.error("expected function call after |>")
			return left
		}
	}

	return left
}

func (p *Parser) parseInfix() ast.Expression {
	left := p.parseUnary()

	for isInfixOp(p.current.Type) {
		tok := p.current
		op := tok.Lexeme
		p.advance()
		right := p.parseUnary()
		left = &ast.InfixExpression{Token: tok, Left: left, Operator: op, Right: right}
	}

	return left
}

func (p *Parser) parseUnary() ast.Expression {
	if p.current.Type == token.Not || p.current.Type == token.Minus {
		tok := p.current
		op := tok.Lexeme
		p.advance()
		right := p.parseUnary()
		return &ast.PrefixExpression{Token: tok, Operator: op, Right: right}
	}
	return p.parseCall()
}

func isInfixOp(t token.Type) bool {
	switch t {
	case token.Plus, token.Minus, token.Star, token.Slash, token.Percent,
		token.Eq, token.NotEq, token.Lt, token.Gt, token.LtEq, token.GtEq,
		token.And, token.Or:
		return true
	}
	return false
}

func (p *Parser) parseCall() ast.Expression {
	expr := p.parseAccess()

	for p.current.Type == token.LParen {
		tok := p.current
		p.advance()

		var args []ast.Expression
		if p.current.Type != token.RParen {
			// First argument
			arg := p.parseExpression()
			if arg != nil {
				args = append(args, arg)
			}
			// Remaining arguments (comma-separated)
			for p.current.Type == token.Comma {
				p.advance()
				arg := p.parseExpression()
				if arg != nil {
					args = append(args, arg)
				}
			}
		}

		if p.current.Type != token.RParen {
			p.error("expected )")
			return expr
		}
		p.advance()

		expr = &ast.CallExpression{Token: tok, Function: expr, Arguments: args}
	}

	return expr
}

func (p *Parser) parseAccess() ast.Expression {
	expr := p.parsePrimary()

	for p.current.Type == token.Dot {
		p.advance()
		if p.current.Type != token.Name && p.current.Type != token.Number {
			p.error("expected field name or index after .")
			return expr
		}
		key := p.current.Lexeme
		tok := p.current
		p.advance()
		expr = &ast.AccessExpression{Token: tok, Left: expr, Key: key}
	}

	return expr
}

func (p *Parser) parsePrimary() ast.Expression {
	switch p.current.Type {
	case token.Number:
		expr := &ast.NumberLiteral{Token: p.current, Value: p.current.Lexeme}
		p.advance()
		return expr

	case token.True:
		expr := &ast.BooleanLiteral{Token: p.current, Value: true}
		p.advance()
		return expr

	case token.False:
		expr := &ast.BooleanLiteral{Token: p.current, Value: false}
		p.advance()
		return expr

	case token.String:
		raw := p.current.Lexeme[1 : len(p.current.Lexeme)-1]
		val := processEscapes(raw)
		expr := &ast.StringLiteral{Token: p.current, Value: val}
		p.advance()
		return expr

	case token.Rune:
		// Strip quotes and parse rune
		val := []rune(p.current.Lexeme[1 : len(p.current.Lexeme)-1])
		var r rune
		if len(val) > 0 {
			r = val[0]
		}
		expr := &ast.RuneLiteral{Token: p.current, Value: r}
		p.advance()
		return expr

	case token.Symbol:
		// Strip colon
		val := p.current.Lexeme[1:]
		expr := &ast.SymbolLiteral{Token: p.current, Value: val}
		p.advance()
		return expr

	case token.Name:
		expr := &ast.Identifier{Token: p.current, Value: p.current.Lexeme}
		p.advance()
		return expr

	case token.LBracket:
		return p.parseList()

	case token.LParen:
		return p.parseGroupOrFunction()

	default:
		p.error(fmt.Sprintf("unexpected token: %s", p.current.Type))
		p.advance()
		return nil
	}
}

func (p *Parser) parseList() ast.Expression {
	tok := p.current
	p.advance()

	// Check for shaped list: [@name ...]
	if p.current.Type == token.Shape {
		shapeName := p.current.Lexeme[1:]
		p.advance()

		var elements []ast.Expression
		for p.current.Type != token.RBracket && p.current.Type != token.EOF {
			// Skip leading comma after shape name
			if p.current.Type == token.Comma {
				p.advance()
				continue
			}
			elem := p.parseExpression()
			if elem != nil {
				elements = append(elements, elem)
			}
			// Optional comma between elements
			if p.current.Type == token.Comma {
				p.advance()
			}
		}

		if p.current.Type != token.RBracket {
			p.error("expected ]")
			return nil
		}
		p.advance()

		return &ast.ShapedListLiteral{Token: tok, Shape: shapeName, Elements: elements}
	}

	// Raw list
	var elements []ast.Expression
	if p.current.Type != token.RBracket {
		// First element
		elem := p.parseExpression()
		if elem != nil {
			elements = append(elements, elem)
		}
		// Remaining elements (comma-separated)
		for p.current.Type == token.Comma {
			p.advance()
			elem := p.parseExpression()
			if elem != nil {
				elements = append(elements, elem)
			}
		}
	}

	if p.current.Type != token.RBracket {
		p.error("expected ]")
		return nil
	}
	p.advance()

	return &ast.ListLiteral{Token: tok, Elements: elements}
}

func (p *Parser) parseGroupOrFunction() ast.Expression {
	tok := p.current
	p.advance()

	// Empty parens: () or () { }
	if p.current.Type == token.RParen {
		p.advance()
		if p.current.Type == token.LBrace {
			body := p.parseBlockStatement()
			return &ast.FunctionLiteral{Token: tok, Parameters: []string{}, Body: body}
		}
		p.error("empty parentheses")
		return nil
	}

	// Check if this looks like function params: (a, b, c) { } or (a) { }
	// Function params are comma-separated names followed by ) {
	if p.current.Type == token.Name {
		// Try to parse as function params
		if p.looksLikeFunctionParams() {
			return p.parseFunctionParams(tok)
		}
	}

	// Otherwise it's a grouped expression: (1 + 2)
	expr := p.parseExpression()
	if p.current.Type != token.RParen {
		p.error("expected )")
		return nil
	}
	p.advance()
	return expr
}

// looksLikeFunctionParams peeks ahead to determine if we're parsing function params
func (p *Parser) looksLikeFunctionParams() bool {
	// Save position
	savedPos := p.pos
	savedCurrent := p.current
	savedPeek := p.peek

	// Try to consume names and commas until we hit ) or something else
	for p.current.Type == token.Name {
		p.advance()
		if p.current.Type == token.Comma {
			p.advance()
		} else {
			break
		}
	}

	// Check if we're at ) followed by {
	isFunc := p.current.Type == token.RParen && p.peek.Type == token.LBrace

	// Restore position
	p.pos = savedPos
	p.current = savedCurrent
	p.peek = savedPeek

	return isFunc
}

func (p *Parser) parseFunctionParams(tok token.Token) ast.Expression {
	var params []string

	// First param
	params = append(params, p.current.Lexeme)
	p.advance()

	// Remaining params (comma-separated)
	for p.current.Type == token.Comma {
		p.advance()
		if p.current.Type != token.Name {
			p.error("expected parameter name")
			return nil
		}
		params = append(params, p.current.Lexeme)
		p.advance()
	}

	if p.current.Type != token.RParen {
		p.error("expected ) after parameters")
		return nil
	}
	p.advance()

	if p.current.Type != token.LBrace {
		p.error("expected { after function parameters")
		return nil
	}
	body := p.parseBlockStatement()

	return &ast.FunctionLiteral{Token: tok, Parameters: params, Body: body}
}

func processEscapes(s string) string {
	var result []rune
	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		if runes[i] == '\\' && i+1 < len(runes) {
			i++
			switch runes[i] {
			case 'n':
				result = append(result, '\n')
			case 't':
				result = append(result, '\t')
			case 'r':
				result = append(result, '\r')
			case '\\':
				result = append(result, '\\')
			case '"':
				result = append(result, '"')
			default:
				result = append(result, runes[i])
			}
		} else {
			result = append(result, runes[i])
		}
	}
	return string(result)
}
