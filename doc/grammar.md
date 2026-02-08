# Aiki Grammar

## Tokens

```
keywords:    let if else while match return export from use true false and or not
sigils:      @ :
delimiters:  ( ) [ ] { }
operators:   + - * / % == != < > <= >= and or not
pipe:        |>
access:      .
assignment:  =
comment:     # to end of line
```

### Token Types

```
NUMBER       42, 3.14, 1/3
BOOLEAN      true, false
STRING       "hello"
RUNE         'a', '€'
SYMBOL       :active, :ok, :error
SHAPE        @point, @user
NAME         x, square, user_name
```

## Productions

### Program Structure

```ebnf
program     = { statement }

statement   = binding | shape | assignment | if | while | match
            | return | export | load | import | expr
```

### Binding and Assignment

```ebnf
binding     = "let" name "=" expr

shape       = "let" "@" name "[" { name | "@" name } "]"

assignment  = name "=" expr
```

`let` creates a new name. `=` without `let` mutates an existing name; error if not found.

Shapes can embed other shapes:
```
let @pet [name age]
let @cat [@pet color]    # embeds pet's fields
```

### Control Flow

```ebnf
if          = "if" expr block [ "else" block ]

while       = "while" expr block

match       = "match" expr "{" { match_arm } "}"

match_arm   = pattern block

pattern     = value | name | "_" | "[" { pattern } "]"
            | "[@" name { pattern } "]"
```

### Functions

```ebnf
function    = "let" name "=" "(" { name } ")" block

lambda      = "(" { name } ")" block

return      = "return" expr
```

Every path through a function must have explicit `return`.

### Expressions

```ebnf
expr        = value | name | access | call | list | lambda
            | "(" expr op expr ")"

access      = expr "." ( name | number )

call        = expr "(" { expr } ")"

pipe_expr   = expr { "|>" call }

list        = "[" { expr } "]"
            | "[@" name { expr } "]"
```

### Operators

```ebnf
op          = "+" | "-" | "*" | "/" | "%"
            | "==" | "!=" | "<" | ">" | "<=" | ">="
            | "and" | "or" | "not"
```

No precedence. Compound expressions require parentheses.

### Modules

```ebnf
export      = "export" "[" { name } "]"

load        = "let" name "=" "load" "(" string ")"

import      = "from" name "use" "[" { name } "]"
```

### Values

```ebnf
value       = number | boolean | string | rune | symbol

number      = digit { digit } [ "/" digit { digit } ]
            | digit { digit } [ "." digit { digit } ]
boolean     = "true" | "false"
string      = '"' { character } '"'
rune        = "'" character "'"
symbol      = ":" name

name        = letter { letter | digit | "_" }
block       = "{" { statement } "}"
```

## Types

Eight types:

| Type     | Example              | Notes                                    |
|----------|----------------------|------------------------------------------|
| number   | `42`, `3/4`, `1.5`   | Rational (exact arithmetic)              |
| boolean  | `true`, `false`      |                                          |
| rune     | `'a'`, `'€'`         | Unicode code point                       |
| string   | `"hello"`            | Immutable rune sequence                  |
| bytes    | `tobytes("hi")`      | Immutable byte sequence (0-255)          |
| symbol   | `:active`, `:ok`     | Atomic, compared by identity             |
| list     | `[1 2 3]`            | Raw or shaped                            |
| function | `(x) { return x }`   | First-class                              |

### Numbers

All numbers are rational (exact):
- Integers: `42` (stored as 42/1)
- Rationals: `1/3`, `3/4`
- Decimal notation: `3.14` (converted to rational)
- Arithmetic is exact: `(1/3) * 3` equals `1`
- Display: `todecimal(1/3 5)` → `"0.33333"`

### Strings and Bytes

Strings are immutable lists of runes:
```
let s = "hello"
s.0         # 'h' (rune)
len(s)      # 5
first(s)    # 'h'
```

Bytes are immutable lists of 0-255:
```
let b = tobytes("hello")
b.0         # 104
len(b)      # 5
```

Conversion:
```
tobytes('é')     # [195 169] (UTF-8)
torune([195 169]) # 'é'
tobytes("hello")  # [104 101 108 108 111]
tostr(bytes)      # "hello"
```

## Lists

Two kinds:

```
[1 2 3]              # raw - positional access only
[@point 10 20]       # shaped - named access, enforced
```

Shaped lists require a shape declaration:

```
let @point [x y]
```

Shapes can embed other shapes:

```
let @pet [name age]
let @cat [@pet color]       # has name, age, color
let @siamese [@cat points]  # has name, age, color, points
```

Access:

```
list.0               # positional (raw or shaped)
point.x              # named (shaped only)
```

Shape inspection:

```
shape([@point 10 20])  # :point
shape([1 2 3])         # :list
```

## Scope

- `let x = 7` creates in current scope
- `x = 9` mutates nearest enclosing `x`, error if none
- `{}` creates new scope
- Inner sees outer, outer never sees inner

## Pipe

```
x |> f() |> g()
```

Left side becomes first argument of right side. If any step returns `[@error reason]`, subsequent steps skip.

## Primitives

Functions implemented in the runtime:

| Category | Functions |
|----------|-----------|
| List     | `first`, `rest`, `len`, `prepend`, `append` |
| Type     | `type`, `inspect`, `shape` |
| Compare  | `equal` |
| Convert  | `tostr`, `tonum`, `tobytes`, `torune`, `todecimal` |
| I/O      | `print`, `read`, `open`, `create`, `close` |
| Math     | `sqrt`, `cos`, `sin`, `random` |
| Bit      | `bit_and`, `bit_or`, `bit_xor`, `bit_not`, `bit_shift` |
| Regex    | `regex`, `regex_all`, `regex_replace` |
| Concurrency | `spawn`, `channel`, `send`, `recv` |
| Canvas   | `canvas`, `draw_line`, `draw_rect`, `draw_circle`, `draw_text`, `clear`, `present` |
| System   | `help`, `quit`, `universe`, `symbols`, `history`, `peek`, `load`, `stack_limit` |

Arithmetic and logic use operators: `+`, `-`, `*`, `/`, `%`, `==`, `!=`, `<`, `>`, `<=`, `>=`, `and`, `or`, `not`.

## Keywords

14 total: `let`, `if`, `else`, `while`, `match`, `return`, `export`, `from`, `use`, `true`, `false`, `and`, `or`, `not`
