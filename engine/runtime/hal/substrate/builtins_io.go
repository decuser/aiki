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

	g.register("read", func(args []value.Value) value.Value {
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

	g.register("ord", func(args []value.Value) value.Value {
		if len(args) != 1 {
			return value.NewError("ord: want 1 argument")
		}
		switch v := args[0].(type) {
		case *value.Rune:
			return value.NewNumber(int64(v.Val), 1)
		case *value.String:
			if len(v.Val) == 0 {
				return value.NewNumber(0, 1)
			}
			return value.NewNumber(int64([]rune(v.Val)[0]), 1)
		default:
			return value.NewError("ord: expected rune or string")
		}
	})

	g.register("range", func(args []value.Value) value.Value {
		if len(args) < 1 || len(args) > 2 {
			return value.NewError("range: want 1 or 2 arguments")
		}
		var start, end int64
		if len(args) == 1 {
			n, ok := args[0].(*value.Number)
			if !ok || !n.Val.IsInt() {
				return value.NewError("range: expected integer")
			}
			start = 0
			end = n.Val.Num().Int64()
		} else {
			s, ok := args[0].(*value.Number)
			if !ok || !s.Val.IsInt() {
				return value.NewError("range: expected integer")
			}
			e, ok := args[1].(*value.Number)
			if !ok || !e.Val.IsInt() {
				return value.NewError("range: expected integer")
			}
			start = s.Val.Num().Int64()
			end = e.Val.Num().Int64()
		}
		var elems []value.Value
		for i := start; i < end; i++ {
			elems = append(elems, value.NewNumber(i, 1))
		}
		return &value.List{Elements: elems}
	})
}
