# Cut 2 — Correct failure semantics

Status: **GATED (environment-limited)**

## Discovery

Investigation established two independent defects:

1. recursive fmt printed parse failures but discarded them;
2. lint's format-first preflight returned an error internally, but top-level CLI
   dispatch discarded lint's return code. The same dispatch defect affected fmt.

Lint printed only one negative fixture because `CheckFormatting` aborted its
walk at the first parse error; the actual lint AST walk never began.

## Result

- Recursive fmt accumulates ordinary malformed-file errors and returns them.
- Lint formatting preflight accumulates ordinary malformed-file errors across
  the walk instead of stopping at the first one.
- Tests use temporary malformed files; no permanent undeclared-invalid fixture
  was added to the repository.
- fmt/lint `Run` results are propagated by the executable.

Declared parse-negative fixtures were already in place from Cut 1, so making
ordinary failures fatal does not intentionally redden the repository gate.
