# 05 — Reconciliation and final gate

**Status:** ACTIVE

A repository search after implementation finds only these non-grammar-package grammar-expression traversals:

- `engine/syntax/parser.go` — generic EBNF interpretation; intentional.
- `cmd/subcommands/dev/enginesmoke/cmd.go` — expression rendering for structural output; intentional.
- `test/invariant/handler_validation_test.go` — constructs a grammar expression to simulate mutation; not a derivation walker.

No consumer-side walker remains for reusable TokenRef, AST-node, terminal-alternative, or newline structural facts.

Decision D3 records the architectural rule and its boundary.

Pending: authoritative full-tree `make validate` in the user's Go 1.24 environment.
