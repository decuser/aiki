**Line Continuation Rules**

A line ends a statement if it ends with:
- a name (`x`, `my_var`)
- a number (`42`, `3.14`)
- a string (`"hello"`)
- a rune (`'a'`)
- a symbol (`:ok`)
- `true` or `false`
- `)` or `]`

A line continues to the next if it ends with:
- an operator (`+`, `-`, `*`, `/`, `<`, `>`, `<=`, `>=`, `and`, `or`, `|>`)
- a comma (`,`)
- an open bracket (`(`, `[`, `{`)

**Braces must be on the same line as `if`, `while`, `match`:**
```aiki
# correct
if x {
    y
}

# error
if x
{
    y
}
```

**Multiple statements on one line use explicit semicolon:**
```aiki
let x = 1; let y = 2; println(x + y)
```
