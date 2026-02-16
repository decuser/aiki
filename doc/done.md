# Aiki Done

## v0.2.7 (Architecture)

### Grammar Cleanup
- Remove `%`, `==`, `!=` from OPERATOR and BINOP ✔
- `modulo()`, `equal()`, `not equal()` replace operators ✔

### Grammar-Evaluator Coupling
- Handler map replaces switch statement ✔
- `ValidateHandlers()` panics on missing handler ✔
- Called from `SetNodeGrammar()` at startup ✔
- No drift possible — grammar change forces handler update ✔

### Eval Package Reorganization
- `node.go` — handler map, EvalNode, all eval logic ✔
- `module.go` — SetNodeGrammar, import/export ✔
- `intrinsics.go` — call dispatch, apply, load ✔
- Dead code removed (evalNodeImport, evalNodeExport) ✔

## v0.2.6 (Parity)

### EBNF Migration
- Lint rewritten with EBNF AST ✔
- Export statement verified ✔
- Import statement verified ✔
- Strict exports parses from strict.ai ✔

### Rich Errors: Phase 1
- `makeError(env, node, ...)` with file:line:source
- Stack traces with `from` lines
- HAL errors annotated with call-site position
- `InspectAtLayer()` for filtered display by layer
- `<main>` frame pushed at entry points
- Layer system (user/strict/hal)
- Tests in `tests/error_test.go`

## v0.2.5

### Architecture
- EBNF grammar-driven parser (grammar.ebnf is source of truth)
- Lexer/parser/fmt/lint all derive from grammar

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
