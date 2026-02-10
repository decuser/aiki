# Aiki Standard Library

## Convention

Success returns value. Failure returns `[@error, reason]`.

## Primitives

Implemented in Go.

### List

| Function | Description |
|----------|-------------|
| `first(list)` | First element. Error if empty. |
| `rest(list)` | List without first. Empty if single. |
| `nth(list, n)` | Element at index n. |
| `len(list)` | Element count. |
| `prepend(list, val)` | Add to start. |
| `append(list, val)` | Add to end. |

### Type

| Function | Description |
|----------|-------------|
| `type(val)` | Returns `:number`, `:boolean`, `:string`, `:rune`, `:bytes`, `:symbol`, `:list`, `:function`, `:handle`. |
| `shape(val)` | Returns shape symbol or `:list` for raw. |
| `equal(a, b)` | Deep equality. |

### Convert

| Function | Description |
|----------|-------------|
| `tostr(val)` | Value to string. |
| `tonum(str)` | String to number. Error if invalid. |
| `todecimal(n, places)` | Number to decimal string. |

### I/O

| Function | Description |
|----------|-------------|
| `print(val...)` | Write to stdout. |
| `read()` | Read line from stdin. `[@end]` at EOF. |
| `open(path)` | Open for reading. Returns handle or error. |
| `create(path)` | Create for writing. Returns handle or error. |
| `fread(handle)` | Read chunk. Returns string, `[@end]`, or error. |
| `fwrite(handle, data)` | Write string/bytes. Returns `true` or error. |
| `fclose(handle)` | Close handle. |

### Math

| Function | Description |
|----------|-------------|
| `sqrt(n)` | Square root. |
| `abs(n)` | Absolute value. |
| `random(n)` | Random integer 0 to n-1. |

### System

| Function | Description |
|----------|-------------|
| `fmt(path)` | Format file or directory. |
| `reset()` | Clear environment, close canvases |

### Internal

| Function | Description |
|----------|-------------|
| `_hash_code(val)` | Hash code for hash map implementation. |

## Prelude

Implemented in Aiki.

### Shapes

```
let @ok [value]
let @error [reason]
let @end []
```

### Iteration

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

## Operators

Arithmetic: `+`, `-`, `*`, `/`, `%`
Comparison: `==`, `!=`, `<`, `>`, `<=`, `>=`
Logic: `and`, `or`, `not`

Division by zero returns `[@error, "division by zero"]`.

## Concurrency

| Function | Description |
|----------|-------------|
| `spawn(function)` | Run function in goroutine, returns `true` |
| `channel()` | Create unbuffered channel |
| `send(channel, value)` | Send value, blocks until received |
| `recv(channel)` | Receive value, blocks until sent |

## Canvas

| Function | Description |
|----------|-------------|
| `canvas(width, height)` | Create window, black bg, white fg |
| `get_bg(canvas)` | Get background color |
| `set_bg(canvas, color)` | Set background color |
| `get_fg(canvas)` | Get foreground color |
| `set_fg(canvas, color)` | Set foreground color |
| `pen_size(canvas, size)` | Set pen size (default 2) |
| `get_pen_size(canvas)` | Get pen size |
| `line(canvas, x1, y1, x2, y2)` | Line |
| `rect(canvas, x, y, width, height)` | Rectangle outline |
| `fill_rect(canvas, x, y, width, height)` | Filled rectangle |
| `circle(canvas, x, y, radius)` | Circle outline |
| `fill_circle(canvas, x, y, radius)` | Filled circle |
| `oval(canvas, x, y, radius_x, radius_y)` | Ellipse outline |
| `fill_oval(canvas, x, y, radius_x, radius_y)` | Filled ellipse |
| `dot(canvas, x, y)` | Pixel |
| `text(canvas, x, y, string)` | Text |
| `clear(canvas)` | Fill with bg |
| `undo(canvas)` | Undo last op |
| `redo(canvas)` | Redo last undo |
| `get_width(canvas)` | Canvas width |
| `get_height(canvas)` | Canvas height |
| `save(canvas, path)` | Save PNG |
| `destroy(canvas)` | Close window |

Colors: `:black`, `:blue`, `:green`, `:cyan`, `:red`, `:magenta`, `:brown`, `:white`, `:gray`, `:bright_blue`, `:bright_green`, `:bright_cyan`, `:bright_red`, `:bright_magenta`, `:yellow`, `:bright_white`, or `[r, g, b]`.

Draw functions accept optional color as last argument to override fg.
