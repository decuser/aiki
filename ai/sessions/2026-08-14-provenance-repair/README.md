# Session 2026-08-14-provenance-repair — historical continuity reconstruction

Status: **COMPLETE**

Prior session: `../2026-08-14-treecheck/`

## Objective

Repair the accidental Git-history discontinuity between the original Aiki
repository and the later post-refactor repository restart, preserving the
actual development lineage from genesis through the refactor and into current
`master`.

The reconstruction was performed outside the authoritative repositories so
that the source repositories remained evidence rather than workspaces.

## Milestones

1. `01-history-reconstruction.md` — historical seam analysis, canonical-line
   reconstruction, side-branch preservation, history cleanup, and validation.

## Result

Aiki now has one continuous canonical history from the original genesis commit
through the pre-refactor/refactor work, the post-refactor state, the repository
restart, and current `HEAD`.

Meaningful abandoned refactor attempts remain reachable as historical side
branches. Historical tags whose names collide with later post-restart tags are
retained under the `pre-restart/` namespace; current tag names remain
canonical.

Temporary scratch artifacts and the retained Thompson working document were
removed from reachable history during reconstruction.

## Validation

The reconstructed repository was validated by:

- exact current-tree preservation across the history rewrite;
- inspection of the complete decorated commit graph;
- preservation of the expected historical side branches;
- absence of the agreed removed artifacts from reachable history;
- clean Git object validation during reconstruction;
- a complete native `make validate` run by the user after adoption, passing all
  Go tests, 408 Aiki tests, grammar coverage, behavior smokes, and engine gold
  checks.

## Next action

None. The provenance repair is closed. Future work should proceed from the
repaired repository as the authoritative baseline.
