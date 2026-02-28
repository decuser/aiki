package substrate

import (
	"bufio"
	"fmt"
	"io"

	"aiki/engine/semantics/value"
)

func halPrint(args []value.Value) value.Value {
	for _, arg := range args {
		if s, ok := arg.(*value.String); ok {
			fmt.Fprint(Stdout, s.Val)
		} else {
			fmt.Fprint(Stdout, arg.Inspect())
		}
	}
	return value.NULL
}

func halRead(args []value.Value) value.Value {
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
		return value.NewError("read: %v", err)
	}
	if len(line) > 0 && line[len(line)-1] == '\n' {
		line = line[:len(line)-1]
	}
	return &value.String{Val: line}
}
