package invariant

import (
	"strings"
	"testing"

	"aiki/engine/runtime/prelude"
	"aiki/engine/syntax"
	"aiki/engine/syntax/grammar"
)

func invariantGrammar(t *testing.T) *grammar.Grammar {
	t.Helper()
	g, err := grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
	if err != nil {
		t.Fatalf("loading grammar: %v", err)
	}
	return g
}

func TestPreludeCatalogJoinsSourceHelpAndDocs(t *testing.T) {
	catalog, err := prelude.LoadCatalog(invariantGrammar(t))
	if err != nil {
		t.Fatalf("LoadCatalog: %v", err)
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

func TestPreludeInvariantRejectsMissingHelp(t *testing.T) {
	helpSource := strings.Replace(prelude.HelpSource,
		"@func truncate\n@template \"truncate(n)\"\n@help \"Truncate number toward zero, returning integer part.\"\n", "", 1)
	if helpSource == prelude.HelpSource {
		t.Fatal("fixture did not remove truncate help")
	}
	_, err := prelude.ValidateSources(invariantGrammar(t), prelude.Source, helpSource, prelude.DocSource)
	if err == nil || !strings.Contains(err.Error(), "missing help for 'truncate'") {
		t.Fatalf("expected missing help failure, got %v", err)
	}
}

func TestPreludeInvariantRejectsMissingDoc(t *testing.T) {
	docSource := strings.Replace(prelude.DocSource,
		"truncate\nTruncates a number toward zero, returning its integer part.\n\ntruncate(n)\n\ntruncate(2.7)     # 2\ntruncate(-2.7)    # -2\ntruncate(0)       # 0\n===\n", "", 1)
	if docSource == prelude.DocSource {
		t.Fatal("fixture did not remove truncate doc")
	}
	_, err := prelude.ValidateSources(invariantGrammar(t), prelude.Source, prelude.HelpSource, docSource)
	if err == nil || !strings.Contains(err.Error(), "missing doc for 'truncate'") {
		t.Fatalf("expected missing doc failure, got %v", err)
	}
}
