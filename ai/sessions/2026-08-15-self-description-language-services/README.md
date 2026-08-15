# Self-description and language services

Status: ACTIVE — Phase I GATED; Phase II Cuts II.0–II.6 GATED; Cut II.7 ACTIVE.

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
15. `15-completion-hover.md` — ACTIVE.

## Current state

Phase I and Phase II Cuts II.0–II.6 are GATED. Xed and nvi live gates are preserved in Milestones 12–13. Cut II.7 adds completion and hover/inspect to the neutral language-service contract and projects them through LSP; implementation and disposable package gates pass, while authoritative validation is pending.

The live Xed PATH lesson is explicit: desktop-launched Xed may inherit only the desktop-session PATH, not `.bashrc`/interactive-shell PATH. Verify the actual process environment via `/proc/<xed-pid>/environ`; ensure the development `aiki` is reachable through a stable path present there.

## Exact next action

Run authoritative `make validate` for Cut II.7. If it passes, mark II.7 GATED and begin II.8, the thin VS Code client / Phase-II completion handoff.
