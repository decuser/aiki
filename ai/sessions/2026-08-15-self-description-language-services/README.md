# Self-description and language services

Status: COMPLETE — Phases I–III and the post-completion out-of-tree selfhost correction are GATED.

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
26. `26-self-interpretation-performance-boundary.md` — SUPERSEDED by measured completion.
27. `27-self-interpretation-complete.md` — GATED.
28. `28-iii6-chars-bootstrap-coupling-fix.md` — GATED; authoritative validate passed.
29. `29-selfhost-out-of-tree-path-imports.md` — GATED; authoritative validate and direct recursive out-of-tree execution passed.

## Current state

Phases I–III are GATED/COMPLETE. Phase III includes the independent runtime/evaluator, self-hosted module loader/cache/export path, scoped HAL bootstrap, behavior-conformance corpus, and a successful full nested self-interpretation proof.

Cut III.6 measurement corrected the earlier performance diagnosis. The dominant cost was repeated string/rune work in the independent lexer; after a linear `string.chars` realization and lexer dispatch cleanup, the nested path exposed two module-identity defects: missing current-directory fallback for path imports and non-canonical cache keys. Both now match the reference loader's observable behavior.

The complete proof executes `1 + 2 * 3` through an Aiki-written interpreter which self-host-loads another Aiki interpreter and returns `9`, preserving Aiki's left-to-right semantics at the third level. The local run completes in roughly 30 seconds.

## Exact next action

None for this session. Out-of-tree recursive self-hosting and profiling were successfully demonstrated; subsequent experiment-framework work is recorded separately.
