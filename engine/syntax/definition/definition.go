// definition/definition.go - loads grammar from EBNF and implements GrammarContract
package definition

import (
	_ "embed"
	"regexp"
	"sort"
	"strings"

	"aiki/engine"
	"aiki/engine/internal"
	"aiki/engine/syntax"
)

//go:embed grammar.ebnf
var grammarSource string

type aikiDefinition struct {
	grammar   *Grammar
	tokens    []syntax.TokenDef
	prodCache map[string]syntax.Production
	observer  engine.Observer
}

func New() syntax.GrammarContract {
	g, err := Parse(grammarSource)
	if err != nil {
		panic("failed to parse grammar.ebnf: " + err.Error())
	}

	d := &aikiDefinition{
		grammar:   g,
		prodCache: make(map[string]syntax.Production),
		observer:  internal.SilentObserver{},
	}
	d.buildTokens()
	return d
}

func (d *aikiDefinition) buildTokens() {
	// Process tokens in a specific order: skip tokens first, then keywords before NAME
	// to ensure proper matching priority
	var skipTokens, keywordToken, nameToken, otherTokens []syntax.TokenDef

	for _, tok := range d.grammar.Tokens {
		def := syntax.TokenDef{
			Name: tok.Name,
			Skip: tok.Skip,
		}

		if tok.Pattern != nil {
			def.Pattern = tok.Pattern
		} else if tok.Literal != "" {
			// Build regex from literals (keywords/operators)
			// Split by whitespace and escape each, join with |
			parts := strings.Fields(tok.Literal)
			var escaped []string
			for _, p := range parts {
				escaped = append(escaped, regexp.QuoteMeta(p))
			}
			// Sort by length descending so longer matches come first
			// This ensures <= matches before <
			sort.Slice(escaped, func(i, j int) bool {
				return len(escaped[i]) > len(escaped[j])
			})
			pattern := "^(" + strings.Join(escaped, "|") + ")"
			// For keywords, add word boundary
			if tok.Name == "KEYWORD" {
				pattern = `^\b(` + strings.Join(escaped, "|") + `)\b`
			}
			def.Pattern = regexp.MustCompile(pattern)
		}

		// Categorize tokens for proper ordering
		if def.Skip {
			skipTokens = append(skipTokens, def)
		} else if def.Name == "KEYWORD" {
			keywordToken = append(keywordToken, def)
		} else if def.Name == "NAME" {
			nameToken = append(nameToken, def)
		} else {
			otherTokens = append(otherTokens, def)
		}
	}

	// Build final token list: skip first, then keywords, then other, then NAME last
	d.tokens = append(d.tokens, skipTokens...)
	d.tokens = append(d.tokens, keywordToken...)
	d.tokens = append(d.tokens, otherTokens...)
	d.tokens = append(d.tokens, nameToken...)
}

func (d *aikiDefinition) GetTokens() []syntax.TokenDef {
	return d.tokens
}

func (d *aikiDefinition) GetProduction(name string) (syntax.Production, bool) {
	// Check cache first
	if prod, ok := d.prodCache[name]; ok {
		return prod, true
	}

	// Look up in grammar
	gProd, ok := d.grammar.Productions[name]
	if !ok {
		return syntax.Production{}, false
	}

	// Convert Expression tree to flat [][]Term format
	prod := d.convertProduction(gProd.Expr)
	d.prodCache[name] = prod
	return prod, true
}

func (d *aikiDefinition) convertProduction(expr Expression) syntax.Production {
	var alternatives [][]syntax.Term

	// Get all alternatives
	alts := d.getAlternatives(expr)
	for _, alt := range alts {
		terms := d.getSequenceTerms(alt)
		alternatives = append(alternatives, terms)
	}

	return syntax.Production{Expressions: alternatives}
}

func (d *aikiDefinition) getAlternatives(expr Expression) []Expression {
	switch e := expr.(type) {
	case *Alternative:
		return e.Exprs
	default:
		return []Expression{expr}
	}
}

func (d *aikiDefinition) getSequenceTerms(expr Expression) []syntax.Term {
	var terms []syntax.Term

	switch e := expr.(type) {
	case *Sequence:
		for _, sub := range e.Exprs {
			terms = append(terms, d.exprToTerms(sub)...)
		}
	default:
		terms = append(terms, d.exprToTerms(expr)...)
	}

	return terms
}

func (d *aikiDefinition) exprToTerms(expr Expression) []syntax.Term {
	switch e := expr.(type) {
	case *Terminal:
		return []syntax.Term{{Value: e.Value, IsSymbol: false}}

	case *Reference:
		return []syntax.Term{{Value: e.Name, IsSymbol: true}}

	case *TokenRef:
		return []syntax.Term{{Value: e.Name, IsSymbol: false}}

	case *Option:
		// For options, we need to get the inner terms and mark them optional
		inner := d.exprToTerms(e.Expr)
		if len(inner) == 1 {
			inner[0].IsOption = true
			return inner
		}
		// For complex options, wrap in a group-like structure
		// This is a simplification - complex options may need special handling
		for i := range inner {
			inner[i].IsOption = true
		}
		return inner

	case *Repetition:
		// For repetitions, get inner terms and mark as repeat
		inner := d.exprToTerms(e.Expr)
		if len(inner) == 1 {
			inner[0].IsRepeat = true
			return inner
		}
		// For complex repetitions, mark all as repeat
		for i := range inner {
			inner[i].IsRepeat = true
		}
		return inner

	case *Group:
		// Groups are transparent - just get inner terms
		return d.exprToTerms(e.Expr)

	case *Sequence:
		var terms []syntax.Term
		for _, sub := range e.Exprs {
			terms = append(terms, d.exprToTerms(sub)...)
		}
		return terms

	case *Alternative:
		// Alternatives at this level indicate nested alternatives
		// The parser handles this via backtracking, so we take first option
		// This is a simplification - may need adjustment
		if len(e.Exprs) > 0 {
			return d.exprToTerms(e.Exprs[0])
		}
		return nil

	default:
		return nil
	}
}

func (d *aikiDefinition) GetStart() string {
	return d.grammar.Start
}

func (d *aikiDefinition) Observe() engine.Observer {
	return d.observer
}

func (d *aikiDefinition) SetObserver(o engine.Observer) {
	d.observer = o
}
