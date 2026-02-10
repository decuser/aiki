# Aiki Design

## Current State (v0.2.3)

Working implementation: lexer, parser, tree-walking evaluator, REPL, file runner, formatter, canvas graphics, concurrency (spawn/channels).

## Vision

A minimal, composable language with a live environment. The language is complete; the system layer (inspection, versioning, TUI) is future work.

## Core Principles

- **One way to do each thing.** Constraint forces clarity.
- **Composition through naming.** Define small pieces, name them, compose.
- **Explicit over implicit.** No magic, no hidden behavior.
- **Inspectability enables knowing.** Everything answers: what are you, what's in you.

## Types

Eight, built from necessity:

| Type | Example | Notes |
|------|---------|-------|
| number | `42`, `3/4` | Rational (exact) |
| boolean | `true`, `false` | |
| rune | `'a'`, `'€'` | Unicode code point |
| string | `"hello"` | Immutable rune sequence |
| bytes | `tobytes("hi")` | Immutable 0-255 sequence |
| symbol | `:ok`, `:error` | Atomic, identity-compared |
| list | `[1, 2, 3]` | Raw or shaped |
| function | `(n) { return n }` | First-class |

## Decisions

**Rational numbers.** No floating point surprises. `0.1 + 0.2 == 0.3` works. `1/3 * 3 == 1`.

**Shaped lists.** `[@point, 10, 20]` enforces structure, enables named access. Raw lists `[1, 2, 3]` for flexibility. Both support `first`, `rest`, `len`.

**Shape composition.** `let @cat [@pet, color]` embeds pet's fields. Structural, not inheritance.

**Symbols are atomic only.** `:ok` is a value, not a list tag. `@` shapes tag lists.

**Comma separators.** `f(a, b)` and `[1, 2, 3]`. Unambiguous, consistent.

**Left-to-right evaluation.** No operator precedence. `1 + 2 * 3` is `9`. Use parens: `1 + (2 * 3)` is `7`.

**`let` creates, `=` mutates.** Catches typos. `countr = 1` errors if `countr` doesn't exist.

**Explicit return.** Every path needs `return`. Enables guard clauses, prevents last-line bugs.

**No `elif`.** Guard clauses with early return. One way to branch.

**One function syntax.** `(params) { body }`. No arrows, no variations.

**`while` only.** Iteration via `map`, `filter`, `each`. One way to loop.

**Empty parens required.** `f()` calls, `f` references. Consistent.

**Pipe operator.** `x |> f() |> g()`. Left becomes first arg. `[@error, ...]` short-circuits.

**Success returns value.** No `[@ok, ...]` wrapper on success. `[@error, reason]` on failure.

**Files as modules.** `from math use [sqrt]`. Filesystem is organization.

**Dot access.** `list.0`, `point.x`, `math.sqrt`. One syntax for all member access.

**No exceptions.** Errors are values. Match or pipe handles them.

**No macros.** Language is closed. Functions are the only extension.

**Primitives vs prelude.** Primitives (Go) touch runtime. Prelude (Aiki) proves the language works.

## Rejected

| Feature | Reason |
|---------|--------|
| Arrow lambdas | Second way to write functions |
| Implicit return | Last-line bugs |
| `for` loop | `while` + functions cover it |
| `elif` | Guard clauses |
| Operator precedence | Parens are explicit |
| Classes/methods | Data doesn't have behavior |
| Exceptions | Errors are values |
| Macros | Language is closed |
| Floats | Rationals are exact |

## Deferred

- Bytecode VM
- Debugger
- Process spawning, sockets
- Time machine versioning
