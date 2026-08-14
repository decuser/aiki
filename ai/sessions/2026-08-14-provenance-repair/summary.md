# Summary — Aiki provenance repair

Aiki's Git history had an accidental discontinuity caused by a repository
restart during recovery from a destructive refactor episode. The project
itself did not restart: development continued from the recovered post-refactor
state.

The historical and current repositories were therefore treated as evidence for
one lineage rather than as two codebases to merge. Their histories were joined
so the canonical path now runs continuously from the original genesis commit,
through the refactor work and recovered post-refactor state, into the current
repository and present `HEAD`.

Abandoned refactor experiments were preserved as side branches because they
are real provenance. Colliding pre-restart version tags were namespaced under
`pre-restart/`. A small set of obvious scratch artifacts and
`forge/Thompson-1968.odt` were removed from reachable history while substantive
historical source, documentation, and intermediate work were retained.

The reconstruction was carried out in an external workspace to avoid
contaminating either source repository during analysis. The adopted repaired
repository subsequently passed the full native `make validate` gate and its
complete decorated graph was inspected. The provenance repair is complete.
