# Native/FFI Separation and Standard-Library Semantic Roles

## Status

Active. Clean implementation pass begun 2026-08-18 from baseline ffe3622. Existing behavioral witnesses are frozen and must remain unchanged through the baseline-preservation phase.

## Motivation

Aiki now has both Aiki-authored library implementations and provider-backed implementations. Profiling has shown that provider-backed code is valuable when it collapses a coarse, load-bearing boundary, but the mechanism by which an implementation crosses into the runtime must not blur what semantics Aiki promises or what implementation the programmer explicitly requested.

The original proposal treated `FFI` as though it were a single semantic category. The Gate 1 inventory showed that this is too coarse. FFI is an implementation mechanism. A provider boundary may exist for acceleration, because a capability is inherently host-mediated, or because interoperability with a foreign/provider abstraction is itself the point.

This project therefore becomes a standard-library and prelude boundary audit/redesign, not merely an `/ffi` naming check.

## Governing Truth Rule

> **The name the programmer asks for must tell the truth about the implementation they receive.**

In particular:

```
X/native -> native Aiki realization of X's portable semantics
X/ffi    -> provider-backed realization
X        -> deliberate public/default facade with an explicit policy
```

A programmer may freely mix native and FFI modules. No declaration or permission ceremony is required merely because FFI is used.

A module or implementation must not silently cross from a promised native realization into provider-backed library semantics.

## Two Different Meanings of Capability

Aiki already uses capability language at the HAL boundary. This proposal introduces a separate programmer-facing concept and keeps the two meanings explicit.

### HAL capability

A HAL capability is authority to cross a runtime or host boundary.

Examples include filesystem access, process creation, terminal state, clocks, networking, and runtime-owned mutable facilities.

HAL answers:

```
Which trusted source may invoke this raw runtime operation?
```

### Semantic capability

A semantic capability is behavior promised to an Aiki programmer.

Examples include:

```
string.split
bytes.digits_from_text
math.sqrt
regex.matches
store.get
```

Semantic policy answers:

```
What behavior does Aiki promise?
Who owns its meaning?
Which realizations are legitimate?
```

The layers are related but not interchangeable:

```
HAL capability      = authority below the boundary
semantic capability = promised behavior above the boundary
realization         = how that behavior is implemented
```

## Semantic Roles

Every shipped public module must have an explicit semantic role. Classification is about the authority for meaning, not merely about source language or directory name.

### `portable`

Aiki defines portable semantics that do not inherently require a host resource or foreign abstraction.

Every portable semantic capability must have a genuine native Aiki path.

Provider-backed realizations may exist as accelerations, but they must preserve the native semantic contract.

Examples intended to fall in this class include ordinary string transformations, collection algorithms, numerical algorithms whose Aiki contract is portable, and pure data conversions.

### `runtime-capability`

The behavior exposes an Aiki-defined runtime facility that cannot be manufactured by ordinary library composition alone.

Examples include runtime-owned mutable store values, channels, evaluator operations, process endpoints, or other facilities whose existence depends on the Aiki runtime.

No fake pure-Aiki implementation is required. The Aiki-facing semantics must nevertheless be explicit, and HAL authority must remain exact.

### `host-capability`

The behavior inherently depends on host state or resources.

Examples include files, processes, networking, terminal state, signals, environment, clocks, and entropy.

No fake native twin is required. Portable policy should specify Aiki-facing behavior and fault shapes where meaningful.

### `interop`

The public purpose is intentional exposure of a foreign/provider abstraction or provider-defined semantics.

An interop module is allowed to be provider-only, but it must be named and documented honestly. It must not masquerade as portable native Aiki semantics.

`regex/ffi` must be classified from its actual contract during this audit rather than mechanically presumed to be acceleration simply because its name ends in `/ffi`.

### `internal`

Implementation support, diagnostics, experiments, profiling seams, or workload-specific helpers that are not part of the public standard-library semantic contract.

This class prevents benchmark-shaped or diagnostic APIs from becoming permanent public affordances merely because they are useful below the boundary.

## Realization Kinds

Semantic role and implementation realization are independent axes.

The relevant realization kinds are:

```
native     Aiki implementation of the library semantics
ffi        provider-backed implementation
intrinsic  irreducible Aiki runtime/value/evaluator primitive
mixed      composition intentionally crosses more than one realization kind
```

`mixed` is acceptable for an honestly named public facade or capability module. It is not acceptable underneath a `/native` name when the mixed portion implements the library semantics that `/native` promises to realize in Aiki.

### Native does not mean reimplementing the evaluator

A genuine native library implementation may use constitutive Aiki intrinsics: operations that are part of the value model, evaluator, or irreducible runtime substrate.

