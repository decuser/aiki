package repl

import (
	"fmt"

	"aiki/runtime/hal"
	"aiki/semantics/value"
	"aiki/version"
)

func init() {
	hal.HAL["help"] = &value.Builtin{
		Name: "help",
		Fn: func(args ...value.Value) value.Value {
			fmt.Fprintf(hal.Stdout, `Aiki %s

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

`, version.Version)
			return value.NULL
		},
	}
}
