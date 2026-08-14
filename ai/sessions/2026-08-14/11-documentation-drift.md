# Milestone 11 - documentation and buglist drift cleanup

Status: **GATED**

## Intent

Remove known documentation/path drift and apply the decision that no semantic items remain open in `buglist.md`.

## Changes

- Replaced `buglist.md` with an empty-state placeholder:
  - no known open bugs or drift items;
  - newly confirmed items should be added as discovered.
- Corrected the README Documentation inventory so every listed path exists in the delivered tree.
- Updated `docs/adding-to-aiki.md` examples from the obsolete `tests/smoke/` layout to `test/behavior/`.
- Updated the obsolete example package path/name from `lib/strings/strings.*` / `package "strings"` to the current `lib/string/string.*` / `package "string"` layout.

## Validation

- Re-scanned README, docs, and AI work records for the known stale path patterns; none remain.
- Reviewed all repository-path mentions in README and Markdown docs. Remaining paths either exist or are explicitly instructions to create a new example file.

## Next step

Run the normal validation and executable-coupling checks against the post-cleanup tree.