For example, native Aiki code may legitimately use primitives required to inspect or construct Aiki values when those primitives are semantic atoms of the language.

It may not outsource its library algorithm to an FFI/provider implementation and still claim to be native.

Thus the audit asks, for every primitive used by native code:

```
Is this constitutive/irreducible Aiki machinery?
or
Is this provider implementation of the library operation itself?
```

## Native-Path Invariant

The core semantic completeness invariant is:

> **Every portable semantic capability in the public standard library must have a genuine native Aiki realization.**

This means an FFI-only portable affordance has three possible dispositions:

1. gain a native Aiki realization;
2. be reclassified, with rationale, as genuine `interop`, `runtime-capability`, or `host-capability`; or
3. move to `internal`/experimental scope if it is not a real public semantic capability.

There is no fourth category of "portable but provider-only because it was convenient."

## Acceleration Invariant

When a provider-backed implementation is an acceleration of portable Aiki semantics:

```
reference/native semantic surface exists
FFI surface matches the native semantic surface
FFI adds no public affordances
FFI omits no public affordances
arity/template shapes agree
return shapes agree
fault/domain behavior agrees
shared behavior contracts pass against both realizations
```

The native implementation is semantic authority.

Provider documentation may mention acceleration or provenance, but it must not define different semantics.

## Programmer-Facing Selection

Explicit selection must remain simple and composable.

Examples:

```
let bytes = import("bytes/native")
let fast_bytes = import("bytes/ffi")
```

A program may use both in the same process.

Bare imports such as:

```
import("bytes")
```

must have an explicit documented default policy. Where a bare alias resolves to `/native`, it must actually provide the truthful native realization.

No additional declaration such as `@uses_ffi` or `system.require(:ffi)` is required.

## FFI Visibility

FFI use is mechanically explicit at import sites, but direct and transitive use should be inspectable without changing normal execution.

Use the non-failing diagnostic command:

```
aiki check --ffi-use program.ai
```

Example:

```
FFI imports:
  string/ffi
  bytes/ffi
```

For a program with no FFI imports:

```
FFI imports: none
```

The diagnostic is visibility, not permission.

## Library and Prelude Audit

Gate 1 inventories the entire shipped `lib/` surface and the prelude, not only files whose names contain `/ffi` or `/native`.

For every public module and significant public affordance, determine:

```
semantic role
semantic authority
realization kind
HAL/raw primitive dependencies
transitive native/ffi dependencies
whether the public name tells the truth
whether every portable semantic capability has a native path
whether public helpers actually belong to the module contract
```

The audit must inspect existing `/native` modules as aggressively as `/ffi` modules. A file named `/native` is not accepted as native merely because of its path.

The prelude must also be checked. Prelude intrinsics are legitimate only when evaluator-privileged, constitutive of Aiki's value model, or otherwise irreducible. Reducible standard-library algorithms belong in `lib/`, where portable semantics require a native path.

## Baseline Inventory Findings and Clean-Pass Disposition

The Gate 1 inventory established these baseline facts. The clean implementation pass uses them as findings to resolve without editing frozen behavioral witnesses.

### `bytes/native`

`bytes/native` currently imports `math/ffi` and invokes raw byte/shaping helpers. Therefore the `/native` label cannot simply be trusted. Each dependency must be separated into constitutive value machinery versus provider-backed library semantics, and provider-backed semantic dependencies must be removed from the native path.

### `math/native` and `math/ffi`

The public surfaces do not match. `math/ffi` currently provides `floor`, `ceil`, and `modulo` that are absent from native; native `sin`, `cos`, and `sqrt` have a different callable shape from the FFI versions.

The pair must be reconciled around one Aiki-defined semantic contract if these operations remain portable acceleration pairs.

### `string` and `string/ffi`

Bare `string` is documented/used as the reference implementation but currently imports `math/ffi` and has provider primitive authority for character/case operations. The audit must produce a truthful native/reference path rather than grandfathering this hidden mixing.

### `hash/native` and `hash/ffi`

`hash/native` is a genuine Aiki implementation, but the baseline `hash/ffi` was not a truthful FFI realization: almost the entire hash-map algorithm was still implemented in Aiki and only bucket modulo crossed to the provider. This pass must make the reverse promise as strong as the native promise. An `/ffi` implementation may compose provider-backed operations in Aiki, but it may not merely reuse the native semantic authority or leave the substantive implementation native while advertising an FFI realization.

### `regex/ffi`

