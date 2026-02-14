# The Aiki Manifesto

## Focus

**Constraints over choice.** One way to do each thing. Features enter by replacing complexity, not adding to it.

**Explicit over implicit.** Left-to-right evaluation. Parens for grouping. `let` creates, `=` mutates. Errors are values, not exceptions.

**Composition over inheritance.** Small pieces, named and combined. Lists and shapes, not classes.

**Simplicity over ease.** Simple removes complexity. Easy hides it.

**Exactness by default.** Rational arithmetic in the core. `1/3 * 3 = 1`. Floats are opt-in, in a separate layer.

**Inspection over abstraction.** The hash map is written in Aiki. The strict layer proves the language works by using it.

## Paradigm

Aiki is pre-paradigmatic. The primitives support multiple styles without privileging one:

- **Recursive.** Functions call themselves. Lists decompose with `first`/`rest`.
- **Iterative.** `while` loops. Explicit state.
- **Functional.** `map`, `filter`, `reduce`, pipes.
- **Imperative.** Sequential statements. Mutation with `=`.
- **Immediate.** REPL. `dot(c, x, y)`. Drawing as primitive.

The language provides the atoms. You choose the chemistry.

## Mechanisms

**Pipes.** `x |> f() |> g()`. Data flows left to right. Errors short-circuit.

**Variadic.** `(...args)` collects. `apply(f, list)` spreads.

**Shapes.** `[@point, x, y]`. Structure without classes.

**Match.** Destructure values, bind names, branch on shape. One construct for what other languages split across switch, instanceof, and unpacking.

**Concurrency.** `spawn(f)` runs. `channel()` connects. `send()`/`recv()` synchronize.

## Architecture

Three layers acknowledge that principles and pragmatism coexist:

| Layer | What | Tradeoff |
|-------|------|----------|
| hal | Go primitives | Speed, OS access |
| strict | Pure Aiki | Inspectable, slower |
| pragmatic | Fast Aiki | Speed, less transparent |

Strict is the default. Pragmatic is opt-in. Same API, different physics.

## Tooling

One formatter. One style. Errors point to source. The REPL is the workbench.

Subcommands are pure Aiki. The tools prove the language works by using it.

Help is projection, not explanation. Derived from grammar, not written as prose.

## Lineage

**Syntax.** BCPL, C, Go. Braces, simplicity, small language philosophy. Familiar costume, different skeleton.

**Data.** Scheme. Lists as universal structure. Composition from atoms.

**Evaluation.** Smalltalk. Strict left-to-right. No operator precedence.

**Flow.** ML, F#. Data flows through pipes. Errors are values like any other. Match on them or let them propagate.

**Kernel.** Forth. Minimal core. Stdlib written in itself.

**Feedback.** BASIC, Logo. Immediate. Drawing as primitive.

**Host.** Go. Goroutines, channels, fmt. We accept the host's physics to enforce our own chemistry.

## Infrastructure

Aiki-grammar is the kernel. Aiki-lang is the first client.

The grammar defines the structural space. Languages are points in that space. The grammar is frozen. Languages can evolve.

Don't build in capabilities you won't surface. If the grammar can express something, users can express it too.
