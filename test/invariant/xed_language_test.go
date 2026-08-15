package invariant

import (
	"encoding/xml"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"aiki/engine/syntax"
	"aiki/engine/syntax/grammar"
)

type xedLanguage struct {
	Contexts []xedContext `xml:"definitions>context"`
}
type xedContext struct {
	ID       string   `xml:"id,attr"`
	Keywords []string `xml:"keyword"`
	Match    string   `xml:"match"`
}

func TestXedLexicalInventoryMatchesGrammar(t *testing.T) {
	data, err := os.ReadFile("../../extra/editors/xed/aiki.lang")
	if err != nil {
		t.Fatal(err)
	}
	var spec xedLanguage
	if err := xml.Unmarshal(data, &spec); err != nil {
		t.Fatal(err)
	}
	contexts := make(map[string]xedContext)
	for _, c := range spec.Contexts {
		contexts[c.ID] = c
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
	gotKeywords := append([]string(nil), contexts["keywords"].Keywords...)
	sort.Strings(gotKeywords)
	sort.Strings(wantKeywords)
	if !reflect.DeepEqual(gotKeywords, wantKeywords) {
		t.Fatalf("Xed keywords=%v grammar=%v", gotKeywords, wantKeywords)
	}

	// GtkSourceView punctuation highlighting is expressed as a regex. Require
	// every grammar operator to be represented literally by the maintained
	// regex; reject old operators that are no longer in the grammar.
	match := contexts["operators"].Match
	for _, op := range wantOperators {
		if !operatorRegexMentions(match, op) {
			t.Errorf("Xed operator regex does not represent grammar operator %q: %q", op, match)
		}
	}
	for _, stale := range []string{"%", "!", "==", "!=", "&&", "||"} {
		if operatorRegexMentions(match, stale) {
			t.Errorf("Xed operator regex retains stale operator %q: %q", stale, match)
		}
	}
}

func operatorRegexMentions(pattern, op string) bool {
	switch op {
	case "|>":
		return strings.Contains(pattern, "\\|>")
	case "<=":
		return strings.Contains(pattern, "<=")
	case ">=":
		return strings.Contains(pattern, ">=")
	case ".":
		return strings.Contains(pattern, ".")
	default:
		// Single-character operators are kept in the final character class.
		start := strings.LastIndex(pattern, "[")
		end := strings.LastIndex(pattern, "]")
		return start >= 0 && end > start && strings.Contains(pattern[start+1:end], op)
	}
}
