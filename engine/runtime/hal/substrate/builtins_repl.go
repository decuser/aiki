package substrate

import (
	"fmt"

	"aiki/engine/runtime/hal"
	"aiki/engine/semantics/value"
)

func halQuit(args []value.Value, ctx *hal.EvalContext) value.Value {
	CloseAllCanvases()
	return value.EXIT
}

func halReset(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 0 {
		return value.NewError("reset: want 0 arguments, got %d", len(args))
	}
	CloseAllCanvases()
	return value.RESET
}

func halHelp(args []value.Value, ctx *hal.EvalContext) value.Value {
	fmt.Fprint(Stdout, `Aiki

Primitives:
  first(list)         - first element
  rest(list)          - all but first
  length(list)        - length
  prepend(list, val)  - add to front
  append(list, val)   - add to end
  type(val)           - type as symbol
  inspect(val)        - string representation
  shape(val)          - shape name or :list
  equal(a, b)         - deep equality
  to_str(val)         - convert to string
  print(val...)       - output
  read()              - input line
  sin(n)              - sine
  cos(n)              - cosine
  sleep(ms)           - pause milliseconds

Control:
  quit()              - exit REPL
  reset()             - reset environment

`)
	return value.EMPTY
}
