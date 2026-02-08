# Aiki Design Rationale

Every decision explained. What was chosen, what was rejected, why.

## Core Principle: One Way to Do Each Thing

Every design question resolves by asking: "Does this add a second way to do something?" If yes, reject it. Constraint forces clarity.

## Core Identity: Composition Through Naming

Everything is about defining small pieces, naming them, and composing them. This drives every decision.

## Eight Types

Built up from necessity:

- **Numbers** - rational (exact arithmetic, integer is n/1)
- **Booleans** - irreducible
- **Runes** - Unicode code points, not bytes. Strings index to runes. Abstraction never leaks.
- **Strings** - immutable rune sequences (immutable list of runes)
- **Bytes** - immutable list of byte values 0-255, for I/O and binary data
- **Symbols** - atomic labels for flags and enums. `:active`, `:ok`. Compared by identity. Not for tagging lists.
- **Lists** - one compound structure. Raw or shaped.
- **Functions** - first-class values

No type was added without justification.

### Why Bytes

Runes are for text. Bytes are for data. File I/O, network I/O, binary formats—these are bytes. The conversion is explicit:

```
let r = 'é'
let b = tobytes(r)    # [195 169] - UTF-8 encoding
let r2 = torune(b)    # 'é'
```

Bytes support `len`, `first`, `rest`, indexing. But `append(bytes 256)` errors—values must be 0-255.

### Why Rational Numbers

No floating point surprises. `0.1 + 0.2 == 0.3` actually works.

- Integers are rationals with denominator 1
- All arithmetic is exact
- `1 / 3` stays `1/3`, simplifies when possible
- `(1 / 3) * 3` equals `1`
- `todecimal(n places)` for display: `todecimal(1/3 5)` → `"0.33333"`
- `1 / 0` returns `[@error "division by zero"]`
- No floats in the language; FFI boundary converts if needed

## Two Kinds of Lists

**Raw lists** hold arbitrary values:
```
[1 2 3]
[10 20]
```

**Shaped lists** have declared structure:
```
let @point [x y]
let p = [@point 10 20]
p.x    # 10
```

Why two kinds:
- Raw lists are flexible, anything goes
- Shaped lists are enforced, named access, catch errors at construction
- Both are still lists underneath - `first()`, `rest()`, `len()` work on both

Why `@`:
- Visual distinction at declaration and construction
- No implicit lookup - you see `@`, you know it's shaped
- `:symbol` stays atomic, doesn't tag lists

### Shape Composition

Shapes can embed other shapes:
```
let @pet [name age]
let @cat [@pet color]       # embeds pet fields
let @siamese [@cat points]  # embeds cat fields

let s = [@siamese "Mochi" 3 "cream" "seal"]
s.name    # "Mochi" - inherited from pet
s.color   # "cream" - inherited from cat
s.points  # "seal"  - siamese's own
```

This is structural composition, not inheritance. The shape flattens. No method dispatch, no polymorphism. Just data with named access.

### Shape Inspection

```
shape([@siamese "Mochi" 3 "cream" "seal"])  # :siamese
shape([1 2 3])                               # :list (raw)
```

Behavioral dispatch is your problem:
```
let speak = (pet) {
    match shape(pet) {
        :dog { return "woof" }
        :cat { return "meow" }
        _    { return "..." }
    }
}
```

No categories, no facts, no inheritance magic. Just match on shape. If you want OO, build it as a library.

## Symbols Are Atomic Only

Symbols are lightweight named values:
```
let status = :active
if status == :ready { }
```

They are not list tags. That's what `@` shapes are for. This separation is explicit - you always know what you're looking at.

## Dot Access

One syntax for all member access:
```
list.0        # positional
point.x       # named field
math.sqrt     # module member
```

Replaces `nth(list 0)`. Unified, readable.

## Binding vs Mutation

```
let x = 7     # creates new binding
x = 9         # mutates existing, error if not found
```

Why require `let`:
- Catches typos (`countr = 1` errors instead of creating)
- Explicit about intent
- Closures work correctly (mutation walks outward)

## Explicit Return Always

Every path through a function must have `return`:
```
let validate = (x) {
    if x < 0 {
        return [@error "negative"]
    }
    return [@ok x]
}
```

Why not implicit return:
- Last-line bugs (add logging, accidentally return wrong value)
- Explicit is readable
- Multiple returns enable clean guard clauses

## Guard Clauses Replace elif

No `elif`. No `cond`. Use guard clauses:
```
let describe = (x) {
    if x < 0 { return "negative" }
    if x == 0 { return "zero" }
    return "positive"
}
```

