# Cut 0 — HAL Redesign Proposal

> **Status: GATED**
>
> Baseline: `v0.4.0-alpha-27`
>
> This is a discussion-phase architecture plan. It is intentionally less detailed
> than an implementation proposal because the purpose of the phase is to discover
> and settle the model before freezing Go interfaces or package boundaries.

## Purpose

Redesign Aiki's HAL so that it describes the host boundary explicitly rather
than merely exposing a registry of native functions.

The present HAL successfully established a boundary between Aiki and Go. The
next step is to make clear, for every crossing:

- what Aiki means;
- what contract the architecture promises;
- what substrate operation realizes that contract;
- what execution context or authority is required;
- which semantics belong in Aiki rather than below the HAL.

The desired result is a smaller, more explicit, more substitutable host
boundary. The redesign should not expand the HAL merely to make it symmetric or
future-proof.

## Primary design target — systems-programmer affordances

The redesign is not an architecture-cleanup exercise for its own sake. Its
primary purpose is to let Aiki expose the ordinary host-facing affordances a
systems programmer needs without letting Go or another substrate silently
define their language semantics.

The design therefore proceeds from the user-facing surface inward:

```text
systems-programmer affordance
        -> Aiki library meaning
        -> canonical HAL contract
        -> substrate realization
```

The initial target includes useful filesystem and directory operations, paths,
program arguments and environment, process execution, clocks/timers, standard
I/O, bytes/binary work, randomness, explicit mutable storage, concurrency, and
stateful host resources when concrete programs require them. The HAL exists to
realize that surface cleanly. It is not itself the user-facing API.

The standing three-name invariant is therefore central:

```text
Aiki name       = programmer meaning
HAL name        = architectural contract
Substrate name  = implementation realization/provenance
```

Example:

```text
file.stat       -> HAL.file.stat       -> go:os.Stat
```


## Governing principles

- **Aiki semantics stay in Aiki whenever Aiki can express them.** Native Aiki
  libraries are real implementations, not ceremonial wrappers around Go.
- **The substrate provides mechanism, not accidental language semantics.** Go
  implementation choices should not silently become Aiki behavior.
- **Authority is explicit.** Access to host facilities should be granted by the
  runtime architecture, not arise accidentally from package globals or path
  topology.
- **Runtime state has an owner.** Process state, resource lifetime, module cache,
  asynchronous faults, and related services belong to a runtime/session rather
  than ambient package state.
- **Generalize only under pressure from concrete computation.** The Thompson
  criterion remains operative: add the smallest general capability a real
  program requires.
- **Pressure cases test the design; they do not define it.** Canvas is an
  important requirement because it is stateful, asynchronous, optional, and
  lifecycle-bearing, but the HAL should not be designed around Canvas.

## The three names at every host crossing

Every host-backed operation should be intelligible at three distinct naming
levels.

### 1. Aiki name — programmer meaning

Example:

```text
file.stat
```

This is what the Aiki programmer sees. It belongs to the language/library
surface and carries the programmer-facing semantics, documentation, composition,
and idiom.

### 2. HAL name — architectural contract

Example:

```text
HAL.file.stat
```

This is the canonical identity of the host crossing. It names the domain and
operation independent of a particular substrate implementation.

The HAL identity is where the architecture can attach or verify such things as:

- capability/domain;
- wrapper/contract;
- required authority;
- context requirements;
- Aiki-level observation/profiling identity;
- error translation rules;
- blocking/cost classification where relevant.

### 3. Substrate name — realization and provenance

Example:

```text
go:os.Stat
```

This identifies how the current substrate realizes the HAL contract. It carries
implementation provenance and host-specific facts such as actual cost, blocking
behavior, platform behavior, IPC, system calls, or library dependencies.

The invariant is:

> **Aiki name = meaning. HAL name = contract. Substrate name = realization.**

These names may be mechanically related, but they are not interchangeable.

## Programmer-facing library surface

The HAL redesign exists to support the expansion of Aiki's practical
usefulness. The architecture is invisible to ordinary programs. What the
programmer sees is a set of library modules with clear operations, sensible
errors, and no surprises.

This section describes the intended user-facing surface. The capabilities
listed here are the most frequently needed operations for programs that
interact with the host — the things a programmer reaches for immediately
when writing anything beyond pure computation.

### What exists today

The baseline (`v0.4.0-alpha-27`) already provides:

```text
file       open, read_text, read_bytes, read_line, write_text, write_bytes,
           close, exists, delete, list, read_at, write_at
system     args, env
time       sleep, after
```

These are thin wrappers over HAL primitives. They work, but the surface is
narrow and some common operations require awkward workarounds or are simply
unavailable.

### What programmers need

The following capabilities are ordered by how frequently real programs need
them, not by architectural importance. Each entry shows what the Aiki
programmer writes, not what the HAL provides underneath.

