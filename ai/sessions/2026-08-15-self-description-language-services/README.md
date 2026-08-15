# Self-description and language services

Status: ACTIVE — Phase I GATED; Phase II Cuts II.0–II.3 GATED; Cut II.4 ACTIVE at final Xed presentation gate.

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
12. `12-xed-support.md` — ACTIVE; static gates complete, live Xed gate pending.

## Current state

Phase I is GATED by authoritative `make validate`. Phase II has established the
neutral language-service diagnostics core, catalog inversion, behavior-neutral
observer/probe seam, and the `aiki lsp` stdio adapter. Xed lexical support is
current and grammar-coupled. Live workstation tracing has confirmed Xed → LSP
initialization → document synchronization → published diagnostics. Two adapter
defects found during that gate (Save As URI transition and zero-width diagnostic
presentation) are corrected, and an idempotent `make install-xed-plugin` target
now owns user-local installation. Cut II.4 remains ACTIVE only until the
corrected visual underline/clear behavior is confirmed in Xed.

## Exact next action

Run `make install-xed-plugin`, restart Xed, and confirm that `let x =` receives
a visible diagnostic underline and that changing it to `let x = 42` clears the
underline. Then run/confirm `make validate`; if both gates pass, mark II.4 GATED
and begin Cut II.5 (symbol/definition service plus nvi tags adapter).
