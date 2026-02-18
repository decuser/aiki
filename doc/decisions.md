<!-- contract
allowed: WHY HIST PLAN

Note: PLAN lines need to include an explicit status sentence.
-->

# Aiki Decisions

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