#### Filesystem — `file` and `path`

The existing `file` module covers open/read/write/close. The gaps are
metadata, directory operations, and path manipulation.

Needed in `file`:

```aiki
let info = file.stat(path)        # -> [@stat, size, modified, is_dir, ...]
                                  #    | [@error, :io, message]
file.rename(old, new)             # -> true | [@error, :io, message]
file.mkdir(path)                  # -> true | [@error, :io, message]
file.mkdir_all(path)              # -> true | [@error, :io, message]
file.remove_all(path)             # -> true | [@error, :io, message]
file.temp()                       # -> path string for a new temp file
file.temp_dir()                   # -> path string for a new temp directory
file.copy(src, dst)               # -> true | [@error, :io, message]
file.size(path)                   # -> integer | [@error, :io, message]
```

`stat` returns a shaped list rather than an opaque object. The programmer
destructures it with ordinary Aiki pattern matching. This is consistent with
the `[@error, domain, message]` convention already established.

`path` is a separate pure Aiki module — no HAL crossing except for operations
that must consult the host (separator, cwd). Everything else is string
manipulation:

```aiki
let p = path.join(dir, name)      # pure string operation
let d = path.dir(p)               # pure: everything before last separator
let b = path.base(p)              # pure: everything after last separator
let e = path.ext(p)               # pure: extension including dot
let clean = path.normalize(p)     # pure: resolve . and ..
let abs = path.is_absolute(p)     # pure: starts with separator

# these two touch the host:
let sep = path.separator()        # -> "/" or "\" depending on host
let here = path.cwd()             # -> current working directory string
```

The principle: `path` is a string library with a host-aware separator.
Only `separator` and `cwd` cross the HAL. Everything else is ordinary Aiki.

#### Process execution — `system`

The existing `system` module has `args` and `env`. The major gap is running
external commands.

```aiki
let result = system.exec(cmd, args)
# -> [@ok, stdout_string, stderr_string, exit_code]
# | [@error, :process, message]

let cwd = system.cwd()            # -> current directory string
system.chdir(path)                # -> true | [@error, :io, message]
system.exit(code)                 # -> does not return

# environment (already exists):
let val = system.env(name)        # -> string | [@error, :environment, message]
let a = system.args()             # -> [string, ...]
```

`system.exec` is the workhorse. It runs a command synchronously, captures
stdout and stderr as strings, and returns the exit code. This covers the
overwhelming majority of process-execution needs: running tests, invoking
compilers, shell commands, text processing pipelines.

For programs that need non-blocking process control:

```aiki
let proc = system.start(cmd, args)  # -> process handle | [@error, ...]
let result = system.wait(proc)      # -> [@ok, exit_code] | [@error, ...]
```

`start`/`wait` is a second tier — useful but less immediately needed than
`exec`. The process handle is an opaque value. Aiki does not need to expose
PID, signals, or other OS process internals unless a concrete program
requires them.

#### Time — `time`

The existing module has `sleep` and `after`. The gaps are knowing what time
it is and basic time arithmetic.

```aiki
let now = time.now()              # -> milliseconds since epoch (integer)
let fmt = time.format(ms, pat)    # -> formatted string
let ms = time.parse(s, pat)       # -> milliseconds | [@error, :time, message]

# already exists:
time.sleep(ms)                    # -> blocks for ms milliseconds
let ch = time.after(ms)           # -> channel that receives true after ms
```

`time.now` crosses the HAL. `time.format` and `time.parse` could be pure
Aiki or HAL-backed depending on how much formatting complexity is justified.
Simple ISO-8601 formatting can be Aiki; locale-aware formatting would need
the host.

Duration arithmetic is just integer arithmetic on milliseconds — no new
abstraction needed. Aiki's exact rationals mean there is no precision loss.

#### Environment — folded into `system`

Environment access (`system.env`, `system.args`) already exists. Two useful
additions:

```aiki
let pairs = system.env_list()     # -> [[@pair, name, value], ...]
let ok = system.env_set(name, v)  # -> true | [@error, :environment, message]
```

These are low priority. Most programs only read environment variables.

### What this looks like in practice

A realistic Aiki program using these facilities:

```aiki
let file = import("file")
let path = import("path")
let system = import("system")
let string = import("string")

# Find all .ai files in the current directory
let here = system.cwd()
let entries = file.list(here)
let ai_files = []
let i = 0
while i < length(entries) {
    if string.ends_with(entries[i], ".ai") {
        ai_files = append(ai_files, entries[i])
    }
    i = i + 1
}

# Run the test suite
let result = system.exec("aiki", ["test", "./..."])
match shape(result) {
    "ok" {
        print("tests passed, exit code: " + to_str(result[2]) + "\n")
    }
    "error" {
        print("test execution failed: " + result[2] + "\n")
    }
}
```

