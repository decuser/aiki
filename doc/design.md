# Aiki Design

A minimal, composable language. The language is complete; the system layer is future work.

## Principles

- **One way to do each thing.** Constraint forces clarity.
- **Composition through naming.** Define small pieces, name them, compose.
- **Explicit over implicit.** No magic, no hidden behavior.
- **Inspectability enables knowing.** Everything answers: what are you, what's in you.

## Keywords (9)

`let` `if` `else` `while` `match` `return` `true` `false` `not`

Note: `and`, `or` are operators.

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

**Keep:** `+` `-` `*` `/` `<` `>` `<=` `>=` `and` `or` `|>` `.` `[]`

**Cut:** `!` -> `not`, `==` -> `equal()`, `!=` -> `not(equal())`, `%` -> `modulo()`

## Modules

**`import("module", :name1, :name2)`** — load module, bind names in current scope.

**`export(:name1, :name2)`** — mark names for external use.

Functions, not keywords. Modules are values.

## Layers

| Layer | What | Visible |
|-------|------|---------|
| **HAL** | Go primitives — `_add`, `_first`, `_print` | No |
| **Prelude** | Aiki wrappers — `+`, `first`, `print` | Yes |

**HAL** is the narrow membrane to Go. Effects live here: canvas, channels, spawn, file I/O. Prefixed with `_`, invisible to user code.

**Prelude** wraps HAL, is the user-facing API. Written in Aiki. The dictionary you start with.

Principle: impurity pushed to edges, purity at center.

## Decisions

**Rational numbers.** No floating point surprises. `1/3 * 3 = 1`.

**Shaped lists.** `[@point, 10, 20]` enforces structure. Resources are shaped lists with HAL registries.

**Left-to-right evaluation.** No precedence. `1 + 2 * 3` is `9`. Use parens.

**`let` creates, `=` mutates.** Catches typos.

**Explicit return.** Every path needs `return`.

**One function syntax.** `(params) { body }`. Rest params via `...args`.

**`while` only.** Iteration via `map`, `filter`, `each`.

**Pipe operator.** `x |> f() |> g()`. Errors short-circuit.

**Bracket indexing.** `list[i]`, `string[i]`. Dot is fields only.

**Errors are values.** User errors: `[@error, reason]`. Runtime errors: stack trace with source.

**Semicolons optional.** Statement separator for one-liners.

## Projections

Grammar is infrastructure. These are derived, not written:

- **Help** — from grammar + layer metadata
- **Errors** — from templates registered per function
- **Lint** — from grammar + style rules

One source, multiple views.

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
| `from`/`export` keywords | Functions are uniform |
| `%`, `==`, `!=` operators | Functions: `modulo()`, `equal()` |
