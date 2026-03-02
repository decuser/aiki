# Aiki Questions

Unresolved design decisions. Each needs a decision before proceeding.

## Types

**Should handles (canvas, channel, file) be opaque types or shaped lists?**
- Opaque: `type(c)` → `:canvas`, can't inspect internals
- Shaped: `[@canvas, id]`, consistent with everything-is-a-list
- Current: opaque types
- Concern: shaped means id leaks to user, what would they do with it?

**Should Bytes exist as a type or only as shaped list?**
- Decided: both - `bytes/pure` uses `[@bytes, ...]`, `bytes/pragmatic` uses opaque type
- Open: is this the right split?

## Modules

**How should stdlib versioning work?**
- User writes `import("math/pure", :sin)`
- Aiki updates, `sin` signature changes
- User code breaks
- Options: version in package name? Stability guarantees? Don't care for now?

**Should prelude have an export statement?**
- Prelude is special - user env encloses it
- Adding export would be consistent
- But prelude isn't imported, it's always there
- Decision: leave as-is or add for consistency?

## Syntax

**Should `not` be a keyword or function?**
- Backlog says: remove from keyword, add to HAL, wrap in prelude
- Implication: `not(x)` instead of `not x`
- Consistent with "operators are functions"
- But reads worse: `if not(empty(list))` vs `if not empty(list)`

**Should there be a `module` keyword separate from `package`?**
- Backlog mentions `module X precise`
- Current: `package "name"`
- Is `module` for something else? Precise mode declaration?

## Semantics

**What is precise mode?**
- Mentioned in backlog and todo
- Seems related to numeric semantics
- Float fast path? Gating numeric functions?
- Needs definition before implementation

**Should operators become function lookups?**
- `x + y` becomes `add(x, y)` lookup in env
- Enables overloading? Or just consistency?
- Performance implications?
- Only in "exact mode"? What's exact mode vs precise mode?

## Runtime

**Canvas process model - fork vs single session?**
- Ebiten requires one RunGame per process
- Fork model: parent shepherds, child runs REPL
- Windows limitation: single session only
- Needs design for cross-platform behavior

**TCO - trampoline or other approach?**
- Deep recursion currently blows stack
- Trampoline is standard approach
- Alternative: detect tail position, reuse frame
- Priority vs other work?

## Tooling

**Should fmt/lint be pure aiki or stay in Go?**
- Backlog mentions moving to pure aiki
- Requires: parse returning AST as shaped lists
- Requires: tooling infrastructure
- Chicken-egg: need tools to build tools

## Process

**What's the bar for alpha?**
- Smoke tests pass?
- Module system complete?
- Help system complete?
- Canvas works?
- All of the above?

**What docs are actually needed?**
- constraints.md - done
- adding-to-aiki.md - done
- Language guide for users?
- Reference for stdlib?
- When to write vs when to wait for stability?
