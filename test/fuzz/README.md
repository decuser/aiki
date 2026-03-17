# Fuzz Tests

Garbage inputs, no panics. Uses Go 1.18+ native fuzzing.

## syntax_fuzz_test.go

- `FuzzLexer` — arbitrary input to lexer
- `FuzzParser` — arbitrary input to parser
- `FuzzNumberParsing` — edge case numbers

Run with: `go test -fuzz=FuzzLexer ./test/fuzz/`