This is mundane, useful, and unsurprising. The programmer does not encounter
the HAL, capability domains, or substrate provenance. They import a module,
call functions, and handle shaped results.

### What does not belong here

The following are explicitly excluded from the initial library surface unless
concrete programs demonstrate the need:

- network sockets / TCP / HTTP;
- signal handling beyond `system.exit`;
- user/group/permission inspection;
- hostname / OS identification;
- inter-process communication primitives;
- shared memory;
- raw ioctl or device access.

These can be added later through the same architecture. The Thompson criterion
applies: add a capability because a real program requires it, not because
another language has it.

## Conceptual responsibilities

The classification phase should distinguish at least the following kinds of
things currently adjacent to the HAL.

### Language intrinsics

Operations whose behavior participates directly in Aiki evaluation and needs
language machinery rather than merely a host effect. Current candidates include
`import`, `export`, `use`, `apply`, `load`, and `spawn`.

The phase must determine which are genuinely evaluator/language concerns and
which contain separable host-backed portions.

### Host capabilities

Irreducible crossings into the host: filesystem access, terminal/process I/O,
time, environment/process arguments, randomness where host-backed, host
representation/bytes, optional graphics or other external resources, and host
observation support.

The HAL should expose the smallest useful mechanism. Rich domain semantics
should remain in Aiki libraries when practical.

### Execution context

The baseline has demonstrated that "environment" contains several different
kinds of context that should not automatically travel together:

```text
lexical bindings
shape vocabulary
source/module provenance
dynamic call state
semantic observation
capability/authority context
```

The redesign should classify these by meaning before deciding whether they
become separate Go interfaces or remain implementation details.

### Runtime services

Longer-lived engine concerns that are neither language values nor ordinary
one-shot host effects, including:

- module resolution/cache lifetime;
- spawned-fault coordination;
- source identity/lifecycle;
- host resource ownership and cleanup;
- profile-label correlation;
- runtime/session process state.

## Canvas as a pressuring requirement

Canvas is not the driver of the redesign. It is one of the strongest tests of
whether the proposed boundary is actually useful.

Canvas pressures the model because it is:

- stateful;
- lifecycle-bearing;
- optional;
- currently realized through IPC and a protocol;
- Aiki-to-host command oriented;
- potentially host-to-Aiki event producing;
- required to coexist naturally with Aiki concurrency and `select`.

The current leading hypothesis is not "HAL.canvas" as a large domain API and
not a universal `Transport` abstraction.

A promising Canvas realization is:

```text
host resource
    + host-backed Aiki communication
    + Aiki-defined canvas protocol
```

The Canvas library would retain domain meanings such as line, rectangle,
clear, color, mouse, resize, and related protocol operations. The generic
runtime/concurrency machinery should not know what Canvas means.

Hard pressure-test criteria:

- `select` must not need Canvas-specific semantics;
- generic channel/concurrency code should not care whether a peer is Aiki code
  or a host service;
- Canvas protocol meanings should live above the generic HAL boundary;
- optional Canvas support must not make Canvas mandatory for every substrate;
- lifecycle/cleanup must have an explicit owner;
- no universal transport abstraction should be introduced until another
  concrete facility proves one is needed.

Host-backed channels are therefore a candidate answer to Canvas's asynchronous
pressure, not a premise of the redesign.

## Authority and trusted code

The current `ScopeUser` / `ScopePrelude` distinction is valuable and should not
be discarded casually.

However, authority should be conceptually independent of filesystem topology.
A library path may bootstrap trust, but the architectural statement should be:

> this module/evaluation context has this declared capability set

rather than:

> this path happens to receive privileged scope

The phase should determine whether `ScopePrelude` remains the right abstraction,
is refined, or becomes an implementation of a more explicit capability policy.

## Runtime/session ownership

Package-global host relationships such as stdin, stdout, user environment, and
resource cleanup should move toward explicit runtime/session ownership where
appropriate.

This matters for:

- tests;
- multiple evaluators in one process;
- notebook/LSP/server embeddings;
- concurrent sessions;
- alternate substrates;
- deterministic host dependency injection.

## Module loading as a boundary test

Module loading currently combines multiple responsibilities:

```text
resolve identity/location
obtain source
assign authority
parse/evaluate Aiki
cache module identity/state
expose exports
```

The redesign should classify these separately. A likely split is:

```text
source resolution/provider     -> host/embedding capability
module evaluation              -> Aiki engine
module authority               -> runtime policy/configuration
module cache/lifetime          -> runtime service
exports                        -> language semantics
```

This is a candidate model, not yet a decision.

## Observation

Preserve the distinction already established by semantic profiling:

```text
semantic observation   = what Aiki computation occurred
substrate observation  = how the host realized it
```

The HAL name provides a stable point at which the two can be correlated without
making Go-level observation the language's semantic account.

