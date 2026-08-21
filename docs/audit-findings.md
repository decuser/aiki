# Engineering Audit Findings

This ledger tracks substantive findings from code audits independently of any
single proposal or session. Stable IDs let later projects close, accept, or
defer findings without reconstructing the review history from chat or commit
messages.

Status vocabulary:

- `OPEN` — credible finding with no disposition yet.
- `ACTIVE` — being addressed by the current bounded project.
- `RESOLVED` — implementation or documentation now closes the finding.
- `ACCEPTED` — understood limitation or tradeoff; no change planned now.
- `DEFERRED` — valid work item intentionally left for a later project.

| ID | Finding | Status | Disposition / project |
|---|---|---|---|
| AF-001 | EBNF `parseIdentifier` assumes ASCII bytes when classifying identifier characters. | RESOLVED | Document the current ASCII assumption in post-grammar hardening closeout. |
| AF-002 | `parseDecorator` advances raw byte positions inside quoted decorator values without maintaining line/column state. | DEFERRED | Separate parser-positioning fix; not part of exceptional-path hardening. |
| AF-003 | Newline analysis creates fresh nullable visiting maps inside FIRST/LAST/CONT traversal. | ACCEPTED | Nullable results are memoized and the current grammar is small; optimize only with evidence. |
| AF-004 | Continuation analysis rejects directly recursive productions rather than solving a recursive fixed point. | ACCEPTED | Failure is loud and safe; Aiki's current expression grammar is non-left-recursive. |
| AF-005 | Parser newline suppression depth can become negative after an unmatched closer, distorting later suppression. | RESOLVED | Post-grammar hardening Cut 1. |
| AF-006 | A failed cached newline analysis silently removes the refined continuation diagnostic from parsing. | RESOLVED | Post-grammar hardening Cut 1: make degradation observable without making parsing unavailable. |
| AF-007 | Evaluator `syntheticHandlerNodes` is a manual parser/evaluator exception list with no source cross-reference. | RESOLVED | Document parser synthesis site in closeout; coverage already makes drift loud. |
| AF-008 | Evaluator terminal-alternative extraction assumed a narrow BINOP expression shape. | RESOLVED | Centralized grammar analysis now derives terminal alternatives structurally. |
| AF-009 | Formatter `printKind` dispatch is mechanically verbose. | ACCEPTED | Direct method-function table caused an initialization cycle; explicit kind dispatch is safe and inspectable. |
| AF-010 | Formatter previously could ignore a `printProgram` error before round-trip verification. | RESOLVED | Fixed during grammar-authority work. |
| AF-011 | Evaluator, lint tests, parser diagnostics, and enginesmoke independently walked/reanalyzed grammar structure. | RESOLVED | Centralized grammar analysis: parse once, derive shared facts once, consume everywhere. |
| AF-012 | Lint's Go-AST test for node-type string literals can miss indirect constants/helpers. | DEFERRED | Current lint uses direct literals; revisit if lint type dispatch becomes indirect. |
| AF-013 | `IsParseNegative` may reopen the same source multiple times during recursive tooling passes. | ACCEPTED | Redundant I/O is immaterial at current tree size; optimize only if measured. |
| AF-014 | Lint path expansion repeats negative-fixture filtering in multiple branches. | RESOLVED | Post-grammar hardening Cut 2. |
| AF-015 | `break_call_smoke` and `break_unary_smoke` intentionally have empty golds but source does not explain why. | RESOLVED | Add fixture comments in closeout. |
| AF-016 | `# @negative parse` can currently appear in arbitrary `.ai` source and exempt it from fmt/lint without smoke pairing. | RESOLVED | Post-grammar hardening Cut 2: constrain declarations to recognized smoke-fixture context. |
| AF-017 | Prelude help covers the callable surface, but full documentation is not 1:1; `truncate` currently has help with no `.doc` entry. | RESOLVED | Gate 2 completion + Gate 3: symmetric prelude help/doc coverage and negative assurance; `truncate` documented. |
| AF-018 | `_seed` and `_random` were classified as host primitives but had no canonical HAL identities. | RESOLVED | Promoted to `HAL.random.seed` and `HAL.random.below`; added random capability, Go provenance, host registration, and reverse host-role invariant. |
| AF-019 | HAL descriptor vocabularies were free-form strings checked only for presence. | RESOLVED | HAL metadata validator now enforces canonical context/effect/blocking/lifetime/optionality/error vocabularies with negative mutation tests. |
| AF-020 | `spawn` fell back to ambient `os.Stderr` when asynchronous fault reporting was unavailable. | RESOLVED | `spawn` now requires an async-fault reporter before launch; runtime-ownership invariant includes the intrinsic implementation. |
| AF-021 | Portable-systems completeness audit found the runtime environment was read-only and child processes inherited ambient host environment rather than the Aiki runtime view. | RESOLVED | Phase 6 remediation: runtime-owned `environ`/`set_env`/`unset_env`; `system.exec` and `process.start` inherit per-start snapshots. |
| AF-022 | Common `io` help/docs lagged the endpoint architecture and described only standard streams/files after process pipes and TCP became generic endpoints. | RESOLVED | Portable-systems final reconciliation: `io` help/docs now name runtime endpoints, process pipes, and TCP explicitly. |

| AF-023 | Portable systems completeness lacked deliberate program termination with an exit status; `quit()` is REPL control and does not provide script/process status semantics. | RESOLVED | Added `system.exit(code)` as evaluator control flow with portable 0..255 status, runner/CLI propagation, and cleanup-before-return regression coverage. |

| AF-024 | Current-facing docs retained pre-native/FFI terminology and deleted module paths after Alpha 2 integration. | RESOLVED | Alpha 2 documentation reconciliation: README/HAL/style/contributor/help surfaces aligned to stdlib policy and current paths. |
| AF-025 | `process.terminate` documentation still described portable signaling as a future phase after the signal module shipped. | RESOLVED | Alpha 2 documentation reconciliation: point to the current `signal` module. |
| AF-026 | `This Is Aiki` retained `commit XXX`, Alpha 1 release labeling, and a pre-systems/pre-native-FFI supplied-module inventory. | RESOLVED | Alpha 2 documentation reconciliation: guide scoped to `v0.4.0-alpha-34` / `8a33879`; RA1 explicitly retains Alpha 1 reference-implementation scope. |
| AF-027 | Several completed project proposals still reported ACTIVE or pending-validation states after their work had been integrated and validated. | RESOLVED | Alpha 2 documentation reconciliation: reconcile only proposals with durable completion evidence; proposed/deferred work remains unchanged. |
| AF-028 | `docs/debug.md` omitted the shipped `debug -stage fmt` formatter view. | RESOLVED | Alpha 2 documentation reconciliation: usage, flags, and formatter example now match the CLI. |
| AF-029 | `This Is Aiki` still described two direct external Go libraries after file locking added `github.com/gofrs/flock`. | RESOLVED | Alpha 2 documentation reconciliation: Alpha 2 guide now names all three direct host-facing dependencies and their roles. |

## Current disposition

`proposals/completed/post-grammar-hardening.md` resolved AF-001, AF-005, AF-006, AF-007,
AF-014, AF-015, and AF-016. All other findings retain the disposition shown
above and remain visible for future projects.
