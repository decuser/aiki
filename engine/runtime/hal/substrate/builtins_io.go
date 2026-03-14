package substrate

import (
	"bufio"
	"fmt"
	"io"

	"aiki/engine/runtime/hal"
	"aiki/engine/semantics/value"
)

func halPrint(args []value.Value, ctx *hal.EvalContext) value.Value {
	for _, arg := range args {
		if s, ok := arg.(*value.String); ok {
			fmt.Fprint(Stdout, s.Val)
		} else {
			fmt.Fprint(Stdout, arg.Inspect())
		}
	}
	return value.EMPTY
}

func halRead(args []value.Value, ctx *hal.EvalContext) value.Value {
	reader := bufio.NewReader(Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			if line == "" {
				return &value.List{
					Elements: []value.Value{},
					Shape:    "end",
				}
			}
			return &value.String{Val: line}
		}
		return value.NewShapedError("io", "read: %v", err)
	}
	if len(line) > 0 && line[len(line)-1] == '\n' {
		line = line[:len(line)-1]
	}
	return &value.String{Val: line}
}
