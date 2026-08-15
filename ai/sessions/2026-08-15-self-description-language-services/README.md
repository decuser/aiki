# Self-description and language services

Status: ACTIVE — Phase I GATED; Phase II GATED/COMPLETE; Phase III Cuts III.0–III.5 locally gated; Cut III.6 BLOCKED on practical parser self-parse performance.

Proposal: `proposals/aiki-self-description-language-services-proposal.md`

Baseline: `v0.4.0-alpha-14-g9c78646` (`9c78646`).

## Milestones

1. `01-grammar-token-authority.md` — GATED.
2. `02-authority-representation-inventory.md` — GATED.
3. `03-independent-lexer.md` — GATED.
4. `04-independent-newline-normalization.md` — GATED.
5. `05-independent-parser.md` — GATED.
6. `06-phase-i-handoff.md` — GATED.
7. `07-phase-i-authoritative-gate.md` — GATED.
8. `08-language-service-inventory.md` — GATED.
9. `09-language-service-diagnostics.md` — GATED.
10. `10-language-service-observation.md` — GATED.
11. `11-lsp-shell.md` — GATED.
12. `12-xed-support.md` — GATED; live Xed underline/clear gate passed.
13. `13-symbol-definition-tags.md` — GATED; authoritative validate and live nvi jump passed.
14. `14-formatting-service.md` — GATED; authoritative validate passed.
15. `15-completion-hover.md` — GATED; authoritative validate passed.
16. `16-vscode-client.md` — GATED.
17. `17-phase-ii-handoff.md` — GATED; Phase II complete.
18. `18-selfhost-runtime-environment.md` — locally gated; authoritative gate pending.
19. `19-selfhost-dynamic-value-construction-gap.md` — SUPERSEDED by the adopted surface in Milestone 20.
20. `20-dynamic-value-surface.md` — locally gated; authoritative gate pending.
21. `21-selfhost-expression-evaluator.md` — locally gated; authoritative gate pending.
22. `22-selfhost-statements-functions.md` — locally gated; authoritative gate pending.
23. `23-module-boundary.md` — SUPERSEDED by the self-hosted module decision.
24. `24-selfhost-modules.md` — locally GATED; authoritative gate pending.
25. `25-selfhost-behavior-conformance.md` — locally GATED; authoritative gate pending.
26. `26-self-interpretation-performance-boundary.md` — BLOCKED on practical nested parser self-parse performance.

## Current state

Phase I and Phase II are GATED. Phase III now has an independent runtime environment, evaluator, self-hosted module loader/cache/export path, scoped HAL bootstrap, and a representative behavior-conformance corpus. `to_symbol` and `shaped` close the dynamic value-construction gaps discovered by self-hosting.

Behavior conformance exposed and corrected two real evaluator issues: recoverable `[@error, ...]` values are no longer conflated with internal halts, and qualified module access in pipeline targets now resolves correctly. Internal evaluator halts use the private `:self_fault` control shape.

Cut III.6 has begun. A traced nested-interpreter attempt successfully entered the inner bootstrap and self-host-loaded the lexer and normalizer, but parser-self-parse exceeded practical local time bounds without producing a semantic fault. This is recorded as a performance boundary rather than a failed self-hosting claim.

## Exact next action

Run authoritative `make validate` for the Phase-III delta through III.5. If it passes, measure the nested parser-self-parse path before changing semantics or weakening the self-interpretation criterion. The immediate III.6 question is performance, not correctness.
