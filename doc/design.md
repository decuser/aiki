# Aiki Design

A minimal, composable language. The language is complete; the system layer is future work.

## Principles

- **One way to do each thing.** Constraint forces clarity.
- **Composition through naming.** Define small pieces, name them, compose.
- **Explicit over implicit.** No magic, no hidden behavior.
- **Inspectability enables knowing.** Everything answers: what are you, what's in you.

## Keywords (11)

`let` `if` `else` `while` `match` `return` `true` `false` `and` `or` `not`

## Types (8)

| Type | Example | Notes |
|------|---------|-------|
| Number | `42`, `3/4` | Rational (exact) |
| Boolean | `true`, `false` | |
| Rune | `'a'` | Unicode code point |
| String | `"hello"` | Immutable |
| Bytes | `to_bytes("hi")` | Immutable 0-255 sequence |
| Symbol | `:ok` | Atomic, identity-compared |
| List | `[1, 2, 3]` | Raw or shaped |
| Function | `(n) { return n }` | First-class |

## Operators

**Keep:** `+` `-` `*` `/` `<` `>` `<=` `>=` `|>` `.` `[]`

**Cut:** `!` -> `not`, `==` -> `equal()`, `!=` -> `not(equal())`, `%` -> `modulo()`

## Layers

| Layer | What | Auto-load |
|-------|------|-----------|
| `hal` | Go primitives — memory, I/O, concurrency, OS | Always |
| `strict` | Pure Aiki — lists, rationals, inspectable | Yes |
| `pragmatic` | Fast Aiki — arrays, floats, hardware speed | Opt-in |

### Runtime Strata

Three layers with distinct responsibilities:

**HAL (Host Abstraction Layer)** - Narrow membrane to Go. Effects live here: canvas, channels, spawn, file I/O. Minimal surface. Everything that touches the outside world.

**Strict** - Pure semantic core. Deterministic evaluation. The canonical specification. All behavior defined here first.

**Pragmatic** - Performance-oriented shadow of Strict. Must be observationally equivalent. Never defines behavior, only optimizes it.

Principle: impurity pushed to edges, purity at center.

**The Twin Pattern:** `strict.X` is pure and inspectable. `pragmatic.X` is fast with same API. No shadowing — namespace separation.

## Decisions

**Rational numbers.** No floating point surprises. `1/3 * 3 = 1`.

**Shaped lists.** `[@point, 10, 20]` enforces structure. Resources are shaped lists with hal registries.

**Left-to-right evaluation.** No precedence. `1 + 2 * 3` is `9`. Use parens.

**`let` creates, `=` mutates.** Catches typos.

**Explicit return.** Every path needs `return`.

**One function syntax.** `(params) { body }`. Rest params via `...args`.

**`while` only.** Iteration via `map`, `filter`, `each`.

**Pipe operator.** `x |> f() |> g()`. Errors short-circuit.

**Bracket indexing.** `list[i]`, `string[i]`. Dot is fields only.

**Errors are values.** No exceptions. `[@error, reason]` on failure.

**Semicolons optional.** Statement separator for one-liners.

## Rejected

| Feature | Reason |
|---------|--------|
| Arrow lambdas | Second way to write functions |
| Implicit return | Last-line bugs |
| `for` loop | `while` + functions cover it |
| `elif` | Guard clauses with early return |
| Operator precedence | Parens are explicit |
| Classes/methods | Composition, not inheritance |
| Exceptions | Errors are values |
| Macros | Language is closed |
| Floats as default | Rationals are exact |
| Block comments | Prevents dead code hiding |
