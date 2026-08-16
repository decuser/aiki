# HAL Redesign — Phase II: Context, Authority, and Ownership

> **Status: GATED (source-inspection/design gate)**
>
> Baseline: `v0.4.0-alpha-27`
>
> Phase I: [`hal-redesign-phase-i.md`](hal-redesign-phase-i.md)

# Cut II.1 — Decompose execution context

**Status: GATED**

The baseline `Env` currently carries several semantically different things:

```text
bindings                 store + outer chain
structural vocabulary    shapes
source provenance        file + source + sourceLines
dynamic execution        stack + stackLimit
observation              semanticProbe
authority                 ScopeUser / ScopePrelude
module declaration       exports + packageName
```

`EvalContext` then adds grammar, active syntax node, evaluator callbacks,
measurement callbacks, profile labels, and async-fault plumbing.

The redesign should treat these as facets, not as one `*Env` capability handed
to any native function that asks for context.

## Propagation model derived from existing behavior

### Ordinary function call

The existing `NewCallEnv(lexicalOuter, caller)` already encodes the right
fundamental distinction:

- lexical bindings and shapes come from the function definition;
- source identity follows the defining lexical chain;
- dynamic stack/limit and semantic probe come from the caller.

The redesign should preserve this split and make it explicit.

### Spawn

Spawn intentionally drops the user's lexical value bindings while retaining:

- prelude vocabulary;
- explicitly supplied arguments;
- fresh dynamic call stack;
- copied stack limit;
- semantic probe;
- shape vocabulary from the function definition;
- defining file provenance for relative import.

This proves that “context” is already decomposable in practice. Source
provenance and shape vocabulary are interpretation metadata, not captured user
state.

A further source-inspection finding is important: `applyUserFunctionIsolated`
currently creates the spawned environment with `NewIsolatedEnclosedEnv` under
the prelude environment. That constructor inherits the outer environment's
`scope`, producing `ScopePrelude`. `evalName` uses that scope when resolving raw
runtime builtins. Therefore the baseline architecture couples isolation to host
privilege and appears to permit a user-defined spawned function to resolve raw
`_` primitives that ordinary user execution cannot resolve.

This is recorded as a **baseline authority leak by construction** and a
migration invariant: authority must be separated from the environment role
used to supply prelude vocabulary. The redesign must not reproduce this
coupling.

### Module evaluation

A module is evaluated in a fresh environment enclosed by the prelude, not by
the importer. It receives its own file/source and its authority is currently
selected from filesystem topology (`lib/` / `contrib/lib/` -> `ScopePrelude`).
Dynamic stack/probe state remains connected through the shared environment
machinery.

### Host invocation

Most host/value primitives need none of `Env`, grammar, evaluator callbacks, or
async-fault machinery, yet the current `ContextCallable` path can receive all of
it. A future host call context should be minimal and explicit: call-site/source
identity, authority, observation correlation, and runtime services only as
needed by that operation. Grammar/Eval/Env belong to language intrinsics, not
the ordinary host HAL.

## Conceptual context facets

Phase II therefore adopts these conceptual facets without yet freezing Go
interfaces:

```text
LexicalContext       bindings and structural/shape vocabulary
SourceContext        source identity + diagnostic source access
DynamicContext       call stack, recursion limit, cancellation/fault relation
ObservationContext   semantic probe + substrate correlation labels
AuthorityContext     granted HAL contracts/capabilities
ModuleContext        package declaration + exports during module evaluation
```

A function definition carries lexical/source/authority meaning. A call supplies
dynamic/observation state. Spawn creates new dynamic state without inventing
new authority.

# Cut II.2 — Runtime/session ownership

**Status: GATED**

The baseline already has a per-`GoRuntime` object, but many host relationships
remain package-global:

```text
Stdout / Stdin
UserEnv
PageOutput
programArgs
GlobalRegistry + module cache
RNG
fileReaders
openCanvases / canvas sessions / bridge state
help registry and related presentation state
```

This prevents the architecture from honestly claiming that a runtime owns its
host relationship.

## Working ownership model

### Runtime

A runtime represents one explicit Aiki host world. By default it owns:

- substrate/HAL implementations and availability;
- authority policy / grants;
- standard input/output endpoints;
- program argument snapshot and environment view;
- working-directory context when introduced;
- module source provider/registry/cache;
- deterministic RNG state and optional entropy source;
- async-fault coordination;
- semantic/substrate observation correlation;
- host-resource manager and cleanup;
- optional host capabilities such as graphics.

Sharing a runtime between embeddings is an explicit choice and therefore makes
shared module/resource identity explicit rather than accidental.

### Evaluation session

A session owns one evaluator interaction over a runtime:

