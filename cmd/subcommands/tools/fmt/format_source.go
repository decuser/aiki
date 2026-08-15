package fmt

import (
	"aiki/engine"
	"aiki/engine/formatter"
	"aiki/engine/syntax/grammar"
)

// FormatSource formats valid Aiki source to the canonical style.
// The implementation lives in engine/formatter so command and language-service
// adapters share one formatting authority.
func FormatSource(g *grammar.Grammar, file, source string) (string, error) {
	return formatter.FormatSource(g, file, source)
}

// FormatSourceWithObserver formats source with the canonical formatter and an
// optional engine observer.
func FormatSourceWithObserver(g *grammar.Grammar, file, source string, observer engine.Observer) (string, error) {
	return formatter.FormatSourceWithObserver(g, file, source, observer)
}
