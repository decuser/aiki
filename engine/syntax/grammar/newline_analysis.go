package grammar

import (
	"fmt"
	"sort"
)

// SurfaceSymbol identifies a token class or literal lexeme in the grammar.
// Exactly one of Token or Lexeme is set.
type SurfaceSymbol struct {
	Token  string
	Lexeme string
}

func tokenSymbol(name string) SurfaceSymbol   { return SurfaceSymbol{Token: name} }
func lexemeSymbol(value string) SurfaceSymbol { return SurfaceSymbol{Lexeme: value} }

// String returns the grammar spelling used in analysis reports.
func (s SurfaceSymbol) String() string {
	if s.Token != "" {
		return s.Token
	}
	return s.Lexeme
}

// NewlineAnalysis describes the structural relationship between the expression
// grammar and the declared newline policy. It is descriptive: policy choices
// such as rejecting unambiguous leading continuations are reported, not judged.
type NewlineAnalysis struct {
	ExpressionEnd      []SurfaceSymbol
	StatementFirst     []SurfaceSymbol
	Continuation       []SurfaceSymbol
	Ambiguous          []SurfaceSymbol
	Overblocked        []SurfaceSymbol
	UncoveredEnd       []SurfaceSymbol
	DeclaredImpossible []SurfaceSymbol
}

// AnalyzeNewlineRule derives the surface sets relevant to newline termination.
//
// ExpressionEnd is LAST(expr): symbols that can end a complete expression.
// StatementFirst is FIRST(statement).
// Continuation contains symbols that can extend an already-complete expression.
// Ambiguous is StatementFirst ∩ Continuation.
// Overblocked is Continuation - StatementFirst: continuations rejected by the
// current preceding-token termination policy even though they cannot begin a
// statement.
// UncoveredEnd is ExpressionEnd - the declared completion set.
// DeclaredImpossible is the declared completion set - ExpressionEnd.
func (g *Grammar) AnalyzeNewlineRule() (*NewlineAnalysis, error) {
	if g == nil {
		return nil, fmt.Errorf("nil grammar")
	}
	a := g.Analysis()
	if a.NewlineError != nil {
		return nil, a.NewlineError
	}
	return a.Newline, nil
}

// deriveNewlineAnalysis performs the newline-specific portion of central
// grammar analysis. It is called only while building Grammar.Analysis.
func deriveNewlineAnalysis(g *Grammar) (*NewlineAnalysis, error) {
	if g == nil {
		return nil, fmt.Errorf("nil grammar")
	}
	if g.Newline == nil {
		return nil, fmt.Errorf("grammar has no newline rule")
	}
	if _, ok := g.Productions["expr"]; !ok {
		return nil, fmt.Errorf("grammar has no expr production")
	}
	if _, ok := g.Productions["statement"]; !ok {
		return nil, fmt.Errorf("grammar has no statement production")
	}

	analyzer := &grammarAnalyzer{g: g, nullableMemo: make(map[string]bool), nullableDone: make(map[string]bool)}

	ends, err := analyzer.lastProduction("expr", make(map[string]bool))
	if err != nil {
		return nil, err
	}
	firstStmt, err := analyzer.firstProduction("statement", make(map[string]bool))
	if err != nil {
		return nil, err
	}
	cont, err := analyzer.contProduction("expr", make(map[string]bool))
	if err != nil {
		return nil, err
	}

	declared := make(symbolSet)
	for _, name := range g.Newline.AfterToken {
		declared.add(tokenSymbol(name))
	}
	for _, value := range g.Newline.AfterLexeme {
		declared.add(lexemeSymbol(value))
	}

	return &NewlineAnalysis{
		ExpressionEnd:      ends.sorted(),
		StatementFirst:     firstStmt.sorted(),
		Continuation:       cont.sorted(),
		Ambiguous:          intersect(firstStmt, cont).sorted(),
		Overblocked:        subtract(cont, firstStmt).sorted(),
		UncoveredEnd:       subtract(ends, declared).sorted(),
		DeclaredImpossible: subtract(declared, ends).sorted(),
	}, nil
}

type symbolSet map[SurfaceSymbol]struct{}

