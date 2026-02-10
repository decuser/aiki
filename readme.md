# Aiki

A minimal, composable programming language.

## Principles

- One way to do each thing
- Explicit over implicit
- Composition through naming

## Example

```aiki
let @point [x, y]

let distance = (p1, p2) {
    let dx = p2.x - p1.x
    let dy = p2.y - p1.y
    return sqrt((dx * dx) + (dy * dy))
}

let a = [@point, 0, 0]
let b = [@point, 3, 4]
print(distance(a, b))
```

## Features

- Eight types: number (rational), boolean, rune, string, bytes, symbol, list, function
- Exact arithmetic: `1/3 * 3 == 1`
- Shaped lists: `[@point, 10, 20]` with named access
- Pipe operator: `x |> f() |> g()` with error short-circuit
- Pattern matching: `match result { [@ok, val] { ... } }`

## Usage

```
aiki                    # REPL
aiki run file.ai        # Run file
aiki fmt file.ai        # Format file
aiki fmt ./...          # Format directory recursively
```

## Documentation

- [design.md](doc/design.md) - Design and rationale
- [grammar.md](doc/grammar.md) - Language grammar
- [stdlib.md](doc/stdlib.md) - Standard library
- [primer.md](doc/primer.md) - Collaboration guide
