# Milestone 18 — Self-host runtime environment

Status: ACTIVE — implementation and disposable execution complete; awaiting the
next authoritative repository gate.

## Intent

Implement Phase III / Cut III.0 using ordinary Aiki capabilities only. The
runtime environment must support lexical lookup, shadowing, update through
outer scopes, enclosure, and shared closure capture without introducing a map,
store, or host-only self-hosting primitive.

## Implementation

Added `selfhost/runtime.ai`.

The environment is represented as an Aiki `@env` whose operations are ordinary
Aiki closures over a mutable lexical `bindings` variable. Bindings themselves
are persistent Aiki list data. Updating an environment reconstructs the binding
list and rebinds the captured `bindings` variable.

This representation matters for closure correctness: multiple interpreted
closures can share the same environment object, and a binding defined after a
closure captures the environment is subsequently visible through that capture.
It therefore avoids the snapshot semantics that a plain immutable
`[@env, bindings, outer]` value would accidentally introduce.

The runtime also defines the planned interpreted closure and return wrapper
shapes, but Cut III.0 gates only environment behavior.

## Validation

A disposable Go-1.23/no-graphics runner executed the real Aiki code and proved:

- missing-name detection;
- definition and lookup;
- child lookup through outer scope;
- child shadowing without changing the outer binding;
- assignment in the nearest owning scope;
- failure to assign an undefined name;
- shared capture of a binding defined after environment capture.

Added `test/invariant/selfhost_runtime_test.go` so the same focused probe runs
under the authoritative `make validate` environment.

## Discovery

The closure-backed environment is sufficient without `store` or a dictionary.
This supports the proposal's instruction not to add a map merely for
self-hosting convenience.

## Next action

Cut III.1 begins with literal/value construction. That work has exposed a
platform-surface sufficiency issue documented in Milestone 19 and must be
resolved before III.1 can honestly gate.
