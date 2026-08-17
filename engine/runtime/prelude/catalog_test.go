package prelude

import (
	"strings"
	"testing"

	"aiki/engine/syntax"
	"aiki/engine/syntax/grammar"
)

func loadTestGrammar(t *testing.T) *grammar.Grammar {
	t.Helper()
	g, err := grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
	if err != nil {
		t.Fatalf("loading grammar: %v", err)
	}
	return g
}

func TestCatalogJoinsPreludeSourceHelpAndDocs(t *testing.T) {
	catalog, err := LoadCatalog(loadTestGrammar(t))
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
		if catalog.Registry.GetDoc(name) == nil {
			t.Errorf("%s missing doc", name)
		}
	}
}

func TestCatalogRejectsMissingPreludeHelp(t *testing.T) {
	helpSource := strings.Replace(
		HelpSource,
		"@func truncate\n@template \"truncate(n)\"\n@help \"Truncate number toward zero, returning integer part.\"\n",
		"",
		1,
	)
	if helpSource == HelpSource {
		t.Fatal("test fixture did not remove truncate help")
	}

	_, err := loadCatalogFromSources(loadTestGrammar(t), Source, helpSource, DocSource)
	if err == nil {
		t.Fatal("expected missing prelude help to fail catalog loading")
	}
	if !strings.Contains(err.Error(), "missing help for 'truncate'") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCatalogRejectsMissingPreludeDoc(t *testing.T) {
	docSource := strings.Replace(
		DocSource,
		"truncate\nTruncates a number toward zero, returning its integer part.\n\ntruncate(n)\n\ntruncate(2.7)     # 2\ntruncate(-2.7)    # -2\ntruncate(0)       # 0\n===\n",
		"",
		1,
	)
	if docSource == DocSource {
		t.Fatal("test fixture did not remove truncate doc")
	}

	_, err := loadCatalogFromSources(loadTestGrammar(t), Source, HelpSource, docSource)
	if err == nil {
		t.Fatal("expected missing prelude doc to fail catalog loading")
	}
	if !strings.Contains(err.Error(), "missing doc for 'truncate'") {
		t.Fatalf("unexpected error: %v", err)
	}
}
