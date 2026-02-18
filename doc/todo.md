# Aiki TODO

## Validation (Do First)

```bash
go test ./...
./aiki examples/canvas.ai
./aiki -e 'print(sum(range(1, 10)))'
```

## Alpha Blockers (Ordered by Execution Priority)

### 1. Cleanup: Terminology + Dead Code
- [x] Delete `pragmatic/` directory
- [x] Remove `LayerPragmatic` constant
- [ ] REPL continuation: detect incomplete input (trailing |>, open delimiters) and prompt for more instead of parsing
- [ ] `LayerStrict` → `LayerPrelude`
- [ ] `SnapshotStrict()` → `SnapshotPrelude()`
- [ ] `shadowed = "strict"` → `shadowed = "prelude"`
- [ ] `loadStrict` → `loadPrelude`
- [X] Remove "used by pragmatic" comment (line 7837)
- [X] style.md: "Strict vs pragmatic" → "Exact vs precise"

### 2. Grammar Cleanup
- [x] Remove `%` from OPERATOR and BINOP
- [x] Remove `==` from OPERATOR and BINOP
- [x] Remove `!=` from OPERATOR and BINOP
- [ ] Remove `not` from KEYWORD
- [ ] Remove `not` from unary_expr (becomes function call)
- [ ] Add `_not` to HAL
- [ ] Add `not` to prelude: `let not = _not`
- [ ] Add `module` to KEYWORD
- [ ] Add `module_decl = "module" NAME [ "precise" ]` to grammar
- [ ] `precise` is contextual, not a keyword

### 3. HAL/Prelude + Operators
- [ ] Rename HAL primitives to `_` prefix (`_add`, `_first`, `_print`, etc.)
- [ ] HAL invisible to user (not in scope)
- [ ] Prelude wraps HAL (`let + = _add`, `let first = _first`, etc.)
- [ ] Operators become function lookups in exact mode
- [ ] Intrinsics (`apply`, `load`, `import`, `export`, `spawn`) unshadowable — error on attempt

### 4. Lint Rules (pairs with #3)
- [ ] Warn on shadowing prelude
- [ ] Error on shadowing intrinsics
- [ ] Case locked on first use
- [ ] `_prefix` export warning

### 5. Type Cleanup
- [ ] `Handle` struct → `[@handle, id]` shaped list
- [ ] `Canvas` struct → `[@canvas, id]` shaped list
- [ ] `Channel` struct → `[@channel, id]` shaped list
- [ ] HAL registries map ids to Go resources
- [ ] Remove `HandleType`, `CanvasType`, `ChannelType` from value types
- [ ] Shape is claim, registry is authority

### 6. Precise Mode
- [ ] `module X precise` parsed from grammar
- [ ] Precise mode skips operator lookup — hardcoded float ops
- [ ] `sin`, `cos`, `sqrt`, `random` only available in precise modules
- [ ] Fast path for numeric-heavy code

### 7. Canvas/Ebiten
- [ ] Fork-on-start for Linux/Mac
- [ ] Parent shepherds child process
- [ ] Child runs REPL with normal terminal I/O
- [ ] `canvas()` opens Ebiten window
- [ ] `destroy()` or window close exits child
- [ ] Parent forks new child — fresh env, no state handoff
- [ ] Windows: warn "canvas limited to one per session", exec normally

### 8. TCO
- [ ] Tail call optimization — trampoline pattern
---

## Alpha Polish

### Documentation
- [ ] Document shadowing rules
- [ ] Document exact vs precise modes

### Tooling
- [ ] `aiki version` as subcommand
- [ ] `aiki validate` — strict check, no execution
- [ ] `aiki --proto` flag for loose mode
- [ ] `--errors=user|prelude|hal` flag for error depth
- [ ] Externalize comment definition into the EBNF grammar
Add a token rule for comments in the grammar and remove the hardcoded "#" check from the lexer. Lexer should treat comment tokens like any other token marked Skip or KeepComments.
- [ ] Remove keyword priority from lexer and encode as grammar disambiguation
Move keyword/operator precedence into the grammar ordering or a token classification field. Lexer should not impose literal before regex ordering; the grammar should define token precedence directly.

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
