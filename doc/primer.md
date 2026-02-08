# Aiki Project Primer

Instructions for collaborating on the Aiki language design. Read grammar.md for syntax, design.md for rationale, this file for how to work with me.

## Project State

v0.2.0 design complete. Implementation not started. Everything is open to challenge if the argument is good.

## How I Work

**Push back.** If something seems wrong, say so. Don't hedge. "That breaks your core principle because X" is better than "you might consider whether..."

**Be direct.** Short answers for short questions. Don't pad.

**Ask hard questions.** "Do you really need that?" and "What does that buy you?" are good questions.

**Show, don't describe.** Write code examples. Let me see what the syntax looks like in practice.

**Challenge redundancy.** If two features do the same thing, one should die. Help me find those.

**Track the changes.** When we make a decision, be ready to summarize what changed.

## Design Values

- One way to do each thing
- Explicit over implicit
- Composition through naming
- No magic, no hidden behavior
- Syntax should be obvious at a glance
- If it needs explaining, it might be wrong
- Inspectability enables knowing

## What's Settled

**Language:**
- Eight types (number, boolean, rune, string, bytes, symbol, list, function)
- Numbers are rational (exact arithmetic, Racket-style)
- Bytes for I/O and binary data, distinct from strings
- Shaped lists with `@` for structure, can embed other shapes
- Symbols (`:name`) for atomic values only
- `let` for binding, `=` for mutation (error if not found)
- Explicit `return` always, multiple returns allowed
- Guard clauses instead of `elif`
- One function syntax: `(params) { return expr }`
- `while` as only loop, iteration via higher-order functions
- Pipe operator with error short-circuit
- Files as modules, explicit export
- Dot access for everything: `list.0`, `point.x`, `math.sqrt`
- Empty parens required for calls
- TCO mandatory, stack depth limits for non-tail recursion

**Standard Library:**
- Primitives implemented in Go (list, type, convert, I/O, math, bit, regex, concurrency, canvas, system)
- Prelude written in Aiki (map, filter, reduce, streams, turtle, etc.)
- Operators for arithmetic, not functions
- Streams are functions that return `[@ok val]` or `[@end]`
- RE2 regex (linear time, safe)

**Concurrency:**
- Go-lite model: `spawn`, `channel`, `send`, `recv`
- Green threads with cooperative scheduling
- Message passing encouraged, shared state discouraged
- Marked experimental

**Environment:**
- REPL as primary mode (no DSL, just Aiki functions)
- Inspect any value
- Hot reload
- File shadows (live system is truth, files are projections)
- Build to executable
- Universe model: loaded, scope, filesystem
- Projects are just lists of files to load

**Graphics:**
- Canvas primitives (Ebiten-based)
- Turtle graphics in prelude
- Logo should be trivial

## What's Open

- Implementation (lexer, parser, evaluator, VM)
- System layer specifics
- Edge cases we haven't hit yet
- Standard library growth (emerges from use)
- TUI design details

## What's Deferred

- Process spawning / IPC
- Sockets / networking
- Full time machine versioning
- Plan 9 style /dev integration

## When We Disagree

Make the case. If I say no, I'll say why. If you're right, I'll change my mind. Don't capitulate just because I pushed back once.

## The Deeper Context

Aiki is grounded in an epistemological stance: information is realized only when a knowing subject apprehends a potential input and undergoes configurational change (K → K′). The programming environment is an information system - its job is to create conditions for apprehension.

This is why inspectability matters. This is why the system must be live. This is why files are shadows, not sources. Opacity prevents proximity. If you can't inspect, you can't align. If you can't align, no knowing.

The language is a learning field, not just a tool.

## Implementation Order

When we start building:

1. **Lexer** - tokens from source (DAL as starting point)
2. **Parser** - AST from tokens
3. **Evaluator** - tree-walking interpreter (simple, slow, correct)
4. **REPL** - interactive, uses evaluator
5. **File runner** - `aiki run file.ai`
6. **Debugger** - step, inspect, breakpoints
7. **Bytecode VM** - replace tree-walker for speed
8. **Executable bundling** - ship without Go installed

Start with 1-5. That's a working language.
