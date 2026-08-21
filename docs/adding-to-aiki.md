# Adding to Aiki

This guide describes how to extend the Aiki language at each layer.

## Architecture Overview

```
┌─────────────────────────────────────────────┐
│  User Code (.ai files)                      │
├─────────────────────────────────────────────┤
│  Lib (lib/*.ai)                             │
│  - Packages written in Aiki                 │
│  - Require: package, export                 │
├─────────────────────────────────────────────┤
│  Prelude (engine/runtime/prelude/)          │
│  - Aiki wrappers around HAL                 │
│  - Always available, no import needed       │
├─────────────────────────────────────────────┤
│  HAL - Hardware Abstraction Layer           │
│  (engine/runtime/hal/substrate/)            │
│  - Go implementations of primitives         │
│  - Names start with underscore: _print      │
├─────────────────────────────────────────────┤
│  Evaluator (engine/semantics/evaluator/)    │
│  - Interprets AST nodes                     │
│  - One handler per grammar production       │
├─────────────────────────────────────────────┤
│  Parser/Lexer (engine/syntax/)              │
│  - Grammar-driven                           │
│  - EBNFX is source of truth                 │
└─────────────────────────────────────────────┘
```
## Executable Invariants

Aiki relies on executable couplings among independently maintained artifacts to
detect implementation and documentation drift.

Eleven couplings keep the distribution self-consistent:

1. grammar <-> evaluator handlers
2. prelude <-> help
3. formatted source <-> AST
4. runtime module discovery <-> shipped modules
5. observable behavior <-> gold files
6. library exports <-> help and documentation
7. language core <-> graphics boundary
8. canvas behavior <-> transcript golds
9. documentation examples <-> stated values
10. documentation entries <-> checked-or-effects disposition
11. shipped modules <-> documentation presence

These checks are intended to compare independent artifacts rather than allow
one artifact to certify another derived from the same assumption.

When extending Aiki, preserve the relevant existing coupling. If a change
introduces a new independently maintained relationship, add an invariant for
that relationship rather than relying on manual review.## Adding a Builtin Function

Example: adding `abs(n)` to return absolute value.

### 1. HAL Implementation

Create or edit a file in `engine/runtime/hal/substrate/`:

```go
// builtins_math.go
func halAbs(args []value.Value, ctx *hal.EvalContext) value.Value {
    if len(args) != 1 {
        return value.NewError("abs: want 1 argument, got %d", len(args))
    }
    
    num, ok := args[0].(*value.Number)
    if !ok {
        return value.NewError("abs: expected number, got %s", args[0].Type())
    }
    
    if num.Sign() < 0 {
        return num.Neg()
    }
    return num
}
```

### 2. Register the HAL function

In `engine/runtime/hal/substrate/register.go`:

```go
// Math
g.register("_abs", halAbs)
```

### 3. Prelude Wrapper

In `engine/runtime/prelude/prelude.ai`:

```
let abs = (n) { _abs(n) }
```

### 4. Help Entry

In `engine/runtime/prelude/prelude.help`:

```
@func abs
@template "abs(n)"
@help "Returns absolute value of number."
```

### 5. Doc Entry

In `engine/runtime/prelude/prelude.doc`:

```
abs
Returns the absolute value of a number.

abs(n)

abs(-5)    # 5
abs(3)     # 3
abs(0)     # 0
===
```

### 6. Test

Create `test/behavior/abs_smoke.ai`:

```
println(abs(-5))
println(abs(5))
println(abs(0))
println("ABS_SMOKE_DONE")
```

Create `test/behavior/abs_smoke.gold`:

```
OUT:5
OUT:5
OUT:0
OUT:ABS_SMOKE_DONE
```

The `.gold` file format:
- `OUT:` - expected stdout
- `IN:` - simulated stdin
- `ERR:` - expected stderr
- `EXIT:` - expected exit code (default 0)

### 7. Verify

```bash
go build ./cmd/aiki
./aiki smoke
```

## Adding New Syntax

Example: adding `unless` as opposite of `if`.

### 1. Grammar Production

In `engine/syntax/grammar.ebnfx`:

Add to statement alternatives:
```
statement = package_stmt | let_stmt | assign_stmt | if_stmt | unless_stmt | ...
```

Add the production:
```
unless_stmt = "unless" expr block [ "else" block ]
    @error "expected 'unless condition { ... }'"
    @template "unless cond { body } | unless cond { body } else { alt }"
    @help "executes body if condition is false"
```

Add keyword:
```
KEYWORD     let if else while match return true false and or not package unless
```

### 2. Grammar Help

In `engine/syntax/grammar.help`:

```
unless_stmt

Executes body if condition is false. Opposite of if.

unless empty(list) {
    print(first(list))
}

unless found {
    print("not found")
} else {
    print("found it")
}
===
```

### 3. Handler Registration

In `engine/semantics/evaluator/evaluator.go`:

