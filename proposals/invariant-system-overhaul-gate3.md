# Proposal: Gate 3 — Invariant System Overhaul

## Status

COMPLETE — the dedicated invariant gate, negative assurance, HAL/engine authority checks, and representation guardrails are integrated into validation.

## Purpose

Separate architectural invariants from ordinary behavioral tests and make them a first-class validation system.

The goal is:

> Every architectural contract Aiki claims should have an executable invariant, and every critical invariant should itself be proven capable of failing.

## Build-System Contract

Add a dedicated target:

`make invariant`

This target runs architectural invariant checks only.

`make validate` depends on `make invariant`.

`make test` asks whether the implementation behaves correctly.

`make invariant` asks whether the architecture is still true.

`make validate` answers both.

## Scope

Invariant checks should cover architectural relationships including:

- HAL identity coherence;
- authority mappings;
- capability membership;
- runtime profile requirements;
- substrate bindings and provenance;
- layer boundaries;
- runtime ownership;
- grammar and evaluator coverage;
- prelude/help/doc completeness;
- module/export/doc coherence;
- primitive-role completeness;
- language-service derivation from engine authorities;
- formatter/AST structural expectations;
- exact-rational numeric representation;
- absence of forbidden float reintroduction;
- left-to-right evaluation rules where structural checks can support them.

Behavioral tests remain responsible for execution semantics.

## Separation

Ordinary tests continue to cover units, regressions, properties, examples, fuzzing, and runtime behavior.

Invariant tests cover architectural relationships, forbidden dependencies, authoritative metadata joins, required completeness, prohibited representations, and cross-layer contracts.

Some concepts may have both kinds of tests.

## HAL as First Full Specimen

The HAL should be the first subsystem fully expressed through the new invariant process.

The invariant system should prove:

`Aiki surface -> HAL identity -> authority -> capability membership -> runtime profile requirement -> substrate binding -> provenance`

It should also enforce that Aiki does not bypass HAL into substrate code, HAL metadata does not become dispatch, authority and availability remain independent, substrate identities do not leak upward, and runtime-owned state does not silently revert to ambient host globals.

## Invariant Inventory

Gate 3 begins by enumerating current architectural promises and classifying each as structural, semantic, or behavioral.

Only structural and architecture-facing semantic claims belong directly in `make invariant`.

Pure execution behavior remains in `make test`.

## Invariant Assurance

A passing invariant suite is not sufficient.

> An invariant is not considered protected until a test proves that violating it causes the invariant gate to fail.

Critical negative tests should exercise the same production invariant path used by the real gate.

Examples include missing prelude docs, unknown HAL operations in capabilities, incomplete primitive roles, grammar/evaluator mismatches, and forbidden float-backed numeric representations.

## Failure Quality

Invariant failures should identify the invariant, stable identity, and broken architectural relationship. Avoid opaque count-only failures where a more specific diagnostic is possible.

## Performance

`make invariant` must remain fast enough for frequent development use. Heavy fuzzing, long property runs, and broad runtime suites remain elsewhere.

## Phases

### Phase 1 — Inventory

Enumerate existing invariants, classify them, and identify missing or duplicated checks.

### Phase 2 — Dedicated Target

Create `make invariant`, move architectural invariant tests behind it, and make `make validate` depend on it without weakening existing validation.

### Phase 3 — HAL Contract Coverage

Express operations, authority, capabilities, profiles, bindings, provenance, and layer boundaries through executable invariants. Add negative tests for critical joins.

### Phase 4 — Engine Authority Coverage

Apply the same treatment to grammar/evaluator, prelude/help/docs, modules/exports/docs, primitive roles, and language services.

### Phase 5 — Representation Guardrails

Protect exact rationals, prohibit float-backed numeric paths, preserve runtime-owned state, and identify other never-regress engine contracts.

### Phase 6 — Reconcile

Review every invariant against the architecture documents, remove only clearly subsumed checks, and document which contracts are enforced by `make invariant`, `make test`, or both.

## Completion Gate

Gate 3 is complete when `make invariant` independently proves Aiki's architectural contracts are intact, every critical invariant has a negative test proving the detector works, `make validate` depends on `make invariant`, and architectural regressions fail clearly and early.
