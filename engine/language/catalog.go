package language

import "aiki/engine/semantics/value"

// HelpEntry is authored help/documentation projected into language services.
type HelpEntry struct {
	Name     string
	Template string
	Summary  string
	Doc      string
}

// Catalog supplies workspace/runtime facts needed by language services without
// coupling the service core to a concrete host substrate.
type Catalog interface {
	VisibleNames(scope value.Scope) []string
	Help(name string) (HelpEntry, bool)
	ModuleSource(currentFile, moduleName string) (path string, source string, ok bool)
}
