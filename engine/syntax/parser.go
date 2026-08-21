package syntax

import (
	"fmt"
	"strings"

	"aiki/engine"
	"aiki/engine/syntax/grammar"
)

// parserNoMatch is the internal speculative mismatch signal. It carries no
// diagnostic payload; the Parser records the best failure separately.
type parserNoMatch struct{}

func (parserNoMatch) Error() string { return "no match" }

var errParserNoMatch error = parserNoMatch{}

type failureExpectationKind uint8

const (
	failureExpectationText failureExpectationKind = iota
	failureExpectationTerminal
)

// ParseFailure captures the best error context for reporting.
type ParseFailure struct {
	Pos               engine.Position // where failure occurred
	Got               string          // actual token
	Expected          string          // what was expected
	Production        string          // innermost production carrying @error metadata
	Stack             []string        // legacy compatibility; parser no longer materializes this stack
	NewlineTerminated bool            // a grammar-declared newline boundary caused this leftover continuation

	expectationKind failureExpectationKind
	terminal        string
}

type Parser struct {
	grammar              *grammar.Grammar
	tokens               []Token
	pos                  int
	furthest             int // furthest position reached
	source               string
	observer             engine.Observer
	errorProduction      string       // innermost active production carrying @error metadata
	failure              ParseFailure // best failure so far, stored in-place
	hasFailure           bool
	syntheticTerminators map[int]bool // filtered-token indexes inserted by the newline policy
	continuationTokens   map[string]bool
	continuationLexemes  map[string]bool
}

func NewParser(g *grammar.Grammar, tokens []Token, source string, observer engine.Observer) *Parser {
	if observer == nil {
		observer = engine.SilentObserver{}
	}

	// Apply the grammar-declared skip/newline policy through the same neutral
	// normalization seam used by cross-implementation conformance.
	filtered, syntheticTerminators := normalizeTokens(g, tokens)

	// The same grammar analysis used by enginesmoke supplies continuation
	// membership for the targeted leftover-token diagnostic. Failure to derive
	// it must never make parsing unavailable; it only disables that refinement.
	rule := g.Newline
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
		if p.hasFailure && p.pos > 0 && p.syntheticTerminators[p.pos-1] && p.isContinuation(tok) {
			p.failure.NewlineTerminated = true
		}
		return nil, p.renderFailure()
	}
	return node, nil
}

func (p *Parser) isContinuation(tok Token) bool {
	return p.continuationTokens[tok.Type] || p.continuationLexemes[tok.Lexeme]
}

func (f *ParseFailure) expectedText() string {
	if f.expectationKind == failureExpectationTerminal {
		return "'" + f.terminal + "'"
	}
	return f.Expected
}

