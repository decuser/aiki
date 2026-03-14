# How Errors Work in Aiki

Aiki has two failure strata: **Faults** (halting) and **Shaped Errors** (recoverable).

## Faults

A Fault is an internal evaluation failure. When one occurs, execution halts immediately. Faults are not representable as Aiki values — user code cannot construct, capture, or inspect them.

**Examples of faults:**
- Wrong arity: `add(1)` when `add` expects two arguments
- Type mismatch: `"hello" + 5`
- Undefined variable: `x` when `x` is not bound
- Division by zero: `1 / 0`
- Index out of bounds: `[1, 2, 3][10]`
- Stack overflow

Faults print a Ruby-style diagnostic and cause nonzero exit:

```
example.ai:5:in 'add': wrong arity: want 2, got 1
    add(1)
        from example.ai:10:in '<main>'
```

## Shaped Errors

A shaped error is a recoverable value: `[@error, :kind, "message"]`. It flows through evaluation like any other value but has special properties:

- **Falsy** — `if err { ... }` does not execute the block
- **Short-circuits pipes** — `x |> f() |> g()` stops at the first shaped error and returns it

**Examples of shaped errors:**
- `import("missing")` → `[@error, :import, "import: cannot find 'missing'"]`
- `to_number("abc")` → `[@error, :parse, "to_number: invalid number 'abc'"]`
- `read()` on IO failure → `[@error, :io, "read: connection closed"]`
- `hash_get(h, key)` when key missing → `[@error, :not_found, "key not found"]`

## Checking for Errors

Use `is_error(x)` to test if a value is a shaped error:

```
let result = import("maybe_missing")
if is_error(result) {
    print("Could not load module")
} else {
    result.some_function()
}
```

Or use `match` on the shape:

```
match shape(result) {
    :error { print("Failed: ", result[1]) }
    _ { use(result) }
}
```

## Pipe Short-Circuiting

Shaped errors propagate through pipes without calling subsequent functions:

```
let process = (x) { x * 2 }
let validate = (x) { 
    if x < 0 { return [@error, :validation, "must be positive"] }
    return x
}

validate(-5) |> process() |> print()
# Prints: [@error, :validation, must be positive]
# process() and print() never called with the error
```

## Creating Shaped Errors

In Aiki code, construct them directly:

```
let divide = (a, b) {
    if b == 0 { return [@error, :math, "division by zero"] }
    return a / b
}
```

The convention is `[@error, :kind, "message"]` where `:kind` is a symbol categorizing the error (`:io`, `:parse`, `:import`, `:validation`, etc.).

## Summary

| Aspect | Fault | Shaped Error |
|--------|-------|--------------|
| Recoverable | No | Yes |
| Halts execution | Yes | No |
| Is an Aiki value | No | Yes (List) |
| Truthiness | N/A | Falsy |
| Pipe behavior | Halts | Short-circuits and returns |
| Exit code | Nonzero | Zero |
| User can create | No | Yes |
