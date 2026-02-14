# Aiki TODO

## Parity: COMPLETE

1. ~~**lint** — rewrite with EBNF AST~~ ✓
2. ~~**export statement** — verified~~ ✓
3. ~~**import statement** — verified~~ ✓
4. ~~**strict exports** — parses from strict.ai~~ ✓
5. **tests** — run full suite to confirm
   ```bash
   go test ./...
   ./aiki examples/canvas.ai
   ./aiki -e 'print(sum(range(1, 10)))'
   ```

## Rich Errors: Phase 1 COMPLETE

- ✓ `makeError(env, node, ...)` throughout eval
- ✓ HAL errors annotated with call-site position
- ✓ Stack traces with `from` lines
- ✓ Layer system (user/strict/pragmatic/hal)
- ✓ `InspectAtLayer()` for filtered display
- ✓ Tests in `tests/error_test.go`

## Rich Errors: Phase 2 (Future)

- [ ] Error template registry (fn + code → template)
- [ ] HAL registers templates at init
- [ ] `@error` shape in grammar for strict/user layers
- [ ] Template interpolation (`{key}` → value)

## Invariants

- Shape is claim, registry is authority
- Canonical (strict) is specification
- Help is projection, not explanation — derived from grammar
- Grammar is infrastructure, language is client
- Errors are projection — templates from grammar/layers

### Error System (add after Invariants, before Alpha Checklist)
- [ ] Error template registry (fn + code → template)
- [ ] HAL registers templates at init
- [ ] `@error` shape in grammar for strict/user layers
- [ ] Template interpolation (`{key}` → value)
- [ ] `--errors=user|strict|hal` flag for error depth

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
- [ ] `--errors=user|strict|hal` flag for error depth

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
- [ ] fmt --rules and lint --rules to print enforced rules

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

### Validation
- [ ] `make build && make test && make fmt && make lint`
- [ ] `./aiki -v`
- [ ] `./aiki` (REPL starts)
- [ ] `./aiki examples/canvas.ai`
- [ ] `./aiki examples/pipeline.ai`


## Future
- [ ] Match pinning (`^name` to match against existing variable's value)
- [ ] Match guards (`pattern if condition { ... }`)
- [ ] Regex module
- [ ] Multi-line REPL
- [ ] Debugger
- [ ] Bytecode VM
- [ ] LSP
- [ ] WASM target
- [ ] `aiki build` — validate + produce artifact
- [ ] `aiki profile` — performance analysis
