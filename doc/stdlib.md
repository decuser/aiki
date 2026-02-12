# Aiki Standard Library

## Convention

Success returns value. Failure returns `[@error, reason]`.

---

## Hal (Go Primitives)

Implemented in Go. Always available.

### List

| Function | Description |
|----------|-------------|
| `first(list)` | First element. Error if empty. |
| `rest(list)` | List without first. Empty if single. |
| `nth(list, n)` | Element at index n. |
| `len(list)` | Element count. |
| `prepend(list, val)` | Add to start. |
| `append(list, val)` | Add to end. |

### Arithmetic

| Function | Description |
|----------|-------------|
| `+`, `-`, `*`, `/` | Basic arithmetic. |
| `modulo(a, b)` | Floored modulo. *(provisional)* |
| `trunc(n)` | Truncate toward zero. *(provisional)* |

### Comparison

| Function | Description |
|----------|-------------|
| `<`, `>`, `<=`, `>=` | Numeric comparison. |
| `eq(a, b)` | Atom equality (numbers, booleans, symbols, runes). *(provisional)* |

### Type

| Function | Description |
|----------|-------------|
| `type(val)` | Returns `:number`, `:boolean`, `:string`, `:rune`, `:bytes`, `:symbol`, `:list`, `:function`. |
| `shape(val)` | Returns shape symbol or `:list` for raw. |
| `inspect(val)` | String representation. |

### Conversion

| Function | Description |
|----------|-------------|
| `tostr(val)` | Value to string. |
| `tonum(str)` | String to number. Error if invalid. |
| `todecimal(n, places)` | Number to decimal string. |
| `tobytes(val)` | String or list to bytes. |
| `torune(val)` | Number to rune. |

### I/O

| Function | Description |
|----------|-------------|
| `print(val...)` | Write to stdout. |
| `open(path)` | Open for reading. Returns handle or error. |
| `create(path)` | Create for writing. Returns handle or error. |
| `fread(handle)` | Read chunk. Returns string, `[@end]`, or error. |
| `fwrite(handle, data)` | Write string/bytes. Returns `true` or error. |
| `fclose(handle)` | Close handle. |

### Concurrency

| Function | Description |
|----------|-------------|
| `spawn(function)` | Run function in goroutine, returns `true`. |
| `channel()` | Create unbuffered channel. |
| `send(channel, value)` | Send value, blocks until received. |
| `recv(channel)` | Receive value, blocks until sent. |

### Canvas

| Function | Description |
|----------|-------------|
| `canvas(w, h)` | Create window. |
| `present(c)` | Flush to GPU. *(provisional)* |
| `destroy(c)` | Close window. |
| `mouse_pos(c)` | Returns `[x, y]` or `nil`. *(provisional)* |
| `mouse_down(c)` | Returns `true/false`. *(provisional)* |
| `load_image(path)` | Load image file. *(provisional)* |

### Module

| Function | Description |
|----------|-------------|
| `import(module, symbols)` | Import from module. *(provisional)* |
| `export(symbols)` | Export from current module. *(provisional)* |

### Control

| Function | Description |
|----------|-------------|
| `load(path)` | Load and run file. |
| `sleep(ms)` | Pause execution. |

### Math

| Function | Description |
|----------|-------------|
| `sqrt(n)` | Square root. |
| `sin(n)` | Sine. |
| `cos(n)` | Cosine. |
| `random(n)` | Random integer 0 to n-1. |

*Note: `sin`, `cos`, `sqrt`, `random` will move to pragmatic module.*

### Internal

| Function | Description |
|----------|-------------|
| `_hash_code(val)` | Hash code for hash map implementation. |

---

## Strict (Pure Aiki)

Implemented in Aiki. Auto-loaded.

### Shapes

```aiki
let @ok [value]
let @error [reason]
let @end []
```

### Higher-Order

| Function | Description |
|----------|-------------|
| `each(list, f)` | Apply f to each. Returns `true`. |
| `map(list, f)` | Transform each. Returns new list. |
| `filter(list, f)` | Keep where f returns true. |
| `reduce(list, acc, f)` | Accumulate with f(acc, elem). |

### List Operations

| Function | Description |
|----------|-------------|
| `range(start, end)` | List from start to end-1. |
| `reverse(list)` | Reversed list. |
| `find(list, f)` | First where f returns true. Error if none. |
| `any(list, f)` | True if any passes. |
| `all(list, f)` | True if all pass. |
| `sum(list)` | Sum of numbers. |
| `max(list)` | Maximum. Error if empty. |
| `min(list)` | Minimum. Error if empty. |

### Comparison

| Function | Description |
|----------|-------------|
| `equal(a, b)` | Deep structural equality. Uses `hal.eq` for atoms. |

### Hash Map

| Function | Description |
|----------|-------------|
| `hash_new()` | Create empty hash. |
| `hash_get(h, key)` | Get value. Error if missing. |
| `hash_put(h, key, val)` | Set value. Returns new hash. |
| `hash_has(h, key)` | Check if key exists. |
| `hash_del(h, key)` | Remove key. Returns new hash. |
| `hash_keys(h)` | List all keys. |
| `hash_values(h)` | List all values. |

### Graphics

| Function | Description |
|----------|-------------|
| `dot(canvas, x, y, color?)` | Pixel. |
| `line(canvas, x1, y1, x2, y2, color?)` | Line (Bresenham). |
| `rect(canvas, x, y, w, h, color?)` | Rectangle outline. |
| `fill_rect(canvas, x, y, w, h, color?)` | Filled rectangle. |
| `circle(canvas, x, y, r, color?)` | Circle outline (midpoint). |
| `fill_circle(canvas, x, y, r, color?)` | Filled circle. |
| `oval(canvas, x, y, rx, ry, color?)` | Ellipse outline. |
| `fill_oval(canvas, x, y, rx, ry, color?)` | Filled ellipse. |
| `poly(canvas, points, color?)` | Polygon outline. *(provisional)* |
| `fill_poly(canvas, points, color?)` | Filled polygon (scanline). *(provisional)* |
| `text(canvas, x, y, string, color?)` | Text (bitmap font). |
| `blit(canvas, image, x, y)` | Copy image. *(provisional)* |
| `blit_sub(canvas, image, sx, sy, sw, sh, dx, dy)` | Copy image region. *(provisional)* |
| `clear(canvas)` | Fill with background. |

Colors: `:black`, `:blue`, `:green`, `:cyan`, `:red`, `:magenta`, `:brown`, `:white`, `:gray`, `:bright_blue`, `:bright_green`, `:bright_cyan`, `:bright_red`, `:bright_magenta`, `:yellow`, `:bright_white`, or `[r, g, b]`.

---

## Pragmatic (Opt-in)

*(Not yet implemented)*

Fast implementations using arrays and floats. Same API as strict.

```aiki
import("pragmatic", [:map, :sum, :line])
```

| Category | Functions |
|----------|-----------|
| Types | `float`, `exact`, `array`, `tolist` |
| Math | `sin`, `cos`, `tan`, `sqrt`, `random` |
| Higher-order | `map`, `filter`, `reduce`, `each` |
| Aggregation | `sum`, `mean`, `min`, `max` |
| Graphics | `dot`, `line`, `rect`, `circle`, `poly`, `text`, `blit` |

---

## REPL Only

| Function | Description |
|----------|-------------|
| `help()` | Show primitives list. |
| `quit()` | Exit REPL. |
| `reset()` | Clear environment, close canvases. |
