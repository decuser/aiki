# Milestone 02 — README completion for alpha audience

Status: GATED

## Intent

Finish the public README as a practical front door for an alpha user without turning it into a roadmap dump or duplicating the curated documentation set.

## Changes

- Added `Alpha Expectations` with concise statements about instability, performance priority, and intentionally incomplete areas.
- Added a deliberately generic `Roadmap` focused on clarity, hardening, tooling, documentation, systems/teaching usefulness, and broader testing of the language's commitments.
- Added a friendly `Forthcoming` note for planned introductory books, particularly for non-programmers and younger learners.
- Added a verified `turtle/simple` example using the actual implicit-state API and symbol color syntax.
- Added a short `Feedback` section welcoming bug reports, reproducible failures, documentation corrections, and discussion of language behavior.
- Performed a coherence pass so the README remains an orientation layer over the deeper curated documents.

## Decisions

### Roadmap scope

The README roadmap is intentionally non-specific. Feature-level plans belong in proposals, issues, and the gated engineering record. The public roadmap should remain true as individual priorities change.

### Alpha limitations

Alpha limitations are stated plainly rather than hidden: the language is usable, but syntax, libraries, tooling, and interfaces may change; performance is not yet a primary optimization target; incomplete areas are not represented as stable commitments.

### Introductory material

The forthcoming books are framed as a slower, playful route into Aiki, particularly for readers who are not already programmers. `turtle/simple` is used as the concrete example because it exposes immediate behavior with minimal machinery.

## Evidence

- `lib/turtle/simple.help` documents `new(dim)`, `forward(distance)`, `right(angle)`, `pen_size(n)`, and `pencolor(color)`.
- `lib/turtle/simple.doc` demonstrates symbol colors such as `pencolor(:red)` and the same implicit-state calling convention used in the README example.
- `extra/samples/pipeline.ai` and both primary ODT language documents referenced by the README are present.
- `git diff --check` reports no whitespace errors.

## Validation boundary

This milestone validates README content and internal references only. Full `make validate` remains a final release gate to be run in the authoritative Linux environment.
