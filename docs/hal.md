# Host Abstraction Layer

Aiki's HAL is the explicit boundary between language/library meaning and the
host substrate. It is intentionally smaller than the runtime primitive registry:
being implemented in Go does not make an operation a host capability.

## Three names

Every irreducible host crossing has three distinct identities:

```text
Aiki name       = programmer meaning
HAL name        = architectural contract
Substrate name  = implementation realization/provenance
```

For example:

```text
file.stat       -> HAL.file.stat       -> go:os.Stat
```

These names are not aliases. Aiki libraries own programmer-facing semantics,
including shaped results and composition. The HAL records the host contract:
authority, required context, blocking/effect/lifetime class, error translation,
semantic observation identity, and substrate provenance. The substrate supplies
mechanism.

A useful systems-programmer operation does not necessarily cross the HAL. The
`path` library, for example, implements joining, base/dir/ext extraction and
normalization in Aiki; only host facts such as the path separator require a host
contract.

## Runtime and library roles

The runtime primitive registry and the standard-library semantic policy answer
different questions. Runtime primitive roles distinguish constitutive/runtime
atoms from provider implementations and canonical host operations. Library
policy, in `engine/runtime/modules/stdlib_policy.go`, separately classifies
shipped modules as portable semantics, runtime capabilities, host capabilities,
provider interop, or internal support, and records whether each realization is
native, FFI, intrinsic, or intentionally mixed.

This separation matters: a Go implementation is not automatically a host
capability, and an FFI realization is not automatically interop. Portable
`X/native` modules remain Aiki implementations transitively; portable `X/ffi`
modules are provider-backed realizations of an explicit native semantic
authority. Host/runtime capabilities do not require fake native twins.

Only irreducible host relationships receive canonical `HAL.<domain>.<operation>`
identities. The compatibility lookup machinery is an implementation detail, not
the architecture.

## Relation to native and FFI modules

HAL authority governs **which trusted source may cross a runtime/host boundary**.
The standard-library policy governs **what semantic realization the programmer
selected**. The two locks are intentionally independent:

```text
portable semantic capability
    -> native Aiki realization (required)
    -> optional provider-backed /ffi realization

runtime/host capability
    -> Aiki-facing contract
    -> HAL authority where irreducible
```

A `/native` module may use constitutive runtime atoms, but it may not implement
its library algorithm by reaching an FFI/provider path. Conversely, a portable
`/ffi` module must actually reach provider primitives and may not fall back to
its native semantic authority. Capability modules may cross HAL directly because
the crossing is part of the capability rather than an acceleration choice.

## Authority

Lexical scope and host authority are independent. `ScopePrelude` remains a
lexical/tooling role; it does not grant host access. Trusted Aiki sources receive
explicit definition-bound authority. For canonical host operations, grants are
expressed by HAL identity (for example `HAL.file.open`), not by the raw substrate
binding (`_file_open`).

Ordinary user code receives no raw host authority. It calls ordinary Aiki
library functions whose definitions carry only the contracts required by their
implementation. Calling an untrusted callback from trusted code does not confer
the caller's host authority. Spawn retains prelude vocabulary and source
provenance without acquiring prelude host privilege.

## Ownership

A `GoRuntime` represents one explicit host world. Runtime-owned state includes
standard I/O endpoints, program arguments and environment view, working
directory, module registry/cache, RNG state, asynchronous-fault coordination,
file auxiliary state, help/session hooks, and host resources. Independent
runtimes therefore do not share those relationships accidentally.

Host resources are explicit Aiki values but their concrete state is owned below
the semantic value layer. A Canvas value, for example, is an opaque handle; the
owning runtime maps it to a substrate `CanvasResource` containing process/session
state, queues, drawing state, and cleanup. `Store` remains an Aiki-defined runtime mutable capability rather than a host resource.

## Canvas pressure case

Canvas is a stress test, not the HAL's organizing abstraction. Drawing meanings
are defined in Aiki and cross the host boundary through one protocol contract:

```text
canvas.line(...)
    -> HAL.canvas.command
    -> Go Canvas/IPC realization
```

The canonical Canvas boundary consists of resource open/close/query operations
and the generic command crossing. `select` and ordinary Aiki concurrency contain
no Canvas-specific semantics. No universal transport abstraction is implied.

## Executable couplings

Repository invariants enforce the architecture rather than leaving it as prose:

- every canonical host contract has a unique HAL identity and substrate
  provenance;
- every declared Aiki binding actually uses the named substrate binding;
- trusted sources receive the canonical HAL authority for host crossings and do
  not receive the raw host binding as authority;
- trusted-source dependency declarations match their actual raw runtime
  dependencies;
- obsolete per-command Canvas compatibility primitives are absent;
- semantic Canvas values cannot regain concrete substrate state.

The detailed design and migration record is retained under
`docs/session-history.md`.
