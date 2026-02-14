# Aiki Done

## v0.2.5

### Architecture
- EBNF grammar-driven parser (grammar.ebnf is source of truth)
- Lexer/parser/fmt/lint all derive from grammar
- Three-layer error system (user/strict/pragmatic/hal)

### Rich Errors
- `makeError(env, node, ...)` with file:line:source
- Stack traces with `from` lines
- HAL errors annotated with call-site position
- `InspectAtLayer()` for filtered display by layer
- `<main>` frame pushed at entry points

### Internals
- `Layer` type on StackFrame and Function
- `PushFrame(name, line, layer)` / `PopFrame()`
- `AnnotateError()` for HAL error decoration
- Lint rewritten with EBNF AST walker
- `strict.Exports()` parses from strict.ai (no hardcoded list)

## v0.2.4

### Syntax
- Bracket indexing `list[i]`, `string[i]`, `f()[0]`
- Dot access for fields only (no `list.0`)
- Rest parameters `(...args)`, `(a, b, ...rest)`
- Semicolon statement separator
- GroupExpression preserves parentheses

### Primitives
- `ord(rune)` — rune to code point
- `apply(fn, list)` — spread list as args
- `shell(cmd)` — run shell command
- Remove `nth` (replaced by bracket indexing)

### CLI
- `-e` flag for one-liner expressions
- `quit()` actually exits REPL

### Internals
- HAL is single source of truth for builtins
- REPL subcommand pattern (`cmd/repl/`)
- Extract version package
- TrackingWriter replaces global state

### Fixes
- Remove blank line on Ctrl-C
- Remove "exit" echo on Ctrl-D

## v0.2.3

- Lexer (state machine, UTF-8)
- Parser (recursive descent, left-to-right)
- Evaluator (tree-walking)
- 8 types (number, boolean, rune, string, bytes, symbol, list, function)
- Rational arithmetic (exact, via big.Rat)
- Shapes with composition
- Pattern matching
- Pipe operator with error short-circuit
- Prelude (map, filter, reduce, range, hash map)
- File I/O (open, create, fread, fwrite, fclose)
- REPL with readline
- File runner
- Formatter with comment preservation
- Comma separators
- Ruby-style error messages
- Canvas primitives (Ebiten)
- Concurrency (spawn, channel, send, recv)
