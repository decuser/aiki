# Aiki Standard Library (v0.2.0)

The library is split into two layers:
1. **Primitives:** Implemented in Go. Cannot be written in Aiki.
2. **Prelude:** Implemented in Aiki. Loaded automatically. User can override.

---

## Primitives

These are the atomic operations required to build everything else.

### List

| Function | Signature | Description |
|----------|-----------|-------------|
| `first`  | `(list)` | Returns the first element. Error if empty. |
| `rest`   | `(list)` | Returns list without first element. Empty if single. |
| `len`    | `(list)` | Returns number of elements. O(1). |
| `prepend`| `(list val)` | Adds `val` to start. Returns new list. O(1). |
| `append` | `(list val)` | Adds `val` to end. Returns new list. O(1) amortized. |

Works on raw lists, shaped lists, strings, and bytes.

### Type

| Function | Signature | Description |
|----------|-----------|-------------|
| `type`   | `(val)` | Returns `:number`, `:boolean`, `:string`, `:rune`, `:bytes`, `:symbol`, `:list`, `:function`. |
| `inspect`| `(val)` | Returns string representation for debugging. Optional depth parameter. |
| `shape`  | `(val)` | Returns shape symbol (`:point`, `:user`, etc.) or `:list` for raw lists. |

No custom toString. These are deterministic and not overridable.

### Compare

| Function | Signature | Description |
|----------|-----------|-------------|
| `equal`  | `(a b)` | Deep equality. Returns `true` or `false`. |

### Convert

| Function | Signature | Description |
|----------|-----------|-------------|
| `tostr`  | `(val)` | Converts value to string. |
| `tonum`  | `(val)` | Parses string to number. Returns `[@ok n]` or `[@error msg]`. |
| `tobytes`| `(val)` | Converts string or rune to UTF-8 bytes. |
| `torune` | `(bytes)` | Converts UTF-8 bytes to rune. Returns `[@ok r]` or `[@error msg]`. |
| `todecimal` | `(n places)` | Converts number to decimal string with given precision. |

### I/O

| Function | Signature | Description |
|----------|-----------|-------------|
| `print`  | `(val)` | Writes to stdout with newline. Returns `true`. |
| `read`   | `()` | Reads line from stdin. Returns string. |
| `open`   | `(path)` | Opens file, returns byte stream function. |
| `create` | `(path)` | Creates file, returns byte sink function. |
| `close`  | `(stream)` | Closes a stream. |

Streams are functions:
```
let f = open("file.txt")
f()      # [@ok <bytes>] or [@end]
close(f)

let out = create("out.txt")
out(bytes)   # write bytes
close(out)
```

### Math

| Function | Signature | Description |
|----------|-----------|-------------|
| `sqrt`   | `(n)` | Square root. |
| `cos`    | `(n)` | Cosine (radians). |
| `sin`    | `(n)` | Sine (radians). |
| `random` | `(n)` | Random integer from 0 to n-1. |

### Bit Operations