func (s symbolSet) add(v SurfaceSymbol) { s[v] = struct{}{} }
func (s symbolSet) addAll(other symbolSet) {
	for v := range other {
		s.add(v)
	}
}
func (s symbolSet) sorted() []SurfaceSymbol {
	out := make([]SurfaceSymbol, 0, len(s))
	for v := range s {
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool {
		// Token classes first, then literal lexemes; lexical order within kind.
		if (out[i].Token != "") != (out[j].Token != "") {
			return out[i].Token != ""
		}
		return out[i].String() < out[j].String()
	})
	return out
}
func intersect(a, b symbolSet) symbolSet {
	out := make(symbolSet)
	for v := range a {
		if _, ok := b[v]; ok {
			out.add(v)
		}
	}
	return out
}
func subtract(a, b symbolSet) symbolSet {
	out := make(symbolSet)
	for v := range a {
		if _, ok := b[v]; !ok {
			out.add(v)
		}
	}
	return out
}

type grammarAnalyzer struct {
	g            *Grammar
	nullableMemo map[string]bool
	nullableDone map[string]bool
}

func (a *grammarAnalyzer) nullableProduction(name string, visiting map[string]bool) (bool, error) {
	if a.nullableDone[name] {
		return a.nullableMemo[name], nil
	}
	if visiting[name] {
		return false, fmt.Errorf("cannot analyze nullable recursive production %q", name)
	}
	p, ok := a.g.Productions[name]
	if !ok {
		return false, fmt.Errorf("unknown production %q", name)
	}
	visiting[name] = true
	v, err := a.nullable(p.Expr, visiting)
	delete(visiting, name)
	if err != nil {
		return false, err
	}
	a.nullableDone[name] = true
	a.nullableMemo[name] = v
	return v, nil
}

func (a *grammarAnalyzer) nullable(e Expression, visiting map[string]bool) (bool, error) {
	switch x := e.(type) {
	case *Sequence:
		for _, child := range x.Exprs {
			n, err := a.nullable(child, visiting)
			if err != nil || !n {
				return false, err
			}
		}
		return true, nil
	case *Alternative:
		for _, child := range x.Exprs {
			n, err := a.nullable(child, visiting)
			if err != nil {
				return false, err
			}
			if n {
				return true, nil
			}
		}
		return false, nil
	case *Repetition, *Option:
		return true, nil
	case *Group:
		return a.nullable(x.Expr, visiting)
	case *Reference:
		return a.nullableProduction(x.Name, visiting)
	case *Terminal, *TokenRef:
		return false, nil
	default:
		return false, fmt.Errorf("unknown grammar expression %T", e)
	}
}

func (a *grammarAnalyzer) firstProduction(name string, visiting map[string]bool) (symbolSet, error) {
	if visiting[name] {
		return nil, fmt.Errorf("cannot analyze FIRST of recursive production %q", name)
	}
	p, ok := a.g.Productions[name]
	if !ok {
		return nil, fmt.Errorf("unknown production %q", name)
	}
	visiting[name] = true
	out, err := a.first(p.Expr, visiting)
	delete(visiting, name)
	return out, err
}

func (a *grammarAnalyzer) first(e Expression, visiting map[string]bool) (symbolSet, error) {
	out := make(symbolSet)
	switch x := e.(type) {
	case *Sequence:
		for _, child := range x.Exprs {
			part, err := a.first(child, visiting)
			if err != nil {
				return nil, err
			}
			out.addAll(part)
			n, err := a.nullable(child, make(map[string]bool))
			if err != nil {
				return nil, err
			}
			if !n {
				break
			}
		}
	case *Alternative:
		for _, child := range x.Exprs {
			part, err := a.first(child, visiting)
			if err != nil {
				return nil, err
			}
			out.addAll(part)
		}
	case *Repetition:
		return a.first(x.Expr, visiting)
	case *Option:
		return a.first(x.Expr, visiting)
	case *Group:
		return a.first(x.Expr, visiting)
	case *Terminal:
		out.add(lexemeSymbol(x.Value))
	case *TokenRef:
		out.add(tokenSymbol(x.Name))
	case *Reference:
		return a.firstProduction(x.Name, visiting)
	default:
		return nil, fmt.Errorf("unknown grammar expression %T", e)
	}
	return out, nil
}

