// Package workspace adapts the Go runtime/module workspace to the neutral
// language.Catalog contract.
package workspace

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"aiki/engine/language"
	"aiki/engine/runtime/hal/substrate"
	"aiki/engine/runtime/help"
	"aiki/engine/runtime/prelude"
	"aiki/engine/semantics/value"
	"aiki/engine/syntax/grammar"
)

var _ language.Catalog = (*Catalog)(nil)

type Catalog struct {
	grammar  *grammar.Grammar
	runtime  *substrate.GoRuntime
	registry *substrate.ModuleRegistry
}

func NewCatalog(g *grammar.Grammar) *Catalog {
	return &Catalog{grammar: g, runtime: substrate.NewGoRuntime()}
}

func (c *Catalog) VisibleNames(scope value.Scope) []string {
	set := map[string]bool{}
	for _, name := range c.runtime.BuiltinNames(scope) {
		set[name] = true
	}
	if scope == value.ScopeUser {
		funcs, err := help.ParseHelpFile("prelude.help", prelude.HelpSource)
		if err == nil {
			for name := range funcs {
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
	funcs, err := help.ParseHelpFile("prelude.help", prelude.HelpSource)
	if err != nil {
		return language.HelpEntry{}, false
	}
	entry, ok := funcs[name]
	if !ok {
		return language.HelpEntry{}, false
	}
	docs, _ := help.ParseDocFile("prelude.doc", prelude.DocSource)
	doc := ""
	if d, found := docs[name]; found {
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
	if substrate.IsPathImport(name) {
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
		r := substrate.NewModuleRegistry(substrate.DefaultModuleRoots(homeDir))
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
