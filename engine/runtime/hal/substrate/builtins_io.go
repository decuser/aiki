package substrate

import (
	"bufio"
	"fmt"
	"io"

	"aiki/engine/semantics/value"
)

func (g *GoRuntime) registerIO() {
	g.register("print", func(args []value.Value) value.Value {
		for _, arg := range args {
			if s, ok := arg.(*value.String); ok {
				fmt.Fprint(Stdout, s.Val)
			} else {
				fmt.Fprint(Stdout, arg.Inspect())
			}
		}
		return value.NULL
	})

	g.register("input", func(args []value.Value) value.Value {
		if len(args) > 0 {
			if s, ok := args[0].(*value.String); ok {
				fmt.Fprint(Stdout, s.Val)
			}
		}
		reader := bufio.NewReader(Stdin)
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return value.NewError("input: %v", err)
		}
		if len(line) > 0 && line[len(line)-1] == '\n' {
			line = line[:len(line)-1]
		}
		return &value.String{Val: line}
	})
}