## Invariants to preserve

Any redesign must preserve at least these properties already earned by the
baseline:

1. ordinary user code cannot directly acquire raw privileged HAL operations;
2. Aiki-written prelude/library implementations remain first-class;
3. exact-number semantics do not depend on host floating point;
4. isolated spawn does not regain parent mutable lexical state through HAL
   context;
5. explicit `Store` and `Channel` capabilities remain explicit across spawn;
6. source provenance required for relative import survives isolated execution;
7. module evaluation does not inherit arbitrary importer bindings;
8. semantic observation remains Aiki-level even when correlated with substrate
   profiling;
9. host capabilities are not added merely for architectural symmetry;
10. executable couplings and validation should become stronger, not weaker.

# Phase/cut structure

Cut 0 records the framing and governing constraints. The redesign then proceeds
through **three phases of four cuts each**: twelve phase cuts, thirteen cuts
total including Cut 0.

## Phase I — Boundary and affordance map

### Cut I.1 — Inventory the current boundary

Record every registered native primitive against the three names, source
location, present Aiki use, context, authority, state/lifetime, and substrate
provenance.

### Cut I.2 — Classify what the registry actually contains

Separate true host crossings from evaluator intrinsics, value/runtime
primitives, optional accelerators, observation/tooling services, execution
context plumbing, and legacy/misplaced responsibilities.

### Cut I.3 — Map systems-programmer affordances

Compare the current public surface with the host-facing operations needed for
ordinary systems work. Distinguish user affordance from the minimum irreducible
HAL mechanism needed to realize it.

### Cut I.4 — Pressure-test and synthesize

Test the classification against module loading, spawn, profiling, file handles,
Canvas, and the Thompson-style rule of adding only capabilities forced by real
programs. Produce the Phase-I boundary model and unresolved decisions for
Phase II.

**Phase-I gate:** every current primitive is accounted for; the systems
programmer target surface is mapped to current coverage/gaps; no pressure case
requires a category that has not been named.

## Phase II — Context, authority, and ownership

### Cut II.1 — Decompose execution context

Separate lexical meaning, shape vocabulary, source/module provenance, dynamic
call state, observation, and authority. Determine what accompanies calls,
spawned computations, modules, and host crossings.

### Cut II.2 — Define runtime/session ownership

Place stdin/stdout, program environment, module cache, asynchronous faults,
resource cleanup, and other process/session state under explicit owners.

### Cut II.3 — Define capability and authority semantics

Settle trusted-library authority, optional capabilities, host resources, and
how current `ScopeUser` / `ScopePrelude` relate to the new model.

### Cut II.4 — Asynchronous/resource pressure gate

Apply the model to Canvas and other existing stateful/asynchronous facilities.
Determine whether host-backed channels suffice for Canvas without making them
a premise of the entire redesign. `select` must remain domain-ignorant.

**Phase-II gate:** context, authority, ownership, and asynchronous host
relationships can be explained without an undifferentiated `*Env`, ambient
package globals, or Canvas-specific generic machinery.

## Phase III — Replacement contract and migration design

### Cut III.1 — Define canonical HAL identities and metadata

Specify the `HAL.<domain>.<operation>` contract model, including authority,
context, error translation, semantic observation, blocking/cost class, and
substrate provenance.

### Cut III.2 — Draft runtime/substrate interfaces

Only now propose concrete Go interfaces, runtime/session objects, source/module
providers, capability registration/discovery, and host-resource hooks.

### Cut III.3 — Build the serial migration plan

Order migration so existing Aiki behavior remains runnable and evidence-gated.
Identify compatibility bridges and executable invariants for every slice.

### Cut III.4 — Review against user affordances and invariants

Trace the systems-programmer surface end-to-end through Aiki name, HAL contract,
and substrate realization. Verify that the architecture enables the intended
affordances rather than merely producing a cleaner diagram.

**Phase-III gate:** a reviewable replacement contract and serial migration plan
exist, preserve the earned language invariants, and cleanly realize the agreed
systems-programmer surface. Implementation begins only after this gate.

# Pressure requirements

Canvas remains a pressuring requirement, not the driver. Module loading, spawn,
profiling, files, explicit mutable Store, channels/select, and the Thompson 7094
work are equally useful architectural tests. A pressure case may force the
recognition of a missing category; it does not get to define the generic model
without broader evidence.

# Completion criterion

The redesign design-phase is complete when every relevant user affordance can be
traced through:

```text
What does the programmer mean?            Aiki name / library semantics
What does the architecture promise?       HAL name / contract
What performs it on this substrate?       substrate name / realization
Who may invoke it?                        authority/capability
What context does it require?             explicit context facets
Who owns state and lifetime?               runtime/session/resource owner
```

and the migration can proceed serially without weakening Aiki's existing
semantic and executable couplings.
