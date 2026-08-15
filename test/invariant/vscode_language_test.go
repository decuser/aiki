package invariant

import (
	"encoding/json"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"aiki/engine/syntax"
	"aiki/engine/syntax/grammar"
)

type textMateGrammar struct {
	Repository map[string]struct {
		Patterns []struct {
			Match string `json:"match"`
		} `json:"patterns"`
	} `json:"repository"`
}

func TestVSCodeLexicalInventoryMatchesGrammar(t *testing.T) {
	data, err := os.ReadFile("../../extra/editors/vscode/syntaxes/aiki.tmLanguage.json")
	if err != nil {
		t.Fatal(err)
	}
	var spec textMateGrammar
	if err := json.Unmarshal(data, &spec); err != nil {
		t.Fatal(err)
	}
	g, err := grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
	if err != nil {
		t.Fatal(err)
	}

	var wantKeywords, wantOperators []string
	for _, td := range g.Tokens {
		parts := strings.Fields(strings.TrimSpace(td.Literal))
		switch td.Name {
		case "KEYWORD":
			wantKeywords = append(wantKeywords, parts...)
		case "OPERATOR":
			wantOperators = append(wantOperators, parts...)
		}
	}
	keywordPatterns := spec.Repository["keyword"].Patterns
	if len(keywordPatterns) != 1 {
		t.Fatalf("VS Code keyword pattern count=%d, want 1", len(keywordPatterns))
	}
	gotKeywords := keywordAlternatives(keywordPatterns[0].Match)
	sort.Strings(gotKeywords)
	sort.Strings(wantKeywords)
	if !reflect.DeepEqual(gotKeywords, wantKeywords) {
		t.Fatalf("VS Code keywords=%v grammar=%v", gotKeywords, wantKeywords)
	}

	operatorPatterns := spec.Repository["operator"].Patterns
	if len(operatorPatterns) != 1 {
		t.Fatalf("VS Code operator pattern count=%d, want 1", len(operatorPatterns))
	}
	match := operatorPatterns[0].Match
	for _, op := range wantOperators {
		if !operatorRegexMentions(match, op) {
			t.Errorf("VS Code operator regex does not represent grammar operator %q: %q", op, match)
		}
	}
	for _, stale := range []string{"%", "!", "==", "!=", "&&", "||"} {
		if operatorRegexMentions(match, stale) {
			t.Errorf("VS Code operator regex retains stale operator %q: %q", stale, match)
		}
	}
}

func keywordAlternatives(pattern string) []string {
	start := strings.Index(pattern, "(?:")
	if start < 0 {
		return nil
	}
	start += len("(?:")
	end := strings.Index(pattern[start:], ")")
	if end < 0 {
		return nil
	}
	return strings.Split(pattern[start:start+end], "|")
}
