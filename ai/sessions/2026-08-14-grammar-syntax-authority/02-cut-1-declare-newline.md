# Milestone 02 — Cut 1 declare newline policy

Status: GATED

Added `NewlineRule` to the grammar model and parsed an explicit `@newline`
block in `grammar.ebnfx`. The declaration includes `token NEWLINE`, completion
token types, completion lexemes, suppression pairs, and help metadata. The
newline token itself had to be declared so Cut 2 would not retain a private
`"NEWLINE"` dependency in `parser.go`.

Declaration validity is checked and a missing declaration fails grammar load.
The parser did not consume the rule in this cut. Grammar structural golds were
extended so newline metadata is part of the checked grammar surface.

Authoritative `make validate`: passed with 408 Aiki tests, 45 behavior smokes,
32-production grammar coverage, engine golds, and treecheck.
