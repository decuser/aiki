# Aiki Style Guide

## Formatting

| Item | Rule |
|------|------|
| Indentation | 4 spaces |
| Braces | Same line |
| Single expression | Braces required |
| Operators | Spaces around (`x + y`) |
| Commas | Space after (`[1, 2, 3]`) |
| Continuation | Comma trails on each line |
| Pipelines | Break after source, `|>` at margin |
| Line length | 80 soft, 100 hard |
| Blank lines | One between functions |
| Parens | No inner space (`foo(x, y)`) |
| Comments | `# ` with space |
| Naming | snake_case |

**Examples:**

```aiki
let process = (data, threshold) {
    if data > threshold {
        return transform(data)
    }
    return data
}

let result = compute(
    arg1,
    arg2,
    arg3
)

data
|> filter(even)
|> map(square)
|> take(10)
|> sum()
```

---

## Conventions

| Item | Convention |
|------|------------|
| Function length | Fits on screen (~50 lines max) |
| Nesting depth | 3-4 levels max, then refactor |
| Variable names | Single letter ok for indices, counters. Otherwise descriptive. |
| Function names | Verbs for actions (`draw`, `send`), nouns for constructors (`canvas`, `channel`), no prefix for predicates (`empty`, `even`) |
| Error handling | `[@error, reason]` for failure, value for success |
| Comments | Never required, never forbidden. Suggested at file start. |
| Module organization | Related functions grouped |
| Strict vs pragmatic | Strict default, pragmatic when performance needed |
| Magic numbers | Name if tunable/config. Inline ok for obvious bounds, indices. |

---

## Naming

| Category | Style | Examples |
|----------|-------|----------|
| Variables | snake_case | `total_count`, `user_input` |
| Functions | snake_case | `find_first`, `to_string` |
| Shapes | @snake_case | `@point`, `@http_response` |
| Symbols | :snake_case | `:ok`, `:not_found` |
| Constants | snake_case | `max_size`, `default_color` |

Abbreviations: Full words preferred (`position` not `pos`), except for common short forms (`x`, `y`, `i`, `n`, `fn`).
