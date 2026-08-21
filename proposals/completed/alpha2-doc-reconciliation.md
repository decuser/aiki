# Aiki Alpha 2 Documentation Reconciliation

## Status
IMPLEMENTED — pending final repository validation

## Baseline
`v0.4.0-alpha-34` / `8a33879`

## Purpose
Reconcile all user-facing and durable repository documentation to the current Aiki architecture after HAL capability enforcement and native/FFI semantic separation.

## Rules
- Code, grammar, module policy, help/doc metadata, and executable behavior are authoritative.
- Do not preserve stale terminology merely for historical continuity in current-facing docs.
- Preserve historical records where they are explicitly historical; remove redundant scratch/prototype cruft.
- Native means a native Aiki realization; FFI means provider-backed realization; bare portable modules resolve native.
- HAL capability/authority and stdlib semantic capability are distinct layers.
- Host/runtime capabilities do not require fake native twins.
- Examples and docs must remain executable where repository invariants require it.

## Phases
1. Inventory and cruft removal.
2. README and top-level orientation.
3. Library `.doc` / `.help` reconciliation.
4. Architecture/developer documentation reconciliation.
5. Examples, experiments, proposals, and internal durable records.
6. Search-based stale-term sweep and executable documentation validation.
7. Final reconciliation for Alpha 2 release candidate.

## Acceptance
- Current-facing docs consistently describe the current module/HAL/native/FFI architecture.
- Removed APIs and obsolete paths are absent from current-facing documentation.
- Historical records remain clearly historical rather than masquerading as current instructions.
- `make validate` or stronger passes at final gate.

## Phase result

- Inventory found no disposable scratch/build cruft in the Alpha 34 baseline; historical experiment analyses, session records, and native/FFI evidence were retained as evidence rather than rewritten.
- README and current Markdown architecture/developer docs now describe Alpha 2, runtime-vs-provider primitive roles, stdlib semantic policy, and explicit native/FFI selection.
- Public library `.help` / `.doc` surfaces were reconciled against actual exports and semantic roles; no missing companion or exported-function documentation gaps remain.
- `This Is Aiki` now tracks Alpha 2 (`v0.4.0-alpha-34` / `8a33879`), while RA1 explicitly remains scoped to the Alpha 1 implementation baseline (`v0.4.0-alpha` / `678aeea`).
- Completed proposals were marked complete only where durable integration/validation evidence exists.
- Current-facing stale-term/path/link sweeps and ODT render checks are clean.
- Final repository validation remains the critical external gate.
