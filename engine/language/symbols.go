package language

import (
	"fmt"
	"sort"

	"aiki/engine"
	languageobserve "aiki/engine/language/observe"
	"aiki/engine/syntax"
)

// Symbol is a source-defined binding known to the language service.
type Symbol struct {
	Name     string
	Kind     string
	Pos      engine.Position
	TopLevel bool
}

// DefinitionResult is the resolved definition for a source position.
type DefinitionResult struct {
	Symbol Symbol
	Found  bool
}

type symbolScope struct {
	parent *symbolScope
	defs   map[string]Symbol
}

func newSymbolScope(parent *symbolScope) *symbolScope {
	return &symbolScope{parent: parent, defs: map[string]Symbol{}}
}
func (s *symbolScope) define(sym Symbol) { s.defs[sym.Name] = sym }
func (s *symbolScope) lookup(name string) (Symbol, bool) {
	for p := s; p != nil; p = p.parent {
		if v, ok := p.defs[name]; ok {
			return v, true
		}
	}
	return Symbol{}, false
}

type symbolIndex struct {
	symbols []Symbol
	refs    map[string]Symbol
}

func posKey(p engine.Position) string { return fmt.Sprintf("%d:%d", p.Line, p.Col) }

// Symbols returns all source-defined symbols in stable source order.
func (s *Service) Symbols(doc Document) ([]Symbol, error) {
	s.observe(languageobserve.EventSymbolsRequested, doc, "")
	s.hit(languageobserve.MetricSymbolsRequest, 1)
	idx, err := s.buildSymbolIndex(doc)
	if err != nil {
		return nil, err
	}
	out := append([]Symbol(nil), idx.symbols...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Pos.Line != out[j].Pos.Line {
			return out[i].Pos.Line < out[j].Pos.Line
		}
		return out[i].Pos.Col < out[j].Pos.Col
	})
	return out, nil
}

// Definition resolves the name token beginning at pos to its lexical definition.
func (s *Service) Definition(doc Document, pos engine.Position) (DefinitionResult, error) {
	s.observe(languageobserve.EventDefinitionRequested, doc, fmt.Sprintf("%d:%d", pos.Line, pos.Col))
	s.hit(languageobserve.MetricDefinitionRequest, 1)
	idx, err := s.buildSymbolIndex(doc)
	if err != nil {
		return DefinitionResult{}, err
	}
	sym, ok := idx.refs[posKey(pos)]
	return DefinitionResult{Symbol: sym, Found: ok}, nil
}

func (s *Service) buildSymbolIndex(doc Document) (*symbolIndex, error) {
	file := doc.Path
	if file == "" {
		file = doc.ID
	}
	lx := syntax.NewLexer(s.grammar, file, doc.Source, nil)
	toks, err := lx.Tokenize()
	if err != nil {
		return nil, err
	}
	p := syntax.NewParser(s.grammar, toks, doc.Source, nil)
	root, err := p.Parse()
	if err != nil {
		return nil, err
	}
	idx := &symbolIndex{refs: map[string]Symbol{}}
	global := newSymbolScope(nil)
	// Aiki top-level lets are mutually visible to structural analysis.
	if root != nil && root.Type == "program" {
		for _, st := range root.Children {
			n := unwrapStatement(st)
			if n != nil && n.Type == "let_stmt" {
				if def := bindingNode(n); def != nil {
					idx.addDef(global, def, kindForBinding(def), true)
				}
			}
		}
	}
	idx.walk(root, global, true)
	return idx, nil
}

func unwrapStatement(n *syntax.Node) *syntax.Node {
	if n != nil && n.Type == "statement" && len(n.Children) > 0 {
		return n.Children[0]
	}
	return n
}
func bindingNode(n *syntax.Node) *syntax.Node {
	for _, ch := range n.Children {
		if ch.Type == "NAME" || ch.Type == "SHAPE" {
			return ch
		}
	}
	return nil
}
func kindForBinding(n *syntax.Node) string {
	if n != nil && n.Type == "SHAPE" {
		return "shape"
	}
	return "binding"
}
func (idx *symbolIndex) addDef(scope *symbolScope, n *syntax.Node, kind string, top bool) Symbol {
	sym := Symbol{Name: n.Value, Kind: kind, Pos: n.Pos, TopLevel: top}
	scope.define(sym)
	idx.symbols = append(idx.symbols, sym)
	idx.refs[posKey(n.Pos)] = sym
	return sym
}
func (idx *symbolIndex) addRef(scope *symbolScope, n *syntax.Node) {
	if sym, ok := scope.lookup(n.Value); ok {
		idx.refs[posKey(n.Pos)] = sym
	}
}

