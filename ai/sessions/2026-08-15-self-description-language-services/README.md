# Self-description and language services

Status: ACTIVE — Phase I GATED; Phase II Cuts II.0–II.7 GATED; Cut II.8 ACTIVE.

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
16. `16-vscode-client.md` — ACTIVE.

## Current state

Phase I and Phase II Cuts II.0–II.7 are GATED. Cut II.8 adds the final Phase-II consumer: a thin VS Code client over `aiki lsp`, plus lexical-only TextMate presentation executable-coupled to the grammar. The extension has an explicit `aiki.server.path` setting so desktop launch does not depend on interactive-shell PATH.

The live Xed PATH lesson remains explicit: desktop-launched editors may inherit only the desktop-session PATH. Inspect the editor process environment rather than inferring it from shell `which`; use a stable executable path or editor launch setting.

VS Code live testing has confirmed extension discovery/highlighting, hover after the explicit-null JSON-RPC correction, and canonical formatting. It also exposed and corrected definition lookup at caret positions inside identifiers. The VS Code installer now builds and installs a VSIX entirely out of tree rather than copying a staged directory into `~/.vscode/extensions`.

## Exact next action

Merge the II.8 definition/installer correction, run authoritative `make validate`, then run `make install-vscode-plugin` and restart VS Code. Recheck Go to Definition with the caret anywhere inside a source-defined name, then complete the remaining live VS Code gate (diagnostic/clear and completion if not yet observed). If all live checks pass, mark II.8 GATED and Phase II complete at its handoff.
