# Aiki Alpha Design

---

## Architecture Philosophy

**Development Priority:**

| Priority | Layer | When to touch |
|----------|-------|---------------|
| 1 | Core language | Almost never. No breaking changes, period. |
| 2 | `hal` | Only if absolute necessity. Wraps Go/OS. |
| 3 | `strict` | First choice for new features. Pure Aiki. |
| 4 | `pragmatic` | Second choice. Fast twin of strict. |

**The Twin Pattern:** Every capability has two implementations. `strict.X` is pure Aiki on lists and rationals, inspectable and slow. `pragmatic.X` is Aiki on arrays and floats, fast with same API. No shadowing. Namespace separation: `strict.map` vs `pragmatic.map`. Extension path: write it in strict first, add pragmatic twin when speed is needed.

---

## Layer Definitions

| Layer | Name | What it is | Auto-load |
|-------|------|------------|-----------|
| `hal` | Hardware Abstraction Layer | Go primitives - memory, I/O, concurrency, canvas window, OS | Always available |
| `strict` | Default Aiki | Lists, rationals, interpreter speed | Yes (default) |
| `pragmatic` | Opt-in fast path | Arrays, floats, hardware speed | No (opt-in) |

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

**Cut:** `from`, `use`, `export` → become `import()` and `export()` functions in hal.

---

## Types (8)

| Type | Notes |
|------|-------|
| Number | Rationals |
| Boolean | true/false |
| String | "..." |
| Rune | '...' |
| Symbol | :name |
| List | [...] |
| Function | (args) { body } |
| Bytes | Binary data |

**Cut:** Handle, Canvas, Channel → become shaped lists with integer ids.

| Resource | Shape |
|----------|-------|
| File handle | `[@handle, id]` |
| Canvas | `[@canvas, id]` |
| Channel | `[@channel, id]` |

Hal maintains registries mapping ids to Go resources. Invalid ids return errors, not crashes.

---

## Operators

**Cut:**

| Operator | Replacement |
|----------|-------------|
| `!` | `not` (keyword) |
| `==` | `equal(a, b)` in strict, `hal.eq` for atoms |
| `!=` | `not(equal(a, b))` |
| `%` | `modulo(a, b)` - floored semantics |

**Keep:**

| Operator | Notes |
|----------|-------|
| `+`, `-`, `*`, `/` | Holy four |
| `<`, `>`, `<=`, `>=` | Comparison |
| `|>` | Pipe |
| `.` | Access |
| `[]` | Indexing |

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

**Invocation:**
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

## Concurrency Model

**Philosophy:** Erlang's safety with Go's performance. Share-nothing, message-passing.

**Primitives (hal):**
```
spawn(fn, args...)   # isolated thread, no closure capture
channel()            # unbuffered channel, returns [@channel, id]
send(ch, val)        # blocks until received
recv(ch)             # blocks until sent
```

**Rules:**
- Spawned functions are orphans - no access to parent scope, no closure capture
- They only see their arguments + read-only globals (strict)
- Channels are the only communication - no shared mutable state
- Go's default channel semantics - pass-by-value
- No mutex tax - single-threaded code has zero locking overhead
- No `move` semantics - not implemented

**Error behavior:**
```aiki
let x = 100
spawn(() { print(x) })   # ERROR: 'x' is not defined in this scope
```

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

**Import:**
```aiki
import("pragmatic", [:map, :sum])
```

**Export:**
```aiki
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

## Monitor Package (Deferred)

**Architecture:**

| Layer | Purpose |
|-------|---------|
| Hal | `canvas`, `present`, mouse |
| Strict/Pragmatic | Drawing primitives |
| Monitor | Dashboard, sparklines, soft-keys, buffer, terminal spawn |

**Three-Tiered Observation:**

| Tier | Trigger | What it does |
|------|---------|--------------|
| Snapshot | `^t` / SIGUSR1 / SNAP button | Single status line |
| Toggle | SIGUSR1 / TRACE button | Stream to terminal |
| Explicit | `profile()` in code | Scoped profiling |

**Pre-alpha:** Log/trace infrastructure only. Visual interface deferred.

---

## Summary

| Item | Count |
|------|-------|
| Keywords | 11 |
| Types | 8 |
| Operators | 10 (`+`, `-`, `*`, `/`, `<`, `>`, `<=`, `>=`, `|>`, `.`, `[]`) |
| Layers | 3 (hal, strict, pragmatic) |
