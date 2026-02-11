package repl

import (
	"fmt"

	"aiki/hal/core"
	"aiki/lang/value"
	"aiki/version"
)

func init() {
	core.HAL["reset"] = &value.Builtin{
		Name: "reset",
		Fn: func(args ...value.Value) value.Value {
			if len(args) != 0 {
				return value.NewError("reset: want 0 arguments, got %d", len(args))
			}
			if core.CloseAllCanvases != nil {
				core.CloseAllCanvases()
			}
			return core.Reset
		},
	}

	core.HAL["help"] = &value.Builtin{
		Name: "help",
		Fn: func(args ...value.Value) value.Value {
			fmt.Fprintf(core.Stdout, `Aiki %s

Primitives:
  first(list)         - first element
  rest(list)          - all but first
  len(list)           - length
  prepend(list val)   - add to front
  append(list val)    - add to end
  type(val)           - type as symbol
  inspect(val)        - string representation
  shape(val)          - shape name or :list
  equal(a b)          - deep equality
  tostr(val)          - convert to string
  tonum(str)          - parse number
  todecimal(n places) - format decimal
  print(val...)       - output
  read()              - input line
  sqrt(n)             - square root
  sin(n)              - sine
  cos(n)              - cosine
  random(n)           - random 0 to n-1
  sleep(ms)           - sleep for n milliseconds

Type help(name) for details.
`, version.Version)
			return value.NULL
		},
	}

	core.HAL["quit"] = &value.Builtin{
		Name: "quit",
		Fn: func(args ...value.Value) value.Value {
			fmt.Println("Goodbye!")
			return value.NULL
		},
	}
}
