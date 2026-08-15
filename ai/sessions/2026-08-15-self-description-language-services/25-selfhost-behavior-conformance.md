# Milestone 25 — Behavioral conformance

Status: GATED locally; authoritative repository gate pending.

## Finding and correction

The first behavior-gold sweep exposed a semantic bug: the self-host evaluator treated recoverable `[@error, ...]` values as halting conditions because it used `is_error()` for internal propagation. Aiki errors are ordinary recoverable values.

Internal evaluator halts now use the distinct `:self_fault` control shape. Recoverable Aiki errors remain values and can be bound, returned, and pattern-matched. `match_smoke.ai` then matched the reference behavior through its recoverable error cases.

The sweep also closed a qualified-pipeline omission: `|> list.filter(...)` now resolves the module export before applying the piped value.

## Acceptance corpus

`test/invariant/selfhost_behavior_acceptance_test.go` defines the durable Phase-III acceptance corpus using existing behavior files spanning:

- exact/left-to-right arithmetic;
- recoverable errors and match patterns;
- qualified pipelines;
- standard-library and HAL-backed modules;
- Unicode regex positions;
- relative imports;
- file effects;
- pure bytes and hash implementations.

Concurrency/select/spawn, debugger break fixtures, interactive input/error handling, and graphics remain explicitly outside the self-host proof scope established by the proposal.

## Additional evidence

Disposable direct comparisons also matched existing behavior gold output for functional closures, recursion, newline policy, string, bytes/ffi, hash/ffi and hash/native, regex, math/native, ord/chr, concatenation, HOF, iteration, pattern literals, hash utilities, and file I/O.
