# Proposal: Post-Grammar Exceptional-Path Hardening

## Status

Implemented on branch `audit/post-grammar-hardening`.

## Purpose

The grammar-authority and centralized-analysis projects are complete and fully
validated. Their subsequent audits found a small set of exceptional-path and
tooling-boundary issues that are independently addressable without reopening
language design.

This project closes those concrete findings while establishing a durable,
repository-wide audit ledger in `docs/audit-findings.md`.

## Governing rule

**Exceptional behavior must be bounded, declared, and observable.**

A malformed token stream must not corrupt later normalization state. A degraded
analysis must be visible to diagnostics without making normal parsing depend on
an optional refinement. Test-only declarations must not become general source
exemptions. Repeated fixture classification should have one implementation
path.

## Audit findings in scope

- AF-005 — newline suppression can become negative.
- AF-006 — newline-analysis diagnostic degradation is silent.
- AF-014 — lint repeats negative-fixture filtering.
- AF-016 — negative-fixture declarations are not scoped to fixtures.
- AF-001, AF-007, AF-015 — cheap documentation/source clarity items closed in
  final reconciliation.

AF-002, AF-003, AF-004, AF-009, AF-012, and AF-013 remain explicitly deferred
or accepted in the audit ledger.

## Non-goals

- Changing valid Aiki syntax or newline policy.
- Making newline analysis mandatory for parsing.
- Generalizing negative fixtures beyond the currently supported `parse` kind.
- Rewriting the linter's overall path discovery architecture.
- Solving decorator source-position tracking (AF-002).
- Performance work without measurements.

## Cuts

### Cut 0 — ledger and proposal

Create the durable audit ledger, assign stable IDs, and bind this project to the
specific findings above. No implementation change.

### Cut 1 — parser exceptional paths

1. Replace aggregate newline suppression depth with delimiter-aware suppression
   tracking that cannot be driven negative by unmatched closing tokens.
2. Preserve all valid-program behavior and existing newline golds.
3. When cached newline analysis is unavailable, parsing remains available but a
   diagnostic-capable observer receives the degradation reason.
4. Add regression tests for unmatched-closer suppression and analysis
   observability.

### Cut 2 — negative-fixture boundary

1. `# @negative parse` is valid only on recognized `*_smoke.ai` fixtures.
   A marker elsewhere is an error, not an exemption.
2. Smoke retains the declaration ↔ observation bidirectional gate.
3. fmt/lint continue to skip valid declared parse-negative smoke fixtures.
4. Consolidate lint's repeated candidate/negative filtering behind one helper.
5. Add tests for valid fixture scope, invalid arbitrary-source scope, and lint
   candidate behavior.

### Cut 3 — reconciliation and cheap audit closures

1. Document the current ASCII assumption in the EBNF identifier scanner.
2. Cross-reference parser-synthesized `TERMINAL` from the evaluator exception.
3. Explain why the call/unary ambiguity smoke fixtures intentionally produce no
   output.
4. Update `docs/testing.md`, `docs/audit-findings.md`, proposal status, and
   session records.
5. Audit the delta against all stable audit IDs; no item may disappear without
   `RESOLVED`, `ACCEPTED`, or `DEFERRED` disposition.

## Acceptance criteria

1. An unmatched suppressing closer cannot make later valid opener context lose
   newline suppression.
2. Normal valid behavior/golds remain unchanged.
3. Newline-analysis degradation is observable to diagnostic-aware observers and
   does not prevent parser construction.
4. A negative marker outside `*_smoke.ai` fails classification and therefore
   cannot silently remove ordinary source from fmt/lint coverage.
5. Declared parse-negative smoke fixtures remain skipped by fmt/lint and checked
   bidirectionally by smoke.
6. Lint has one source-candidate classification path rather than three copies of
   negative filtering.
7. The audit ledger contains every combined post-branch audit finding with a
   stable disposition.
8. Repository validation passes with no intended behavior-gold changes.

## Validation

Each cut gets focused Go tests where possible. The final authoritative gate is:

```text
make validate
```

No gold should require reblessing for this project.

## Outcome

Implemented as proposed. The parser now uses delimiter-aware suppression state;
newline-analysis degradation is observable through the optional diagnostic
observer extension; negative declarations are restricted to smoke fixtures;
and lint has one source-candidate classification path. Cheap audit clarity
items AF-001, AF-007, and AF-015 were also closed.

No language behavior or gold baseline was intentionally changed. Final project
status is complete subject to the authoritative repository `make validate` gate.
