# Aiki Style Guide

## Naming

| Category | Style | Examples |
|----------|-------|----------|
| Variables | snake_case | `total_count`, `user_input` |
| Functions | snake_case | `find_first`, `to_string` |
| Shapes | @snake_case | `@point`, `@http_response` |
| Symbols | :snake_case | `:ok`, `:not_found` |
| Fixed values | SCREAMING_SNAKE | `MAX_SIZE`, `DEFAULT_PORT` |
| Internal | _leading | `_bucket_index`, `_HASH_SIZE` |

Full words preferred. Exceptions: `x`, `y`, `i`, `n`, `fn`, loop counters.

## Naming Rules (Enforced)

- **snake_case only.** Mixed case (`firstName`, `FirstName`) is a syntax error.
- **Case locked on first use.** `MAX_SIZE` stays `MAX_SIZE`. Inconsistency is an error.
- **SCREAMING is all-or-nothing.** No `MAX_size`. If you scream, scream the whole thing.
- **`_prefix` is distinct.** `_name` and `name` are different identifiers. You can't have both.
- **Case propagates through imports.** If a module exports `MAX_INT`, you reference `MAX_INT`.

## Shadowing

- **Allowed, always warned.** You can shadow imports, library functions, operators.
- **Export `_name` warned.** Exporting internal-marked names triggers a warning.
- **You own the consequences.** The tools witness your choices, don't prevent them.

## Formatting (Enforced by `aiki fmt`)

- **Tabs for indentation.** Not spaces.
- **One statement per line.** Semicolons normalized away.
- **Control structures expanded.** `if x { y }` becomes multiline.
- **Trailing comma iff multiline.** `[1, 2, 3]` but `[\n  a,\n  b,\n]`.
- **Braces same line as keyword.** `if x {` not `if x\n{`.
- **`#` for comments.** No block comments.

## Conventions

- **Function length.** Fits on screen. If scrolling, refactor.
- **Nesting depth.** Shallow is better. Extract, early return, pipeline.
- **Error handling.** Return `[@error, reason]` on failure. Return value on success.
- **Comments.** Never required, never forbidden. File header if helpful.
- **Exact vs precise.** Exact by default. Precise when profiling says so.

## Command Strictness

| Command | Behavior |
|---------|----------|
| `aiki` | REPL. No rules. Workbench. |
| `aiki file` | Strict. Lint errors bail. Unused = error. |
| `aiki --proto file` | Loose. Warnings only. Runs anyway. TDD mode. |
| `aiki validate file` | Strict check. No execution. |

Default is strict. Opt into looseness with `--proto`.

## Error Depth

| Flag | Shows |
|------|-------|
| `--errors=user` | Your code only. Default. |
| `--errors=strict` | Your code + strict library. |
| `--errors=hal` | Everything. Full stack to primitives. |

Stack traces filter by layer. Deeper = more detail.

## Examples

See `strict.ai`.
