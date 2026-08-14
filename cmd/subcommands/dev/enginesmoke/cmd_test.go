package enginesmoke

import (
	"testing"

	"aiki/engine/syntax"
	"aiki/engine/syntax/grammar"
)

func TestCollectProductionsCountsOnlyGrammarNodes(t *testing.T) {
	g := &grammar.Grammar{Productions: map[string]*grammar.Production{
		"program":   {Name: "program"},
		"statement": {Name: "statement"},
	}}
	ast := &syntax.Node{Type: "program", Children: []*syntax.Node{
		{Type: "statement", Children: []*syntax.Node{{Type: "NUMBER"}}},
	}}
	covered := map[string]bool{}
	collectProductions(ast, g, covered)
	if !covered["program"] || !covered["statement"] {
		t.Fatalf("expected grammar productions covered, got %#v", covered)
	}
	if covered["NUMBER"] {
		t.Fatalf("token/value node must not count as a grammar production")
	}
}
