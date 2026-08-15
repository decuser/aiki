package lint

import (
	"aiki/engine/language"
	"aiki/engine/language/workspace"
	"aiki/engine/semantics/value"
	"aiki/engine/syntax/grammar"
)

// Diagnostic is the language-service diagnostic used by the lint CLI.
type Diagnostic = language.Diagnostic

// LintSource is the CLI compatibility seam over editor-independent structural
// language analysis. New consumers should use engine/language.Service.
func LintSource(g *grammar.Grammar, file string, source string, lintScope value.Scope) ([]Diagnostic, error) {
	return language.AnalyzeStructure(g, file, source, lintScope, workspace.NewCatalog(g))
}
