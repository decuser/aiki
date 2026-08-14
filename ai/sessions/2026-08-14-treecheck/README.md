# Session 2026-08-14-treecheck — distribution orphan detection

Status: **COMPLETE**

Prior session: `../2026-08-14-review/`

## Objective

Add `aiki treecheck`, a developer invariant that identifies files in a single
Aiki source tree which have no recognized structural disposition and are not
explicitly allowed as standalone distribution artifacts.

The immediate motivation is overlay-based delivery: additions and
modifications merge naturally, but a renamed or deleted old file can remain
behind. `treecheck` should detect such orphan candidates from the current tree
without requiring a previous tree or a generated full-file manifest.

## Milestones

1. `01-treecheck-invariant.md` — distribution disposition invariant, first-run
   orphan cleanup, CLI/validation integration, and evidence gate.
2. `02-native-validation.md` — complete `make validate` pass in the normal
   Go 1.24/Ebiten development environment.

## Established design constraints

- Inspect the current source tree, not repository history.
- Prefer programmatic structural relationships over a complete file manifest.
- Use a small root allow file only for intentional standalone classes/artifacts.
- Include tracked files and untracked non-ignored files when Git metadata is
  available, so overlays are checked before staging.
- Report structurally contradictory artifacts separately from orphan candidates.
- Keep `treecheck` alongside developer commands such as `fmt` and `lint`.

## Current state

`aiki treecheck` is implemented and wired into `make check` / `make validate`.
The first run found and removed eight pre-existing stale artifacts. The current
authoritative tree is clean under the invariant.

## Next action

None. The native `make validate` gate passed. No further treecheck design work
is planned unless use exposes a missing structural family.
