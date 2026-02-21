<!-- contract
allowed: WHY HIST PLAN

Note: PLAN lines need to include an explicit status sentence.
-->

# Aiki Decisions

## 2026 02 18 Tooling, drift control
PLAN Doclint as documentation contract enforcer. Status implemented for doc and work trees.
WHY Documentation contracts prevent drift between design decisions work tracking and implementation.

PLAN Smoke subcommand and transcript goldens as canary contracts. Status implemented for core math, recursion, iteration, pipeline, functional, IO, error, import, canvas, ord, and hash surfaces.
WHY Smoke enforces behavioral contracts at the language surface and catches drift beyond Go tests.

## 2026 02 18 Print semantics and naming

PLAN Print is concatenative with no implicit separators; callers must print spaces and newlines explicitly. Status implemented in runtime/hal.  
WHY Removing implicit spaces aligns IO with explicit semantics and makes print behavior contractible.

PLAN "_" is a valid discard identifier in naming rules without runtime magic. Status implemented in tools/lint only.  
WHY This enables a conventional ignore binding while keeping evaluation and case policy simple and explicit.


## 2026 02 18 Tooling, drift control
PLAN Doclint as documentation contract enforcer. Status implemented for doc and work trees.
WHY Documentation contracts prevent drift between design decisions work tracking and implementation.

## 2026 02 17 Grammar loading boundary and package stabilization

HIST Embedded grammar via go embed. [syntax/grammar_loader.go]
HIST Introduced syntax.GetGrammar as canonical loader. [syntax/grammar_loader.go GetGrammar]
HIST Updated tools and tests to use GetGrammar. [tools and tests]

WHY A single grammar entry point prevents drift and avoids filesystem dependence for grammar.
WHY Cached grammar reduces repeated parse cost and removes constructor ambiguity.
WHY Package boundaries keep syntax free of meaning and keep tools free of language rules.

## 2026 02 15 Architecture direction

PLAN Operators as functions in exact mode. Status not implemented. [doc/todo.md]
WHY Operator lookup enables shadowing and keeps semantics uniform with normal calls.
PLAN Precise mode per module. Status not implemented. [doc/todo.md]
WHY Precise mode enables float fast paths without changing exact default semantics.
PLAN Name based modules via resolver. Status not implemented. [doc/todo.md]
WHY Module meaning should not depend on storage layout.

PLAN Resources as shaped lists. Status not implemented. [doc/todo.md]
WHY Shape is claim and registry is authority. Effects remain in runtime.
