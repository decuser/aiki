# Aiki Backlog

## Validation

- go test ./...
- ./aiki examples/canvas.ai
- ./aiki -e 'print(sum(range(1, 10)))'

## Alpha Blockers

### 1. Cleanup: Terminology + Dead Code
- Delete pragmatic directory
- Remove LayerPragmatic constant
- REPL continuation detection
- LayerStrict to LayerPrelude rename sweep
- SnapshotStrict to SnapshotPrelude
- shadowed = strict to prelude
- loadStrict to loadPrelude
- Style terminology alignment

### 2. Grammar Cleanup
- Replace filesystem based import resolution with resolver interface
- Remove not from keyword and unary_expr
- Add _not to HAL and wrap in prelude
- Add module keyword
- Add module_decl grammar production
- Make precise contextual

### 3. HAL and Prelude and Operators
- Rename HAL primitives with underscore prefix
- HAL invisible to user scope
- Prelude wraps HAL
- Operators become function lookups in exact mode
- Intrinsics unshadowable

### 4. Lint Rules
- Warn on shadowing prelude
- Error on shadowing intrinsics
- Case locking on first use
- Underscore export warning

### 5. Type Cleanup
- Handle struct to shaped list
- Canvas struct to shaped list
- Channel struct to shaped list
- HAL registries map ids to resources
- Remove dedicated type constants

### 6. Precise Mode
- module X precise grammar
- Float fast path
- Numeric function gating

### 7. Canvas and Ebiten
- Fork model for Linux and Mac
- Parent shepherd model
- Child runs REPL
- Windows single session limitation

### 8. TCO
- Trampoline tail call optimization

## Post Alpha

### Error System
- Error template registry
- Template interpolation

### Help System
- Help derived from grammar
- Help indexed by syntactic unit

### Subcommands as Pure Aiki
- Move fmt to pure Aiki
- Move lint to pure Aiki
- Plugin architecture

## Future

### Language
- Match pinning
- Match guards
- Inexact numeric regime

### Tooling
- Multi line REPL
- Debugger
- LSP

### Runtime
- Bytecode VM
- WASM target

### Homoiconicity
- parse returns AST as shaped lists
- unparse
- eval over shaped AST
- quote mechanism
- Tooling written in Aiki
