# Invariant Tests

Verify that Aiki's declared architecture is still true.

`make invariant` is the dedicated architectural gate. `make test` excludes the
invariant and boundary packages; `make check` and therefore `make validate`
require both behavioral tests and architectural invariants.

## Rule

An invariant belongs here when it protects a structural relationship, layer
boundary, authoritative metadata join, prohibited representation, or another
architecture-level contract.

A critical invariant is not considered protected merely because the valid tree
passes. It must also have a negative test that violates the contract and proves
the same detector rejects the invalid state with a useful diagnostic.

## Critical assured contracts

- HAL metadata identity/capability/profile coherence, including negative graph mutations.
- HAL runtime host-binding coverage and three-name contract, including missing-binding mutation.
- capability availability and required-profile gating, including missing-operation mutation.
- primitive architectural roles and runtime registration-role coverage, including missing/wrong-role mutations.
- grammar/token handler coverage, including deliberately missing handler tests.
- prelude source/help/doc coverage, including missing-help and missing-doc mutations.
- library export/help/doc coverage, including missing and phantom entry mutations.
- concrete substrate dependency confinement to the composition root, including injected-import mutation.
- runtime-owned I/O/system state, including injected ambient-process-global mutation.
- exact-number core representation: no float path in `engine/semantics` or `engine/syntax`, including injected-float mutation.

## Other structural checks

The gate also checks documentation disposition, graphics confinement, opaque
Canvas semantic values, shipped-module discovery, help signature agreement, and
editor lexical/client coupling. These remain structural distribution checks;
negative mutation coverage may be added when a concrete regression risk warrants
promoting one to a critical assured contract.

## Behavioral/conformance separation

Execution-heavy self-host conformance and executable documentation examples live
under `test/conformance` and run with `make test`. Exact-number execution behavior
lives under `test/contract`. They support architectural claims but do not belong
in the fast structural gate.
