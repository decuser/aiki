# 05 — Independent Aiki parser and syntax projection

Status: ACTIVE — implementation complete; authoritative gate pending.

## Result

Added `selfhost/parser.ai`, an independent hand-written recursive-descent parser
covering the authoritative Aiki grammar, including package declarations, shape
definitions, assignment, if/else, while, match patterns, select cases/default,
functions/rest parameters, lists/shapes, postfix operations, pipelines, unary
forms, and left-to-right binary-expression structure.

It consumes only the Aiki normalizer's token stream. It does not call or inspect
the Go parser.

## Neutral syntax projection

No new parse-gold format was invented. The neutral projection is the existing
human-readable grammar-tree surface already stored in
`test/structure/engine/*.ai.parse.gold`:

```text
indentation + node kind + line:column + optional terminal/token lexeme
```

This surface is grammar-shaped rather than a Go-struct serialization. Reusing
it avoids a second parse expectation corpus. The ordinary engine gates require
the Go parser to match those files; the new self-host invariant requires the
Aiki parser to match the same files.

`test/invariant/selfhost_parser_conformance_test.go` runs the complete independent
Aiki front end over every engine structure input and compares its tree with the
corresponding parse gold.

## Evidence

All selfhost `.ai` files were first lexed and parsed by the real Go syntax
engine in a disposable Go-1.23 syntax-only copy.

The disposable non-graphics Aiki runtime then executed:

```text
Aiki lexer -> Aiki normalizer -> Aiki parser -> Aiki tree projection
```

for all ten existing engine structure programs. Every projection matched its
`.parse.gold` file exactly, including `09_grammar_coverage_engine.ai` and the
select grammar fixture.

Authoritative Go-1.24 validation remains pending.
