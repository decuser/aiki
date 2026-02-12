# Aiki TODO

## Done (v0.2.3)

- [x] Lexer (state machine, UTF-8)
- [x] Parser (recursive descent, left-to-right)
- [x] Evaluator (tree-walking)
- [x] 8 types (number, boolean, rune, string, bytes, symbol, list, function)
- [x] Rational arithmetic (exact, via big.Rat)
- [x] Shapes with composition
- [x] Pattern matching
- [x] Pipe operator with error short-circuit
- [x] Pipe auto-unwrap `[@ok, val]`
- [x] Prelude (map, filter, reduce, range, hash map)
- [x] File I/O (open, create, fread, fwrite, fclose)
- [x] REPL with readline
- [x] File runner
- [x] Formatter with comment preservation
- [x] Comma separators (calls, lists, params, patterns)
- [x] Ruby-style error messages with source line
- [x] Canvas primitives (Ebiten)
- [x] Concurrency (spawn, channel, send, recv)
- [~] Test suite

---

## Done (v0.2.4)

### Syntax
- [x] Bracket indexing `list[i]`, `string[i]`
- [x] Dot access for fields only `point.x` (no `list.0`)
- [x] Rest parameters `(...args)`, `(a, b, ...rest)`
- [x] Semicolon statement separator
- [x] GroupExpression preserves parentheses in formatter

### Primitives
- [x] `ord(rune)` - rune to code point
- [x] `apply(fn, list)` - spread list as function args
- [x] `shell(cmd)` - run shell command, returns `[@ok, output]` or `[@error, msg]`
- [x] Remove `nth` (replaced by bracket indexing)

### CLI
- [x] `-e` flag for one-liner expressions
- [x] `quit()` actually exits REPL (ExitSignal)

### Internals
- [x] HAL is single source of truth for builtins
- [x] Remove `BuiltinNames` map from eval.go
- [x] Specials (`apply`, `load`) use `Fn: nil` in HAL
- [x] Parser handles `f()[0]`, `f().field` chaining
- [x] Formatter handles IndexExpression, RestParam, GroupExpression
- [x] Lint handles RestParam in function scope

### Fixes
- [x] Remove blank line on Ctrl-C
- [x] Remove "exit" echo on Ctrl-D (EOFPrompt)
- [x] strict.ai exports list added
- [x] pragmatic.ai `nth()` → bracket indexing

---

## Codebase Cleanup

- [x] Rename `builtins` → `hal`
- [x] Rename `prelude` → `strict`
- [x] REPL subcommand pattern (`cmd/repl/` with `builtin.go`, `cmd.go`, `session.go`)
- [x] Extract version package (`version/version.go`)
- [x] TrackingWriter replaces global state
- [x] Move REPL builtins (`reset`, `help`, `quit`) to `cmd/repl/builtin.go`

---

## Pending (v0.2.4)

- [ ] Fix `^D` literal echo (readline issue)
- [ ] Update help() text with new primitives
- [ ] Formatter adds blank line before final statement (cosmetic)
- [ ] Formatter collapses export to one line (cosmetic)
- [ ] Run full test suite
- [ ] Test all examples
- [ ] Version bump to v0.2.4
- [ ] Update changelog

---

## Pre-Alpha Stabilization

### Core Cleanup

- [ ] 1. Remove `!`, `==`, `!=`, `%` from lexer/parser
- [ ] 2. Remove `from`, `use`, `export` keywords from lexer/parser
- [ ] 3. Add `modulo` (floored), `trunc`, `eq` (atoms only) to hal
- [ ] 4. Add `import()`, `export()` functions to hal

### Type Consolidation

- [ ] 5. Remove Handle, Canvas, Channel as distinct types
- [ ] 6. Implement shaped list resources:
  - [ ] `[@handle, id]` for file handles
  - [ ] `[@canvas, id]` for canvas windows
  - [ ] `[@channel, id]` for channels
- [ ] 7. Add hal registries mapping ids to Go resources

### Strict Layer

- [ ] 8. Move `range`, `sum`, `max`, `min`, `reverse`, `find`, `any`, `all` to strict
- [ ] 9. Add `equal` (deep structural, uses `hal.eq`) to strict

### Concurrency

- [ ] 10. Enforce spawn isolation (orphan scopes, no closure capture)

### Canvas

- [ ] 11. Add `mouse_pos`, `mouse_down` to hal
- [ ] 12. Add `present` to hal (flush to GPU)
- [ ] 13. Add drawing algorithms to strict:
  - [ ] `line` (Bresenham)
  - [ ] `rect`, `fill_rect`
  - [ ] `circle`, `fill_circle` (midpoint)
  - [ ] `poly`, `fill_poly` (scanline)
  - [ ] `text` (bitmap font)
  - [ ] `blit`, `blit_sub`

### Pragmatic Module

- [ ] 14. Create `pragmatic` module scaffold
- [ ] 15. Add `float`, `exact`, `array`, `tolist` to pragmatic
- [ ] 16. Move `sin`, `cos`, `tan`, `sqrt`, `random` to pragmatic
- [ ] 17. Add vectorized ops (`array + array`, `array * scalar`, etc.)
- [ ] 18. Add pragmatic twins:
  - [ ] `map`, `filter`, `reduce`, `each`
  - [ ] `sum`, `mean`, `min`, `max`
- [ ] 19. Add pragmatic drawing twins (array-backed)

### Modules

- [ ] 20. Create `regex` module (`match`, `find`, `replace`)

### Profiling Foundation

- [ ] 21. Add log/trace infrastructure:
  - [ ] Segmented ring buffer
  - [ ] `^t` / SIGUSR1 snapshot
  - [ ] Terminal stream toggle
  - [ ] Filter support

---

## Future

- [ ] Debugger (breakpoints, stepping)
- [ ] Monitor visual interface (Ebiten dashboard, soft-keys, sparklines)
- [ ] Bit primitives
- [ ] Bytecode VM
- [ ] Multi-line REPL
- [ ] LSP
- [ ] WASM target

---

## Architecture Reference

| Layer | What | Auto-load |
|-------|------|-----------|
| `hal` | Go primitives - memory, I/O, concurrency, canvas | Always |
| `strict` | Pure Aiki - lists, rationals, inspectable | Yes |
| `pragmatic` | Fast Aiki - arrays, floats, hardware speed | Opt-in |

**Keywords (11):** `let`, `if`, `else`, `while`, `match`, `return`, `true`, `false`, `and`, `or`, `not`

**Types (8):** Number, Boolean, String, Rune, Symbol, List, Function, Bytes

**The Twin Pattern:** `strict.X` is pure, `pragmatic.X` is fast. Same API, different speed.

**No shadowing.** Namespace separation: `strict.map` vs `pragmatic.map`
