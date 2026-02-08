# Aiki

A minimal, composable programming language with a live environment.

## Philosophy

**Composition through naming.** Everything is about defining small pieces, naming them, and composing them into larger pieces.

- One way to do each thing
- Explicit over implicit
- No magic, no hidden behavior
- Inspectable, live, transparent

## Quick Example

```
let @point [x y]
let @user [name email]

let distance = (p1 p2) {
    let dx = p2.x - p1.x
    let dy = p2.y - p1.y
    return sqrt((dx * dx) + (dy * dy))
}

let u = [@user "jdoe" "j@d.com"]
let result = u |> validate() |> save()

match result {
    [@ok val]     { print("saved") }
    [@error msg]  { print(msg) }
}
```

## Key Features

- **Eight types**: number (rational), boolean, rune, string, bytes, symbol, list, function
- **Exact arithmetic**: `1/3 * 3 == 1`, no floating point surprises
- **Shaped lists**: `[@point 10 20]` with named access and enforcement
- **Shape composition**: `let @cat [@pet color]` embeds pet's fields
- **Explicit bindings**: `let` creates, `=` mutates
- **One function syntax**: `(params) { return expr }`
- **Guard clauses**: explicit `return` enables flat conditionals
- **Pipe operator**: `x |> f() |> g()` with error short-circuit
- **Files as modules**: filesystem is organization
- **Streams as functions**: `open("file")` returns a callable
- **Go-lite concurrency**: `spawn`, `channel`, `send`, `recv`
- **Canvas primitives**: turtle graphics, Logo in 50 lines
- **Live environment**: REPL, inspection, hot reload, build to executable

## Documentation

- [grammar.md](grammar.md) - Language syntax specification
- [design.md](design.md) - Design rationale and philosophy
- [stdlib.md](stdlib.md) - Standard library (primitives + prelude)
- [primer.md](primer.md) - Project collaboration guide

## Status

v0.2.0 design complete. Implementation not started.
