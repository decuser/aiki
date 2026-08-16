# Milestone 30 — HAL redesign M5: systems-programmer surface

> Status: ACTIVE — implementation complete, awaiting authoritative Go test checkpoint
>
> Branch: `hal-redesign`
>
> Builds on gated M1-M4.

## Purpose

Realize the agreed first-order systems-programmer affordances through the
three-name model without expanding the HAL one-for-one where Aiki can own the
semantics.

## Implemented surface

### File

Added Aiki wrappers and canonical host contracts for:

- `file.stat`
- `file.rename`
- `file.mkdir`
- `file.mkdir_all`
- `file.remove_all`
- `file.temp`
- `file.temp_dir`
- `file.copy`
- `file.size`

`file.stat` returns `[@stat, size, modified_ms, is_dir]` or an `:io` shaped
error. Existing and new relative path operations resolve against the owning
runtime's working directory.

### System/process

Added:

- `system.cwd`
- `system.chdir`
- `system.exec`

Working directory is runtime-owned. `system.chdir` does not mutate the Go
process CWD. `system.exec` runs synchronously in the runtime working directory
and returns `[@ok, stdout, stderr, exit_code]` for a successfully started
process, including nonzero exit status; start failure returns a `:process`
shaped error.

### Time

Added `time.now`, returning integer milliseconds since the Unix epoch.

### Path

Added `lib/path/path.ai` with:

- host-backed `path.separator`
- `path.cwd` as composition over `system.cwd`
- pure-Aiki `join`, `dir`, `base`, `ext`, `normalize`, and `is_absolute`

The pure path operations deliberately do not delegate their semantics to Go's
`filepath` package. The initial absolute-path contract remains the design's
simple "begins with host separator" rule.

## Architectural evidence

M5 adds 14 real compatibility primitives, moving the classified registry from
117 to 131 entries. Canonical host contracts move from 18 to 32.

Every new host-backed Aiki wrapper has:

1. programmer-facing Aiki name;
2. canonical HAL identity/metadata;
3. Go substrate provenance and binding;
4. exact trusted-source authority grant.

The self-host bootstrap captures the new raw primitives so native and
self-hosted execution do not silently diverge.

`path.cwd` demonstrates non-one-to-one layering: it adds an Aiki affordance
without adding a second HAL crossing. The rest of the path manipulation layer
is Aiki computation.

## Validation prepared

- focused Go tests cover file metadata/directory/copy/temp operations;
- focused Go tests cover `time.now`;
- focused Go tests prove runtime-local CWD drives relative file bindings without
  changing process CWD;
- focused Go tests cover `system.exec` stdout/stderr/nonzero-exit/runtime-CWD;
- Aiki `path_test.ai` covers portable path composition/normalization behavior;
- role invariant expects 131 compatibility primitives and 32 canonical host
  bindings;
- three-name invariant expects 32 host contracts;
- help/doc entries were updated for file/system/time and added for path;
- `git diff --check` is clean.

This execution environment cannot run the repository's Go 1.24 test suite, so
M5 is not marked GATED here.

## Deliberately not included

- asynchronous process `start`/`wait`;
- environment enumeration/mutation;
- locale-heavy time formatting/parsing;
- networking/signals/general IPC;
- Canvas protocol/resource-representation redesign (M6).

These remain subject to concrete-program pressure rather than architectural
symmetry.

## Exact next action

Merge the M5 delivery and run `go test ./...`. A full `make validate` is not
required for this checkpoint. Because M5 changes Go substrate code, rebuild
with `make` before using the Aiki executable for manual validation.
