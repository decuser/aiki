package nativeffi

import (
	"os"
	"path/filepath"
	"testing"

	"aiki/engine/syntax"
	"aiki/engine/syntax/grammar"
)

func TestNeutralContractSourcesParse(t *testing.T) {
	g, err := grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
	if err != nil {
		t.Fatal(err)
	}
	files, err := filepath.Glob("*_contract.ai")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("no native/FFI contract sources found")
	}
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		lx := syntax.NewLexer(g, file, string(data), nil)
		tokens, err := lx.Tokenize()
		if err != nil {
			t.Errorf("%s lex: %v", file, err)
			continue
		}
		p := syntax.NewParser(g, tokens, string(data), nil)
		if _, err := p.Parse(); err != nil {
			t.Errorf("%s parse: %v", file, err)
		}
	}
}
