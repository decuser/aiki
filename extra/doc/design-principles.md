# Aiki Design Principles

Aiki is a minimal composable programming language with a live environment. Its design goal is to make language behavior inspectable enough to support learning.

Aiki treats a programming environment as an information system. Syntax, runtime behavior, errors, help, libraries, and examples all shape the learner’s proximity to the language.

## Principles

### Constraints over choice

Aiki intentionally limits the number of ways to express core operations. Fewer forms make behavior easier to inspect, explain, and remember.

### Explicit over implicit

Aiki avoids hidden behavior where possible. Grouping, return, mutation, imports, and error handling should be visible in the program text.

### Composition through naming

Programs are built from small named pieces that can be read, inspected, combined, and reused.

### Simplicity over ease

Aiki prefers simple rules over conveniences that hide complexity. The goal is not to make every expression shorter, but to make behavior legible.

### Exactness by default

Numbers use exact rational arithmetic in the core language. Floating point behavior is available through explicit library support.

### Inspectability enables learning

The prelude and libraries are part of the learning surface. Users should be able to ask what a function is, where it comes from, and how it behaves.

## Lineage

Aiki draws selectively from several programming traditions.

| Area | Influence | Use in Aiki |
|---|---|---|
| Syntax | BCPL, C, Go | Braces, small language feel, familiar surface |
| Data | Scheme | Lists as universal structure |
| Evaluation | Smalltalk | Left to right evaluation |
| Flow | ML, F# | Pipes and values for recoverable errors |
| Kernel | Forth | Small core with language level library surface |
| Feedback | BASIC, Logo | Immediate interaction and drawing |
| Host | Go | Runtime implementation and system interface |

## Types

Aiki uses a small set of core types:

| Type | Purpose |
|---|---|
| Number | Exact rational arithmetic |
| Boolean | Truth values |
| Rune | Unicode code point |
| String | Immutable rune sequence |
| Bytes | Immutable byte sequence for I/O |
| Symbol | Atomic identity value |
| List | Universal compound structure, raw or shaped |
| Function | First class callable value |

## Syntax Commitments

Aiki keeps syntax deliberately small.

| Commitment | Rationale |
|---|---|
| `let` creates bindings | Creation is explicit |
| `=` mutates existing bindings | Mutation is distinct from creation |
| Explicit `return` | Function exits are visible |
| One function syntax | No duplicate forms for functions |
| One loop form, `while` | Iteration can be built from while and functions |
| One conditional form, `if` and `else` | Branching remains uniform |
| No operator precedence | Evaluation order remains visible |
| Parentheses for grouping | Grouping is explicit |
| Pipes for composition | Data flow can be read left to right |
| Errors as values | Recoverable failure remains inspectable |

## Architecture Commitments

Aiki has two main implementation layers.

### HAL

The HAL provides the gateway to the Go runtime. HAL names are internal primitives and begin with an underscore.

### Aiki Layer

The prelude and libraries wrap HAL capabilities in Aiki code. This keeps much of the usable language surface inspectable from within Aiki itself.

### Grammar

The grammar is the source of truth. It is embedded in the binary and paired with grammar help.

### Evaluator

The evaluator uses one handler per grammar production. Startup validation checks that grammar productions and evaluator handlers remain aligned.

### Tools

Tools are projections over the language. They should not contain independent language rules.

## Supported Styles

Aiki is intentionally pre paradigmatic. Its primitives support several programming styles without committing the language to one dominant paradigm.

| Style | Support |
|---|---|
| Recursive | `first`, `rest`, functions |
| Iterative | `while` |
| Functional | `map`, `filter`, pipes |
| Imperative | Mutation with `=` |
| Immediate | REPL and drawing |

## Rejected Features

Aiki rejects some familiar features because they add hidden behavior or duplicate existing forms.

| Feature | Reason |
|---|---|
| Arrow lambdas | Second way to write functions |
| Implicit return | Function exits become less visible |
| `for` loop | `while` and functions cover iteration |
| `elif` | Guard clauses and nested conditionals cover branching |
| Operator precedence | Invisible ordering rule |
| Classes and methods | Data and behavior remain separate |
| Exceptions | Recoverable errors are values |
| Macros | The language surface remains closed |
| Default IEEE floats | Exact rationals are the default |
| Pointers | Pointer behavior breaks inspectability |

