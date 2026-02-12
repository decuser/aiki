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
- Exact arithmetic: `1/3 * 3` is exactly `1`
- No operator precedence: `1 + 2 * 3` is `9` (left-to-right)
- Shaped lists: `[@point, 10, 20]` with named access
- Pipe operator: `x |> f() |> g()` with error short-circuit
- Pattern matching: `match result { [@ok, val] { ... } }`
- Canvas graphics (Ebiten)
- Concurrency (spawn, channels)

## Architecture

| Layer | What | Auto-load |
|-------|------|-----------|
| `hal` | Go primitives - memory, I/O, concurrency, canvas | Always |
| `strict` | Pure Aiki - lists, rationals, inspectable | Yes |
| `pragmatic` | Fast Aiki - arrays, floats, hardware speed | Opt-in |

## Usage

```
aiki                    # REPL
aiki file.ai            # Run file
aiki fmt file.ai        # Format file
aiki fmt ./...          # Format directory recursively
aiki lint file.ai       # Lint file
```

## Documentation

- [design.md](doc/design.md) - Design and rationale
- [grammar.md](doc/grammar.md) - Language grammar
- [stdlib.md](doc/stdlib.md) - Standard library
- [style.md](doc/style.md) - Style guide
- [todo.md](doc/todo.md) - Roadmap

## Status

Working: lexer, parser, evaluator, REPL, file runner, formatter, linter, canvas, concurrency.

## License

MIT
