package workspace

import (
	"testing"

	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
	"aiki/engine/syntax/grammar"
)

func testCatalog(t *testing.T) *Catalog {
	t.Helper()
	g, err := grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
	if err != nil {
		t.Fatal(err)
	}
	return NewCatalog(g)
}

func TestVisibleNamesIncludesAuthoredPreludeSurface(t *testing.T) {
	c := testCatalog(t)
	seen := map[string]bool{}
	for _, name := range c.VisibleNames(value.ScopeUser) {
		seen[name] = true
	}
	for _, want := range []string{"print", "length", "inspect", "import", "use", "export"} {
		if !seen[want] {
			t.Fatalf("visible names missing %q", want)
		}
	}
}

func TestHelpProjectsPreludeAuthority(t *testing.T) {
	c := testCatalog(t)
	h, ok := c.Help("print")
	if !ok {
		t.Fatal("missing print help")
	}
	if h.Template == "" || h.Summary == "" {
		t.Fatalf("incomplete print help: %#v", h)
	}
	if _, ok := c.Help("definitely_missing"); ok {
		t.Fatal("unexpected help for missing name")
	}
}
