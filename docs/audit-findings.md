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
| AF-017 | Prelude help covers the callable surface, but full documentation is not 1:1; `truncate` currently has help with no `.doc` entry. | OPEN | Discovered during Gate 2 authority centralization. Do not silently strengthen the runtime invariant inside a behavior-preserving refactor. |

## Current disposition

`proposals/post-grammar-hardening.md` resolved AF-001, AF-005, AF-006, AF-007,
AF-014, AF-015, and AF-016. All other findings retain the disposition shown
above and remain visible for future projects.
