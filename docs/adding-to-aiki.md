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
    
    if num.Val.Sign() < 0 {
        return &value.Number{Val: new(big.Rat).Neg(num.Val)}
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

Example: creating a `string` package.

### 1. Create Package File

Create `lib/string/string.ai`:

```
package "string"

let upper = (s) { _str_upper(s) }
let lower = (s) { _str_lower(s) }
let split = (s, delim) { _str_split(s, delim) }
let join = (list, delim) { _str_join(list, delim) }

export(:upper, :lower, :split, :join)
```

### 2. HAL Support (if needed)

If functions need Go implementation, add to `engine/runtime/hal/substrate/`:

```go
// builtins_string.go
func halStrUpper(args []value.Value, ctx *hal.EvalContext) value.Value {
    // ...
}
```

Register in `register.go`.

### 3. Help File

Create `lib/string/string.help`:

```
@func upper
@template "upper(s)"
@help "Returns string in uppercase."

@func lower
@template "lower(s)"
@help "Returns string in lowercase."
```

### 4. Doc File

Create `lib/string/string.doc`:

```
upper
Returns a new string with all characters in uppercase.

upper(s)

upper("hello")    # "HELLO"
upper("Hi There") # "HI THERE"
===
lower
Returns a new string with all characters in lowercase.

lower(s)

lower("HELLO")    # "hello"
===
```

### 5. Usage

```
import("strings", :upper, :lower)
print(upper("hello"))
```

Or with module:

```
let str = import("strings")
print(str.upper("hello"))
```

## Validation

The system enforces consistency:

1. **Grammar ↔ Evaluator**: Every production must have a handler. Startup panics if missing.

2. **Grammar ↔ grammar.help**: Every production/token must have a help entry. Load fails if mismatch.

3. **Prelude ↔ prelude.help**: Every exported function must have help. Load fails if mismatch.

4. **Package ↔ export**: Imports only see exported names. Import fails if name not exported.

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
    exact.ai           # Exact math package
    inexact.ai         # Float math package
tests/
  smoke/               # Smoke tests
```

