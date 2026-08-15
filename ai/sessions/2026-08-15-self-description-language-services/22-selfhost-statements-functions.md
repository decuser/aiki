# Milestone 22 — Self-host statements, closures, and calls

Status: GATED locally; authoritative repository gate pending.

Implemented statement/control-flow evaluation: `let`, assignment, interpreted shape declarations, `if`/`else`, `while`, `match`, list-pattern binding, blocks, and return wrappers. Extended the interpreted environment with shape metadata.

Implemented self-host closure calls, lexical capture, recursion, rest parameters, ordinary host-prelude calls, and pipelines. `selfhost/host_prelude.ai` explicitly installs real prelude function values into an interpreted root environment; an invariant couples its name inventory to `prelude.ai` so it cannot become a second vocabulary authority.

Disposable cross-implementation probes agree on assignment/control flow/shape access/pattern matching and on recursion/rest/pipeline cases, including a combined result of `141`.

Added `selfhost_evaluator_program_test.go` and `selfhost_prelude_bridge_test.go` for the authoritative gate.

## Follow-up — authoritative validation exposed formatter non-convergence

The first authoritative `make validate` after merging Cuts III.1–III.3 reached
`aiki lint ./...` and failed its format-idempotence precheck for
`selfhost/host_prelude.ai` and `selfhost/runtime.ai`. No structural lint
diagnostics were emitted.

Reproduction showed that the formatter's first pass could alter top-level
source-line spacing in a way that changed the blank-line decision made by a
second pass. `host_prelude.ai` lost an extra blank line on pass two, while
`runtime.ai` gained canonical separation between adjacent top-level function
bindings on pass two.

This is a formatter fixed-point defect exposed by the new Phase-III source,
not an evaluator defect. `FormatSourceWithObserver` now iterates a bounded
number of AST-preserving formatter passes and returns only a stable canonical
representation. A regression test covers the adjacent top-level declaration
patterns that exposed the defect.

Authoritative validation remains pending after this correction.
