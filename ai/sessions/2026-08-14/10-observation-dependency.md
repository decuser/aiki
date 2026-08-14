# Milestone 10 - remove profiling dependency from value

Status: **GATED**

## Intent

Remove the profiling-specific dependency from `engine/semantics/value` to the top-level `engine` package without changing profiling behavior or forcing a repository-wide API rename.

## Changes

- Added leaf package `engine/observe` containing the neutral observation/profiling contracts and data types:
  - `SemanticKind`
  - `SemanticSite`
  - `SemanticProbe`
  - `AttributionProbe`
  - `SemanticCounts`
  - `SemanticSiteCount`
  - `SemanticMeasurement`
  - `ProfileLabels`
- Converted `engine/profiling.go` to compatibility aliases over `engine/observe` so existing engine-facing callers remain unchanged.
- Changed `value.Env` to store and expose `observe.SemanticProbe` directly.

## Architectural result

Profiling no longer causes `value` to depend on the top-level `engine` package. The remaining `value -> engine` import is pre-existing and unrelated to profiling: `tailcall.go` uses `engine.Position`.

The new dependency shape for profiling is:

```text
value -> observe
engine -> observe
```

with `observe` as a leaf package importing no higher-level Aiki packages.

## Validation

A disposable copy of the source had only its Go version line lowered from 1.24 to 1.23 so the container's installed toolchain could compile the affected leaf packages. Product `go.mod` was not changed.

Passed:

```text
go test ./engine/observe ./engine ./engine/semantics/value
```

## Caveat

This milestone intentionally does not migrate all existing `engine.Semantic*` call sites to `observe.Semantic*`; compatibility aliases keep that migration optional and separate.

## Next step

Clean documentation and buglist drift without adding capability.
