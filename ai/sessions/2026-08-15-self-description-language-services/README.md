# Self-description and language services

Status: ACTIVE — Phase I GATED; Phase II GATED/COMPLETE; Phase III Cut III.0 ACTIVE; Cut III.1 BLOCKED on dynamic native-value construction.

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
18. `18-selfhost-runtime-environment.md` — ACTIVE; implementation/disposable execution complete, authoritative gate pending.
19. `19-selfhost-dynamic-value-construction-gap.md` — BLOCKED; dynamic symbol/shaped-list construction requires an explicit language-level decision.

## Current state

Phase I and Phase II are GATED. Xed, nvi, and VS Code have all passed their live client gates over the shared language-service authorities.

Phase III has begun. `selfhost/runtime.ai` implements lexical environments as ordinary Aiki closures over persistent binding lists, preserving shadowing, enclosing assignment, and shared capture without `store` or a map. A focused invariant test is present for the authoritative repository gate.

Cut III.1 exposed a deliberate self-hosting boundary issue: ordinary user-level Aiki cannot dynamically construct an arbitrary native symbol from a source lexeme, and arbitrary shaped-list construction is likewise not exposed at user/prelude level (`_make_shaped_list` exists only below the HAL boundary). Hard-coded symbol tables, wrapper values, or delegating through `load()` would weaken or invalidate the intended proof.

## Exact next action

Review the dynamic-value construction finding in Milestone 19 and deliberately choose a general Aiki-level surface for constructing native symbols and shaped lists. Then resume Cut III.1 using only that ordinary public/prelude-level capability. The current III.0 implementation should be included in the next authoritative `make validate` run.
