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
	return ValidateSources(g, Source, HelpSource, DocSource)
}

// ValidateSources validates a prelude source/help/doc triple through the same
// catalog construction path used by the runtime. It exists so architectural
// invariant tests can deliberately mutate one authority without duplicating the
// join logic.
func ValidateSources(g *grammar.Grammar, source, helpSource, docSource string) (*Catalog, error) {
	info, err := modules.AnalyzeSource(g, "engine/runtime/prelude/prelude.ai", source)
	if err != nil {
		return nil, fmt.Errorf("analyzing prelude source: %w", err)
	}

	funcs, err := help.ParseHelpFile("prelude.help", helpSource)
	if err != nil {
		return nil, err
	}
	docs, err := help.ParseDocFile("prelude.doc", docSource)
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
	if err := validateHelpCoverage(declared, funcs); err != nil {
		return nil, err
	}
	if err := validateDocCoverage(declared, docs); err != nil {
		return nil, err
	}
	sort.Strings(names)

	registry := help.NewRegistry()
	registry.Merge(funcs, docs)
	return &Catalog{Registry: registry, Names: names}, nil
}

func validateHelpCoverage(declared map[string]bool, entries map[string]help.FuncEntry) error {
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

func validateDocCoverage(declared map[string]bool, entries map[string]help.DocEntry) error {
	var problems []string
	for name := range declared {
		if _, ok := entries[name]; !ok {
			problems = append(problems, fmt.Sprintf("missing doc for '%s'", name))
		}
	}
	for name := range entries {
		if !declared[name] {
			problems = append(problems, fmt.Sprintf("orphan doc entry '%s'", name))
		}
	}
	if len(problems) == 0 {
		return nil
	}
	sort.Strings(problems)
	return fmt.Errorf("prelude doc mismatch:\n  %s", strings.Join(problems, "\n  "))
}
