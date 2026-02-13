# Aiki TODO

## Invariants

- Shape is claim, registry is authority. Tagged lists like `[@canvas, id]` are handles, not permission. The host registry validates on use—same pattern as bounds checking.
- Canonical (strict until we rename it) is specification. Pragmatic must pass Canonical equivalence tests. Never define behavior in Pragmatic first.

## Alpha Checklist

### Core Cleanup
- [ ] Remove `==`, `!=`, `%` from lexer/parser
- [ ] Remove `from`, `use`, `export` keywords
- [ ] Add `modulo`, `trunc`, `eq` (atoms only) to hal
- [ ] Add `import()`, `export()` functions to hal

### Type Consolidation
- [ ] Resources as shaped lists: `[@handle, id]`, `[@canvas, id]`, `[@channel, id]`
- [ ] Hal registries mapping ids to Go resources

### Layers
- [ ] Move `equal` to strict (uses `hal.eq`)
- [ ] Move drawing algorithms to strict (Bresenham, midpoint, scanline)
- [ ] Move `sin`, `cos`, `sqrt`, `random` to pragmatic
- [ ] Create pragmatic module (float, array, vectorized ops)

### Concurrency
- [ ] Enforce spawn isolation (no closure capture)

### Loose Ends
- [ ] Update help() text with new primitives
- [ ] Run full test suite
- [ ] Test all examples

## Future

- [ ] `fmt --rules` and `lint --rules` to print enforced rules
- [ ] Regex module
- [ ] Multi-line REPL
- [ ] Debugger
- [ ] Bytecode VM
- [ ] LSP
- [ ] WASM target
