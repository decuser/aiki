# Aiki TODO

## Validation (Do First)

```bash
go test ./...
./aiki examples/canvas.ai
./aiki -e 'print(sum(range(1, 10)))'
```

## Alpha Blockers (Ordered)

### 1. Grammar Cleanup
- [x] Remove `%` from OPERATOR and BINOP
- [x] Remove `==` from OPERATOR and BINOP
- [x] Remove `!=` from OPERATOR and BINOP
- [ ] Add `module` to KEYWORD
- [ ] Add `module_decl = "module" NAME [ "precise" ]` to grammar
- [ ] `precise` is contextual, not a keyword

### 2. Grammar-Evaluator Coupling
- [x] Replace switch statement with handler map
- [x] `init()` panics on missing handler for any production
- [x] Test verifies all grammar productions have handlers
- [x] No drift possible — grammar change forces handler update

### 3. HAL/Prelude + Operators
- [ ] Rename HAL primitives to `_` prefix (`_add`, `_first`, `_print`, etc.)
- [ ] HAL invisible to user (not in scope)
- [ ] Prelude wraps HAL (`let + = _add`, `let first = _first`, etc.)
- [ ] Operators become function lookups in exact mode
- [ ] Intrinsics (`apply`, `load`, `import`, `export`) unshadowable — error on attempt
- [x] Rename `strict/` directory to `prelude/`
- [x] Rename `strict.ai` to `prelude.ai`
- [ ] Update terminology (prelude not strict, no pragmatic)
- [ ] Delete `pragmatic/` directory

### 4. Type Cleanup
- [ ] `Handle` struct → `[@handle, id]` shaped list
- [ ] `Canvas` struct → `[@canvas, id]` shaped list
- [ ] `Channel` struct → `[@channel, id]` shaped list
- [ ] HAL registries map ids to Go resources
- [ ] Remove `HandleType`, `CanvasType`, `ChannelType` from value types
- [ ] Shape is claim, registry is authority

### 5. Precise Mode
- [ ] `module X precise` parsed from grammar
- [ ] Precise mode skips operator lookup — hardcoded float ops
- [ ] `sin`, `cos`, `sqrt`, `random` only available in precise modules
- [ ] Fast path for numeric-heavy code

### 6. Canvas/Ebiten
- [ ] Fork-on-start for Linux/Mac
- [ ] Parent shepherds child process
- [ ] Child runs REPL with normal terminal I/O
- [ ] `canvas()` opens Ebiten window
- [ ] `destroy()` or window close exits child
- [ ] Parent forks new child — fresh env, no state handoff
- [ ] Windows: warn "canvas limited to one per session", exec normally

### 7. Tail Call Optimization
- [ ] TCO (tail call optimization) — trampoline pattern

## Alpha Polish

### Documentation
- [ ] Update design.md with new layers (HAL/Prelude)
- [ ] Document shadowing rules
- [ ] Document exact vs precise modes

### Tooling
- [ ] `aiki version` as subcommand
- [ ] `aiki validate` — strict check, no execution
- [ ] `aiki --proto` flag for loose mode
- [ ] `--errors=user|strict|hal` flag for error depth

### Lint Rules
- [ ] Warn on shadowing prelude
- [ ] Error on shadowing intrinsics
- [ ] Case locked on first use
- [ ] `_prefix` export warning

## Post-Alpha

### Error System (Phase 2)
- [ ] Error template registry
- [ ] HAL registers templates at init
- [ ] Template interpolation

### Help System
- [ ] Help derived from grammar
- [ ] Help indexed by syntactic unit

### Subcommands as Pure Aiki
- [ ] Move fmt to pure Aiki
- [ ] Move lint to pure Aiki
- [ ] Plugin architecture

## Future

### Language
- [ ] Match pinning (`^name`)
- [ ] Match guards (`pattern if condition`)
- [ ] `inexact` numeric regime — symbolic irrationals

### Tooling
- [ ] Multi-line REPL
- [ ] Debugger
- [ ] LSP

### Runtime
- [ ] Bytecode VM
- [ ] WASM target

### Homoiconicity

- [ ] `parse(string)` — returns AST as shaped lists
- [ ] Shape names match grammar production names (e.g., `[@if_stmt, ...]`, `[@let_stmt, ...]`)
- [ ] `unparse(ast)` — returns source string from shaped AST
- [ ] `eval(ast)` — executes shaped AST
- [ ] Quote mechanism (`quote(expr)` returns shape, not value)
- [ ] AST manipulation in pure Aiki — tooling writes itself