func (idx *symbolIndex) walk(n *syntax.Node, scope *symbolScope, top bool) {
	if n == nil {
		return
	}
	switch n.Type {
	case "program":
		for _, ch := range n.Children {
			idx.walk(ch, scope, true)
		}
	case "statement":
		for _, ch := range n.Children {
			idx.walk(ch, scope, top)
		}
	case "expr_stmt", "expr", "pipe_expr", "infix_expr", "unary_expr", "primary", "list_literal", "call", "index", "return_stmt", "if_stmt", "while_stmt", "block":
		for _, ch := range n.Children {
			idx.walk(ch, scope, false)
		}
	case "let_stmt":
		def := bindingNode(n)
		if def != nil && !top {
			idx.addDef(scope, def, kindForBinding(def), false)
		}
		for _, ch := range n.Children {
			if ch != def {
				idx.walk(ch, scope, false)
			}
		}
	case "assign_stmt":
		for _, ch := range n.Children {
			if ch.Type == "NAME" {
				idx.addRef(scope, ch)
			} else {
				idx.walk(ch, scope, false)
			}
		}
	case "func_literal":
		child := newSymbolScope(scope)
		if params := n.ChildByType("params"); params != nil {
			idx.bindNames(params, child, "parameter")
		}
		if b := n.ChildByType("block"); b != nil {
			idx.walk(b, child, false)
		}
	case "match_stmt":
		seenExpr := false
		for i := 0; i < len(n.Children); i++ {
			ch := n.Children[i]
			if ch.Type == "expr" && !seenExpr {
				idx.walk(ch, scope, false)
				seenExpr = true
				continue
			}
			if ch.Type == "pattern" {
				arm := newSymbolScope(scope)
				idx.bindPattern(ch, arm)
				if i+1 < len(n.Children) && n.Children[i+1].Type == "block" {
					idx.walk(n.Children[i+1], arm, false)
					i++
				}
			}
		}
	case "select_stmt":
		for _, ch := range n.Children {
			if ch.Type == "select_case" {
				arm := newSymbolScope(scope)
				var bind *syntax.Node
				for _, p := range ch.Children {
					if p.Type == "NAME" && bind == nil {
						bind = p
					}
					if p.Type == "expr" {
						idx.walk(p, scope, false)
					}
				}
				if bind != nil {
					idx.addDef(arm, bind, "binding", false)
				}
				if b := ch.ChildByType("block"); b != nil {
					idx.walk(b, arm, false)
				}
			} else if ch.Type == "select_default" {
				if b := ch.ChildByType("block"); b != nil {
					idx.walk(b, newSymbolScope(scope), false)
				}
			}
		}
	case "postfix_expr":
		for _, ch := range n.Children {
			idx.walk(ch, scope, false)
		}
	case "access":
		// Field names are not lexical references.
	case "NAME":
		idx.addRef(scope, n)
	default:
		for _, ch := range n.Children {
			idx.walk(ch, scope, false)
		}
	}
}
func (idx *symbolIndex) bindNames(n *syntax.Node, scope *symbolScope, kind string) {
	if n == nil {
		return
	}
	if n.Type == "NAME" {
		idx.addDef(scope, n, kind, false)
		return
	}
	for _, ch := range n.Children {
		idx.bindNames(ch, scope, kind)
	}
}
func (idx *symbolIndex) bindPattern(n *syntax.Node, scope *symbolScope) {
	if n == nil {
		return
	}
	if n.Type == "NAME" && n.Value != "_" {
		idx.addDef(scope, n, "pattern", false)
		return
	}
	for _, ch := range n.Children {
		idx.bindPattern(ch, scope)
	}
}
