package prelude

import (
	"testing"

	"aiki/engine/syntax"
	"aiki/engine/syntax/grammar"
)

func TestCatalogJoinsPreludeSourceHelpAndDocs(t *testing.T) {
	g, err := grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
	if err != nil {
		t.Fatalf("loading grammar: %v", err)
	}
	catalog, err := LoadCatalog(g)
	if err != nil {
		t.Fatalf("LoadCatalog: %v", err)
	}
	if len(catalog.Names) == 0 {
		t.Fatal("prelude catalog has no names")
	}
	for _, name := range catalog.Names {
		if catalog.Registry.GetHelp(name) == nil {
			t.Errorf("%s missing help", name)
		}
	}
	if len(catalog.Registry.Docs) == 0 {
		t.Fatal("prelude catalog has no documentation entries")
	}
}
