package language

import "aiki/engine/semantics/value"

// Catalog supplies workspace/runtime facts needed by structural language
// analysis without coupling the service core to a concrete host substrate.
type Catalog interface {
	BuiltinNames(scope value.Scope) []string
	ModuleSource(currentFile, moduleName string) (path string, source string, ok bool)
}
