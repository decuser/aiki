# Alpha release prep — 2026-08-14

Status: ACTIVE

## Intent

Prepare the repaired Aiki repository for its first public GitHub alpha release without conflating alpha maturity with public-artifact quality.

## Baseline

- Baseline archive: `aiki(10).tgz`
- Branch: `master`
- Tag at session start: `v0.4.0-alpha`
- Prior session: `../2026-08-14-provenance-repair/`
- Historical provenance repair is complete.

## Current state

Milestones 01–03 are gated:

- public README reframed around Aiki alpha rather than SPLASH-E;
- Linux-first Getting Started path established;
- alpha limitations, generic roadmap, forthcoming introductory books, and feedback guidance added;
- AI/authorship boundary and invariant framework documented;
- BSD 3-Clause licensing and copyright presentation confirmed;
- public/sensitive-information audit completed;
- historical scratch files `out` and `output` removed from all reachable Git history;
- repository integrity and release-surface checks completed.

## Validation status

Repository/history hygiene checks are gated. Full `make validate` cannot be completed in this sandbox because the required Go 1.24 toolchain is not locally available and network access is unavailable for toolchain retrieval.

Final release validation must therefore be run in the authoritative Linux environment from the exact candidate tree after these release-prep changes are committed.

## Next action

1. Commit the alpha-release-prep changes.
2. Run `make validate` in the authoritative Linux environment.
3. Move/recreate `v0.4.0-alpha` on the validated release commit if HEAD has advanced.
4. Force-update the private/local `origin` history because the `out`/`output` scrub rewrote commit IDs.
5. Publish the validated repository to GitHub and create the alpha release.
