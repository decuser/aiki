# Structure Tests

Verify the engine produces correct intermediate representations at each stage.

## engine/

Gold-file tests for lexer, parser, and evaluator output. Each `.ai` file has corresponding `.lex.gold`, `.parse.gold`, and `.eval.gold` files.

Run via: `aiki enginesmoke`

## Grammar coverage

The structural engine suite must exercise every production declared in `grammar.ebnfx`. `aiki enginesmoke --coverage` checks this without comparing or writing golds. Both `--check` and `--gold` also enforce complete production coverage before proceeding.

Golds preserve a structure already established as correct; they are not themselves proof of correctness. See `docs/testing.md`.
