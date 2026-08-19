# Aiki Alpha 1

Alpha 1 is the first public milestone of Aiki, a small experimental programming language implemented in Go and designed to make computational relationships visible.

This release establishes the language's core semantics, runtime model, standard library, tooling baseline, and the architectural conventions that later releases build on. 

Note: This document was created retrospectively at the time of the Alpha 2 release to describe the earlier Alpha 1 milestone.

## Language

Alpha 1 defines the small Aiki surface:

- exact rational numbers by default
- left-to-right binary evaluation, with parentheses for grouping
- first-class functions
- recursion and iteration
- pipelines and composition
- lists and shaped data
- symbols, strings, runes, booleans, and numbers
- `if` / `else`, `while`, `match`, and `return`
- falsehood limited to `false` and `[]`; all other values are true
- no classes or object hierarchy

The language is intentionally constrained. Its design favors explicit relationships, inspectability, and composition through naming over hidden machinery or large surface area.

## Runtime and modules

Alpha 1 includes:

- a Go-hosted interpreter
- a grammar-driven parser and evaluator
- a module system and standard library
- isolated spawned computations with message-passing concurrency
- channels with `spawn`, `channel`, `send`, and `recv`
- shaped values
- recoverable error values
- a HAL-backed substrate for runtime services

The standard library already includes general-purpose facilities for strings, lists, bytes, math, files, system information, storage, graphics, and turtle-style drawing.

## Tooling

The first alpha includes a practical command-line environment for working with Aiki programs, including:

- formatting
- parser/evaluator debugging
- executable help and documentation
- validation and gold-file checks
- semantic profiling
- examples and smoke tests
- build and distribution targets

The repository treats documentation, examples, grammar, and behavior checks as executable parts of the implementation rather than as separate descriptive artifacts.

## Grammar and implementation coupling

Aiki's grammar is coupled directly to the implementation.

The repository checks relationships such as:

- grammar productions to parser/evaluator handling
- formatter behavior to the AST
- modules to exports
- help and documentation to library affordances
- examples to observed results
- behavior to gold files

This coupling is intended to keep the implementation receivable: the relationships that define the language should remain visible enough to inspect and verify.

## Numbers and evaluation

Numbers are exact rationals unless an operation explicitly enters an inexact/provider-backed domain.

Binary expressions evaluate strictly from left to right.

For example:

```aiki
1 + 2 * 3
```

evaluates as:

```text
(1 + 2) * 3
= 9
```

This is a deliberate semantic choice, not a parser limitation.

## Concurrency

Alpha 1 includes isolated spawned computation and message-passing channels.

Spawned computations do not share ordinary mutable execution state. Communication occurs explicitly through channels, preserving a small and inspectable concurrency model.

## Graphics

The release includes canvas and turtle facilities sufficient for small graphical programs and examples.

These are part of the same language/runtime environment rather than a separate demonstration layer.

## Validation

Alpha 1 already places substantial emphasis on validation.

The repository includes:

- Go tests
- Aiki tests
- gold tests
- grammar/engine smoke checks
- structural checks
- documentation/example checks
- semantic profiling
- distribution checks

The purpose is not only to catch defects, but to make architectural claims executable where practical.

## Project status

Alpha 1 is a public experimental release.

The language is intentionally small, but the implementation is already substantial enough to support real programs, concurrency, modules, graphics, exact arithmetic, executable documentation, and architectural validation.

The release is intended for inspection, experimentation, and feedback rather than compatibility-sensitive production use.

Later Alpha milestones may tighten implementation boundaries, expand systems capabilities, improve portability and tooling, and make more of the language self-describing while preserving the core semantic character established here.
