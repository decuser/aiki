## Style

| Item | Rule |
|------|------|
| Indentation | 4 spaces |
| Braces | Same line |
| Single expression | Braces required |
| Operators | Spaces around (`x + y`) |
| Commas | Space after, trailing required (`[1, 2, 3,]`) |
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
    data
}

let result = compute(
    arg1,
    arg2,
    arg3,
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
| Variable names | Single letter ok for indices, counters. Otherwise say the name. |
| Function names | Verbs for actions (`draw`, `send`), nouns for constructors (`canvas`, `array`), no prefix for predicates (`empty`, `even`) |
| Error handling | `nil` for absence, `[@error, msg]` for failure |
| Comments | Never required, never forbidden. Suggested at start of file. |
| Module organization | Related groups |
| Strict vs pragmatic | Strict default, pragmatic when performance needed |
| Magic numbers | Name if tunable/config. Inline ok for obvious bounds, indices. |
