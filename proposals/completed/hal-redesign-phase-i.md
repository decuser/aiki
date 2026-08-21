# HAL Redesign — Phase I: Boundary and Affordance Map

> **Status: GATED (source-inspection/design gate)**
>
> Baseline: `v0.4.0-alpha-27`
>
> Cut 0: [`hal-redesign-cut0.md`](hal-redesign-cut0.md)
>
> Full primitive inventory: [`hal-redesign-phase-i-inventory.md`](hal-redesign-phase-i-inventory.md)

## Phase-I question

Before replacing the HAL, determine what the existing boundary actually contains
and whether it can cleanly realize the host-facing affordances expected by an
Aiki systems programmer.

The organizing invariant is the three-name chain:

```text
Aiki name       = programmer meaning
HAL name        = architectural contract
Substrate name  = implementation realization/provenance
```

The redesign is judged from the programmer surface inward. It is not a registry
cleanup project whose success is measured by a prettier Go interface.

# Cut I.1 — Inventory the current boundary

**Status: GATED**

The current Go substrate registers **117 `_` primitives** through one
`GoRuntime` registry. Every registration was traced to its Go implementation
file and to production Aiki/prelude use where present. The complete table is in
the companion inventory document.

The registry groups are:

```text
Canvas                                    17
File I/O                                  12
Math                                       9
Evaluation-context intrinsics              9
Test framework                             8
List                                       7
Bytes                                      7
Type                                       6
Convert                                    6
Regex                                      6
Bits                                       6
String                                     5
REPL                                       5
Store                                      4
Host program environment                   3
Semantic profiling                         3
I/O                                        2
Time                                       2
                                         ---
                                         117
```

This immediately establishes that the present registry is **not synonymous
with a host HAL**. It is a common native-function registry containing several
architecturally different things.

Adjacent state was also inventoried because it affects the boundary even though
it is not registered as a primitive:

- `RuntimeContract` currently consists of `Execute`, `HasBuiltin`, and
  `GetBuiltin`;
- `EvalContext` carries `Env`, syntax node, grammar, semantic probe,
  measurement callbacks, profile labels, async-fault plumbing, and evaluator
  callback;
- `GoRuntime` owns the builtin registry, profile-label state/cache, and
  async-fault channel;
- package globals currently include stdin/stdout, REPL `UserEnv`, pageable
  output, program arguments, the module registry/cache, random generator,
  file-reader cache, and Canvas session state.

These adjacent items become Phase-II ownership/context work rather than being
forced into the primitive inventory.

# Cut I.2 — Classify what the registry actually contains

**Status: GATED**

The 117 registrations fall into the following architectural classes. The
boundaries are intentionally semantic rather than package-based.

## A. True host effects and host resources

Examples:

```text
print/read        -> host standard I/O
sleep/after       -> host clock/timer
system.env        -> process environment
file.*            -> filesystem/file handles
canvas.*          -> optional host-owned graphical session
random            -> host/substrate nondeterminism
```

These are candidates for actual `HAL.<domain>.<operation>` contracts because
some irreducible host relationship remains after Aiki-level semantics are
pushed upward.

## B. Language/evaluator intrinsics

Examples:

```text
apply
import/use/export
load
spawn
channel/send/recv
```

These are currently native registrations but are not ordinary host services.
They participate in evaluation, scope, module semantics, isolation, or Aiki
concurrency. Some contain host-backed portions, but their language semantics
must not be defined as generic HAL effects.

## C. Language/value operations currently implemented natively

Examples:

```text
first/rest/length/prepend/append
shape/type/equal/inspect/conversions
bytes representation operations
Store operations
bit operations
floor/ceil/truncate/modulo
```

These manipulate Aiki values and semantics. Their Go implementation may be
necessary or efficient, but that does not make them host capabilities. The new
architecture needs a place for irreducible/native language primitives that is
conceptually separate from the host HAL.

## D. Optional library accelerators / FFI realizations

Examples:

