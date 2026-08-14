# Alpha release prep — 2026-08-14

Status: COMPLETE

## Intent

Prepare the repaired Aiki repository for its first public GitHub alpha release without conflating alpha maturity with public-artifact quality.

## Baseline

- Baseline archive: `aiki(10).tgz`
- Branch: `master`
- Tag at session start: `v0.4.0-alpha`
- Prior session: `../2026-08-14-provenance-repair/`
- Historical provenance repair is complete.

## Current state

Milestones 01–04 are gated:

- public README reframed around Aiki alpha rather than SPLASH-E;
- Linux-first Getting Started path established;
- alpha limitations, generic roadmap, forthcoming introductory books, and feedback guidance added;
- AI/authorship boundary and invariant framework documented;
- BSD 3-Clause licensing and copyright presentation confirmed;
- public/sensitive-information audit completed;
- historical scratch files `out` and `output` removed from all reachable Git history;
- repository integrity and release-surface checks completed;
- final `make validate` passed in the authoritative Linux environment;
- rewritten `master` and tags were synchronized successfully to the private Odin remote.

## Validation status

COMPLETE. Repository/history hygiene checks were gated in the reconstruction environment, and the final release candidate subsequently passed `make validate` in the authoritative Linux environment. The rewritten history and tags were also synchronized to the private Odin remote.

## Next action

None for this session. The repository is prepared for public alpha publication. GitHub publication and optional release-binary packaging may proceed as separate work.
