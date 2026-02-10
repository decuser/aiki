# Aiki

A minimal, composable programming language. v0.2.3

## Thesis

**One way to do each thing.** Constraint forces clarity.

**Composition, not inheritance.** Shapes embed (`let @cat [@pet, color]`). Functions compose (`|>`). Data and behavior stay separate.

## What Makes It Distinctive

1. **Constraint philosophy** — One loop construct (`while`). One function syntax. No `elif`. No operator precedence. The language is deliberately small and closed. No macros, no metaprogramming. Functions are the only extension mechanism.

2. **Rational-only arithmetic** — No floats. `1/3` is a value, not a division operation. `1/3 * 3 == 1`. `0.1 + 0.2 == 0.3`. Exact computation.

3. **No operator precedence** — Left-to-right evaluation. `1 + 2 * 3` is `9`. Parentheses are mandatory for grouping. You never remember precedence rules.

4. **Shaped lists as universal structure** — No classes, no structs, no records. Just lists, optionally tagged. `[@point, 10, 20]` gets named access (`p.x`, `p.y`). Hash maps are implemented in the language itself.

5. **Error short-circuit in pipes** — `[@error, reason]` propagates automatically through `|>`. Not just errors-as-values; the pipe operator knows about errors.

6. **Closed language** — The prelude (standard library) is written in pure Aiki, proving the language is complete without needing extension mechanisms.

## Lineage

```
Scheme ─────────── rationals, lists as data, first-class functions
    │
    ├── ML/F# ──── pipe operator, pattern matching
    │
    ├── Erlang ─── :symbol syntax, error tuples, tagged data
    │
    ├── Go ─────── braces, explicit return, spawn/channels, pragmatic minimalism
    │
    ├── Smalltalk ─ left-to-right evaluation, no precedence
    │
    └───────────── AIKI
```

Scheme's data philosophy + Go's pragmatism + Elixir's error conventions + Smalltalk's evaluation model + a constraint philosophy that's its own.

## Types

Eight, built from necessity:

| Type | Example | Notes |
|------|---------|-------|
| number | `42`, `3/4`, `0.5` | Rational (exact) |
| boolean | `true`, `false` | |
| rune | `'a'`, `'€'` | Unicode code point |
| string | `"hello"` | Immutable rune sequence |
| bytes | `tobytes("hi")` | Immutable byte sequence |
| symbol | `:ok`, `:error` | Atomic, identity-compared |
| list | `[1, 2, 3]` | Raw or shaped |
| function | `(n) { return n * 2 }` | First-class |

## Syntax Sample

```aiki
let @point [x, y]

let distance = (p1, p2) {
    let dx = p2.x - p1.x
    let dy = p2.y - p1.y
    return sqrt((dx * dx) + (dy * dy))
}

let origin = [@point, 0, 0]
let target = [@point, 3, 4]
print(distance(origin, target))
```

## Implementation

Tree-walking interpreter in Go. ~8000 lines.

- State machine lexer (UTF-8)
- Recursive descent parser (left-to-right, no precedence)
- Tree-walking evaluator
- Prelude in pure Aiki (map, filter, reduce, hash maps)
- Formatter with comment preservation
- Canvas graphics (Ebiten)
- Concurrency (spawn, channels)

## Status

Working: lexer, parser, evaluator, REPL, file runner, formatter, canvas, concurrency.

Next: debugger, regex, bit primitives.

Future: bytecode VM, LSP, WASM target.
