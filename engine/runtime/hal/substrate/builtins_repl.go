package substrate

import (
	"fmt"
	"time"

	"aiki/engine/semantics/value"
)

func (g *GoRuntime) registerREPL() {
	// sleep - pause execution for milliseconds
	g.Register("sleep", func(args []value.Value) (value.Value, error) {
		if len(args) != 1 {
			return value.NullValue(), fmt.Errorf("sleep: want 1 argument")
		}
		if args[0].Type != value.Number {
			return value.NullValue(), fmt.Errorf("sleep: expected number (milliseconds)")
		}
		ms := args[0].Data.(float64)
		time.Sleep(time.Duration(ms) * time.Millisecond)
		return value.True(), nil
	})

	// help - display help text
	g.Register("help", func(args []value.Value) (value.Value, error) {
		helpText := `Aiki

Primitives:
  first(list)         - first element
  rest(list)          - all but first
  length(list)        - length
  prepend(list, val)  - add to front
  append(list, val)   - add to end
  type(val)           - type as symbol
  shape(val)          - shape name or :list
  equal(a, b)         - deep equality
  to_str(val)         - convert to string
  print(val...)       - output (no newline)
  read()              - read line from stdin
  modulo(a, b)        - remainder
  ord(rune)           - character code
  sqrt(n), sin(n), cos(n), abs(n) - math
  random(max)         - random int [0, max)
  sleep(ms)           - pause execution
  open(path)          - open file for reading
  create(path)        - create file for writing
  fread(h), fwrite(h, s), fclose(h) - file I/O

REPL:
  help()              - this message
  quit()              - exit REPL
`
		fmt.Fprint(Stdout, helpText)
		return value.NullValue(), nil
	})

	// quit - exit REPL
	g.Register("quit", func(args []value.Value) (value.Value, error) {
		return value.NewSymbol("exit"), nil
	})

	// exit - exit REPL
	g.Register("exit", func(args []value.Value) (value.Value, error) {
		return value.NewSymbol("exit"), nil
	})

	// reset - reset environment
	g.Register("reset", func(args []value.Value) (value.Value, error) {
		return value.NewSymbol("reset"), nil
	})
}
