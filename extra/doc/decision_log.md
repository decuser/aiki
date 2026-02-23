# Aiki Decision Log

## Deferred

20250221-decision: deferred TCO, rationale: required for recursion-heavy code, not yet implemented
20250221-decision: deferred inter-process communication, rationale: spawn/channels exist, cross-process IPC post-alpha
20250221-decision: deferred sockets/networking, rationale: post-alpha
20250221-decision: deferred time travel versioning, rationale: post-alpha
20250221-decision: deferred TUI, rationale: console first

## Architecture

20250221-decision: adopted GrammarContract/Observer/engine structure as target, rationale: clean boundaries from failed refactor
20250221-decision: structural reorg to reference/engine/extra/cmd, rationale: isolate exemplar for shadow implementation
20250221-decision: doclint file-authoritative, rationale: removed ini dependency
20250221-decision: debug command with gold/check, rationale: comparison infrastructure for engine rebuild

## Syntax

20250222-decision: unary binds tighter than infix via grammar nesting, rationale: -1 + 2 must mean (-1) + 2, visible in grammar structure not precedence table

## Violations to Fix

20250222-decision: fix evaluator imports HAL directly, rationale: should mediate through RuntimeContract interface
20250222-decision: fix evaluator imports os/fmt for spawn error logging, rationale: route through HAL.Stderr
