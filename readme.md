# Aiki

A minimal, composable programming language.

## Principles

- One way to do each thing.
- Explicit over implicit. No operator precedence. No magic.
- Composition through functions and shaped lists. No classes.
- Inspectability enables knowing. The stdlib is written in Aiki.

Aiki is a learning field. Constraints force clarity.

## Use

```
aiki                        # REPL
aiki file.ai                # run strict (default)
aiki --proto file.ai        # run loose (TDD mode)
aiki validate file.ai       # check without running
aiki fmt file.ai            # format
aiki lint file.ai           # lint
aiki help [topic]           # help
aiki version                # version
aiki clean                  # remove generated files
aiki --errors=strict file   # show errors through strict layer
aiki --errors=hal file      # show full stack to primitives
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
range(1, 10) |> map((x) { x * x }) |> filter((x) { x > 10 }) |> println()

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
| hal | Go primitives — I/O, concurrency, canvas |
| strict | Pure Aiki — lists, hash maps, algorithms |
| pragmatic | Fast Aiki — arrays, floats (opt-in) |

Strict is default. Errors show your code by default, `--errors=hal` for full depth.

## Docs

- [Design](design.md) — decisions and rationale
- [Manifesto](manifesto.md) — philosophy
- [Style](style.md) — naming and formatting rules
