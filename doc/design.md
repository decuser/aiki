# Aiki Design

## Vision

A minimal, composable language with a live environment. The language is complete; the system layer (inspection, versioning, TUI) is future work.

## Core Principles

- **One way to do each thing.** Constraint forces clarity.
- **Composition through naming.** Define small pieces, name them, compose.
- **Explicit over implicit.** No magic, no hidden behavior.
- **Inspectability enables knowing.** Everything answers: what are you, what's in you.

---

## Architecture

**Development Priority:**

| Priority | Layer | When to touch |
|----------|-------|---------------|
| 1 | Core language | Almost never. No breaking changes, period. |
| 2 | `hal` | Only if absolute necessity. Wraps Go/OS. |
| 3 | `strict` | First choice for new features. Pure Aiki. |
| 4 | `pragmatic` | Second choice. Fast twin of strict. |

**Layer Definitions:**

| Layer | Name | What it is | Auto-load |
|-------|------|------------|-----------|
| `hal` | Hardware Abstraction Layer | Go primitives - memory, I/O, concurrency, canvas, OS | Always |
| `strict` | Default Aiki | Lists, rationals, interpreter speed | Yes |
| `pragmatic` | Opt-in fast path | Arrays, floats, hardware speed | No |

**The Twin Pattern:** Every capability has two implementations. `strict.X` is pure Aiki on lists and rationals, inspectable and slow. `pragmatic.X` is Aiki on arrays and floats, fast with same API. No shadowing. Namespace separation: `strict.map` vs `pragmatic.map`. Extension path: write it in strict first, add pragmatic twin when speed is needed.

---

## Keywords (11)

| Keyword | Purpose |
|---------|---------|
| `let` | Binding |
| `if` | Conditional |
| `else` | Conditional branch |
| `while` | Loop |
| `match` | Pattern matching |
| `return` | Early exit |
| `true` | Boolean literal |
| `false` | Boolean literal |
| `and` | Logic (short-circuit) |
| `or` | Logic (short-circuit) |
| `not` | Logic |

---

## Types (8)

| Type | Example | Notes |
|------|---------|-------|
| Number | `42`, `3/4` | Rational (exact) |
| Boolean | `true`, `false` | |
| Rune | `'a'`, `'€'` | Unicode code point |
| String | `"hello"` | Immutable rune sequence |
| Bytes | `tobytes("hi")` | Immutable 0-255 sequence |
| Symbol | `:ok`, `:error` | Atomic, identity-compared |
| List | `[1, 2, 3]` | Raw or shaped |
| Function | `(n) { return n }` | First-class |

**Resources as shaped lists:**

| Resource | Shape |
|----------|-------|
| File handle | `[@handle, id]` |
| Canvas | `[@canvas, id]` |
| Channel | `[@channel, id]` |

Hal maintains registries mapping ids to Go resources. Invalid ids return errors, not crashes.

---

## Operators

**Keep:**

| Operator | Notes |
|----------|-------|
| `+`, `-`, `*`, `/` | Arithmetic |
| `<`, `>`, `<=`, `>=` | Comparison |
| `\|>` | Pipe |
| `.` | Access |
| `[]` | Indexing |

**Cut:**

| Operator | Replacement |
|----------|-------------|
| `!` | `not` (keyword) |
| `==` | `equal(a, b)` in strict, `eq` in hal for atoms |
| `!=` | `not(equal(a, b))` |
| `%` | `modulo(a, b)` - floored semantics |

---

## Hal - The Machine (Go)

Only what you literally cannot compute in Aiki.

| Category | Primitives |
|----------|------------|
| Memory | `first`, `rest`, `nth`, `append`, `len`, `cons` |
| Arithmetic | `+`, `-`, `*`, `/`, `modulo`, `trunc` |
| Comparison | `<`, `>`, `<=`, `>=`, `eq` (atoms only) |
| Type | `type`, `inspect`, `shape` |
| Conversion | `tostr`, `tonum`, `tobytes`, `torune` |
| I/O | `print`, `open`, `fread`, `fwrite`, `fclose` |
| Concurrency | `spawn`, `channel`, `send`, `recv` |
| Canvas | `canvas`, `present`, `destroy`, `mouse_pos`, `mouse_down`, `load_image` |
| Module | `import`, `export` |
| Control | `load`, `help`, `quit` |

---

## Strict - Pure Aiki (Auto-loaded)

Built on hal. Inspectable. Proves the language works.

| Category | Functions |
|----------|-----------|
| Higher-order | `map`, `filter`, `reduce`, `each` |
| List ops | `range`, `reverse`, `find`, `any`, `all` |
| Aggregation | `sum`, `max`, `min` |
| Comparison | `equal` (deep structural, uses `hal.eq`) |
| Graphics | `dot`, `line`, `rect`, `fill_rect`, `circle`, `fill_circle`, `oval`, `fill_oval`, `poly`, `fill_poly`, `text`, `blit`, `blit_sub`, `clear` |
| Hash map | List-based implementation |

---

## Pragmatic - Fast Aiki (Opt-in)

Aiki on arrays and floats. Same API as strict, hardware speed.

```aiki
import("pragmatic", [:map, :sum, :line])
```

