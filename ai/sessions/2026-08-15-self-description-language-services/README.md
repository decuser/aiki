# Self-description and language services

Status: ACTIVE — Phase I GATED; Phase II Cuts II.0–II.5 GATED; Cut II.6 ACTIVE.

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
14. `14-formatting-service.md` — ACTIVE.

## Current state

Phase I and Phase II Cuts II.0–II.5 are GATED. The real Xed environment confirmed `.ai` recognition, grammar-coupled highlighting, LSP initialization/document synchronization, visible diagnostics, diagnostic clearing, Save-As transition handling, and the idempotent installer. The live nvi gate also passed: a tags file generated for `extra/samples/string_utils.ai` resolved `my_slice` with `^]`. Cut II.6 now extracts the canonical formatter into a neutral engine capability and projects it through the language service/LSP without duplicating formatting logic.

## Exact next action

Complete Cut II.6: extract the canonical formatter below command adapters, add `Service.Format`, expose LSP `textDocument/formatting`, prove CLI/service equivalence and preserve the existing parse-preservation safety gate. Then run the authoritative `make validate`.
