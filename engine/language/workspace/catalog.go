// Package workspace adapts the engine module workspace to the neutral
// language.Catalog contract.
package workspace

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"aiki/engine/language"
	"aiki/engine/runtime/modules"
	"aiki/engine/runtime/prelude"
	"aiki/engine/runtime/primitives"
	"aiki/engine/semantics/value"
	"aiki/engine/syntax/grammar"
)

var _ language.Catalog = (*Catalog)(nil)

type Catalog struct {
	grammar        *grammar.Grammar
	registry       *modules.ModuleRegistry
	preludeCatalog *prelude.Catalog
}

func NewCatalog(g *grammar.Grammar) *Catalog {
	return &Catalog{grammar: g}
}

func (c *Catalog) authoredPrelude() *prelude.Catalog {
	if c.preludeCatalog != nil {
		return c.preludeCatalog
	}
	catalog, err := prelude.LoadCatalog(c.grammar)
	if err != nil {
		return nil
	}
	c.preludeCatalog = catalog
	return catalog
}

func (c *Catalog) VisibleNames(scope value.Scope) []string {
	set := map[string]bool{}
	for _, name := range primitives.NamesForScope(scope) {
		set[name] = true
	}
	if scope == value.ScopeUser {
		if catalog := c.authoredPrelude(); catalog != nil {
			for _, name := range catalog.Names {
				set[name] = true
			}
		}
	}
	out := make([]string, 0, len(set))
	for name := range set {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

func (c *Catalog) Help(name string) (language.HelpEntry, bool) {
	catalog := c.authoredPrelude()
	if catalog == nil {
		return language.HelpEntry{}, false
	}
	entry := catalog.Registry.GetHelp(name)
	if entry == nil {
		return language.HelpEntry{}, false
	}
	doc := ""
	if d := catalog.Registry.GetDoc(name); d != nil {
		doc = d.Doc
	}
	return language.HelpEntry{Name: name, Template: entry.Template, Summary: entry.Help, Doc: doc}, true
}

func (c *Catalog) ModuleSource(currentFile, name string) (string, string, bool) {
	path := c.resolveModulePath(currentFile, name)
	if path == "" {
		return "", "", false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", "", false
	}
	return path, string(data), true
}

func (c *Catalog) resolveModulePath(currentFile, name string) string {
	if modules.IsPathImport(name) {
		path := name
		if !strings.HasSuffix(path, ".ai") {
			path += ".ai"
		}
		if currentFile != "" && currentFile != "<unknown>" {
			candidate := filepath.Join(filepath.Dir(currentFile), path)
			if _, err := os.Stat(candidate); err == nil {
				return candidate
			}
		}
		if _, err := os.Stat(path); err == nil {
			return path
		}
		return ""
	}
	if c.registry == nil {
		homeDir, _ := os.UserHomeDir()
		r := modules.NewModuleRegistry(modules.DefaultModuleRoots(homeDir))
		if err := r.Scan(c.grammar); err != nil {
			return ""
		}
		c.registry = r
	}
	path, _, ok := c.registry.Resolve(name)
	if !ok {
		return ""
	}
	return path
}