`regex/ffi` is provider-backed and explicitly describes RE2/Go-regexp semantics. It must not be forced into an acceleration pair by filename convention. The audit must decide whether Aiki intends to own portable regex semantics, in which case a native path is required, or intentionally exposes provider regex interoperability, in which case it should remain provider-only and be classified/documented as interop.

### `store`

`store` exposes an Aiki runtime capability: mutable indexed storage backed by a runtime value. Core operations such as `new`, `get`, `set`, `length`, and `snapshot` are not ordinary acceleration FFI and do not require a fake pure implementation.

`store.digits_to_text` is a performance-shaped conversion helper and `store.checksum` is diagnostic/workload behavior. They must be audited separately from the store capability itself. They should not become permanent Store semantics merely because the runtime can implement them efficiently.

### `path`, `turtle`, and capability-bottomed native modules

Do not classify a library as mixed merely because its native Aiki semantics eventually use an irreducible host capability.

`turtle` and `turtle/simple` are native Aiki abstractions: movement, heading, coordinate transforms, state transitions, and their mathematical work are implemented in Aiki. Their drawing endpoint legitimately bottoms out in the `canvas` capability. Their slow/reference math path is therefore `math/native`, not `math/ffi`.

This is allowed:

```
turtle -> math/native
       -> canvas capability -> HAL/substrate
```

This is not allowed for the default/native path:

```
turtle -> math/ffi
```

unless the programmer explicitly selects an FFI realization.

The rule is:

> Native means that the library algorithm is implemented in Aiki wherever a native semantic path exists. Native code may bottom out in irreducible runtime or host capabilities; it may not silently substitute provider-backed library algorithms.

`path` remains a separate case because its current surface intentionally combines portable path manipulation with host-sensitive behavior and must be audited affordance by affordance.

## Declarative Metadata

Add one module-layer authority for semantic classification. Do not infer role from `/native`, `/ffi`, HAL primitive names, or provider provenance.

A candidate shape is:

```
type StdlibSemanticRole string

const (
    RolePortable          StdlibSemanticRole = "portable"
    RoleRuntimeCapability StdlibSemanticRole = "runtime-capability"
    RoleHostCapability    StdlibSemanticRole = "host-capability"
    RoleInterop           StdlibSemanticRole = "interop"
    RoleInternal          StdlibSemanticRole = "internal"
)

type StdlibRealization string

const (
    RealizationNative    StdlibRealization = "native"
    RealizationFFI       StdlibRealization = "ffi"
    RealizationIntrinsic StdlibRealization = "intrinsic"
    RealizationMixed     StdlibRealization = "mixed"
)
```

The final representation should stay as small as possible. Function-level metadata should be introduced only where module-level classification cannot honestly express a mixed public contract.

Policy keys should use source-declared canonical packages. Registry-generated bare aliases may derive their policy from the canonical source rather than duplicating facts.

## Relationship to HAL

HAL remains the authority layer for raw primitives.

```
HAL: which trusted source may call _string_split?
stdlib semantic policy: what does string.split promise, and which realizations are allowed?
```

Raw primitive authority remains exact:

```
raw primitive used => authority granted
authority granted  => raw primitive used
```

The stdlib policy must not turn HAL into a library-design registry, and HAL role metadata must not be treated as sufficient proof that a module is semantically native or FFI.

## Coarse Provider Boundaries

Provider-backed acceleration should collapse coarse semantic boundaries.

Good:

```
split this whole string
join this whole list
decode this whole digit payload
```

Bad:

```
call FFI once per character
call FFI once per cell
call FFI once per tiny arithmetic step
```

Performance is a reason to choose an FFI realization, not a semantic role.

## Enforcement Targets

The project should end with invariants covering at least:

```
all shipped modules classified
portable semantic capabilities have native paths
/native realizations do not depend on provider-backed library semantics
/ffi realizations do not fall back to their native semantic authority
acceleration references/semantic authorities exist
native and FFI public surfaces match where FFI is acceleration
help/doc public surfaces match those contracts
shared behavior contracts pass against native and FFI
HAL authority remains exact
bare aliases/default facades resolve according to declared policy
```

Expected failures should be architectural and specific, for example:

```
bytes/native imports provider-backed math/ffi
math/ffi exports floor but math/native does not
string/native uses provider primitive _string_split
portable module foo/ffi has no native realization
store.checksum is public but classified internal-only
```

## Proposed Cuts

### Gate 1: Inventory and declare semantic roles

Audit all shipped library modules plus prelude. Record semantic role, realization, native-path status, and residual misclassifications. Replace filename-based assumptions with declarative metadata.

No intentional user-visible behavior change is required merely to establish the authority, but false classifications discovered during inventory must be recorded rather than hidden.

