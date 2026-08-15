// Package language provides editor-independent Aiki language services.
// It depends inward on Aiki language authorities; protocol/editor adapters
// depend on this package, never the reverse.
package language

import (
	"errors"

	"aiki/engine"
	languageobserve "aiki/engine/language/observe"
	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
	"aiki/engine/syntax/grammar"
)

// Document is the neutral source unit consumed by language services.
type Document struct {
	ID      string
	Path    string
	Source  string
	Version int
}

// Service provides editor-independent queries over Aiki source.
type Service struct {
	grammar  *grammar.Grammar
	catalog  Catalog
	observer languageobserve.Observer
	probe    languageobserve.Probe
}

func NewService(g *grammar.Grammar, catalogs ...Catalog) *Service {
	var catalog Catalog
	if len(catalogs) > 0 {
		catalog = catalogs[0]
	}
	return &Service{grammar: g, catalog: catalog}
}

// SetObserver installs optional behavior-neutral service observation.
func (s *Service) SetObserver(observer languageobserve.Observer) { s.observer = observer }

// SetProbe installs optional service instrumentation.
func (s *Service) SetProbe(probe languageobserve.Probe) { s.probe = probe }

func (s *Service) observe(kind languageobserve.EventKind, doc Document, detail string) {
	if s.observer != nil {
		s.observer.ObserveLanguage(languageobserve.Event{Kind: kind, DocumentID: doc.ID, Detail: detail})
	}
}

func (s *Service) hit(metric languageobserve.Metric, n int64) {
	if s.probe != nil {
		s.probe.HitLanguage(metric, n)
	}
}

// Diagnostics returns lexical, parse, or structural findings for doc.
// Syntax failure stops structural analysis because there is no valid tree to
// inspect. Diagnostics use Aiki-native source positions.
func (s *Service) Diagnostics(doc Document, scope value.Scope) []Diagnostic {
	s.observe(languageobserve.EventDiagnosticsRequested, doc, "")
	s.hit(languageobserve.MetricDiagnosticsRequest, 1)
	file := doc.Path
	if file == "" {
		file = doc.ID
	}
	s.hit(languageobserve.MetricLexRun, 1)
	lx := syntax.NewLexer(s.grammar, file, doc.Source, nil)
	tokens, err := lx.Tokenize()
	if err != nil {
		d := diagnosticFromSyntaxError(err, "lex")
		s.diagnostic(doc, d)
		return []Diagnostic{d}
	}
	s.hit(languageobserve.MetricParseRun, 1)
	parser := syntax.NewParser(s.grammar, tokens, doc.Source, nil)
	node, err := parser.Parse()
	if err != nil {
		d := diagnosticFromSyntaxError(err, "parse")
		s.diagnostic(doc, d)
		return []Diagnostic{d}
	}
	s.hit(languageobserve.MetricAnalysisRun, 1)
	diagnostics := analyzeNode(s.grammar, file, node, scope, s.catalog)
	for _, d := range diagnostics {
		s.diagnostic(doc, d)
	}
	return diagnostics
}

func (s *Service) diagnostic(doc Document, d Diagnostic) {
	s.hit(languageobserve.MetricDiagnostic, 1)
	s.observe(languageobserve.EventDiagnosticProduced, doc, d.Source+":"+d.Severity)
}

func diagnosticFromSyntaxError(err error, fallback string) Diagnostic {
	var sourceErr *syntax.SourceError
	if errors.As(err, &sourceErr) {
		return Diagnostic{
			Pos:      sourceErr.Pos,
			Severity: "error",
			Source:   sourceErr.Kind,
			Message:  sourceErr.Message,
		}
	}
	return Diagnostic{
		Pos:      engine.Position{},
		Severity: "error",
		Source:   fallback,
		Message:  err.Error(),
	}
}
