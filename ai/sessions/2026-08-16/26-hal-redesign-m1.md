# Milestone 26 — HAL redesign implementation M1

Status: **GATED**

Implementation begins from exact baseline `6e0b103` on branch `hal-redesign`.

## Purpose

Add canonical host-operation metadata beside current behavior. M1 must make the
three-name relationship executable without changing Aiki semantics or beginning
the M2 registry split.

## Current slice

The first slice introduces:

- architectural `hal.HostOperation` / `hal.AikiBinding` metadata;
- substrate bindings for 18 already-settled host crossings: standard I/O,
  time, program args/environment, and the 12 existing file operations;
- `registerHost` as compatibility registration plus contract attachment;
- runtime inspection through `HostOperations()`;
- an invariant test that checks HAL identity uniqueness, primitive registration,
  Go substrate provenance, and the declared Aiki wrapper -> primitive dependency.

Canvas is intentionally not canonized in M1. Phase III explicitly leaves its
minimal acquisition/protocol contract to the Canvas migration. `_module_roots`
is likewise still a runtime/module service rather than a host contract.

## Behavioral constraint

All execution continues through the existing `_` compatibility registry. The
new metadata must not alter evaluation, authority, errors, resource behavior, or
public Aiki names.

## Validation evidence

Completed in the delivery environment:

- `gofmt` on all changed Go files;
- `git diff --check` — clean;
- independent source-coupling probe — all 20 declared Aiki bindings resolve to
  wrappers that contain their declared compatibility primitive.

Authoritative Go tests could not run because the environment has Go 1.23.2 and
no network access to fetch the required Go 1.24 toolchain. A disposable copied-
tree probe with the module version lowered was also blocked because Ebiten is
not cached and cannot be downloaded. These are environment limitations, not
passing or failing test evidence.

## Gate result

The user merged the M1 drop and reported `make validate` passed on the authoritative
Aiki tree on 2026-08-16. M1 is therefore GATED. The next implementation cut is M2:
split the compatibility registry by semantic role without changing public behavior.

