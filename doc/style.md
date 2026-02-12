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

Fixed values: convention, not enforcement. The name is the warning.

## Conventions

- **Function length.** Fits on screen. If scrolling, refactor.
- **Nesting depth.** Shallow is better. Extract, early return, pipeline.
- **Error handling.** Return `[@error, reason]` on failure. Return value on success.
- **Comments.** Never required, never forbidden. File header if helpful.
- **Strict vs pragmatic.** Strict by default. Pragmatic when profiling says so.

## Formatting

Run `aiki fmt`. One style. No debates.

## Examples

See `strict.ai`.
