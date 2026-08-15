package syntax

import (
	"fmt"
	"strings"

	"aiki/engine"
	"aiki/engine/syntax/grammar"
)

// ParseFailure captures the best error context for reporting.
type ParseFailure struct {
	Pos               engine.Position // where failure occurred
	Got               string          // actual token
	Expected          string          // what was expected
	Stack             []string        // production stack (outermost first)
	NewlineTerminated bool            // a grammar-declared newline boundary caused this leftover continuation
}

type Parser struct {
	grammar              *grammar.Grammar
	tokens               []Token
	pos                  int
	furthest             int // furthest position reached
	source               string
	observer             engine.Observer
	stack                []string      // current production stack
	failure              *ParseFailure // best failure so far
	syntheticTerminators map[int]bool  // filtered-token indexes inserted by the newline policy
	continuationTokens   map[string]bool
	continuationLexemes  map[string]bool
}

func NewParser(g *grammar.Grammar, tokens []Token, source string, observer engine.Observer) *Parser {
	if observer == nil {
		observer = engine.SilentObserver{}
	}

	// Compile the grammar-declared newline policy once for this token stream.
	// The parser applies the policy generically; Aiki's particular completion
	// tokens, lexemes, and suppression delimiters live only in grammar.ebnfx.
	rule := g.Newline
	afterToken := make(map[string]struct{}, len(rule.AfterToken))
	for _, name := range rule.AfterToken {
		afterToken[name] = struct{}{}
	}
	afterLexeme := make(map[string]struct{}, len(rule.AfterLexeme))
	for _, lexeme := range rule.AfterLexeme {
		afterLexeme[lexeme] = struct{}{}
	}
	suppressOpen := make(map[string]string, len(rule.SuppressIn))
	suppressClose := make(map[string]struct{}, len(rule.SuppressIn))
	for _, pair := range rule.SuppressIn {
		suppressOpen[pair[0]] = pair[1]
		suppressClose[pair[1]] = struct{}{}
	}
	endsStatement := func(tok Token) bool {
		if _, ok := afterToken[tok.Type]; ok {
			return true
		}
		_, ok := afterLexeme[tok.Lexeme]
		return ok
	}

	// The same grammar analysis used by enginesmoke supplies continuation
	// membership for the targeted leftover-token diagnostic. Failure to derive
	// it must never make parsing unavailable; it only disables that refinement.
	continuationTokens := make(map[string]bool)
	continuationLexemes := make(map[string]bool)
	analysis := g.Analysis()
	if analysis.NewlineError != nil {
		if diagnostics, ok := observer.(engine.DiagnosticObserver); ok {
			diagnostics.OnDiagnostic("grammar-newline-analysis", analysis.NewlineError.Error(), rule.Pos)
		}
	}
	if analysis.Newline != nil {
		for _, symbol := range analysis.Newline.Continuation {
			if symbol.Token != "" {
				continuationTokens[symbol.Token] = true
			} else {
				continuationLexemes[symbol.Lexeme] = true
			}
		}
	}

	// Filter @skip tokens and normalize physical newlines into statement
	// terminators according to the grammar-declared policy.
	filtered := make([]Token, 0, len(tokens))
	syntheticTerminators := make(map[int]bool)
	var suppressStack []string
	for _, tok := range tokens {
		if def, ok := g.GetToken(tok.Type); ok && def.Skip {
			continue
		}

		// Suppression is delimiter-aware rather than an aggregate depth. An
		// unmatched or mismatched closer cannot drive state negative or cancel a
		// later legitimate opener; the grammar parser will report the malformed
		// delimiter itself downstream.
		if len(suppressStack) > 0 && tok.Lexeme == suppressStack[len(suppressStack)-1] {
			suppressStack = suppressStack[:len(suppressStack)-1]
		} else if closer, ok := suppressOpen[tok.Lexeme]; ok {
			suppressStack = append(suppressStack, closer)
		} else if _, ok := suppressClose[tok.Lexeme]; ok {
			// Unmatched/mismatched closer: ignore for normalization state.
		}

		if tok.Type == rule.Token {
			if len(suppressStack) == 0 && len(filtered) > 0 && endsStatement(filtered[len(filtered)-1]) {
				syntheticTerminators[len(filtered)] = true
				filtered = append(filtered, Token{
					Type:   "DELIMITER",
					Lexeme: ";",
					Pos:    tok.Pos,
				})
			}
			continue
		}
		filtered = append(filtered, tok)
	}
	return &Parser{
		grammar:              g,
		tokens:               filtered,
		pos:                  0,
		furthest:             0,
		source:               source,
		observer:             observer,
		syntheticTerminators: syntheticTerminators,
		continuationTokens:   continuationTokens,
		continuationLexemes:  continuationLexemes,
	}
}

func (p *Parser) Parse() (*Node, error) {
	node, err := p.parseProduction(p.grammar.Start)
	if err != nil {
		return nil, p.renderFailure()
	}
	if p.pos < len(p.tokens) {
		// Leftover tokens - record as failure. If the immediately preceding
		// terminator was synthesized from a grammar-declared newline and the
		// leftover token is a grammar-derived continuation, retain that provenance
		// so the diagnostic can explain the actual surface rule.
		tok := p.tokens[p.pos]
		p.recordFailure(tok.Pos, tok.Lexeme, "end of input")
		if p.failure != nil && p.pos > 0 && p.syntheticTerminators[p.pos-1] && p.isContinuation(tok) {
			p.failure.NewlineTerminated = true
		}
		return nil, p.renderFailure()
	}
	return node, nil
}