// renderFailure formats the best failure using grammar metadata.
func (p *Parser) renderFailure() error {
	if !p.hasFailure {
		return &SourceError{Kind: "parse", Message: "parse error", Rendered: "parse error"}
	}

	f := &p.failure
	line := engine.GetSourceLine(p.source, f.Pos.Line)
	caret := engine.FormatCaret(f.Pos.Col)

	// The parser records the innermost active production carrying @error
	// metadata at failure time. This is the only production-stack fact needed
	// for rendering, so speculative parsing does not materialize stack copies.
	prodName := f.Production
	var meta grammar.Meta
	if prodName != "" {
		if prod, ok := p.grammar.GetProduction(prodName); ok {
			meta = prod.Meta
		}
	}
	// Preserve rendering compatibility for ParseFailure values manually built
	// with the legacy Stack field. Normal parser operation never populates it.
	if prodName == "" {
		for i := len(f.Stack) - 1; i >= 0; i-- {
			name := f.Stack[i]
			if prod, ok := p.grammar.GetProduction(name); ok && prod.Meta.Error != "" {
				prodName = name
				meta = prod.Meta
				break
			}
		}
	}

	expected := f.expectedText()

	// For closing delimiters like '}' or ')', use the raw terminal error
	// since "expected '}'" is clearer than a production error.
	isClosingDelim := expected == "'}'" || expected == "')'" || expected == "']'"

	var detail strings.Builder
	if f.NewlineTerminated && p.grammar.Newline != nil {
		detail.WriteString("the previous newline ended the statement")
		if help := p.grammar.Newline.Meta.Help; help != "" {
			detail.WriteString(fmt.Sprintf("\n\nnewline: %s", help))
		}
		detail.WriteString(fmt.Sprintf("\n\n'%s' continues an expression, but the newline terminated the expression before it.", f.Got))
		detail.WriteString(fmt.Sprintf("\nPlace '%s' before the newline it continues.", f.Got))
	} else if prodName != "" && meta.Error != "" && !isClosingDelim {
		detail.WriteString(fmt.Sprintf("%s: %s", prodName, meta.Error))
		if meta.Template != "" {
			detail.WriteString(fmt.Sprintf("\n\nSyntax: %s", meta.Template))
		}
	} else {
		detail.WriteString(fmt.Sprintf("expected %s", expected))
		if f.Got != "" && f.Got != "EOF" && f.Got != "end of input" {
			detail.WriteString(fmt.Sprintf(", got '%s'", f.Got))
		}
	}

	message := detail.String()
	var rendered strings.Builder
	rendered.WriteString(fmt.Sprintf("%s:%d:%d:\n", p.getFile(), f.Pos.Line, f.Pos.Col))
	rendered.WriteString(line + "\n")
	rendered.WriteString(caret + "\n")
	rendered.WriteString(message)

	return &SourceError{Kind: "parse", Pos: f.Pos, Message: message, Rendered: rendered.String()}
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
	// Only update if this is further than previous best. Use >= to prefer later
	// failures at the same position, preserving the existing context rule.
	tokenPos := p.pos
	if p.hasFailure && tokenPos < p.furthest {
		return
	}

	// Keep the best failure in-place. The current @error-bearing production is
	// maintained incrementally by parseProduction, so speculative failure
	// recording performs no production-stack copy and no failure allocation.
	p.failure.Pos = pos
	p.failure.Got = got
	p.failure.Expected = expected
	p.failure.expectationKind = failureExpectationText
	p.failure.terminal = ""
	p.failure.Production = p.errorProduction
	p.failure.Stack = nil
	p.failure.NewlineTerminated = false
	p.hasFailure = true
	p.furthest = tokenPos
}

func (p *Parser) recordTerminalFailure(pos engine.Position, got, terminal string) {
	tokenPos := p.pos
	if p.hasFailure && tokenPos < p.furthest {
		return
	}

	p.failure.Pos = pos
	p.failure.Got = got
	p.failure.Expected = ""
	p.failure.expectationKind = failureExpectationTerminal
	p.failure.terminal = terminal
	p.failure.Production = p.errorProduction
	p.failure.Stack = nil
	p.failure.NewlineTerminated = false
	p.hasFailure = true
	p.furthest = tokenPos
}

func (p *Parser) parseProduction(name string) (*Node, error) {
	prod, ok := p.grammar.GetProduction(name)
	if !ok {
		return nil, fmt.Errorf("undefined production: %s", name)
	}

	// Track only the innermost active production whose metadata can affect a
	// rendered failure. The former full production stack existed solely to
	// recover this one fact after speculative parsing.
	previousErrorProduction := p.errorProduction
	if prod.Meta.Error != "" {
		p.errorProduction = name
	}
	defer func() { p.errorProduction = previousErrorProduction }()

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
	return nil, errParserNoMatch
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
		p.recordTerminalFailure(tok.Pos, "end of input", term.Value)
		return nil, errParserNoMatch
	}
	tok := p.tokens[p.pos]
	if tok.Lexeme != term.Value {
		p.recordTerminalFailure(tok.Pos, tok.Lexeme, term.Value)
		return nil, errParserNoMatch
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
		return nil, errParserNoMatch
	}
	tok := p.tokens[p.pos]
	if !p.matchesToken(tok, ref.Name) {
		p.recordFailure(tok.Pos, tok.Lexeme, ref.Name)
		return nil, errParserNoMatch
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
