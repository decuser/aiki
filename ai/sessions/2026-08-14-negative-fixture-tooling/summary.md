# Summary — Negative-fixture tooling

The grammar-authority project intentionally introduced smoke specimens that do
not parse. Their presence exposed a deeper tooling problem: source tools had no
explicit concept of an intentionally negative specimen, fmt swallowed ordinary
parse failures, lint stopped its formatting preflight at the first parse failure,
and the executable discarded several subcommand return codes.

The design follows the same authority discipline as the grammar work. Intent is
now a declaration (`# @negative parse`); the smoke gold remains evidence. Smoke
checks the two against one another, so neither a reblessed regression nor a stale
negative declaration can silently become accepted behavior. fmt and lint consume
only the declaration relevant to them: parse-negative specimens are excluded
because successful parsing is not part of their contract. Ordinary malformed
source remains a real error.

The fix also restores process-level truthfulness. Recursive fmt and lint
formatting preflight traverse past individual malformed files and return the
accumulated failures, and CLI dispatch propagates integer subcommand statuses.
Thus a command that reports failure can no longer quietly exit zero.

The first full-tree gate after these changes exposed one additional pre-existing
linter drift: `use("list")` was resolved by a hand-written filesystem heuristic
rather than the runtime module registry, so lint could not see exports from the
actual nested package `lib/list/list.ai`. Cut 4 replaces that duplicate resolver
for public package names with `ModuleRegistry` resolution while retaining the
runtime distinction for relative path imports. The failure was therefore useful
evidence that the repaired traversal was reaching source it had previously not
checked.