func (a *grammarAnalyzer) lastProduction(name string, visiting map[string]bool) (symbolSet, error) {
	if visiting[name] {
		return nil, fmt.Errorf("cannot analyze LAST of recursive production %q", name)
	}
	p, ok := a.g.Productions[name]
	if !ok {
		return nil, fmt.Errorf("unknown production %q", name)
	}
	visiting[name] = true
	out, err := a.last(p.Expr, visiting)
	delete(visiting, name)
	return out, err
}

func (a *grammarAnalyzer) last(e Expression, visiting map[string]bool) (symbolSet, error) {
	out := make(symbolSet)
	switch x := e.(type) {
	case *Sequence:
		for i := len(x.Exprs) - 1; i >= 0; i-- {
			child := x.Exprs[i]
			part, err := a.last(child, visiting)
			if err != nil {
				return nil, err
			}
			out.addAll(part)
			n, err := a.nullable(child, make(map[string]bool))
			if err != nil {
				return nil, err
			}
			if !n {
				break
			}
		}
	case *Alternative:
		for _, child := range x.Exprs {
			part, err := a.last(child, visiting)
			if err != nil {
				return nil, err
			}
			out.addAll(part)
		}
	case *Repetition:
		return a.last(x.Expr, visiting)
	case *Option:
		return a.last(x.Expr, visiting)
	case *Group:
		return a.last(x.Expr, visiting)
	case *Terminal:
		out.add(lexemeSymbol(x.Value))
	case *TokenRef:
		out.add(tokenSymbol(x.Name))
	case *Reference:
		return a.lastProduction(x.Name, visiting)
	default:
		return nil, fmt.Errorf("unknown grammar expression %T", e)
	}
	return out, nil
}

// cont derives tokens that can extend a derivation which is already complete at
// the right edge of e. For a right-edge repetition or option, FIRST(body) is an
// extension; nested continuations within that body are extensions as well.
func (a *grammarAnalyzer) contProduction(name string, visiting map[string]bool) (symbolSet, error) {
	if visiting[name] {
		return nil, fmt.Errorf("cannot analyze continuations of recursive production %q", name)
	}
	p, ok := a.g.Productions[name]
	if !ok {
		return nil, fmt.Errorf("unknown production %q", name)
	}
	visiting[name] = true
	out, err := a.cont(p.Expr, visiting)
	delete(visiting, name)
	return out, err
}

func (a *grammarAnalyzer) cont(e Expression, visiting map[string]bool) (symbolSet, error) {
	out := make(symbolSet)
	switch x := e.(type) {
	case *Sequence:
		suffixNullable := true
		for i := len(x.Exprs) - 1; i >= 0 && suffixNullable; i-- {
			child := x.Exprs[i]
			part, err := a.cont(child, visiting)
			if err != nil {
				return nil, err
			}
			out.addAll(part)
			n, err := a.nullable(child, make(map[string]bool))
			if err != nil {
				return nil, err
			}
			suffixNullable = suffixNullable && n
		}
	case *Alternative:
		for _, child := range x.Exprs {
			part, err := a.cont(child, visiting)
			if err != nil {
				return nil, err
			}
			out.addAll(part)
		}
	case *Repetition:
		first, err := a.first(x.Expr, visiting)
		if err != nil {
			return nil, err
		}
		out.addAll(first)
		nested, err := a.cont(x.Expr, visiting)
		if err != nil {
			return nil, err
		}
		out.addAll(nested)
	case *Option:
		first, err := a.first(x.Expr, visiting)
		if err != nil {
			return nil, err
		}
		out.addAll(first)
		nested, err := a.cont(x.Expr, visiting)
		if err != nil {
			return nil, err
		}
		out.addAll(nested)
	case *Group:
		return a.cont(x.Expr, visiting)
	case *Reference:
		return a.contProduction(x.Name, visiting)
	case *Terminal, *TokenRef:
		// Atomic symbols cannot extend themselves.
	default:
		return nil, fmt.Errorf("unknown grammar expression %T", e)
	}
	return out, nil
}
