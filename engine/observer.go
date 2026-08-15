package engine

// Observer receives events during lexing, parsing, evaluation, and formatting.
// All parameters are primitives to avoid coupling to value types.
type Observer interface {
	OnLex(token string, lexeme string, pos Position)
	OnParse(production string, depth int, pos Position)
	OnEval(node string, result string, scope int, pos Position)
	OnEffect(action string, target string, pos Position)
	OnFormat(method string, output string, node string, depth int)
}

// DiagnosticObserver is an optional observer extension for non-fatal engine
// diagnostics. Components use a type assertion so diagnostic observability can
// be added without making it part of the core Observer contract.
type DiagnosticObserver interface {
	OnDiagnostic(kind string, message string, pos Position)
}
