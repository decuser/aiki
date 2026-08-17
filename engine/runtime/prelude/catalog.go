package prelude

import (
	"fmt"
	"sort"
	"strings"

	"aiki/engine/runtime/help"
	"aiki/engine/runtime/modules"
	"aiki/engine/syntax/grammar"
)

// Catalog is the engine-owned derived view of the authored prelude source,
// help, and documentation artifacts. The artifacts remain authoritative; the
// catalog joins them by procedure identity for runtime and tooling consumers.
type Catalog struct {
	Registry *help.Registry
	Names    []string
}

// LoadCatalog derives and validates the prelude surface from the real Aiki
// source plus its authored help and documentation.
func LoadCatalog(g *grammar.Grammar) (*Catalog, error) {
	info, err := modules.AnalyzeSource(g, "engine/runtime/prelude/prelude.ai", Source)
	if err != nil {
		return nil, fmt.Errorf("analyzing prelude source: %w", err)
	}

	funcs, err := help.ParseHelpFile("prelude.help", HelpSource)
	if err != nil {
		return nil, err
	}
	docs, err := help.ParseDocFile("prelude.doc", DocSource)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(info.Functions))
	declared := make(map[string]bool, len(info.Functions))
	for name := range info.Functions {
		if strings.HasPrefix(name, "_") {
			continue
		}
		declared[name] = true
		names = append(names, name)
	}
	if err := validateCoverage(declared, funcs); err != nil {
		return nil, err
	}
	sort.Strings(names)

	registry := help.NewRegistry()
	registry.Merge(funcs, docs)
	return &Catalog{Registry: registry, Names: names}, nil
}

func validateCoverage(declared map[string]bool, entries map[string]help.FuncEntry) error {
	var problems []string
	for name := range declared {
		if _, ok := entries[name]; !ok {
			problems = append(problems, fmt.Sprintf("missing help for '%s'", name))
		}
	}
	for name := range entries {
		if !declared[name] {
			problems = append(problems, fmt.Sprintf("orphan help entry '%s'", name))
		}
	}
	if len(problems) == 0 {
		return nil
	}
	sort.Strings(problems)
	return fmt.Errorf("prelude help mismatch:\n  %s", strings.Join(problems, "\n  "))
}
