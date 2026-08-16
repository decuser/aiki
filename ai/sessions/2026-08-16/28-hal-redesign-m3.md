# Milestone 28 — HAL redesign implementation M3

Status: **GATED**

## Purpose

Move ambient host/process state toward explicit runtime/session ownership without
changing the Aiki-facing API. M3 is intentionally being executed in serial
ownership slices rather than as a single global-state rewrite.

## Slice M3.a — runtime process/RNG/file auxiliary state

Implemented:

- `system.args` now reads a per-`GoRuntime` copied argument snapshot;
- `system.env` now reads through a per-runtime environment-view function whose
  default is `os.LookupEnv`;
- `random.seed/random` operate on a per-runtime RNG rather than package-global
  `rng`; default construction preserves the baseline time-seeded behavior;
- line-oriented file-reader auxiliary state is per-runtime rather than the
  package-global `fileReaders` map;
- the runner carries program arguments into the runtime it creates instead of
  mutating a package-global argument vector.

These stateful operations are registered as bound runtime method values. The
universal builtin calling convention is unchanged; pure/value primitives remain
ordinary free functions.

## Executable evidence added

`runtime_ownership_test.go` exercises independent runtime argument snapshots,
environment views, explicitly seeded RNG streams, and file-reader auxiliary
state.

## Slice M3.b — I/O and REPL/session ownership

Implemented:

- Aiki standard input/output endpoints are per-runtime and configured with `SetIO`;
- REPL pageable help output is runtime-owned rather than package-global;
- the REPL user environment used by `delete()` is attached to its runtime/session;
- runner smoke and engine-smoke capture I/O through the runtime rather than by
  mutating package globals;
- runtime-isolation tests cover input, output, and REPL user environments.

The CLI line reader still reads process `os.Stdin`; that is REPL UI input rather
than the Aiki `read()` host endpoint and is intentionally outside HAL ownership.

## Slice M3.c — module/help service ownership

Implemented:

- module registry, module cache, and prelude help/doc registry are owned by
  `GoRuntime`;
- `import` and `use` are runtime-bound intrinsics because module resolution/cache
  now belongs to the runtime that executes them;
- `_module_roots`, `help()`, and `doc()` consult their owning runtime;
- REPL reset rebuilds the module registry on the same runtime;
- `RunExpr` now explicitly initializes the runtime module registry rather than
  depending on ambient prior initialization;
- an isolation test proves separate runtimes expose separate module-root sets.

## Slice M3.d — Canvas owner/lifecycle tracking

Implemented:

- open Canvas resources are tracked by the `GoRuntime` that created them;
- quit/reset and runner/REPL lifecycle cleanup close canvases through that owner;
- package-global `openCanvases` tracking is removed.

The lower-level Canvas bridge/session maps remain substrate implementation state.
Their representation/protocol migration is intentionally reserved for M6; M3 does
not redesign Canvas semantics or transport.

The Aiki test framework counters remain a runtime/tooling service with package
state. They were classified separately in M2 and are not one of the host-world
ownership requirements named by M3; moving them is deferred until the tooling
cleanup where it can be done without widening the runner API merely for symmetry.

## Validation status

The user reported `go test ./...` clean after M3.a and its stale-test correction.
For M3.b-d, the user reported `go test ./...` clean on the authoritative tree after the two contained build-fallout corrections (`enginesmoke` help registry ownership and a stale REPL import). M3 is therefore GATED. Full `make validate` was not required at this checkpoint.
