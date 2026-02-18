# Aiki Decisions


## 2026-02-17: Grammar Loading and Package Boundaries

### Grammar

- The canonical grammar is embedded in the binary.
- GetGrammar is the sole entry point for the canonical grammar.
- Tools must use GetGrammar and must not load grammar from filesystem.

### Boundaries

- syntax is structure only.
- semantics is meaning only and operates on AST.
- runtime is capability only and has no structural authority.
- resolver is location only and is not in the semantic pipeline.
- tools are projections and contain no language rules.

### Tests

- Integration tests live in semantics integration.
- Integration tests use integration_test package to avoid import cycles.
Architectural decisions with context. Why we did what we did.

## 2026-02-15: Architecture Overhaul

### Layers

Two layers, not three:

| Layer | What | Visible | Shadowable |
|-------|------|---------|------------|
| **HAL** | `_add`, `_first`, `_print` — Go primitives | No | No (not in scope) |
| **Prelude** | `+`, `first`, `print` — Aiki wrappers | Yes | Yes, warn |

- HAL is underground, `_`-prefixed, invisible to user
- Prelude wraps HAL, is the user-facing API
- User code sits on prelude
- `_` is convention, not enforcement — user can define `_foo`
- Pragmatic layer eliminated
- Strict renamed to prelude (Forth heritage — the dictionary you start with)

### Intrinsics

| Name | What | Shadowable |
|------|------|------------|
| `apply`, `load`, `import`, `export` | Evaluator mechanics, need AST/env access | No, error |

Not a layer. Evaluator internals that are callable values. User sees them as primitives. Can't shadow because they can't be reimplemented in Aiki.

### Operators as Functions

- `+`, `-`, `*`, `/`, `<`, `>`, etc. become prelude functions wrapping HAL
- `let + = _add` in prelude
- Shadowable in exact mode (lookup cost)
- Hardcoded in precise mode (fast, not shadowable)
- Scheme-ish flexibility

### Exact vs Precise

| Term | Meaning |
|------|---------|
| **exact** | Rationals, default, shadowable operators, lookup cost |
| **precise** | Floats, `module X precise` declaration, hardcoded ops, fast |
| **inexact** | Irrationals, symbolic — future, not current scope |

- Literals are potential — functions realize them
- Module declaration controls realization
- Canvas/HAL is always fast (it's Go)

### Grammar Cleanup ✔

Removed:
- `%` → `modulo()` function
- `==` → `equal()` function  
- `!=` → `not(equal())`

To add (with precise mode):
- `module` keyword
- `module_decl = "module" NAME [ "precise" ]`
- `precise` is contextual, not a keyword

Keywords: 9 currently (`let if else while match return true false not`)
- `and`, `or` are operators
- `module` added when precise mode lands

### Type Cleanup

User-visible: 8 (unchanged)
- number, boolean, rune, string, bytes, symbol, list, function

Resources become shaped lists:
- `Handle` → `[@handle, id]`
- `Canvas` → `[@canvas, id]`
- `Channel` → `[@channel, id]`

HAL maintains registries. Shape is claim, registry is authority.

### Grammar-Evaluator Coupling ✔

- Handler map, not switch statement
- `ValidateHandlers()` panics on missing handler
- Called from `SetNodeGrammar()` at startup
- Grammar changes → immediate failure until handler added
- No drift

### Canvas/Ebiten

- Fork-on-start for Linux/Mac
- Parent shepherds, child runs REPL
- `canvas()` opens window, `destroy()` exits child
- Parent forks new child — fresh env, no state handoff
- Windows: warn "canvas limited to one per session", exec normally
- Sidesteps Ebiten's one-RunGame-per-process constraint

### Shadowing Rules

| Category | Shadowable | Behavior |
|----------|------------|----------|
| HAL | No | Not in scope, can't name it |
| Prelude | Yes | Warn |
| Intrinsics | No | Error |
| User code | Yes | No warning |

### TCO

- Required in v0.2.1 design
- Silently dropped during v0.2.3 implementation  
- Now tracked as future work (trampoline pattern)

## Invariants

- Grammar is source of truth
- Evaluator must handle all productions or panic ✔
- Shape is claim, registry is authority
- HAL is invisible — prelude is the API
- Exact is default — precise is opt-in
- Intrinsics are unshadowable
- Two layers: HAL (Go) and Aiki (prelude + user)

## Historical Notes

### v0.2.1 → v0.2.2
- "The Way" added: success returns value, failure returns `[@error, reason]`

### v0.2.2 → v0.2.3
- Implementation begun
- TCO dropped (silently)
- Streams-as-functions → handle-based I/O

### v0.2.3 → v0.2.4
- Three-layer architecture introduced (later simplified)
- Bracket indexing replaced `list.0`

### v0.2.4 → v0.2.5
- EBNF grammar-driven architecture
- Rich error system

### v0.2.5 → v0.2.6
- Two-layer architecture (HAL/Prelude)
- Operators as functions
- Exact/precise terminology
- Grammar-evaluator coupling
- Canvas fork-on-start design

### v0.2.6 → v0.2.7
- Grammar cleanup: `%`, `==`, `!=` removed
- Grammar-evaluator coupling implemented
- Eval package reorganized (node.go, module.go, intrinsics.go)
