package substrate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
	"aiki/engine/syntax/grammar"
)

// ModuleRegistry maps package names to their file paths and caches loaded modules.
type ModuleRegistry struct {
	paths   map[string]string        // package name -> file path
	modules map[string]*value.Module // package name -> loaded module
	roots   []string                 // directories to scan
	seen    map[string]bool          // absolute paths already scanned
}

// GlobalRegistry is the module registry used by import.
var GlobalRegistry *ModuleRegistry

// NewModuleRegistry creates a new registry with the given roots.
func NewModuleRegistry(roots []string) *ModuleRegistry {
	return &ModuleRegistry{
		paths:   make(map[string]string),
		modules: make(map[string]*value.Module),
		roots:   roots,
		seen:    make(map[string]bool),
	}
}

// Scan scans all roots for package declarations.
func (r *ModuleRegistry) Scan(g *grammar.Grammar) error {
	for _, root := range r.roots {
		if err := r.scanDir(root, g); err != nil {
			// Skip non-existent directories
			if os.IsNotExist(err) {
				continue
			}
			return err
		}
	}
	return nil
}

func (r *ModuleRegistry) scanDir(dir string, g *grammar.Grammar) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())
		
		// Get absolute path to detect duplicates
		absPath, err := filepath.Abs(path)
		if err != nil {
			absPath = path
		}
		
		if r.seen[absPath] {
			continue
		}
		r.seen[absPath] = true
		
		if entry.IsDir() {
			// Recurse into subdirectories
			if err := r.scanDir(path, g); err != nil {
				return err
			}
			continue
		}

		// Only scan .ai files
		if !strings.HasSuffix(entry.Name(), ".ai") {
			continue
		}

		// Extract package name from file
		pkgName, err := r.extractPackageName(path, g)
		if err != nil {
			// Not a package file, skip
			continue
		}
		if pkgName == "" {
			continue
		}

		// Check for conflicts
		if existingPath, exists := r.paths[pkgName]; exists {
			return fmt.Errorf("duplicate package '%s': found in '%s' and '%s'", pkgName, existingPath, path)
		}

		r.paths[pkgName] = path
	}

	return nil
}

func (r *ModuleRegistry) extractPackageName(path string, g *grammar.Grammar) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	source := string(data)

	// Quick check: does it contain "package"?
	if !strings.Contains(source, "package") {
		return "", nil
	}

	// Parse to find package declaration
	lexer := syntax.NewLexer(g, path, source, nil)
	tokens, err := lexer.Tokenize()
	if err != nil {
		return "", err
	}

	parser := syntax.NewParser(g, tokens, source, nil)
	ast, err := parser.Parse()
	if err != nil {
		return "", err
	}

	// Find package_stmt in AST
	return findPackageName(ast), nil
}

func findPackageName(node *syntax.Node) string {
	if node == nil {
		return ""
	}

	if node.Type == "package_stmt" {
		for _, child := range node.Children {
			if child.Type == "STRING" {
				name := child.Value
				// Remove quotes
				if len(name) >= 2 && name[0] == '"' && name[len(name)-1] == '"' {
					return name[1 : len(name)-1]
				}
				return name
			}
		}
	}

	for _, child := range node.Children {
		if name := findPackageName(child); name != "" {
			return name
		}
	}

	return ""
}

// Lookup returns the file path for a package name.
func (r *ModuleRegistry) Lookup(name string) (string, bool) {
	path, ok := r.paths[name]
	return path, ok
}

// GetCached returns a cached module if available.
func (r *ModuleRegistry) GetCached(name string) (*value.Module, bool) {
	mod, ok := r.modules[name]
	return mod, ok
}

// Cache stores a loaded module.
func (r *ModuleRegistry) Cache(name string, mod *value.Module) {
	r.modules[name] = mod
}

// IsPathImport returns true if the import string looks like a path.
func IsPathImport(name string) bool {
	return strings.HasPrefix(name, "./") || strings.HasPrefix(name, "../")
}