Why no `elif`:
- Nested `if` is ugly but unambiguous
- Guard clauses flatten naturally with explicit return
- Deep nesting is a signal to refactor
- One way to branch

## One Function Syntax

```
let square = (n) { return n * n }
map(nums, (n) { return n * n })
```

Named and anonymous use same syntax. No arrow lambdas.

Why no arrows:
- Adds second way to write functions
- Saves two characters, adds a concept
- Not worth it

## Recursion

### Tail Call Optimization

Required. Without it, functional patterns blow the stack:
```
let sum = (list acc) {
    if len(list) == 0 { return acc }
    return sum(rest(list) (acc + first(list)))  # tail position
}
```

This must not grow the stack. TCO makes it a loop internally.

### Stack Limits

Non-tail recursion still grows the stack. Protection:
- Configurable stack depth limit (default 10,000)
- Overflow returns `[@error "stack overflow"]`
- `stack_limit(n)` to configure
- Debugger shows call chain on overflow

## One Loop

`while` is the only loop construct:
```
while x > 0 {
    x = x - 1
}
```

Iteration uses higher-order functions:
```
each(nums, (n) { print(n) })
nums |> map((n) { return n * n })
```

Why no `for`:
- `while` covers all cases
- Iteration is a function, not syntax
- One way to loop

## Empty Parens Required

```
validate()           # call
let f = validate     # reference
u |> validate()      # call in pipe
```

Consistency everywhere. Parens mean call.

## Pipe Operator

```
u |> validate() |> transform() |> save()
```

Left-to-right composition. Left side becomes first argument of right side.

Error short-circuit: if any step returns `[@error reason]`, rest of pipe skips.

## Errors Are Values

No exceptions. Functions return `[@ok value]` or `[@error reason]`. Pipe automates checking. `match` destructures.

## Files as Modules

No packages. Files are modules. Filesystem is organization.

```
let math = load("math")
math.sqrt(16)

from math use [sqrt]
sqrt(16)
```

## Primitives vs Prelude

The standard library has two layers:

**Primitives** - Implemented in the runtime (Go). Cannot be written in Aiki. These touch memory, I/O, or runtime internals:
- List: `first`, `rest`, `len`, `prepend`, `append`
- Type: `type`, `inspect`, `shape`
- Compare: `equal`
- Convert: `tostr`, `tonum`, `tobytes`, `torune`, `todecimal`
- I/O: `print`, `read`, `open`, `create`, `close`
- Math: `sqrt`, `cos`, `sin`, `random`
- Bit: `bit_and`, `bit_or`, `bit_xor`, `bit_not`, `bit_shift`
- Regex: `regex`, `regex_all`, `regex_replace`
- Concurrency: `spawn`, `channel`, `send`, `recv`
- Canvas: `canvas`, `draw_line`, `draw_rect`, `draw_circle`, `draw_text`, `clear`, `present`
- System: `help`, `quit`, `universe`, `symbols`, `history`, `peek`

**Prelude** - Written in Aiki. Loaded by default. User can read, understand, replace:
- `map`, `filter`, `reduce`, `each`
- `range`, `reverse`
- `find`, `any`, `all`
- `as_lines`, `as_runes` (stream wrappers)
- Turtle graphics (built on canvas primitives)

Why this split:
- Primitives are what you literally can't compute without
- Prelude proves the language works - it's written in itself
- User can read the source, understand how `map` works
- User can replace any prelude function
- Forth spirit: the language builds itself

Arithmetic uses operators (`+`, `-`, `*`, `/`, `%`), not functions. One way to add.

## No Language Extension

No macros. No metaprogramming. Functions are the only extension mechanism. The language is closed. The library is open.

## Concurrency: Go-Lite

Minimal threading with channels. Marked experimental.

Four primitives:
```
spawn(f)      # green thread, returns immediately
channel()     # unbuffered channel
send(ch val)  # blocks until received
recv(ch)      # blocks until sent
```

No new syntax. Just builtins.

### Known Limitations

- No protection against data races
- Shared mutable state is dangerous
- No preemption - long-running computations block
- Channel operations are the only yield points

Best practice: Don't share mutable state. Communicate through channels.

### Why Go-Lite

Go's model (goroutines + channels) is well-understood and maps cleanly to composition philosophy. Message passing through channels is encouraged. Shared mutable state is possible but discouraged.

This model will evolve based on real usage patterns.

### Process-Level Parallelism

Deferred. V1 is single-process with green threads and channels. Process spawning, IPC, sockets come later.