- grammar/evaluator instance;
- prelude/user environments;
- dynamic execution state;
- current source/evaluation interaction;
- REPL-specific presentation hooks such as paging when applicable.

The normal CLI/REPL case may remain one session over one runtime. The
separation exists so tests, servers, notebooks, and other embeddings can choose
whether to share a runtime deliberately.

### Resource ownership

File, Canvas, and future process handles expose Aiki-level resource capability,
but their concrete host state and cleanup belong to the runtime resource
manager. Auxiliary state such as `fileReaders` must follow the resource owner,
not live in package globals.

This also exposes a current substrate leak: core `value.File` embeds `*os.File`,
and `value.Canvas` embeds Go `image/color`, Go channels, and Canvas command
structures. The redesigned semantic value layer should retain Aiki-visible
resource identity/inspectability without importing the concrete substrate
representation into the core value package.

`Store` is explicitly **not** a host resource. It is an Aiki mutable capability
and remains part of language/runtime semantics.

# Cut II.3 — Capability and authority model

**Status: GATED**

The current `ScopeUser` / `ScopePrelude` mechanism serves two unrelated roles:
lexical/environment role and authority to resolve raw native primitives. The
redesign separates them.

## Working authority decisions

1. **Canonical grants are expressed in HAL identities**, e.g.
   `HAL.file.open`, `HAL.file.stat`, or domain-level sets where appropriate.
2. **Authority is lexical/definition-bound**, not dynamically inherited from
   the caller. A trusted library function retains the authority with which it
   was defined; calling an untrusted callback from trusted code does not grant
   that callback the caller's host authority.
3. **Ordinary user modules receive no raw host-operation grants.** They use
   public Aiki library functions and explicit resource values.
4. **Trusted Aiki libraries receive only the grants they need.** The file
   library should not gain Canvas operations merely because both are blessed
   libraries.
5. **Filesystem path may bootstrap policy, but does not define authority.** A
   module resolver may identify a canonical trusted module and the runtime
   policy assigns its grants.
6. **Prelude language privileges and host authority are separate.** The prelude
   may need evaluator intrinsics without thereby receiving every host
   capability.
7. **Resource values remain explicit capabilities.** Passing a file/process/
   graphics resource across spawn is an explicit program relationship; the
   concrete resource remains owned by the runtime.
8. **Optional host capability is normal absence**, not a new architectural
   category. The runtime declares which HAL contracts are available.

This model preserves Aiki's existing useful pattern: programmer-facing library
functions can be ordinary closures defined in a trusted module. The function's
lexical authority lets it cross the HAL; the user gets the meaningful Aiki API,
not the raw crossing.

# Cut II.4 — Stateful/asynchronous pressure gate

**Status: GATED**

## Existing precedent: `time.after`

The baseline already supports a host-produced event through ordinary Aiki
concurrency:

- Go `time.AfterFunc` sends `true` into `value.NewEventChannel()`;
- the event channel is receive-only from Aiki's perspective;
- evaluator `select` uses the same receive machinery for that channel;
- `TestSessionSelectWithTimeAfter` validates competition between an ordinary
  Aiki channel and the host-produced timer event.

Therefore **host-backed selectable event endpoints are existing Aiki behavior**.
No Canvas-specific select semantics are needed or justified.

## Canvas pressure result

Canvas still does not define the generic architecture. It pressures four
requirements simultaneously:

```text
optional host capability
host-owned resource lifetime
Aiki -> host commands
host -> Aiki asynchronous events
```

The leading Canvas direction remains:

```text
trusted Aiki canvas library
        -> minimal host resource/session acquisition
        -> ordinary Aiki communication endpoints where they fit
        -> Aiki-defined protocol/domain meaning
```

The exact acquisition contract and whether a separate opaque lifetime token is
necessary remain Phase-III contract questions. What is settled is:

- generic `select` remains Canvas-ignorant;
- domain operations such as line/rect/color should not require one HAL
  primitive each if the Aiki library can express the protocol;
- no universal `Transport` type is introduced from Canvas alone;
- host-backed channels are evidence-backed machinery, not a speculative new
  concurrency subsystem;
- concrete Canvas process/session state belongs to runtime resource ownership,
  not the semantic `Canvas` value implementation.

# Phase-II gate result

**GATED by source inspection and architectural consistency.**

The design now distinguishes:

```text
lexical meaning
source provenance
execution state
observation state
authority
module evaluation state
runtime/session ownership
resource ownership
```

and can explain ordinary calls, spawn, module evaluation, host invocation,
files, timers/select, and Canvas without relying on one undifferentiated
`*Env` or on package-global host state as an architectural premise.

## Phase-III entry requirements

The replacement contract must now encode these decisions while retaining the
three-name invariant and the systems-programmer affordance target. It must also
provide a serial migration that fixes the spawn authority coupling rather than
merely renaming it.