| Category | Functions |
|----------|-----------|
| Types | `float`, `exact`, `array`, `tolist` |
| Math | `sin`, `cos`, `tan`, `sqrt`, `random` |
| Higher-order | `map`, `filter`, `reduce`, `each` (array-backed) |
| Aggregation | `sum`, `mean`, `min`, `max` (vectorized) |
| Graphics | `dot`, `line`, `rect`, `circle`, `poly`, `text`, `blit` (array-backed) |
| Vectorized ops | `array + array`, `array * scalar`, etc. |

**Rules:**
- No implicit conversion - `exact + float = ERROR`
- `3.14` stays Rational - floats require explicit `float(3.14)`
- Arrays are homogeneous floats

---

## Concurrency

Minimal threading with channels. Go-lite model.

**Primitives (hal):**

```aiki
spawn(f)       # green thread, returns immediately
channel()      # unbuffered channel
send(ch, val)  # blocks until received
recv(ch)       # blocks until sent
```

No new syntax. Just builtins.

**Example:**

```aiki
let ch = channel()

spawn(() {
    send(ch, 42)
})

let val = recv(ch)
print(val)
```

**Known Limitations:**
- No protection against data races
- Shared mutable state is dangerous
- No preemption - long-running computations block
- Channel operations are the only yield points

**Best practice:** Don't share mutable state. Communicate through channels.

**Why Go-Lite:** Go's model (goroutines + channels) is well-understood and maps cleanly to composition philosophy. Message passing through channels is encouraged. Shared mutable state is possible but discouraged.

---

## Canvas

**Hal (the machine):**

| Function | Purpose |
|----------|---------|
| `canvas(w, h)` | Create window, returns `[@canvas, id]` |
| `present(c)` | Flush pixel array to GPU |
| `destroy(c)` | Cleanup |
| `mouse_pos(c)` | Returns `[x, y]` or `nil` |
| `mouse_down(c)` | Returns `true/false` |
| `load_image(path)` | Returns `[@image, id]` |

**Strict (Aiki algorithms):**

| Function | Implementation |
|----------|----------------|
| `dot` | Pixel write |
| `line` | Bresenham |
| `rect`, `fill_rect` | Loop |
| `circle`, `fill_circle` | Midpoint algorithm |
| `oval`, `fill_oval` | Ellipse algorithm |
| `poly`, `fill_poly` | Scanline fill |
| `text` | Bitmap font lookup |
| `blit`, `blit_sub` | Pixel copy |

**Pragmatic:** Same API, array-backed implementations.

---

## Module System

```aiki
import("pragmatic", [:map, :sum])
export([:foo, :bar])
```

Both are hal functions, not keywords.

---

## Regex Module

Separate module, not hal.

```aiki
import("regex", [:match, :find, :replace])

match("^hello", text)           # true/false
find("\\d+", text)              # list of matches
replace("foo", text, "bar")     # new string
```

---

## Decisions

**Rational numbers.** No floating point surprises. `0.1 + 0.2 == 0.3` works. `1/3 * 3 == 1`.

**Shaped lists.** `[@point, 10, 20]` enforces structure, enables named access. Raw lists `[1, 2, 3]` for flexibility. Both support `first`, `rest`, `len`.

**Shape composition.** `let @cat [@pet, color]` embeds pet's fields. Structural, not inheritance.

**Symbols are atomic only.** `:ok` is a value, not a list tag. `@` shapes tag lists.

**Comma separators.** `f(a, b)` and `[1, 2, 3]`. Unambiguous, consistent.

**Left-to-right evaluation.** No operator precedence. `1 + 2 * 3` is `9`. Use parens: `1 + (2 * 3)` is `7`.

**`let` creates, `=` mutates.** Catches typos. `countr = 1` errors if `countr` doesn't exist.

**Explicit return.** Every path needs `return`. Enables guard clauses, prevents last-line bugs.

**No `elif`.** Guard clauses with early return. One way to branch.

**One function syntax.** `(params) { body }`. No arrows, no variations.

**`while` only.** Iteration via `map`, `filter`, `each`. One way to loop.

**Empty parens required.** `f()` calls, `f` references. Consistent.

**Pipe operator.** `x |> f() |> g()`. Left becomes first arg. `[@error, ...]` short-circuits.

**Success returns value.** No `[@ok, ...]` wrapper on success. `[@error, reason]` on failure.

**Dot access.** `list.0`, `point.x`, `math.sqrt`. One syntax for all member access.

**No exceptions.** Errors are values. Match or pipe handles them.

**No macros.** Language is closed. Functions are the only extension.

---

## Rejected

| Feature | Reason |
|---------|--------|
| Arrow lambdas | Second way to write functions |
| Implicit return | Last-line bugs |
| `for` loop | `while` + functions cover it |
| `elif` | Guard clauses |
| Operator precedence | Parens are explicit |
| Classes/methods | Data doesn't have behavior |
| Exceptions | Errors are values |
| Macros | Language is closed |
| Floats | Rationals are exact |

---

## Summary

| Item | Count |
|------|-------|
| Keywords | 11 |
| Types | 8 |
| Operators | 10 (`+`, `-`, `*`, `/`, `<`, `>`, `<=`, `>=`, `\|>`, `.`, `[]`) |
| Layers | 3 (hal, strict, pragmatic) |