## I/O and Streams

### Core Primitives

```
open(path)       # returns byte stream (function)
create(path)     # returns byte sink (function)
close(stream)
```

### Streams Are Functions

A stream is a function that returns `[@ok value]` or `[@end]` on each call:
```
let f = open("file.txt")
match f() {
    [@ok chunk] { process(chunk) }
    [@end]      { done() }
}
close(f)
```

No new type. Closures handle state internally.

### Stream Wrappers (Prelude)

```
let lines = as_lines(open("file.txt"))
lines()    # [@ok "first line"] or [@end]

let runes = as_runes(open("file.txt"))
runes()    # [@ok 'h'] or [@end]
```

Composition at multiple levels. User can build `as_csv`, `as_json`, etc.

## Regex

Three primitives, RE2 semantics (linear time, no backtracking):
```
regex(string pattern)           # [@ok [full groups...]] or [@error "no match"]
regex_all(string pattern)       # [@ok [matches...]] or [@error "no match"]
regex_replace(string pattern replacement)  # new string
```

RE2 limitations vs Perl: no backreferences, no lookahead/lookbehind. 95% of patterns work. The 5% that don't were usually dangerous anyway.

## Bit Operations

Six functions, integers only:
```
bit_and(a b)       # bitwise AND
bit_or(a b)        # bitwise OR
bit_xor(a b)       # bitwise XOR
bit_not(a)         # bitwise NOT (two's complement)
bit_shift(a n)     # left if n > 0, right if n < 0
```

Non-integer input returns `[@error "requires integer"]`.

Functions, not operators. Rare enough that call overhead doesn't matter.

## Canvas

Primitives for graphics (implemented via Ebiten):
```
canvas(width height)              # create window, return handle
draw_line(c x1 y1 x2 y2 color)
draw_rect(c x y w h color)
draw_circle(c x y r color)
draw_text(c x y text color)
clear(c color)
present(c)                        # flush to screen
close(c)
```

Colors: `:red`, `:blue`, etc. or `[@rgb 255 128 0]`

Turtle graphics is prelude, not primitive. Logo in 50 lines.

## Why a Live System

The language spec above is complete. You can write Aiki, run files, ship executables. Nothing requires the system layer.

But Aiki is designed to support a live environment - not because liveness is convenient, but because of how knowing works.

### The Epistemological Grounding

Information is realized only when a knowing subject apprehends a potential input and undergoes configurational change (K → K′). For apprehension to occur, the subject needs proximity - the ability to encounter, examine, and align with structure.

A programming environment is an information system. Its job is to create conditions for apprehension. Opacity prevents proximity. If you can't inspect, you can't align. If you can't align, no K → K′.

### What This Requires

**Inspectability.** Everything answers the same questions:
- What are you? (type, shape)
- What's in you? (contents, fields)
- Where did you come from? (provenance, version)

**Liveness.** The system reshapes through use. You're inside it, not outside sending commands. Development and execution are the same activity.

**Transparency.** Unlike Smalltalk's opaque image, the system must be approachable from outside. Files are shadows - human-readable projections of current state. You can always escape, inspect with external tools, rebuild from text.

**Versioning.** History is structural. Every mutation increments a version. Old versions exist until garbage collected. You can ask "where did this come from?" and get an answer. (Full time machine deferred; hooks preserved.)

### The Layering

```
┌─────────────────────────────┐
│   Aiki code you write       │  <- language layer (this spec)
├─────────────────────────────┤
│   Live environment          │  <- system layer
│   (REPL, inspection,        │
│    versioning, shadows)     │
├─────────────────────────────┤
│   Runtime (Go)              │  <- implementation
└─────────────────────────────┘
```

The language runs on the system. The system enables the learning field. But you can ignore the system layer entirely - files in, program out.

### Files as Shadows

The live system is the truth. Files are projections:
- Save casts the current configuration to disk
- Load restores a configuration from disk
- Files are diffable, versionable, readable without the runtime
- You're never trapped inside

This gives Smalltalk's liveness with Forth's transparency.

### What This Enables

