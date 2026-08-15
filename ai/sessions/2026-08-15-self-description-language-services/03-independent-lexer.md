# 03 — Independent Aiki lexer

Status: ACTIVE — implementation complete; authoritative gate pending.

## Result

Added `selfhost/lexer.ai`, a character-level lexer written in ordinary Aiki. It
does not call the Go lexer and does not use regex. It implements the grammar's
pattern languages directly and applies maximal-munch classification, including
keyword/name and operator/delimiter conflicts.

The lexer emits shaped `@token [kind, lexeme, line, col]` values and exposes a
deterministic human-readable token projection.

## Authority coupling

The lexer necessarily copies token names, keyword/operator/delimiter tables,
and skip metadata. `selfhost/check_lexer_authority.ai` compares those copied
facts with the Aiki-native projection derived from `grammar.ebnfx` by Cut I.0.
The check required an Aiki-native recursive list comparator because the public
`equal()` operation intentionally returns false for lists.

## Conformance corpus

`test/conformance/syntax/lex/` contains four reviewed ASCII fixture families:

- basic names/binding/call/newline;
- comments, tabs, carriage-return behavior, and physical newline runs;
- number/string/rune/symbol/shape literals;
- maximal-munch and keyword/name conflicts.

`test/invariant/selfhost_lexer_conformance_test.go` requires both the Go lexer
and the Aiki lexer to equal the same `.tokens` projection.

## Implementation discovery

Aiki `and`/`or` are not short-circuit guards. Initial scanner code used a
C/Go-style bounds idiom such as `(i < n) and source[i]...`; disposable runtime
execution correctly exposed the out-of-bounds fault. The scanner now uses
bounds-safe accessor predicates, preserving Aiki evaluation semantics rather
than assuming short-circuit behavior.

## Evidence

A disposable validation runtime was built from a copy of the tree using local
Go 1.23, with only Ebiten replaced by a no-op graphics stub. This is not an
authoritative release gate, but it exercised the real Aiki lexer/parser,
evaluator, HAL, prelude, module system, and the repository's current Aiki
sources. All four lexical fixtures matched exactly.

The authoritative invariant remains pending because the actual repository
requires Go 1.24.
