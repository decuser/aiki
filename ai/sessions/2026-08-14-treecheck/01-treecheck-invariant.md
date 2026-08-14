# 01 — Distribution tree invariant

Status: **GATED**

## Intent

Add a single-tree invariant that detects files with no recognized distribution
relationship, primarily to catch stale paths left behind by overlay-based
merges after renames or deletions.

## Implemented

Added `aiki treecheck` under `cmd/subcommands/tools/treecheck/`.

The checker:

- examines tracked files plus untracked non-ignored files when Git metadata is
  available;
- ignores tracked paths that no longer exist in the working tree;
- falls back to a filesystem walk when Git metadata is unavailable;
- infers standard-library `.ai`/`.help`/`.doc` relationships;
- recognizes Aiki-native tests, behavior/visual smoke specimens and golds,
  engine specimens and stage golds, grammar/prelude artifacts, sample programs,
  profiling drivers, Go implementation/test source, and direct path references
  from already-justified text files;
- resolves Aiki package helpers imported by already-justified Aiki programs;
- reports structural contradictions separately from orphan candidates;
- uses root `treecheck.allow` only for intentional standalone artifacts.

The command is wired into `make check`, and therefore `make validate`.

## Root exception policy

`treecheck.allow` is intentionally small. It currently admits standalone
provenance/design families (`ai/`, `bootstraps/`, `proposals/`), editor support,
a profiling baseline pattern, and the intentionally empty regression README.
It is an exception list, not a complete repository manifest.

## Existing cruft found and removed

The first run found eight stale artifacts already present in the repository.
They were verified as unreferenced/superseded and removed rather than
whitelisted:

- `test/behavior/basic_io_smoke.in` — obsolete sidecar from the old smoke-input
  format; current smoke transcripts carry authored `IN:` directives.
- `test/behavior/down.ai`
- `test/behavior/mutual_even_odd.ai`
- `test/behavior/tail_if.ai`
- `test/behavior/tail_match.ai`
- `test/behavior/tail_sum.ai`
  — old recursion/TCO behavior programs superseded by canary coverage and not
  discovered by the current smoke runner.
- `test/structure/engine/07_bytes_engine.ai.lex.gold`
- `test/structure/engine/07_bytes_engine.ai.parse.gold`
  — structural golds whose source specimen no longer exists and which still
  referenced the removed `bytes/pragmatic` module.

`test/behavior/README.md` was corrected to state that `*_smoke.ai` specimens
have golds while helper modules may legitimately be unpaired.

## Validation

The container cannot download the repository-required Go 1.24 toolchain, so
compilation used the established disposable Go-1.23 harness with local stubs
for Ebiten/readline. No harness files are part of the product tree.

Completed successfully:

- `go test ./cmd/subcommands/tools/treecheck`
- CLI canary for `aiki treecheck`
- `go test -race ./cmd/subcommands/tools/treecheck ./test/canary`
- broad non-visual Go suites across `cmd/...`, `engine/...`, boundary, canary,
  contract, fuzz, property
- invariant tests except `TestDocExamplesExecutable`, whose real-canvas path is
  not runnable with the disposable graphics stub
- `./aiki lint ./...`
- Aiki-native tests: **408/408 passed**
- behavior smokes: **34/34 passed**
- grammar production coverage: **32/32 across 10 inputs**
- engine gold check: **10/10 inputs passed**
- harness-built `aiki treecheck` run against the authoritative uploaded tree
  with its actual Git metadata: **430 files, 388 structurally justified, 42
  explicitly allowed; no findings**

## Overlay workflow proof

A fresh copy of the uploaded baseline was overlaid with the completed tree
without deleting old paths. `treecheck` reported exactly the two orphaned
engine golds and six orphan behavior/input files listed above. After removing
those eight paths, the same overlaid tree passed. This directly exercises the
merge workflow the invariant was designed to protect.

## Design result

Overlay delivery does not need a full historical manifest to catch ordinary
stale-path debris. The current tree can enforce a distribution disposition
invariant: inferred structural participation plus a small explicit exception
set. Absence from a delivered overlay still does not itself imply deletion,
but an old path that no longer participates in the current repository
structure is now detectable during normal validation.

## Next action

Run `make validate` with the normal Go 1.24/Ebiten environment. If that is
green, no further treecheck work is planned before use exposes additional
structural families worth recognizing.