```text
Unicode case conversion
regexp-backed regex
inexact trig/sqrt
some bytes conversions
```

These are substrate implementations of library behavior, often paired with
Aiki/native alternatives. They need provenance and perhaps contract metadata,
but they should not be confused with host authority merely because Go performs
them.

## E. Observation, tooling, and session services

Examples:

```text
semantic profiling primitives
REPL help/reset/delete/quit/doc
test framework primitives
module-root exposure
```

These are engine/tooling services. Some may legitimately cross an embedding or
runtime boundary, but they are not systems-programmer host affordances.

## F. Context/ownership concerns masquerading as callable machinery

No registered primitive is classified solely as “execution context,” but the
implementation of imports, spawn, profiling, REPL operations, and module roots
shows substantial context/ownership carried through `EvalContext`, `Env`, and
package globals. This is an explicit Phase-II category rather than being hidden
inside the other five.

### Phase-I classification conclusion

The redesign should not replace today's 117-entry registry with a more ornate
117-entry HAL. It should first separate:

```text
language/evaluator primitive
host HAL contract
native/accelerated library realization
runtime/tooling service
execution context / authority / ownership
```

Only the second of those is the host HAL in the architectural sense used by
Cut 0.

# Cut I.3 — Systems-programmer affordance map

**Status: GATED**

The design target is the ordinary host-facing work a systems programmer should
be able to perform from Aiki. The baseline already has a useful substrate, but
it is incomplete and uneven at the public library surface.

| Affordance | Baseline Aiki surface | Important gaps | Likely irreducible host mechanism |
|---|---|---|---|
| Standard input/output | `print`, `println`, `read`, `input` | explicit/runtime-owned streams; possibly stderr when demanded | reader/writer endpoints owned by runtime |
| Files | `file.open/read_*/write_*/close/read_at/write_at` | stat/metadata, rename/copy, truncate/seek if forced | file-handle + filesystem operations |
| Directories | `file.list`, `exists`, `delete` | mkdir, mkdir_all, recursive removal, temp dirs | filesystem directory operations |
| Paths | ad hoc strings | join/base/dir/ext/normalize/is_absolute; cwd/separator host facts | mostly pure Aiki; cwd/separator host query |
| Program arguments | `system.args` | none essential | runtime/session program-argument snapshot |
| Environment | `system.env` | enumerate/set only if concrete use requires | process/session environment service |
| Process execution | none | synchronous `exec`; later `start/wait` if required | process creation, pipes/capture, wait |
| Working directory | implicit process CWD | `cwd`, `chdir` | runtime/process directory service |
| Time | `sleep`, `after` | `now`; formatting/parsing as justified | clock/timer source |
| Randomness | `random.seed/random` | ownership/determinism policy | runtime RNG / host entropy when requested |
| Bytes/binary | bytes module + bits module | no immediate architectural gap identified | mainly language/native primitives, not host HAL |
| Mutable indexed state | `store` | no immediate gap identified | explicit Aiki capability, not host service |
| Concurrency | spawn/channel/send/recv/select | host-connected event endpoints must preserve same semantics | language concurrency + narrowly host-backed endpoints |
| Stateful host resources | Canvas is current example | generic ownership/lifetime account | resource/session creation + minimal communication mechanism |
| Module source access | `import/load` | disentangle source acquisition from evaluation/authority/cache | source provider / filesystem realization |

The initial surface intentionally does **not** add networking, signals, user/group
inspection, raw devices, shared memory, or generalized IPC merely because
systems languages often have them. Those require concrete pressure.

A key design consequence is that **user affordance and HAL operation need not be
one-to-one**. `path.join`, for example, belongs entirely in Aiki even though it
is a systems-programming affordance. Conversely, one narrow host operation may
support a rich Aiki module.

# Cut I.4 — Pressure tests and synthesis

**Status: GATED**

The classification was tested against the hardest existing cases rather than
against hypothetical future facilities.

## Module loading

