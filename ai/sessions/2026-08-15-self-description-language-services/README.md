# Self-description and language services

Status: ACTIVE — Phase I GATED; Phase II Cuts II.0–II.4 GATED; Cut II.5 ACTIVE at nvi handoff gate.

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
13. `13-symbol-definition-tags.md` — ACTIVE; implementation/local gates complete, authoritative validate + live nvi gate pending.

## Current state

Phase I and Phase II Cuts II.0–II.4 are GATED. The real Xed environment confirmed `.ai` recognition, grammar-coupled highlighting, LSP initialization/document synchronization, visible diagnostics, diagnostic clearing, Save-As transition handling, and the idempotent installer. Cut II.5 now adds one neutral symbol/definition authority with LSP and classic-tags projections. Local package tests pass in the disposable validation harness; authoritative validation and a live nvi tag jump remain.

## Exact next action

Run `make validate`. If it passes, generate a tags file with `./aiki tags -o tags <path>`, open an Aiki source file in nvi with `:set tags=./tags`, and confirm `^]` on a top-level name jumps to its definition. Then mark II.5 GATED and begin II.6 formatting service.