| Function | Signature | Description |
|----------|-----------|-------------|
| `bit_and` | `(a b)` | Bitwise AND. Integers only. |
| `bit_or`  | `(a b)` | Bitwise OR. Integers only. |
| `bit_xor` | `(a b)` | Bitwise XOR. Integers only. |
| `bit_not` | `(a)` | Bitwise NOT (two's complement). Integers only. |
| `bit_shift` | `(a n)` | Left shift if n > 0, right if n < 0. Integers only. |

Non-integer input returns `[@error "requires integer"]`.

### Regex

| Function | Signature | Description |
|----------|-----------|-------------|
| `regex` | `(str pattern)` | Match pattern. Returns `[@ok [full groups...]]` or `[@error "no match"]`. |
| `regex_all` | `(str pattern)` | Find all matches. Returns `[@ok [matches...]]` or `[@error "no match"]`. |
| `regex_replace` | `(str pattern repl)` | Replace matches. Returns new string. |

RE2 semantics: linear time, no backtracking, no backreferences.

### Concurrency (Experimental)

| Function | Signature | Description |
|----------|-----------|-------------|
| `spawn` | `(f)` | Start green thread. Returns immediately. |
| `channel` | `()` | Create unbuffered channel. |
| `send` | `(ch val)` | Send value. Blocks until received. |
| `recv` | `(ch)` | Receive value. Blocks until sent. |

Known limitations:
- No protection against data races
- Shared mutable state is dangerous
- Channel operations are the only yield points

Best practice: Communicate through channels. Don't share mutable state.

### Canvas

| Function | Signature | Description |
|----------|-----------|-------------|
| `canvas` | `(w h)` | Create window, return handle. |
| `draw_line` | `(c x1 y1 x2 y2 color)` | Draw line. |
| `draw_rect` | `(c x y w h color)` | Draw rectangle. |
| `draw_circle` | `(c x y r color)` | Draw circle. |
| `draw_text` | `(c x y text color)` | Draw text. |
| `clear` | `(c color)` | Clear canvas. |
| `present` | `(c)` | Flush to screen. |
| `close` | `(c)` | Close window. |

Colors: `:red`, `:blue`, `:green`, `:white`, `:black`, etc. or `[@rgb 255 128 0]`.

### System

| Function | Signature | Description |
|----------|-----------|-------------|
| `help` | `()` or `(fn)` | Show help or describe function. |
| `quit` | `()` | Exit. |
| `universe` | `()` | List all loaded modules. |
| `symbols` | `()` | List bindings in current scope. |
| `history` | `()` | Expression history. |
| `peek` | `(path)` | Inspect file without loading. |
| `load` | `(path)` | Load and execute file. |
| `stack_limit` | `(n)` | Set recursion depth limit. |

### Note on Arithmetic

Arithmetic uses operators, not functions:
- `+`, `-`, `*`, `/`, `%` for math
- `==`, `!=`, `<`, `>`, `<=`, `>=` for comparison
- `and`, `or`, `not` for logic

One way to add numbers: `a + b`.

All arithmetic is exact (rational). Division by zero returns `[@error "division by zero"]`.

---

## Prelude

Written in Aiki. Proves the language. User can read, modify, replace.

### each

Side effects over a list.

```
let each = (list f) {
    let i = 0
    while i < len(list) {
        f(list.i)
        i = i + 1
    }
    return true
}
```

### map

Transform each element.

```
let map = (list f) {
    let result = []
    let i = 0
    while i < len(list) {
        result = append(result f(list.i))
        i = i + 1
    }
    return result
}
```

### filter

Keep elements that pass a test.

```
let filter = (list f) {
    let result = []
    let i = 0
    while i < len(list) {
        if f(list.i) {
            result = append(result list.i)
        }
        i = i + 1
    }
    return result
}
```

### reduce

Accumulate a single value.

```
let reduce = (list acc f) {
    let result = acc
    let i = 0
    while i < len(list) {
        result = f(result list.i)
        i = i + 1
    }
    return result
}
```

### range

Generate a list of numbers.

```
let range = (start end) {
    let result = []
    let i = start
    while i < end {
        result = append(result i)
        i = i + 1
    }
    return result
}
```

### reverse

Reverse a list.

```
let reverse = (list) {
    let result = []
    let i = len(list) - 1
    while i >= 0 {
        result = append(result list.i)
        i = i - 1
    }
    return result
}
```

### find

Find first element matching a predicate.

```
let find = (list f) {
    let i = 0
    while i < len(list) {
        if f(list.i) {
            return [@ok list.i]
        }
        i = i + 1
    }
    return [@error "not found"]
}
```

### any

True if any element passes.

```
let any = (list f) {
    let i = 0
    while i < len(list) {
        if f(list.i) {
            return true
        }
        i = i + 1
    }
    return false
}
```

### all

True if all elements pass.

```
let all = (list f) {
    let i = 0
    while i < len(list) {
        if not f(list.i) {
            return false
        }
        i = i + 1
    }
    return true
}
```

### as_lines

Wrap byte stream to yield lines.

```
let as_lines = (stream) {
    let buffer = []
    return () {
        while true {
            match stream() {
                [@ok chunk] {
                    # accumulate bytes, split on newline
                    # return [@ok line] when complete
                }
                [@end] {
                    if len(buffer) > 0 {
                        let line = tostr(buffer)
                        buffer = []
                        return [@ok line]
                    }
                    return [@end]
                }
            }
        }
    }
}
```

### as_runes

Wrap byte stream to yield runes.

```
let as_runes = (stream) {
    let buffer = []
    return () {
        # accumulate bytes until valid UTF-8 rune
        # return [@ok rune] or [@end]
    }
}
```

### Turtle Graphics

Built on canvas primitives.

```
let @turtle [canvas x y angle]

let turtle = () {
    let c = canvas(400 400)
    clear(c :black)
    return [@turtle c 200 200 0]
}

let forward = (t dist) {
    let rad = t.angle * 3.14159 / 180
    let nx = t.x + (dist * cos(rad))
    let ny = t.y + (dist * sin(rad))
    draw_line(t.canvas t.x t.y nx ny :white)
    t.x = nx
    t.y = ny
    present(t.canvas)
    return t
}

let right = (t degrees) {
    t.angle = t.angle + degrees
    return t
}

let left = (t degrees) {
    t.angle = t.angle - degrees
    return t
}
```

Usage:
```
let t = turtle()
t |> forward(100) |> right(90) |> forward(100)
```

---

## Why This Split

**Primitives** touch runtime internals:
- `first`, `rest`, `prepend`, `append` manipulate list memory
- `type`, `inspect`, `shape` query runtime type information
- `print`, `read`, `open`, `create` perform I/O
- `sqrt`, `cos`, `sin`, `random` call system libraries
- `spawn`, `channel`, `send`, `recv` manage green threads
- `canvas`, `draw_*` interface with graphics system
- `regex*` use Go's regexp engine

You cannot write `first` in Aiki without `first`.

**Prelude** is pure Aiki:
- `map` is just `while` + `append`
- `filter` is just `while` + `if` + `append`
- `as_lines` wraps a stream function
- `turtle` wraps canvas primitives
- Every function is readable, understandable, replaceable

The split follows from necessity, not convention.

---

## Usage

Prelude loads automatically. Override by defining your own:

```
# my faster map using some trick
let map = (list f) {
    # your implementation
}
```

Your `map` shadows the prelude. The prelude version still exists but is unreachable in your scope.
