package substrate

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"aiki/engine/runtime/help"
	"aiki/engine/runtime/libpath"
	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
	"aiki/engine/syntax/grammar"
)

// ModuleHelp holds help and doc data for a module.
type ModuleHelp struct {
	Funcs map[string]help.FuncEntry // function name -> help entry
	Docs  map[string]help.DocEntry  // function name -> doc entry
}

// ModuleRegistry maps package names to their file paths and caches loaded modules.
type ModuleRegistry struct {
	paths    map[string]string        // package name -> file path
	modules  map[string]*value.Module // package name -> loaded module
	helpData map[string]*ModuleHelp   // package name -> help data
	roots    []string                 // directories to scan
	seen     map[string]bool          // absolute paths already scanned
}

// GlobalRegistry is the module registry used by import.
var GlobalRegistry *ModuleRegistry

// DistributionModuleRoots returns the directories, relative to a distribution
// root, that hold shipped modules. Everything under them is subject to the
// documentation invariants, including vendored modules: a module that ships
// inside the distribution appears in help() and doc() like any other, so a
// user cannot tell which modules are held to the standard.
//
// The roots import also scans - the working directory and the user library -
// are deliberately outside this set. A developer's own modules are their own
// business.
func DistributionModuleRoots() []string {
	return []string{"lib", "vendor"}
}

// DefaultModuleRoots returns every directory import scans, in order. It is the
// distribution roots plus the working directory and the user library, neither
// of which belongs to the distribution.
func DefaultModuleRoots(homeDir string) []string {
	roots := []string{"."}
	roots = append(roots, DistributionModuleRoots()...)
	if homeDir != "" {
		roots = append(roots, filepath.Join(homeDir, ".aiki", "lib"))
	}
	return roots
}

// NewModuleRegistry creates a new registry with the given roots.
func NewModuleRegistry(roots []string) *ModuleRegistry {
	return &ModuleRegistry{
		paths:    make(map[string]string),
		modules:  make(map[string]*value.Module),
		helpData: make(map[string]*ModuleHelp),
		roots:    roots,
		seen:     make(map[string]bool),
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

		// Load sibling .help and .doc files (required for lib packages)
		if err := r.loadModuleHelp(pkgName, path); err != nil {
			return err
		}
	}

	return nil
}

// loadModuleHelp loads .help and .doc files adjacent to the .ai file.
// For lib packages, both files are required.
func (r *ModuleRegistry) loadModuleHelp(pkgName, aiPath string) error {
	// Derive sibling paths: foo.ai -> foo.help, foo.doc
	base := strings.TrimSuffix(aiPath, ".ai")
	helpPath := base + ".help"
	docPath := base + ".doc"

	mh := &ModuleHelp{
		Funcs: make(map[string]help.FuncEntry),
		Docs:  make(map[string]help.DocEntry),
	}

	// Check if this is a lib package (requires both files)
	requireDocs := libpath.IsBlessedLibPath(aiPath)

	// Load .help file
	helpData, helpErr := os.ReadFile(helpPath)
	if helpErr != nil {
		if requireDocs {
			return fmt.Errorf("package '%s' missing required %s", pkgName, helpPath)
		}
	} else {
		funcs, err := help.ParseHelpFile(helpPath, string(helpData))
		if err != nil {
			return fmt.Errorf("package '%s': %v", pkgName, err)
		}
		mh.Funcs = funcs
	}

	// Load .doc file
	docData, docErr := os.ReadFile(docPath)
	if docErr != nil {
		if requireDocs {
			return fmt.Errorf("package '%s' missing required %s", pkgName, docPath)
		}
	} else {
		docs, err := help.ParseDocFile(docPath, string(docData))
		if err != nil {
			return fmt.Errorf("package '%s': %v", pkgName, err)
		}
		mh.Docs = docs
	}

	// Store help data
	r.helpData[pkgName] = mh
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

// ListPackages returns all registered package names, sorted.
func (r *ModuleRegistry) ListPackages() []string {
	names := make([]string, 0, len(r.paths))
	for name := range r.paths {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// GetModuleHelp returns help data for a package, or nil if none.
func (r *ModuleRegistry) GetModuleHelp(pkgName string) *ModuleHelp {
	return r.helpData[pkgName]
}

// HasPackage returns true if the package is registered.
func (r *ModuleRegistry) HasPackage(name string) bool {
	_, ok := r.paths[name]
	return ok
}

// IsPathImport returns true if the import string looks like a path.
func IsPathImport(name string) bool {
	return strings.HasPrefix(name, "./") || strings.HasPrefix(name, "../")
}
