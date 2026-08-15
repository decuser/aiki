# Milestone 02 — Parser Exceptional Paths

**Status:** GATED (focused validation)

Resolved AF-005 by replacing aggregate newline suppression depth with a stack of
expected closing delimiters derived from the grammar rule. Unmatched/mismatched
closers no longer corrupt later suppression state.

Resolved AF-006 by adding the optional `engine.DiagnosticObserver` extension.
If cached newline analysis has `NewlineError`, `NewParser` reports
`grammar-newline-analysis` to diagnostic-aware observers while continuing to
construct the parser normally. Debug tracing implements the extension.

Added regression tests for both behaviors. `engine/syntax` and
`engine/syntax/grammar` pass in a disposable Go-1.23-compatible copy.
