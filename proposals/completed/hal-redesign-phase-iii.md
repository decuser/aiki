# HAL Redesign — Phase III: Replacement Contract and Migration Design

> **Status: GATED (design gate)**
>
> Baseline: `v0.4.0-alpha-27`
>
> Phase II: [`hal-redesign-phase-ii.md`](hal-redesign-phase-ii.md)

# Cut III.1 — Canonical HAL identities and metadata

**Status: GATED**

The replacement architecture gives a canonical identity only to **irreducible
host crossings**. Native implementation alone does not create a HAL operation.

The naming invariant remains:

```text
Aiki name       = programmer meaning
HAL name        = architectural contract
Substrate name  = implementation realization/provenance
```

The relationship is not required to be one-to-one. A pure Aiki affordance may
have no HAL name; one Aiki library operation may compose several HAL contracts;
several Aiki functions may share one narrow host operation.

## Host-operation contract record

The architectural contract for a host operation needs enough metadata to be
inspectable and executable without freezing a particular Go representation.
The working fields are:

```text
HAL identity        HAL.<domain>.<operation>
required authority  exact operation or named grant set
context facets      source/callsite, observation, runtime service as required
effect class         bounded call / resource acquisition / async source / termination
blocking class       nonblocking / may block / waits externally
state/lifetime       stateless / returns resource / operates on resource
optionality          constitutive / optional substrate capability
error contract       Fault vs shaped recoverable error and translation rules
semantic observation Aiki-level work identity
substrate binding    provenance name + implementation callable
```

`HAL.file.stat` is an example of such an identity. `_file_stat` is at most a
migration-era implementation binding; it is not the architecture name.

## Separate registries/namespaces

Phase-I classification requires at least conceptual separation of:

### Language/evaluator intrinsics

`apply`, import/use/export, load, spawn, and related evaluator-sensitive
operations. These use an intrinsic context and may coordinate host services,
but they are not looked up as generic host effects.

### Native language/value primitives

Operations required to implement Aiki values efficiently or irreducibly:
list/value operations, exact-number helpers, bits, Store, Channel machinery,
conversions, etc. These are implementation primitives, not host authority.

### Native/FFI library providers

Regex, Unicode, inexact math, and similar optional native realizations. They
carry implementation provenance but do not acquire host authority merely by
being written in Go.

### Host HAL operations

Only irreducible host relationships such as filesystem, process environment /
execution, clocks/timers, standard streams, source provision, and optional host
resources.

### Runtime/tooling services

Module cache, help/test/REPL presentation, profiling coordination, resource
cleanup, and similar services are owned/invoked directly by runtime/tooling
architecture rather than masquerading as user host capabilities.

This is the central structural correction to the current one-registry model.

# Cut III.2 — Runtime/substrate contract shape

**Status: GATED**

The following is a design shape, not frozen Go API.

## Runtime

A runtime owns one explicit host world:

```text
Runtime
  HostOperations      canonical HAL contracts + current substrate bindings
  AuthorityPolicy     grants for prelude/trusted modules
  ModuleSystem        source provider + registry/cache
  IO                   stdin/stdout/stderr-like endpoints
  ProcessContext       args, environment view, working-directory context
  Clock                time source/timers
  Random               deterministic RNG state / optional entropy source
  Resources            runtime-scoped host-resource table + cleanup
  AsyncFaults          spawned-fault coordination
  Observation          semantic/substrate correlation
```

This does not imply each line becomes a Go interface. It records ownership so
package globals cannot silently become architecture.

## Intrinsic context

Evaluator-sensitive operations receive the context they actually need:

```text
IntrinsicContext
  lexical/module environment as required
  grammar/parser/evaluator services as required
  source provenance
  dynamic execution state
  observation
  runtime reference
```

This is where import/load/apply/spawn belong.

## Host-call context

Ordinary HAL operations receive a much smaller context:

```text
HostCallContext
  source/callsite identity
  authority grant
  observation correlation
  runtime/resource access required by that contract
```

They do not receive arbitrary `*Env`, grammar, or evaluator callbacks.

## Substrate binding

A substrate binds canonical HAL identities to realizations:

```text
HAL.file.stat   -> go:os.Stat
HAL.time.now    -> go:time.Now
HAL.process.exec -> go:os/exec.Command...
```

Substrate provenance is inspectable for profiling/debugging but does not define
the Aiki semantics.

## Module/source provider boundary

Module loading is decomposed as:

```text
resolve module identity        runtime/module registry
obtain source                  source provider (host/embedding)
assign authority               runtime authority policy
tokenize/parse/evaluate        language engine
collect exports                language semantics
cache module identity/state    runtime module service
```

This allows an embedded source provider to replace filesystem acquisition
without changing import semantics.

## Host resource representation

Core semantic values should not embed concrete Go host objects. A file, Canvas
session, or process handle may retain its Aiki-visible kind and inspectability,
but concrete state is runtime-owned and referenced opaquely. The exact Go value
representation remains an implementation choice for migration design.

Language-native `Store` is explicitly outside this host-resource mechanism.

## Host-backed asynchronous endpoints

The existing event-channel precedent is preserved. A host service may produce
an Aiki channel endpoint whose peer is substrate/runtime code. The channel
retains ordinary Aiki receive/select semantics. Domain protocols remain above
the generic concurrency machinery.

# Cut III.3 — Serial migration plan

**Status: GATED**

Implementation is intentionally ordered so each slice can be validated against
the current language before removing compatibility machinery.

## Migration M1 — Add contract metadata beside current behavior

- introduce canonical HAL identities/descriptors for true host crossings only;
- record substrate provenance names;
- add executable validation that every host binding has one canonical identity
  and that public Aiki wrappers document/map their host dependencies;
- retain existing `_` calls as compatibility plumbing.

**Gate:** no behavior change; current suite/golds pass; three-name mapping is
machine-checkable for migrated host operations.

## Migration M2 — Split the native registry by semantic role

- separate evaluator intrinsics from host HAL operations;
- separate language/value primitives and native/FFI providers from host HAL;
- move REPL/test/profiling service entry points out of the generic host
  namespace where appropriate;
- retain compatibility aliases until callers migrate.

**Gate:** user visibility and public library behavior unchanged; raw host
operations remain inaccessible to ordinary user code.

## Migration M3 — Introduce runtime/session ownership

- move program args, I/O endpoints, module registry/cache, RNG, async faults,
  file auxiliary state, Canvas/session tracking, and resource cleanup under
  explicit owners;
- remove package-global state serially;
- establish runtime-scoped opaque host-resource references without changing
  Aiki-facing behavior.

**Gate:** independent runtimes/tests do not share mutable host state unless
explicitly configured to do so.

## Migration M4 — Split context and authority

- replace blanket `EvalContext` use with intrinsic context vs minimal host-call
  context;
- introduce explicit authority grants keyed by canonical HAL identity/domain;
- assign trusted-module grants through runtime policy;
- preserve lexical authority through function calls;
- correct spawn so isolated user computation does not acquire prelude host
  authority merely to see prelude vocabulary.

**Gate:** explicit tests prove ordinary user code and user-spawned code cannot
resolve raw host operations; trusted library functions retain only declared
grants; callbacks do not inherit caller privilege.

## Migration M5 — Realize the systems-programmer surface

Migrate and fill the agreed first-order affordances through the new contract:

```text
standard I/O
filesystem files + metadata/directories
path library (pure Aiki where possible)
args/environment/working-directory view
process exec, then start/wait if still required
time now/sleep/after
random runtime ownership
```

**Gate:** public API examples and executable documentation trace each host-backed
affordance through Aiki name -> HAL identity -> substrate provenance.

## Migration M6 — Canvas pressure migration

