# Aiki TODO

## Invariants

- Shape is claim, registry is authority. Tagged lists like `[@canvas, id]` are handles, not permission. The host registry validates on use—same pattern as bounds checking.
- Canonical (strict) is specification. Pragmatic must pass Canonical equivalence tests. Never define behavior in Pragmatic first.
- Help is projection, not explanation. Derived from grammar + paradigm contexts + fmt rules + lint rules.
- Grammar is infrastructure, language is client. Aiki-grammar is the kernel; Aiki-lang is the first client.

## Alpha Checklist

### Core Cleanup
- [ ] Remove `==`, `!=`, `%` from lexer/parser
- [ ] Remove `from`, `export` keywords (use functions)
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

### Tooling
- [ ] `aiki version` as subcommand (not `-v` flag)
- [ ] `aiki validate` — strict check, no execution
- [ ] `aiki --proto` flag for loose mode
- [ ] `aiki clean` — remove generated cruft
- [ ] `aiki help [topic]` — help system entry point
- [ ] Auto-fmt in memory on `aiki run` (original positions for errors)

### Subcommands as Pure Aiki
- [ ] Move fmt to pure Aiki (register as subcommand)
- [ ] Move lint to pure Aiki (register as subcommand)
- [ ] Move validate to pure Aiki
- [ ] Move test to pure Aiki
- [ ] Move help to pure Aiki
- [ ] Plugin architecture: drop .ai file in, it registers

### Help System
- [ ] Help derived from grammar (mechanically)
- [ ] Help indexed by syntactic unit
- [ ] Paradigm contexts as source truth (recursive, iterative, functional, immediate)
- [ ] Paradigm contexts shadowable by user
- [ ] `fmt --rules` and `lint --rules` to print enforced rules

### Lint Rules
- [ ] Case locked on first use
- [ ] SCREAMING all-or-nothing
- [ ] `_prefix` export warning
- [ ] Shadow warning
- [ ] Unused = error (strict), warn (proto)
- [ ] Case propagates through imports

### Loose Ends
- [ ] Update help() text with new primitives
- [ ] Run full test suite
- [ ] Test all examples

## Future

- [ ] Regex module
- [ ] Multi-line REPL
- [ ] Debugger
- [ ] Bytecode VM
- [ ] LSP
- [ ] WASM target
- [ ] `aiki build` — validate + produce artifact
- [ ] `aiki test` — run tests
- [ ] `aiki profile` — performance analysis
