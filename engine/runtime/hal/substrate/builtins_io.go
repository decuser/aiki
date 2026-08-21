package substrate

import (
	"bufio"
	"fmt"
	"io"

	"aiki/engine/runtime/hal"
	"aiki/engine/semantics/value"
)

func (g *GoRuntime) halPrint(args []value.Value, ctx *hal.EvalContext) value.Value {
	for _, arg := range args {
		if s, ok := arg.(*value.String); ok {
			fmt.Fprint(g.stdout, s.Val)
		} else {
			fmt.Fprint(g.stdout, arg.Inspect())
		}
	}
	return value.EMPTY
}

func (g *GoRuntime) halRead(args []value.Value, ctx *hal.EvalContext) value.Value {
	return g.halIOReadLine([]value.Value{&value.Symbol{Val: "stdin"}}, ctx)
}

func (g *GoRuntime) halIORead(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("io.read: want 2 arguments, got %d", len(args))
	}
	count, ok := args[1].(*value.Number)
	if !ok || !count.IsInt() || count.Sign() < 0 {
		return value.NewFault("io.read: count must be a non-negative integer")
	}
	n := count.Int64Value()
	if n == 0 {
		return &value.Bytes{Val: []byte{}}
	}
	if n > int64(^uint(0)>>1) {
		return value.NewShapedError("io", "read count too large")
	}
	reader, _, shapedErr := g.ioReader(args[0])
	if shapedErr != nil {
		return shapedErr
	}
	buf := make([]byte, int(n))
	read, err := reader.Read(buf)
	if err != nil && err != io.EOF {
		return value.NewShapedError("io", "read: %v", err)
	}
	if read == 0 && err == io.EOF {
		return &value.List{Shape: "end", Elements: []value.Value{}}
	}
	return &value.Bytes{Val: buf[:read]}
}

func (g *GoRuntime) halIOReadLine(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("io.read_line: want 1 argument, got %d", len(args))
	}
	reader, _, shapedErr := g.ioReader(args[0])
	if shapedErr != nil {
		return shapedErr
	}
	buffered, ok := reader.(*bufio.Reader)
	if !ok {
		buffered = bufio.NewReader(reader)
	}
	line, err := buffered.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			if line == "" {
				return &value.List{Shape: "end", Elements: []value.Value{}}
			}
			return &value.String{Val: line}
		}
		return value.NewShapedError("io", "read_line: %v", err)
	}
	if len(line) > 0 && line[len(line)-1] == '\n' {
		line = line[:len(line)-1]
	}
	if len(line) > 0 && line[len(line)-1] == '\r' {
		line = line[:len(line)-1]
	}
	return &value.String{Val: line}
}

func (g *GoRuntime) halIOWrite(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("io.write: want 2 arguments, got %d", len(args))
	}
	writer, _, shapedErr := g.ioWriter(args[0])
	if shapedErr != nil {
		return shapedErr
	}
	var data []byte
	switch v := args[1].(type) {
	case *value.String:
		data = []byte(v.Val)
	default:
		var err error
		data, err = bytesFromAiki(args[1])
		if err != nil {
			return value.NewFault("io.write: %v", err)
		}
	}
	written, err := writer.Write(data)
	if err != nil {
		return value.NewShapedError("io", "write: %v", err)
	}
	if written != len(data) {
		return value.NewShapedError("io", "short write: wrote %d of %d bytes", written, len(data))
	}
	return value.TRUE
}

func (g *GoRuntime) halIOClose(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("io.close: want 1 argument, got %d", len(args))
	}
	switch v := args[0].(type) {
	case *value.File:
		return g.halFileClose(args, ctx)
	case *value.Endpoint:
		resource, ok := g.endpointResource(v)
		if !ok {
			return value.NewShapedError("io", "endpoint does not belong to this runtime")
		}
		if err := resource.close(); err != nil {
			return value.NewShapedError("io", "close: %v", err)
		}
		return value.TRUE
	default:
		return value.NewShapedError("io", "expected file or endpoint, got %s", args[0].Type())
	}
}
