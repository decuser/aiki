# Aiki TODO

## Done (v0.2.3)

- [x] Lexer (state machine, UTF-8)
- [x] Parser (recursive descent, left-to-right)
- [x] Evaluator (tree-walking)
- [x] 8 types (number, boolean, rune, string, bytes, symbol, list, function)
- [x] Rational arithmetic (exact, via big.Rat)
- [x] Shapes with composition
- [x] Pattern matching
- [x] Pipe operator with error short-circuit
- [x] Pipe auto-unwrap `[@ok, val]`
- [x] Prelude (map, filter, reduce, range, hash map)
- [x] File I/O (open, create, fread, fwrite, fclose)
- [x] REPL with readline
- [x] File runner
- [x] Formatter with comment preservation
- [x] Comma separators (calls, lists, params, patterns)
- [x] Ruby-style error messages with source line
- [~] Test suite
- [x] Canvas primitives (Ebiten)
- [x] Concurrency (spawn, channel, send, recv)

## Next
- [ ] Contemplate and decide - remove !, !=, == in favor of not, not equal, equal?
- [ ] Debugger
- [ ] Regex primitives
- [ ] Bit primitives

## Future

- [ ] Module system (from/use/export)
- [ ] Bytecode VM
- [ ] Multi-line REPL
- [ ] LSP
- [ ] WASM target