func (p *Parser) isContinuation(tok Token) bool {
	return p.continuationTokens[tok.Type] || p.continuationLexemes[tok.Lexeme]
}

// renderFailure formats the best failure using grammar metadata.
func (p *Parser) renderFailure() error {
	if p.failure == nil {
		return fmt.Errorf("parse error")
	}

	f := p.failure
	line := engine.GetSourceLine(p.source, f.Pos.Line)
	caret := engine.FormatCaret(f.Pos.Col)

	// Find the most relevant production in stack with @error
	// Search from innermost (end) to outermost (start) - prefer specific errors
	var prodName string
	var meta grammar.Meta
	for i := len(f.Stack) - 1; i >= 0; i-- {
		name := f.Stack[i]
		if prod, ok := p.grammar.GetProduction(name); ok {
			if prod.Meta.Error != "" {
				prodName = name
				meta = prod.Meta
				break
			}
		}
	}

	// For closing delimiters like '}' or ')', use the raw terminal error
	// since "expected '}'" is clearer than a production error
	isClosingDelim := f.Expected == "'}'" || f.Expected == "')'" || f.Expected == "']'"

	// Build error message
	var msg strings.Builder
	msg.WriteString(fmt.Sprintf("%s:%d:%d:\n", p.getFile(), f.Pos.Line, f.Pos.Col))
	msg.WriteString(line + "\n")
	msg.WriteString(caret + "\n")

	if f.NewlineTerminated && p.grammar.Newline != nil {
		msg.WriteString("the previous newline ended the statement")
		if help := p.grammar.Newline.Meta.Help; help != "" {
			msg.WriteString(fmt.Sprintf("\n\nnewline: %s", help))
		}
		msg.WriteString(fmt.Sprintf("\n\n'%s' continues an expression, but the newline terminated the expression before it.", f.Got))
		msg.WriteString(fmt.Sprintf("\nPlace '%s' before the newline it continues.", f.Got))
	} else if prodName != "" && meta.Error != "" && !isClosingDelim {
		msg.WriteString(fmt.Sprintf("%s: %s", prodName, meta.Error))
		if meta.Template != "" {
			msg.WriteString(fmt.Sprintf("\n\nSyntax: %s", meta.Template))
		}
	} else {
		msg.WriteString(fmt.Sprintf("expected %s", f.Expected))
		if f.Got != "" && f.Got != "EOF" && f.Got != "end of input" {
			msg.WriteString(fmt.Sprintf(", got '%s'", f.Got))
		}
	}

	return fmt.Errorf("%s", msg.String())
}

// getFile returns the source file name (or "<input>" if unknown).
func (p *Parser) getFile() string {
	if len(p.tokens) > 0 && p.tokens[0].Pos.File != "" {
		return p.tokens[0].Pos.File
	}
	return "<input>"
}

// recordFailure records a failure if it's at or beyond the furthest position.
func (p *Parser) recordFailure(pos engine.Position, got, expected string) {
	// Only update if this is further than previous best
	// Use >= to prefer later failures at the same position (more context)
	tokenPos := p.pos
	if p.failure != nil && tokenPos < p.furthest {
		return
	}

	// Copy current stack
	stack := make([]string, len(p.stack))
	copy(stack, p.stack)

	p.failure = &ParseFailure{
		Pos:      pos,
		Got:      got,
		Expected: expected,
		Stack:    stack,
	}
	p.furthest = tokenPos
}

func (p *Parser) parseProduction(name string) (*Node, error) {
	prod, ok := p.grammar.GetProduction(name)
	if !ok {
		return nil, fmt.Errorf("undefined production: %s", name)
	}

	// Push onto stack
	p.stack = append(p.stack, name)
	defer func() { p.stack = p.stack[:len(p.stack)-1] }()

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
	// Don't record failure here - let the deeper failures stand
	// They have more specific information about what was expected
	return nil, fmt.Errorf("no alternative matched")
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
		tok := p.peek()
		p.recordFailure(tok.Pos, "end of input", fmt.Sprintf("'%s'", term.Value))
		return nil, fmt.Errorf("unexpected end of input")
	}
	tok := p.tokens[p.pos]
	if tok.Lexeme != term.Value {
		p.recordFailure(tok.Pos, tok.Lexeme, fmt.Sprintf("'%s'", term.Value))
		return nil, fmt.Errorf("expected '%s'", term.Value)
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
		tok := p.peek()
		p.recordFailure(tok.Pos, "end of input", ref.Name)
		return nil, fmt.Errorf("unexpected end of input")
	}
	tok := p.tokens[p.pos]
	if !p.matchesToken(tok, ref.Name) {
		p.recordFailure(tok.Pos, tok.Lexeme, ref.Name)
		return nil, fmt.Errorf("expected %s", ref.Name)
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
		// Return EOF token with position at end of last token
		if len(p.tokens) > 0 {
			last := p.tokens[len(p.tokens)-1]
			return Token{
				Type:   "EOF",
				Lexeme: "",
				Pos: engine.Position{
					File: last.Pos.File,
					Line: last.Pos.Line,
					Col:  last.Pos.Col + len(last.Lexeme),
				},
			}
		}
		return Token{Type: "EOF"}
	}
	return p.tokens[p.pos]
}
