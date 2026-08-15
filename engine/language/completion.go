package language

import (
	"fmt"
	"sort"

	"aiki/engine"
	languageobserve "aiki/engine/language/observe"
	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
)

// CompletionItem is a neutral completion candidate.
type CompletionItem struct {
	Name   string
	Kind   string
	Detail string
}

// HoverResult is authored help or known semantic information for a name.
type HoverResult struct {
	Name          string
	Signature     string
	Summary       string
	Documentation string
	Found         bool
}

// Completion returns names visible at pos. Source-defined names come from the
// lexical scope model; runtime/prelude names come from the catalog authority.
func (s *Service) Completion(doc Document, pos engine.Position, scope value.Scope) ([]CompletionItem, error) {
	s.observe(languageobserve.EventCompletionRequested, doc, fmt.Sprintf("%d:%d", pos.Line, pos.Col))
	s.hit(languageobserve.MetricCompletionRequest, 1)
	idx, err := s.buildSymbolIndex(doc)
	if err != nil {
		return nil, err
	}
	file := doc.Path
	if file == "" {
		file = doc.ID
	}
	lx := syntax.NewLexer(s.grammar, file, doc.Source, nil)
	tokens, err := lx.Tokenize()
	if err != nil {
		return nil, err
	}
	var lexical []Symbol
	for _, tok := range tokens {
		if tok.Type != "NAME" || tok.Pos.Line != pos.Line {
			continue
		}
		start := tok.Pos.Col
		end := start + len(tok.Lexeme)
		if pos.Col >= start && pos.Col <= end {
			lexical = idx.visible[posKey(tok.Pos)]
			break
		}
	}
	set := map[string]CompletionItem{}
	for _, sym := range lexical {
		set[sym.Name] = CompletionItem{Name: sym.Name, Kind: sym.Kind, Detail: "source-defined " + sym.Kind}
	}
	if s.catalog != nil {
		for _, name := range s.catalog.VisibleNames(scope) {
			item := CompletionItem{Name: name, Kind: "builtin"}
			if h, ok := s.catalog.Help(name); ok {
				item.Detail = h.Template
			}
			set[name] = item
		}
	}
	out := make([]CompletionItem, 0, len(set))
	for _, item := range set {
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// Hover returns authored help for runtime/prelude names or known source
// definition information for lexical bindings.
func (s *Service) Hover(doc Document, pos engine.Position) (HoverResult, error) {
	s.observe(languageobserve.EventHoverRequested, doc, fmt.Sprintf("%d:%d", pos.Line, pos.Col))
	s.hit(languageobserve.MetricHoverRequest, 1)
	file := doc.Path
	if file == "" {
		file = doc.ID
	}
	lx := syntax.NewLexer(s.grammar, file, doc.Source, nil)
	tokens, err := lx.Tokenize()
	if err != nil {
		return HoverResult{}, err
	}
	var target *syntax.Token
	for i := range tokens {
		tok := &tokens[i]
		if tok.Type != "NAME" || tok.Pos.Line != pos.Line {
			continue
		}
		if pos.Col >= tok.Pos.Col && pos.Col <= tok.Pos.Col+len(tok.Lexeme) {
			target = tok
			break
		}
	}
	if target == nil {
		return HoverResult{}, nil
	}
	idx, err := s.buildSymbolIndex(doc)
	if err != nil {
		return HoverResult{}, err
	}
	if sym, ok := idx.refs[posKey(target.Pos)]; ok {
		return HoverResult{Name: sym.Name, Signature: sym.Kind + " " + sym.Name, Summary: fmt.Sprintf("defined at %d:%d", sym.Pos.Line, sym.Pos.Col), Found: true}, nil
	}
	if s.catalog != nil {
		if h, ok := s.catalog.Help(target.Lexeme); ok {
			return HoverResult{Name: h.Name, Signature: h.Template, Summary: h.Summary, Documentation: h.Doc, Found: true}, nil
		}
	}
	return HoverResult{}, nil
}
