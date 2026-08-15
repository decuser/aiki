# Milestone 21 — Self-host expression evaluator

Status: GATED locally; authoritative repository gate pending.

Implemented `selfhost/evaluator.ai` expression semantics over the independent Phase-I AST: native literals, lists, dynamic symbols/shapes, grouping, unary operations, left-to-right binary evaluation, indexing, interpreted shape field access, and function-literal closure construction.

Cross-implementation probes agree on representative expressions. Notably `1 + 2 * 3` evaluates to `9`, preserving Aiki's left-to-right rule. The second implementation also exposed and corrected the fact that `and`/`or` return operand values (`true and 7` -> `7`) rather than coerced booleans.

Added `selfhost_evaluator_expr_test.go` for the authoritative gate.
