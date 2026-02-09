# Aiki Standard Library (v0.2.2)

The library is split into two layers:
1. **Primitives:** Implemented in Go. Cannot be written in Aiki.
2. **Prelude:** Implemented in Aiki. Loaded automatically. User can override.

## The Way

Functions that can fail return either the value (success) or `[@error reason]` (failure). Success is not wrapped.
```
find([1 2 3] isEven)   # 2, or [@error "not found"]
max([5 3 8])           # 8, or [@error "empty list"]
open("file.txt")       # handle, or [@error reason]
```

The pipe operator recognizes this:
- `[@error ...]` short-circuits the pipeline
- `[@ok ...]` auto-unwraps (for compatibility)
- Raw values pass through

**Success needs no announcement.**

---

## Primitives

These are the atomic operations required to build everything else.

### List

| Function | Signature | Description |
|----------|-----------|-------------|
| `first`  | `(list)` | Returns the first element. Error if empty. |
| `rest`   | `(list)` | Returns list without first element. Empty if single. |
| `len`    | `(list)` | Returns number of elements. O(1). |
| `prepend`| `(list val)` | Adds `val` to start. Returns new list. |
| `append` | `(list val)` | Adds `val` to end. Returns new list. |
| `nth`    | `(list n)` | Returns element at index n. |

Works on raw lists, shaped lists, strings, and bytes.

### Type

| Function | Signature | Description |
|----------|-----------|-------------|
| `type`   | `(val)` | Returns `:number`, `:boolean`, `:string`, `:rune`, `:bytes`, `:symbol`, `:list`, `:function`, `:handle`. |
| `inspect`| `(val)` | Returns string representation for debugging. |
| `shape`  | `(val)` | Returns shape symbol (`:point`, `:error`, etc.) or `:list` for raw lists. |

### Compare

| Function | Signature | Description |
|----------|-----------|-------------|
| `equal`  | `(a b)` | Deep equality. Returns `true` or `false`. |

### Convert

| Function | Signature | Description |
|----------|-----------|-------------|
| `tostr`  | `(val)` | Converts value to string. |
| `tonum`  | `(str)` | Parses string to number. Returns number or `[@error reason]`. |
| `todecimal` | `(n places)` | Converts number to decimal string with given precision. |

### I/O

| Function | Signature | Description |
|----------|-----------|-------------|
| `print`  | `(val...)` | Writes to stdout with newline. Returns `null`. |
| `read`   | `()` | Reads line from stdin. Returns string or `[@end]` at EOF. |
| `open`   | `(path)` | Opens file for reading. Returns handle or `[@error reason]`. |
| `create` | `(path)` | Creates file for writing. Returns handle or `[@error reason]`. |
| `fread`  | `(handle)` | Reads chunk from file. Returns string, `[@end]`, or `[@error reason]`. |
| `fwrite` | `(handle data)` | Writes string or bytes to file. Returns `true` or `[@error reason]`. |
| `fclose` | `(handle)` | Closes file handle. Returns `true`. |

Example:
```
let h = create("out.txt")
fwrite(h "hello\n")
fclose(h)

let r = open("out.txt")
let data = fread(r)
fclose(r)
```

### Math

| Function | Signature | Description |
|----------|-----------|-------------|
| `sqrt`   | `(n)` | Square root. |
| `cos`    | `(n)` | Cosine (radians). |
| `sin`    | `(n)` | Sine (radians). |
| `random` | `(n)` | Random integer from 0 to n-1. |

### System

| Function | Signature | Description |
|----------|-----------|-------------|
| `help`   | `()` | Show help. |
| `quit`   | `()` | Exit. |
| `fmt`    | `(path)` | Format file or directory (`./...` for recursive). |

### Note on Arithmetic

Arithmetic uses operators, not functions:
- `+`, `-`, `*`, `/`, `%` for math
- `==`, `!=`, `<`, `>`, `<=`, `>=` for comparison
- `and`, `or`, `not` for logic

All arithmetic is exact (rational). Division by zero returns `[@error "division by zero"]`.

---

## Prelude

Written in Aiki. Proves the language. User can read, modify, replace.

### Shapes
```
let @ok [value]      # success wrapper (for compatibility)
let @error [reason]  # failure
let @end []          # stream termination
```

### each

Side effects over a list. Returns `true`.
```
each([1 2 3] print)
```

### map

Transform each element. Returns new list.
```
map([1 2 3] (n) { return n * 2 })   # [2 4 6]
```

### filter

Keep elements that pass a test. Returns new list.
```
filter([1 2 3 4] (n) { return n > 2 })   # [3 4]
```

### reduce

Accumulate a single value.
```
reduce([1 2 3] 0 (acc n) { return acc + n })   # 6
```

### range

Generate a list of numbers.
```
range(1 5)   # [1 2 3 4]
```

### reverse

Reverse a list.
```
reverse([1 2 3])   # [3 2 1]
```

### find

Find first element matching a predicate.

Returns: value, or `[@error "not found"]`
```
find([1 2 3 4] (n) { return n > 2 })   # 3
find([1 2 3] (n) { return n > 10 })    # [@error "not found"]
```

### any

True if any element passes.
```
any([1 2 3] (n) { return n > 2 })   # true
```

### all

True if all elements pass.
```
all([1 2 3] (n) { return n > 0 })   # true
```

### sum

Sum of numbers in list.
```
sum([1 2 3 4])   # 10
```

### max

Maximum value in list.

Returns: value, or `[@error "empty list"]`
```
max([3 1 4 1 5])   # 5
max([])            # [@error "empty list"]
```

### min

Minimum value in list.

Returns: value, or `[@error "empty list"]`
```
min([3 1 4 1 5])   # 1
min([])            # [@error "empty list"]
```

### Hash Map

A hash map implemented in pure Aiki. O(1) average lookup.

| Function | Signature | Description |
|----------|-----------|-------------|
| `hash_new` | `()` | Create empty hash. |
| `hash_get` | `(h key)` | Get value. Returns value or `[@error "key not found"]`. |
| `hash_put` | `(h key val)` | Set value. Returns new hash. |
| `hash_has` | `(h key)` | Check if key exists. Returns boolean. |
| `hash_del` | `(h key)` | Remove key. Returns new hash. |
| `hash_keys` | `(h)` | List all keys. |
| `hash_values` | `(h)` | List all values. |

Example:
```
let h = hash_new()
let h = hash_put(h "name" "Mochi")
let h = hash_put(h "age" 3)
hash_get(h "name")   # "Mochi"
hash_has(h "color")  # false
```

---

## Why This Split

**Primitives** touch runtime internals:
- `first`, `rest`, `append` manipulate list memory
- `type`, `inspect`, `shape` query runtime type information
- `print`, `read`, `open`, `fread` perform I/O
- `sqrt`, `cos`, `sin`, `random` call system libraries

You cannot write `first` in Aiki without `first`.

**Prelude** is pure Aiki:
- `map` is just `while` + `append`
- `filter` is just `while` + `if` + `append`
- `hash_*` is lists of lists with a hash function
- Every function is readable, understandable, replaceable

The split follows from necessity, not convention.
