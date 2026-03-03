# Aiki Constraints

## 1. IDENTITY

Aiki is a minimal composable programming language with a live environment.
Aiki is not a product. It is an epistemological stance.
It is code that does not rot, because it hides nothing.
The language is a learning field, not just a tool.

## 2. PRINCIPLES

Constraints over choice. One way to do each thing.
Explicit over implicit. No magic, no hidden behavior.
Composition through naming. Small pieces, named and combined.
Simplicity over ease. Simple removes complexity. Easy hides it.
Exactness by default. Rationals in the core. Floats are opt-in.
Inspectability enables knowing. The prelude proves the language. Everything answers: what are you, what's in you.

## 3. EPISTEMOLOGY

A programming environment is an information system whose job is to create conditions for apprehension.
Opacity prevents proximity. If you can't inspect, you can't align.
A system built on strict constraints—rational math, explicit grouping, universal data structures—remains comprehensible.

## 4. LINEAGE

Syntax: BCPL → C → Go. Braces, simplicity, small language philosophy. Familiar costume, different skeleton.
Data: Scheme. Lists as universal structure. Composition from atoms. We reject the macro.
Evaluation: Smalltalk. Left-to-right. No operator precedence.
Flow: ML, F#. Pipes. Errors are values.
Kernel: Forth. Minimal core. Stdlib written in itself.
Feedback: BASIC, Logo. Immediate. Drawing as primitive.
Host: Go. Goroutines, channels, fmt. We accept the host's physics to enforce our own chemistry.

## 5. TYPES

Eight types, each added from necessity:
- Number (rational, exact arithmetic)
- Boolean
- Rune (Unicode code point)
- String (immutable rune sequence)
- Bytes (immutable 0-255 for I/O)
- Symbol (atomic, identity-compared, not list tags)
- List (raw or shaped, one compound structure)
- Function (first-class)

## 6. SYNTAX

let creates, = mutates, error if not found.
Explicit return required in every path.
One function syntax. No arrows.
One loop: while. Iteration via map, filter, each.
One conditional: if/else. Guard clauses replace elif.
Empty parens required for calls.
Left-to-right evaluation. No precedence. Parens for grouping.
Pipe operator. Left becomes first arg. Errors short-circuit.
Success returns value. Failure returns [@error, reason].

## 7. ARCHITECTURE

Two layers: HAL and Aiki.
HAL provides gateway to Go runtime. Invisible to user.
Prelude uses HAL to provide the language. The dictionary you start with.
Syntax is structure. Semantics is meaning. Runtime is capability.
Grammar is source of truth, embedded in binary.
Evaluator uses handler map. ValidateHandlers ensures coverage.
Tools are projections. They contain no language rules.

## 8. PARADIGM

Pre-paradigmatic. Primitives support multiple styles:
- Recursive (first/rest)
- Iterative (while)
- Functional (map, filter, pipes)
- Imperative (mutation with =)
- Immediate (REPL, drawing)

## 9. REJECTED

Arrow lambdas — second way to write functions.
Implicit return — last-line bugs.
for loop — while + functions cover it.
elif — guard clauses.
Operator precedence — invisible rules that fool you.
Classes/methods — data doesn't have behavior.
Exceptions — errors are values.
Macros — language is closed.
IEEE floats — rationals are exact.
Pointers — breaks inspectability.
