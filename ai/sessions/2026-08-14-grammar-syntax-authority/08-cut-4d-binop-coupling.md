# Milestone 08 — Cut 4d BINOP coupling

Status: GATED

Removed the duplicated binary-operator membership list from evaluator helpers.
`SetGrammar` now derives operator membership from the grammar's `BINOP`
production. Evaluator code retains the semantic implementations and profiling
categories.

Bidirectional validation proves every grammar `BINOP` has semantic treatment
and every semantic binary operator belongs to `BINOP`. Negative tests exercise
both drift directions.

Authoritative `make validate`: passed.
