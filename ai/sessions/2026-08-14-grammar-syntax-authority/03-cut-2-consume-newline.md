# Milestone 03 — Cut 2 consume grammar newline policy

Status: GATED

Replaced the parser's private newline vocabulary with generic application of
`g.Newline`. Completion token/lexeme sets and suppression pairs are compiled
from grammar metadata. `isComplete` and hardcoded Aiki newline/suppression
membership were deleted from `parser.go`.

A unit test mutates the grammar newline declaration in memory and proves the
same token stream changes parse behavior, demonstrating that the parser truly
consumes the grammar rule rather than carrying it alongside a hidden rule.

All eleven Cut 0 probes preserved their baseline parse behavior. Authoritative
`make validate`: passed with 408 Aiki tests, 45 behavior smokes, grammar
coverage, engine golds, and treecheck.
