# Self-description and language services

Status: ACTIVE — Phase I implemented; authoritative validation correction pending re-run.

Proposal: `proposals/aiki-self-description-language-services-proposal.md`

Baseline: `v0.4.0-alpha-14-g9c78646` (`9c78646`).

## Milestones

1. `01-grammar-token-authority.md` — ACTIVE; implementation complete, authoritative executable gate pending.
2. `02-authority-representation-inventory.md` — GATED.
3. `03-independent-lexer.md` — ACTIVE; implementation complete, authoritative gate pending.
4. `04-independent-newline-normalization.md` — ACTIVE; implementation complete, authoritative gate pending.
5. `05-independent-parser.md` — ACTIVE; implementation complete, authoritative gate pending.
6. `06-phase-i-handoff.md` — ACTIVE; implementation complete, authoritative gate pending.

## Current state

Phase I implementation is complete through Cut I.4. The independent Aiki lexer,
newline normalizer, and parser agree with all reviewed fixtures under a
disposable non-graphics validation runtime. The Go normalization seam has also
passed the existing syntax-package tests in a disposable local-Go copy.

The first authoritative `make validate` attempt exposed a fixture-packaging
error: raw lexical/newline conformance inputs used `.ai` extensions and were
therefore (correctly) passed through `aiki fmt`, although several are
intentionally non-program fragments. They are now `*.input`; fixture content
and reviewed token projections are unchanged. The second attempt reached
`treecheck` and exposed missing distribution relationships for the new
self-host/conformance artifacts plus one lint-only helper name. Those
relationships are now structural treecheck rules (not allow-list exceptions),
and the helper has been renamed. Phase II has not started.

## Exact next action

On the authoritative development machine run:

```text
make validate
```

If it passes, mark Phase I GATED and begin Phase II Cut II.0. If it fails,
correct Phase I from this same tree before advancing.