- remove the one-native-operation-per-drawing-command assumption where Aiki can
  own the protocol;
- move concrete Canvas session state under runtime resources;
- use ordinary Aiki communication semantics for host events where applicable;
- keep `select` generic;
- do not introduce universal transport machinery unless another concrete
  facility demonstrates the need.

**Gate:** Canvas/turtle behavior remains executable while generic HAL and
concurrency code contain no Canvas operation semantics.

## Migration M7 — Remove compatibility layer and strengthen couplings

- remove obsolete `_` host aliases and package globals;
- update selfhost/bootstrap, docs, help, lint/tooling, and architecture docs;
- make Aiki/HAL/substrate name/provenance coverage executable;
- run full validation/race/smoke/gold/tree checks.

**Gate:** replacement architecture is the only path and all earned invariants
remain green.

M1-M7 are implementation migrations, not additional design cuts. Each would be
worked as serial evidence-gated implementation cuts when this design is
accepted for implementation.

# Cut III.4 — User-affordance and invariant review

**Status: GATED**

The design was traced from the systems-programmer surface through the three-name
model.

## Filesystem metadata

```text
Aiki:       file.stat(path)
HAL:        HAL.file.stat
substrate:  go:os.Stat
```

`file.stat` defines the Aiki shaped result and error semantics. The HAL contract
defines required file authority and host-error translation. Go's `os.FileInfo`
is not an Aiki value.

## Pure path work

```text
Aiki:       path.join(a, b)
HAL:        none
substrate:  none
```

This is an important success case: systems-programmer usefulness does not imply
host crossing.

## Process execution

```text
Aiki:       system.exec(cmd, args)
HAL:        HAL.process.exec
substrate:  go:os/exec.Command + pipes/wait
```

The Aiki library owns the shaped result and public convention; the HAL contract
owns process authority, blocking/error behavior, and realization boundary.

## Current time

```text
Aiki:       time.now()
HAL:        HAL.time.now
substrate:  go:time.Now
```

Formatting/parsing remains Aiki-side unless concrete requirements justify host
facilities.

## Timer/select

```text
Aiki:       time.after(ms)
HAL:        HAL.time.after
substrate:  go:time.AfterFunc -> host-backed Aiki event channel
```

`select` has no time-specific branch.

## Canvas pressure

A likely shape is intentionally less settled:

```text
Aiki:       canvas.line(...), canvas events, etc.
HAL:        minimal optional resource/session crossing(s)
substrate:  Go IPC/Ebiten session realization
```

The line/rectangle/color protocol should remain Aiki-visible/library meaning
where practical. The exact acquisition HAL identity is deferred to the Canvas
migration because the architecture does not need to invent a universal
transport today.

# Preservation review

The design preserves or strengthens the baseline commitments:

1. exact-number semantics remain language semantics rather than host-number
   behavior;
2. Aiki libraries remain real implementations;
3. ordinary user code receives no raw host authority;
4. spawn isolation no longer relies on a privilege-bearing scope distinction;
5. Store and ordinary channels remain explicit Aiki capabilities;
6. host resources become explicit runtime-owned capabilities without concrete
   Go objects in core values;
7. relative source provenance remains distinct from captured mutable state;
8. module evaluation remains isolated from arbitrary importer bindings;
9. semantic observation remains distinct from substrate observation;
10. Canvas cannot infect generic `select`/HAL code with domain semantics;
11. new systems affordances are added because real programs need them, not for
    architectural symmetry;
12. the three-name relationship becomes executable project evidence rather than
    prose convention.

# Phase-III and design-project gate

**GATED.**

Cut 0 plus Phases I-III now provide a coherent design account and migration
sequence. No replacement HAL code has been implemented in these design phases.

The next project action, if accepted, is implementation Migration M1: add
canonical host-operation metadata and executable three-name/provenance coverage
beside existing behavior, with no semantic change.