```go
var handlers = map[string]handlerFunc{
    // ...
    "unless_stmt": (*Evaluator).evalUnless,
    // ...
}
```

### 4. Implementation

In `engine/semantics/evaluator/statements.go`:

```go
func (e *Evaluator) evalUnless(node *syntax.Node, env *value.Env) value.Value {
    var cond value.Value
    var thenBlock, elseBlock *syntax.Node
    
    for _, child := range node.Children {
        switch child.Type {
        case "expr":
            cond = e.Eval(child, env)
        case "block":
            if thenBlock == nil {
                thenBlock = child
            } else {
                elseBlock = child
            }
        }
    }
    
    if value.IsError(cond) {
        return cond
    }
    
    // unless = if NOT condition
    if !value.IsTruthy(cond) {
        return e.Eval(thenBlock, env)
    } else if elseBlock != nil {
        return e.Eval(elseBlock, env)
    }
    
    return value.EMPTY
}
```

### 5. Test

Create `test/behavior/unless_smoke.ai`:

```
let x = 5
unless x > 10 {
    println("small")
}
unless x < 10 {
    println("big")
} else {
    println("also small")
}
println("UNLESS_SMOKE_DONE")
```

Create `test/behavior/unless_smoke.gold`:

```
OUT:small
OUT:also small
OUT:UNLESS_SMOKE_DONE
```

### 6. Verify

```bash
go build ./cmd/aiki
./aiki smoke
```

## Adding a Library Package

Before adding a package, decide what semantic role it has and record that role in
`engine/runtime/modules/stdlib_policy.go`.

Portable semantics and host/runtime capabilities are different contracts:

- a portable `X/native` module implements the public library semantics in Aiki;
- an optional `X/ffi` module is provider-backed and must preserve the native
  export, signature, help/doc, and behavior contract;
- host/runtime capabilities and deliberate provider interop do not require fake
  native twins, but their role must be declared honestly;
- bare `X` resolves to `X/native` when a native module is present and no explicit
  bare package overrides that default.

### 1. Add the Native Semantic Authority

For a portable package, create `lib/example/native.ai`:

```
package "example/native"

let twice = (n) {
    n + n
}

export(:twice)
```

A `/native` package may use constitutive Aiki runtime atoms exposed by the
prelude, but it must not import `/ffi` modules or call provider-role primitives
to implement its library algorithm.

Add matching `native.help` and `native.doc` files and tests against the bare
`example` import so the default path exercises the native implementation.

### 2. Add an FFI Acceleration Only if Needed

If a coarse provider boundary materially helps, add `lib/example/ffi.ai`:

```
package "example/ffi"

let twice = (n) {
    _example_twice(n)
}

export(:twice)
```

Register `_example_twice` as a provider primitive and grant it only to the FFI
source through HAL authority policy. Declare `example/ffi` as portable/ffi with
`example/native` as its semantic authority.

The acceleration must expose the same public function names and callable shapes
as native. Help/doc surface must match, and the same behavior contract should be
run against both realizations. Provider-specific semantics belong in an `interop`
module instead of being disguised as acceleration.

### 3. Add a Capability or Interop Module Honestly

If the behavior intrinsically depends on a host resource (files, processes,
terminal state, clocks, etc.), classify it as a host/runtime capability and use
HAL authority directly as needed. If the purpose is to expose a provider's own
semantics, classify it as interop. Neither case needs a fake pure-Aiki twin.

### 4. Verify

Run the relevant focused tests while cutting the change, then the repository
validation gate. The stdlib policy invariants check declaration completeness,
truthful `/native` and `/ffi` naming, native-path purity, acceleration surface
parity, and help/doc parity.

## Validation

The system enforces consistency through executable couplings rather than relying
on this document as authority. Among them:

1. grammar productions ↔ evaluator handlers and structural engine coverage;
2. grammar/prelude/library exports ↔ help and documentation surfaces;
3. formatted source ↔ parsed structure;
4. shipped modules ↔ registry discovery and distribution-tree relationships;
5. behavior ↔ blessed smoke transcripts;
6. HAL identities ↔ substrate registration, provenance, and exact authority;
7. stdlib semantic policy ↔ truthful `/native`, `/ffi`, capability, interop, and
   bare-default behavior.

Use focused checks while cutting a change, then `make validate` for the normal
repository gate. Use `make rigorous` when the work warrants the stronger gate.

## File Locations

```
engine/
  syntax/
    grammar.ebnfx      # Grammar source of truth
    grammar.help       # Grammar documentation
  semantics/
    evaluator/         # AST interpretation
    value/             # Value types
  runtime/
    hal/substrate/     # Go primitives
    prelude/
      prelude.ai       # Aiki wrappers
      prelude.help     # Quick reference
      prelude.doc      # Full documentation
lib/
  math/
    native.ai          # Portable Aiki semantic authority
    ffi.ai             # Provider-backed realization where declared
test/
  behavior/            # Behavior smoke specimens and golds
```

