# Milestone 05 — Cut 4a evaluator coupling

Status: GATED

Made grammar/evaluator dispatch validation bidirectional over the real
AST-producing set. Correct accounting is:

- 32 grammar productions;
- 6 production-referenced token node classes: `NAME`, `NUMBER`, `STRING`,
  `RUNE`, `SYMBOL`, `SHAPE`;
- synthetic parser node `TERMINAL`;
- 39 evaluator dispatch entries total.

`BINOP` is a production, not a `TokenRef`; this corrected an initial test
assumption. Dead lexical handlers for `KEYWORD`, `OPERATOR`, `DELIMITER`, and
`NEWLINE` were removed. Negative tests prove both missing-handler and
extra-handler drift is rejected.

Authoritative `make validate`: passed after the representation-correct test fix.
