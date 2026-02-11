# Aiki Pre-Alpha Stabilization

Priority order by ease of integration.

---

## Core Cleanup

- [ ] 1. Remove `!`, `==`, `!=`, `%` from lexer/parser
- [ ] 2. Remove `from`, `use`, `export` keywords from lexer/parser
- [ ] 3. Add `modulo` (floored), `trunc`, `eq` (atoms only) to hal
- [ ] 4. Add `import()`, `export()` functions to hal

## Type Consolidation

- [ ] 5. Remove Handle, Canvas, Channel as distinct types
- [ ] 6. Implement shaped list resources:
  - [ ] `[@handle, id]` for file handles
  - [ ] `[@canvas, id]` for canvas windows
  - [ ] `[@channel, id]` for channels
- [ ] 7. Add hal registries mapping ids to Go resources

## Renaming

- [ ] 8. Rename `builtins` → `hal`
- [ ] 9. Rename `prelude` → `strict`

## Strict Layer

- [ ] 10. Move `range`, `sum`, `max`, `min`, `reverse`, `find`, `any`, `all` to strict
- [ ] 11. Add `equal` (deep structural, uses `hal.eq`) to strict

## Concurrency

- [ ] 12. Enforce spawn isolation (orphan scopes, no closure capture)

## Canvas

- [ ] 13. Add `mouse_pos`, `mouse_down` to hal
- [ ] 14. Add `present` to hal (flush to GPU)
- [ ] 15. Add drawing algorithms to strict:
  - [ ] `line` (Bresenham)
  - [ ] `rect`, `fill_rect`
  - [ ] `circle`, `fill_circle` (midpoint)
  - [ ] `poly`, `fill_poly` (scanline)
  - [ ] `text` (bitmap font)
  - [ ] `blit`, `blit_sub`

## Pragmatic Module

- [ ] 16. Create `pragmatic` module scaffold
- [ ] 17. Add `float`, `exact`, `array`, `tolist` to pragmatic
- [ ] 18. Move `sin`, `cos`, `tan`, `sqrt`, `random` to pragmatic
- [ ] 19. Add vectorized ops (`array + array`, `array * scalar`, etc.)
- [ ] 20. Add pragmatic twins:
  - [ ] `map`, `filter`, `reduce`, `each`
  - [ ] `sum`, `mean`, `min`, `max`
- [ ] 21. Add pragmatic drawing twins (array-backed)

## Modules

- [ ] 22. Create `regex` module (`match`, `find`, `replace`)

## Profiling Foundation

- [ ] 23. Add log/trace infrastructure:
  - [ ] Segmented ring buffer
  - [ ] `^t` / SIGUSR1 snapshot
  - [ ] Terminal stream toggle
  - [ ] Filter support

---

## Deferred (Post-Alpha)

- [ ] Monitor visual interface (Ebiten dashboard, soft-keys, sparklines)
- [ ] Debugger (breakpoints, stepping)

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
