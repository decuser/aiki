package grammar

import "sort"

// Analysis is the cached structural knowledge derived from a Grammar.
//
// The grammar tree remains the authority. Analysis is its reusable derived
// view: consumers may compile these facts into representations suited to their
// work, but should not walk grammar expressions to rediscover them.
type Analysis struct {
	productions          map[string]struct{}
	tokenRefs            map[string]struct{}
	astNodeTypes         map[string]struct{}
	terminalAlternatives map[string]map[string]struct{}

	// Newline is present when Aiki-style newline analysis can be derived.
	// NewlineError records why it could not be derived without making general
	// structural analysis unavailable.
	Newline      *NewlineAnalysis
	NewlineError error
}

// Analysis returns the cached analysis, deriving it once for grammars that
// were constructed manually rather than parsed by the EBNF loader.
func (g *Grammar) Analysis() *Analysis {
	if g == nil {
		return nil
	}
	if g.analysis == nil {
		g.Reanalyze()
	}
	return g.analysis
}

// Reanalyze explicitly rebuilds derived grammar knowledge after deliberate
// mutation. Normal parsed/loaded grammars are analyzed once and never need it.
func (g *Grammar) Reanalyze() *Analysis {
	if g == nil {
		return nil
	}
	g.analysis = analyzeGrammar(g)
	return g.analysis
}

// ProductionNames returns the named grammar productions in lexical order.
func (a *Analysis) ProductionNames() []string {
	if a == nil {
		return nil
	}
	return sortedSet(a.productions)
}

// TokenRefs returns the token classes referenced directly by productions.
// The returned set is a copy and may be modified by callers.
func (a *Analysis) TokenRefs() map[string]struct{} {
	if a == nil {
		return map[string]struct{}{}
	}
	return copySet(a.tokenRefs)
}

// ASTNodeTypes returns node types the grammar itself can produce: named
// productions plus production-referenced token classes. Parser-synthesized
// nodes are intentionally not grammar facts. The returned set is a copy and
// may be modified by callers.
func (a *Analysis) ASTNodeTypes() map[string]struct{} {
	if a == nil {
		return map[string]struct{}{}
	}
	return copySet(a.astNodeTypes)
}

// TerminalAlternatives returns literal terminals structurally contained by a
// named production. The returned set is a copy and may be modified by callers.
func (a *Analysis) TerminalAlternatives(production string) map[string]struct{} {
	if a == nil {
		return map[string]struct{}{}
	}
	return copySet(a.terminalAlternatives[production])
}

func analyzeGrammar(g *Grammar) *Analysis {
	a := &Analysis{
		productions:          make(map[string]struct{}, len(g.Productions)),
		tokenRefs:            make(map[string]struct{}),
		astNodeTypes:         make(map[string]struct{}, len(g.Productions)+8),
		terminalAlternatives: make(map[string]map[string]struct{}, len(g.Productions)),
	}

	// One central structural walk. References are recorded, not expanded: every
	// named production is itself visited exactly once below.
	var walk func(string, Expression)
	walk = func(production string, expr Expression) {
		switch x := expr.(type) {
		case *TokenRef:
			a.tokenRefs[x.Name] = struct{}{}
			a.astNodeTypes[x.Name] = struct{}{}
		case *Terminal:
			terms := a.terminalAlternatives[production]
			if terms == nil {
				terms = make(map[string]struct{})
				a.terminalAlternatives[production] = terms
			}
			terms[x.Value] = struct{}{}
		case *Sequence:
			for _, child := range x.Exprs {
				walk(production, child)
			}
		case *Alternative:
			for _, child := range x.Exprs {
				walk(production, child)
			}
		case *Repetition:
			walk(production, x.Expr)
		case *Option:
			walk(production, x.Expr)
		case *Group:
			walk(production, x.Expr)
		case *Reference:
			// The referenced production has its own visit.
		}
	}

	for name, production := range g.Productions {
		a.productions[name] = struct{}{}
		a.astNodeTypes[name] = struct{}{}
		walk(name, production.Expr)
	}

	a.Newline, a.NewlineError = deriveNewlineAnalysis(g)
	return a
}

func copySet(in map[string]struct{}) map[string]struct{} {
	out := make(map[string]struct{}, len(in))
	for name := range in {
		out[name] = struct{}{}
	}
	return out
}

func sortedSet(in map[string]struct{}) []string {
	out := make([]string, 0, len(in))
	for name := range in {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}
