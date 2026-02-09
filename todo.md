# Aiki TODO

## Done (v0.2.1)
- [x] Lexer (state machine, UTF-8)
- [x] Parser (recursive descent, left-to-right)
- [x] Evaluator (tree-walking)
- [x] 8 types (number, boolean, rune, string, bytes, symbol, list, function)
- [x] Rational arithmetic (exact, via big.Rat)
- [x] Shapes with composition
- [x] Pattern matching (match statement)
- [x] Pipe operator (|>)
- [x] Prelude (map, filter, reduce, range, etc.)
- [x] Hash map (pure Aiki, O(1) lookup)
- [x] REPL with readline (liner)
- [x] File runner
- [x] Test suite
- [x] Escape sequences in strings
- [x] File I/O (open, create, fread, fwrite, fclose)
- [x] Pipe auto-unwrap [@ok val]
- [x] "The way": success returns value, failure returns [@error reason]
- [x] Formatter shows changed files

## Next
- [ ] Module system (from/use/export)
- [ ] Canvas primitives
- [ ] Error handling improvements
- [ ] Inline function literals as arguments

## Future
- [ ] Bytecode VM (replace tree-walking)
- [ ] LSP (language server for editors)
- [ ] Debugger
- [ ] Regex primitives
- [ ] Bits primitives
- [ ] Concurrency (go-lite: spawn, channels)
- [ ] WASM target
- [ ] Self-hosting (maybe)

## Bugs / Polish
- [ ] Parser: inline function literals `reduce(list 0 (a b) { ... })`
- [ ] REPL: multi-line input
- [ ] Better error messages with line/column
- [ ] `quit()` should actually exit

## Documentation
- [ ] Update design.md with hash map section
- [ ] Tutorial / examples
- [ ] API reference for builtins
