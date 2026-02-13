# Aiki

A minimal, composable programming language.

## Principles

- One way to do each thing.
- The design prioritizes inspectability over convenience, explicitness over magic. 
- Constraints force clarity.

Aiki is a learning field. The design prioritizes inspectability over convenience, explicitness over magic. Constraints force clarity.

No operator precedence. No implicit behavior. No classes. Composition through functions and shaped lists.

## Use
```
aiki              # REPL
aiki run file.ai  # run a file
aiki fmt file.ai  # format
aiki -e '1 + 1'   # evaluate expression
```

## Examples
```
-- recursive
let factorial = (n) { if n < 2 { return 1 } else { return n * factorial(n - 1) }

-- imperative
let sum = 0
each([1, 2, 3], (x) { sum = sum + x })
println(sum)  -- 6

-- functional pipeline
map([1, 2, 3], (x) { x * x }) |> filter((x) { x > 1 }) |> println()

-- canvas
let c = canvas(400, 400)
circle(c, 200, 200, 50, :red)
sleep(1000)
destroy(c)
```
