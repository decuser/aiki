# Milestone 11 — spawned relative-import provenance

Status: ACTIVE

## Finding

Repository-wide Aiki tests exposed a context-dependent failure in the Thompson monitor/service tests. The same tests passed when launched from the experiment directory but failed from repository root with `unknown shape: error` after entering the spawned machine service.

The service worker performs relative imports after crossing `spawn`. `applyUserFunctionIsolated` correctly constructs an isolated environment and copies shape vocabulary, but it did not preserve the defining function's file provenance. Relative import resolution therefore had no lexical file anchor and fell back to the process working directory.

This is an Aiki engine defect, not a Thompson-specific path problem.

## Correction

`applyUserFunctionIsolated` now copies only the defining environment's file provenance into the isolated call environment. This is lexical metadata used for source-relative resolution; it does not reconnect the spawned computation to outer mutable bindings or execution state.

A runner regression test defines a worker module whose spawned function imports a sibling module with `import("./dep")`. The main program is executed independently of that module directory, so the test fails under the old behavior and succeeds only when spawned relative imports retain their defining-file anchor.

The service source also carries the prior lint rename `runtime.INIT` -> `runtime.init` so the delivered tree is reproducible.

## Gate

Not yet GATED in the container. Required local evidence:

1. `go test ./...`
2. rebuild Aiki (`make`)
3. `./aiki lint ./...`
4. `./aiki treecheck`
5. `./aiki test ./...`
6. experiment `./run.sh`

A rebuild is required because the correction changes Go engine code.
