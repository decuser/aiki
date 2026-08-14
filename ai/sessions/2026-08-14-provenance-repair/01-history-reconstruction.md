# Milestone 01 — historical continuity reconstruction

Status: **GATED**

## Intent

Restore Aiki's real historical continuity without treating the repository
restart as a new project lineage and without reintroducing obsolete source.
The task was provenance repair, not code recovery or feature merging.

## Evidence and reconstruction

Analysis of the original and current repositories established a defensible
continuous lineage:

```text
genesis
  |
  | original development
  |
pre-refactor / refactor work
  |
  | canonical refactor-clean line
  |
post-refactor state
  |
  | repository restart boundary
  |
current history
  |
HEAD
```

The old canonical line ends in the post-refactor state represented in the
repaired history by `eb894f9` (`post refactor`, historical branch
`history/refactor-clean`). The current post-restart line begins immediately
above it at `7964e5b` (`v0.1.0, post refactor`). The reconstructed graph makes
that ancestry explicit rather than leaving two disconnected repositories.

The earlier abandoned refactor experiments were preserved as side history,
including:

- `history/engine-refactor-attempt` / `history/old-master`;
- `history/archive-work`.

They remain branches from their historical fork and were not falsely merged
into the canonical development line.

## Tag treatment

Current post-restart tag names remain authoritative. Earlier tags that would
collide with later reused names were retained under `pre-restart/`, preserving
historical landmarks without creating ambiguous canonical tags.

## Historical hygiene

The reconstruction removed temporary or accidental repository artifacts from
all reachable history:

```text
diff.txt
diffs.txt
diff1.txt
diff2.txt
engine-refactor-diff.txt
dep.txt
engine.zip
forge/Thompson-1968.odt
```

This cleanup was deliberately narrow. Intermediate source states, broken work,
old documentation, tags, and abandoned development branches were retained as
project provenance.

## Validation evidence

During reconstruction:

- the resulting canonical history had a single genesis/root lineage into
  current `master`;
- the reconstructed current `HEAD` tree matched the supplied current `HEAD`
  tree exactly;
- the agreed removed files were absent from all reachable refs;
- reconstruction-only refs were removed;
- `git fsck --full` completed cleanly.

After adoption, the user inspected:

```text
git log --graph --decorate --oneline --all --date-order
```

and confirmed the expected continuous canonical line and preserved side
branches.

The user then ran the complete native validation workflow:

```text
make validate
```

which passed, including:

- Go build, formatting, lint, and treecheck;
- all Go tests;
- 408 Aiki tests, 408 passed;
- 32 grammar productions across 10 engine inputs;
- 34 behavior smoke tests;
- engine gold checks.

## Result

The repository restart is now represented as a continuity point in Aiki's
history rather than an artificial second genesis. The repaired repository is
the authoritative baseline for subsequent Aiki work.
