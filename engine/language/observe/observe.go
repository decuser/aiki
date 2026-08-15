// Package observe defines neutral language-service observation and
// instrumentation contracts. It imports no language-service implementation
// package so service internals can depend downward on these contracts.
package observe

// EventKind names observable language-service work.
type EventKind string

const (
	EventDiagnosticsRequested EventKind = "diagnostics_requested"
	EventDiagnosticProduced   EventKind = "diagnostic_produced"
)

// Event describes service work without exposing parser or protocol structs.
type Event struct {
	Kind       EventKind
	DocumentID string
	Detail     string
}

// Observer receives language-service events. Observers must not affect
// service results.
type Observer interface {
	ObserveLanguage(Event)
}

// Metric names a measured service operation.
type Metric string

const (
	MetricDiagnosticsRequest Metric = "diagnostics_request"
	MetricLexRun             Metric = "lex_run"
	MetricParseRun           Metric = "parse_run"
	MetricAnalysisRun        Metric = "analysis_run"
	MetricDiagnostic         Metric = "diagnostic"
)

// Probe receives service measurements. Implementations used by concurrent
// clients must be safe for concurrent use.
type Probe interface {
	HitLanguage(Metric, int64)
}
