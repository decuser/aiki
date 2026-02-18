# Aiki

A minimal, composable programming language.

## Principles

- One way to do each thing.
- Explicit over implicit. No operator precedence. No magic.
- Composition through functions and shaped lists. No classes.
- Inspectability enables knowing. The prelude is written in Aiki.

Aiki is a learning field. Constraints force clarity.

## Use

```
aiki                        # REPL
aiki file.ai                # run file
aiki -e 'print(1 + 2)'      # eval expression
aiki fmt file.ai            # format
aiki lint file.ai           # lint
```

## Examples

```
# recursive
let factorial = (n) {
    if n < 2 {
        return 1
    }
    return n * factorial(n - 1)
}

# imperative
let sum = 0
each([1, 2, 3], (x) { sum = sum + x })
println(sum)  # 6

# functional pipeline
range(1, 10) |> map((x) { return x * x }) |> filter((x) { return x > 10 }) |> println()

# shaped data
let @point [x, y]
let origin = [@point, 0, 0]
println(origin.x)  # 0

# canvas
let c = canvas(400, 400)
circle(c, 200, 200, 50, :red)
sleep(1000)
destroy(c)
```

## Layers

| Layer | What |
|-------|------|
| HAL | Go primitives — I/O, concurrency, canvas |
| Prelude | Aiki wrappers — lists, hash maps, algorithms |

Prelude is user-facing. HAL is invisible.

## Docs

- [Design](design.md) — decisions and rationale
- [Manifesto](manifesto.md) — philosophy
- [Style](style.md) — naming and formatting rules

## Architecture

syntax is structure.
semantics is meaning.
runtime is capability.
resolver is location.
tools are projections.

Filesystem has no semantic authority.
