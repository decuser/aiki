package grammar

import (
	"fmt"
	"strings"
)

// Load parses the EBNFX grammar and help file, merges them, and validates 1:1 match.
func Load(ebnfxFile, ebnfxSource, helpFile, helpSource string) (*Grammar, error) {
	parser := NewParser(ebnfxFile, ebnfxSource)
	g, err := parser.Parse()
	if err != nil {
		return nil, fmt.Errorf("parsing %s: %w", ebnfxFile, err)
	}

	helpEntries, err := ParseHelp(helpFile, helpSource)
	if err != nil {
		return nil, fmt.Errorf("parsing %s: %w", helpFile, err)
	}

	if err := mergeHelp(g, helpEntries); err != nil {
		return nil, err
	}

	if err := validate(g, helpEntries); err != nil {
		return nil, err
	}

	return g, nil
}

type HelpEntry struct {
	Name string
	Doc  string
}

func ParseHelp(file, source string) (map[string]HelpEntry, error) {
	entries := make(map[string]HelpEntry)
	parts := strings.Split(source, "\n===\n")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		lines := strings.SplitN(part, "\n", 2)
		name := strings.TrimSpace(lines[0])
		if name == "" {
			continue
		}
		var doc string
		if len(lines) > 1 {
			doc = strings.TrimSpace(lines[1])
		}
		if _, exists := entries[name]; exists {
			return nil, fmt.Errorf("%s: duplicate entry for '%s'", file, name)
		}
		entries[name] = HelpEntry{Name: name, Doc: doc}
	}
	return entries, nil
}

func mergeHelp(g *Grammar, help map[string]HelpEntry) error {
	for name, prod := range g.Productions {
		if entry, ok := help[name]; ok {
			prod.Meta.Doc = entry.Doc
			g.Productions[name] = prod
		}
	}
	for i, tok := range g.Tokens {
		if entry, ok := help[tok.Name]; ok {
			g.Tokens[i].Meta.Doc = entry.Doc
		}
	}
	return nil
}

func validate(g *Grammar, help map[string]HelpEntry) error {
	var errors []string
	grammarNames := make(map[string]bool)
	for name := range g.Productions {
		grammarNames[name] = true
	}
	for _, tok := range g.Tokens {
		grammarNames[tok.Name] = true
	}

	for name := range grammarNames {
		if _, ok := help[name]; !ok {
			errors = append(errors, fmt.Sprintf("missing help for '%s'", name))
		}
	}
	for name := range help {
		if !grammarNames[name] {
			errors = append(errors, fmt.Sprintf("orphan help entry '%s'", name))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("grammar/help mismatch:\n  %s", strings.Join(errors, "\n  "))
	}
	return nil
}
