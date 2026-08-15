# Milestone 04 — Reconciliation

**Status:** ACTIVE — final repository gate pending

Resolved the cheap audit clarity items:

- AF-001: documented the current byte/ASCII assumption in EBNF identifier scanning.
- AF-007: cross-referenced parser synthesis of `TERMINAL` from evaluator coverage policy.
- AF-015: documented why call/unary ambiguity smokes intentionally have empty output golds.

Updated `docs/testing.md`, `docs/audit-findings.md`, and the proposal outcome.
All selected findings are `RESOLVED`; AF-002/012 remain `DEFERRED` and
AF-003/004/009/013 remain `ACCEPTED`.

`git diff --check` is clean. Final status depends on the user's Go-1.24
`make validate`; no gold changes are intended.