### Gate 2: Make native paths truthful and complete

For every portable semantic capability, establish a genuine native Aiki realization. Remove transitive provider-backed library dependencies from `/native` implementations. Resolve FFI-only affordances by native implementation, honest interop/capability classification, or removal from public stdlib scope.

This is the core semantic repair gate.

### Gate 3: Enforce native/FFI surface parity for acceleration pairs

Where FFI is acceleration, enforce exact public semantic surface parity in both directions.

### Gate 4: Enforce help/doc parity and default-facade policy

Document native, FFI, capability, and interop roles accurately. Ensure bare aliases/default facades have an explicit and truthful policy.

### Gate 5: Shared behavior contracts and residual lib/prelude cleanup

Run the same semantic contracts against native and accelerated realizations. Remove or relocate workload-specific helpers. Recheck prelude irreducibility and module boundaries.

### Gate 6: FFI-use diagnostic

Add a non-failing user-facing diagnostic that reports direct and transitive FFI/provider-backed module imports without changing import mechanics.

### Final reconciliation

Re-audit every shipped module and prelude against the final rules, reconcile the proposal to implementation, and run the strongest repository gate available.

## Clean-Pass Architecture on `native-ffi-separation-clean`

The clean branch currently realizes the policy as follows, while preserving all 395 pre-existing behavioral witnesses byte-for-byte:

```
portable/native + optional FFI acceleration:
  bits, bytes, hash, string

portable/native only:
  list, number, turtle, turtle/simple

portable native math plus explicit provider interop:
  math/native, math/ffi

runtime capability:
  io, store

host capability:
  canvas, file, net, path, process, random, signal, system, term, time

provider interop:
  regex/ffi

internal/tooling:
  profile, selfhost/bootstrap, test, string/case_table
```

Bare names backed by `/native` resolve to the native implementation. Portable native modules are checked for direct and transitive FFI reachability. Portable FFI acceleration modules are checked against a native semantic authority for exported/signature/help/doc parity and may not import that authority as a fallback implementation. `hash/ffi` was converted from a mostly-Aiki implementation with provider modulo into a provider-backed realization of the same hash representation and contract.

The prelude contains no provider-role library primitives. Its raw operations remain limited to evaluator/value/runtime atoms and irreducible services. `turtle` and `turtle/simple` remain native Aiki algorithms using `math/native`; their drawing endpoint is the canvas capability.

The clean history preserves a dedicated regression checkpoint at `1f4e906` (`native-ffi-preservation-candidate`) where baseline `store.digits_to_text` and `store.checksum` still exist and all 395 pre-existing witness blobs are unchanged. Post-preservation commit `e5e4fe3` then removes those misplaced workload/performance helpers as an explicit API cleanup, updates Four-Way Life to compose `store.snapshot` with `bytes/ffi` and its own Aiki checksum, and retires only the directly obsolete Store tests. This separation keeps the preservation claim independently testable even though the final API is intentionally smaller.

`aiki check --ffi-use file.ai` reports direct and transitive FFI module use without changing execution or requiring declarations. The primitive architecture role formerly named `native` is now `runtime`, reserving that category for constitutive language/value/runtime atoms and leaving `provider` as the truthful label for FFI implementation primitives.

Clean-pass evidence is intentionally split: frozen baseline witnesses remain unchanged; new architecture tests live in new files; a neutral relational corpus compares native and FFI realizations; and disposable mutation checks prove the architectural guards detect native→FFI leakage, FFI→native fallback, default-facade hijacking, provider-role mislabeling, and turtle acquiring FFI transitively.

## Non-Goals

This proposal does not require every portable module to have an FFI implementation.

It does not require fake native implementations for genuine host or runtime capabilities.

It does not prohibit deliberate provider interoperability.

It does not make FFI the default.

It does not prohibit programs from mixing native and FFI imports.

It does not make FFI a permission system.

It does not require constitutive evaluator/value primitives to be reimplemented as ordinary Aiki library algorithms.

## Final Rules

Put these rules in the repository architecture:

```
Aiki owns the semantics of every portable standard-library capability.

Every portable semantic capability has a genuine native Aiki path.

A /native name must truthfully select that native realization.
A /ffi name must truthfully select a provider-backed realization.
Programs may mix them freely.

FFI acceleration may not add, remove, or change the native semantic contract.

Genuine runtime/host capabilities and deliberate provider interoperability are
not acceleration. They must be classified and documented honestly rather than
forced into fake native/FFI pairs.

HAL governs authority below the boundary.
Stdlib semantic policy governs promises above the boundary.

FFI visibility is diagnostic; import mechanics remain unchanged.
```
