# Proposals

`proposals/` is organized by **disposition**, not chronology. The top-level
directory is an index only; proposal documents belong in one of these
directories:

```text
active/       work currently being designed or executed
completed/    implemented or fully gated work retained as design history
superseded/   proposals replaced by a later controlling design
parked/       credible proposals not currently active
```

## Lifecycle

A new proposal starts in `active/`. Closing a project includes moving its
proposal to the correct disposition directory in the same closeout commit.

`active/` is therefore the authoritative live-work surface. A proposal should
not remain there merely because its document once used the word `ACTIVE`.
Repository state, validation evidence, and durable session history determine
disposition.

Do not create proposal files directly under `proposals/`. `aiki treecheck`
enforces this layout.

A parked proposal is not rejected. It is retained because the idea remains
credible but is not part of current work. A superseded proposal is retained to
show design history and should identify the controlling replacement where one
exists.

Historical references in `docs/session-history.md` may continue to name the
path that existed when the event was recorded; do not rewrite history merely
for path tidiness. Current-facing documentation and proposal cross-references
should use the disposition path.
