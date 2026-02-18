# Aiki Design


## Architecture

Aiki is stratified by authority.

syntax is structure only. It defines the grammar, lexer, parser, and AST.
semantics is meaning only. It evaluates AST.
runtime is capability only. It hosts effects and resource registries.
resolver is location only. It maps module names to sources.
tools are projections. They project the language into views like fmt, lint, and help.

Hard boundaries.

syntax has no meaning.
semantics operates only on AST.
runtime has no structural or semantic authority.
resolver is not in the semantic pipeline.
tools contain no language rules.

Filesystem and network are storage, not semantics.

## Grammar

The canonical grammar is embedded and loaded through one public entry point.

The grammar file is embedded in the binary.
GetGrammar returns the canonical grammar.
The grammar is loaded once and cached.
Tools use the canonical grammar through GetGrammar.
No tool loads grammar from the filesystem.

Grammar defines structure.
Evaluator assigns meaning.
Runtime performs effects.
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

## Modules

In Aiki, the filesystem never determines meaning.

Aiki imports names.
Resolvers locate modules.
Filesystem and network are implementation details, not semantics.

Invariants.

import refers only to declared module names.
resolver maps names to code and nothing else.
filesystem is never meaning.
collisions must be explicit or fail.
names beginning with underscore are internal. all others are exported.

Only one keyword is required. module.

Module identity is semantic.

A module declares its identity inside the file.

module math

This name is the module.

File paths do not matter.
Directory layout does not matter.
Moving or renaming files does not change meaning.

Export is implicit by naming.

names beginning with underscore are internal.
all other names are exported.

Imports refer only to declared module names.

use("math")

Imports resolve by module name, not path.

no relative imports.
no directory based visibility.
no path semantics of any kind.
no inference.

Resolution is delegated to a pluggable resolver.

Resolver decides where a module is loaded from.
This resolver is outside language semantics.

Runtime loading is AST based.

Resolver returns source text or a cached AST.
Evaluator loads the AST into an environment.

local vs remote makes no semantic difference.
module identity is stable.
tooling works over AST, not paths.

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
