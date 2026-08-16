# Milestone 27 — HAL redesign implementation M2

Status: **GATED**

## Purpose

Split the old one-registry native surface by architectural responsibility while
preserving all compatibility names, scope visibility, and execution behavior. This
is classification made executable; it is not yet the M4 authority redesign.

## Registry roles

The 117 compatibility primitives are now installed into five disjoint registries:

```text
intrinsic   evaluator/language-sensitive machinery                  9
native      language/value primitives implemented in Go             40
provider    native/FFI realizations of library behavior              14
host        host-facing compatibility operations                     37
service     runtime/tooling/session services                         17
                                                                  -----
                                                                    117
```

Canonical M1 host metadata remains a stricter subset of the host-role registry:
18 contracts. Canvas remains host-role compatibility machinery without prematurely
freezing its M6 contract. Randomness is placed with host/runtime-facing operations
because Phase III assigns RNG ownership to Runtime; M3 will settle that ownership.

## Behavioral constraint

`RuntimeContract`, `Execute`, `HasBuiltin`, `GetBuiltin`, and `BuiltinNames` retain
the existing compatibility behavior. `import`, `use`, and `export` continue to be
the only unprefixed runtime intrinsics available in user scope. No authority rule
changes in M2.

## Executable coupling

`TestCompatibilityRegistrySeparatedByRole` proves all 117 names are classified once,
checks the exact role counts, samples compatibility lookup across every role, and
proves M1's canonical host-binding count remains 18.

## Gate evidence

The user merged the M2/M3.a delivery plus the narrow stale-test correction and
reported the requested `go test ./...` checkpoint clean. The split therefore has
compile/test evidence on the authoritative tree. Full `make validate` was not
required for this intermediate migration.