- REPL as primary mode (you're inside, not outside)
- Hot reload (reshape without restart)
- Time travel (roll back to previous versions - deferred)
- Inspection (query any value's type, contents, history)
- Build (bag current versions into static executable)

### What This Doesn't Change

The language spec. Eight types, shaped lists, `let`/`=`, explicit return, one function syntax, `while`, pipes, modules. All unchanged. The system layer is addition, not modification.

## REPL Design

The REPL is just Aiki. No magic commands, no DSL.

**Not this:**
```
:help
:load file.ai
:inspect x
```

**Just Aiki:**
```
help()
load("file.ai")
inspect(x)
```

Functions, not commands. Same language everywhere. Extension lives in files you load:
```
# my-repl.ai
let ls = () { return symbols() }
let r = () { return load("main.ai") }
```

Then: `load("my-repl.ai")` and use your shortcuts.

## Universe Model

Three zones visible in the environment:

1. **Loaded** — current project, live, executable
2. **Scope** — current bindings (REPL state)
3. **Filesystem** — everything else, browsable but inert

Operations:
```
peek(path)              # inspect without loading
load(path)              # bring into project
from mod use [syms]     # selective import
```

### Projects

A project is just a list of files to load:
```
# project.ai
load("utils.ai")
load("math.ai")
load("main.ai")
```

Or:
```
aiki --project project.aiki
```

No namespacing beyond filenames. No package manager. Files in directories.

## TUI Vision

Console first. TUI later.

```
┌─────────────────────────────────────────────────┐
│ [search: ___________]                           │
├──────────────┬──────────────────────────────────┤
│ Navigator    │ Inspector                        │
│              │                                  │
│ ▼ my-project │ (shows selected item)            │
│   ▼ utils.ai │                                  │
│     helper   │                                  │
│   ▼ math.ai  │                                  │
│     sqrt     │                                  │
│ ▼ filesystem │                                  │
│   ▼ ~/lib/   │                                  │
│     strings  │                                  │
│ ▼ <scope>    │                                  │
│   x          │                                  │
├──────────────┴──────────────────────────────────┤
│ > _                                             │
│ console                                         │
└─────────────────────────────────────────────────┘
```

- Search filters the tree
- Click item → inspector shows it
- Console is always top-level scope
- Navigator shows loaded, filesystem, and scope
- No modes, no context switching

## System Invariants

The environment must never violate these rules.

1. **Shapes are immutable contracts.** Once declared, a shape cannot be redefined. Existing instances keep their original shape. Redeclaration is an error, not an update.

2. **Shape enforcement is total.** `[@point 10]` errors at construction. No partial construction, no bypass, no casting raw lists to shapes.

3. **Formatting is automatic and unavoidable.** The environment formats on save. This is not optional. Canonical formatting is part of the system, not a suggestion.

4. **Prelude stays shallow.** If a prelude function doesn't fit on a screen, break it into named steps. Scale reintroduces opacity.

5. **Pipe privileges `[@ok]`/`[@error]`.** This is explicit semantic bias. Other error conventions exist outside pipe ergonomics. The pipe does one thing.

6. **File shadows are perfectly reversible.** Save then load produces identical configuration. Always. If this breaks, trust is gone.

7. **Provenance is trivial.** Any value answers "where did you come from?" without ceremony. If this query is expensive or awkward, the epistemic claim collapses.

8. **Debugger shows user code by default.** Stepping into higher-order functions shows user code, not prelude plumbing, unless explicitly requested.

9. **No custom toString.** `type()`, `inspect()`, `shape()` are deterministic, not customizable. Want custom display? Write a function.

10. **Sandbox via capability.** Security is capability-based. A sandboxed environment simply doesn't provide I/O functions. You can't call what doesn't exist in your scope.

## Rejected Features

| Feature | Reason |
|---------|--------|
| Arrow lambdas | Second way to write functions |
| Implicit return | Last-line bugs, unclear |
| `for` loop | `while` + functions cover it |
| `elif` / `cond` | Guard clauses solve it |
| Bare `=` for creation | Typos become bugs |
| `:symbol` for list tags | `@` shapes are explicit |
| Multiple compound types | Lists handle everything |
| Operator precedence | Parens required, no ambiguity |
| Classes/methods | Data doesn't have behavior |
| Exceptions | Errors are values |
| Macros | Language is closed |
| Math functions for arithmetic | Operators handle it |
| Pointers | Breaks inspectability |
| IEEE floats | Rationals are exact |
| REPL meta-commands | Just use functions |
| Context/namespace switching | Qualified names, no modes |
| Full OO system | Build it as a library if needed |
| Async/await | Go-lite concurrency instead |
| DSL/sublanguages | REPL is just Aiki |

## Deferred Features

| Feature | Status |
|---------|--------|
| Process spawning / IPC | After v1 |
| Sockets / networking | After v1 |
| Full time machine | Hooks preserved, implementation later |
| Plan 9 style /dev | Nice idea, needs use case |
