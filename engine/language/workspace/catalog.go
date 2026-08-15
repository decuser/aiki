// Package workspace adapts the Go runtime/module workspace to the neutral
// language.Catalog contract.
package workspace

import (
	"os"
	"path/filepath"
	"strings"

	"aiki/engine/language"
	"aiki/engine/runtime/hal/substrate"
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

func (c *Catalog) BuiltinNames(scope value.Scope) []string {
	return c.runtime.BuiltinNames(scope)
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