Current `_import`/`_load` combine source resolution, filesystem access,
lex/parse/evaluation, module authority, module cache, relative-source context,
and exports. This confirms that module loading must be decomposed across host
source provision, language evaluation, authority policy, and runtime cache
ownership. It should not remain one giant HAL operation.

## Spawn

`spawn` is a language/runtime intrinsic, not a host capability simply because
the Go implementation uses a goroutine. The recently validated relative-import
provenance work also proves that lexical/source interpretation context and
mutable lexical state are distinct things. Phase II must preserve that
separation.

## Profiling

Semantic profiling and Go CPU profiling already demonstrate the three-name
idea in another form: Aiki-level work identity and substrate realization are
correlated at the boundary without being equated. Observation belongs in the
contract metadata/runtime-service model, not in the definition of every Aiki
library function.

## File handles

`value.File` already behaves like an explicit host-backed resource value, but
its buffered-reader cache is package-global. This is a clean example of the
difference between **resource capability at the Aiki surface** and **lifetime /
auxiliary state ownership in the runtime**.

## Canvas

Canvas remains a pressure requirement rather than the architecture's driver.
The present 17 native Canvas operations are conspicuously domain-rich relative
to the proposed narrow boundary. They are evidence that protocol meaning may be
able to move upward into Aiki.

However, the asynchronous side is no longer hypothetical. The baseline already
contains `time.after`, which creates a receive-only `value.Channel` from a Go
`time.AfterFunc` callback. Existing `select` uses the ordinary channel receive
path, and `TestSessionSelectWithTimeAfter` demonstrates that a host-produced
event competes normally with Aiki channels.

Therefore Phase II may investigate host-backed channels for Canvas from an
**existing architectural precedent**, while still refusing to generalize them
into a universal transport abstraction without additional evidence.

## Thompson 7094 pressure

The 7094 project remains evidence for the governing expansion rule. Real
requirements forced bits, Store, select, system arguments, and related systems
substrate without requiring a large new syntactic surface. The HAL redesign
should preserve that ability to expand through small capabilities and Aiki
libraries rather than domain-specific syntax or host leakage.

# Phase-I model

The boundary now has a sufficiently concrete shape to proceed:

```text
SYSTEMS PROGRAMMER
      |
      | Aiki name: meaning / useful affordance
      v
Aiki prelude + libraries
      |
      | HAL name only where an irreducible host crossing remains
      v
HAL.<domain>.<operation>
      |
      | contract: authority, context, errors, observation, cost/blocking
      v
substrate realization
      |
      | substrate name: provenance / actual host behavior
      v
Go / OS / embedded service
```

Alongside that vertical path are three non-HAL architectural concerns that must
be made explicit in Phase II:

```text
language/evaluator primitives
execution context + authority
runtime/session + resource ownership
```

Native accelerators are a fourth implementation category: they may use the
three-name machinery for provenance/observation, but their existence does not
turn an Aiki operation into a host capability.

# Phase-I gate result

**GATED by source inspection and internal consistency.**

- all 117 current registered primitives are present in the inventory;
- all 117 have an implementation location and architectural classification;
- adjacent `EvalContext`, runtime, registry, and package-global state is
  explicitly separated for Phase II;
- the initial systems-programmer affordance surface is mapped to baseline
  coverage and concrete gaps;
- module loading, spawn, profiling, file resources, Canvas, timers/select, and
  the Thompson work fit the named categories without requiring a new category;
- no Go interface or implementation migration has been proposed prematurely.

## Phase-II entry questions

Phase II must now settle, from the baseline rather than from abstraction alone:

1. Which facets of `Env`/`EvalContext` constitute lexical meaning, dynamic
   execution state, source provenance, observation, and authority?
2. What exactly is a runtime/session, and which current globals belong to it?
3. How are host capabilities granted to trusted Aiki libraries without making
   filesystem path the semantic definition of authority?
4. How are resource lifetime and cleanup owned?
5. Can the existing receive-only event-channel precedent extend cleanly to the
   Canvas pressure case while leaving `select` domain-ignorant?
