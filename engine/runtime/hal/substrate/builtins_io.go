package substrate

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"aiki/engine/semantics/value"
)

func (g *GoRuntime) registerIO() {
	// Print - writes arguments contiguously with no implicit separators
	g.Register("print", func(args []value.Value) (value.Value, error) {
		for _, a := range args {
			if a.Type == value.String {
				fmt.Fprint(Stdout, a.Data)
			} else {
				fmt.Fprint(Stdout, a.Inspect())
			}
		}
		if f, ok := Stdout.(*os.File); ok {
			f.Sync()
		}
		return value.NullValue(), nil
	})

	// Read - reads a line from stdin
	g.Register("read", func(args []value.Value) (value.Value, error) {
		reader := bufio.NewReader(os.Stdin)
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				if line == "" {
					return value.NewList(nil), nil
				}
				return value.NewString(line), nil
			}
			return value.NullValue(), fmt.Errorf("read: %v", err)
		}
		if len(line) > 0 && line[len(line)-1] == '\n' {
			line = line[:len(line)-1]
		}
		return value.NewString(line), nil
	})

	// open - open file for reading
	g.Register("open", func(args []value.Value) (value.Value, error) {
		if len(args) != 1 {
			return value.NullValue(), fmt.Errorf("open: want 1 argument")
		}
		if args[0].Type != value.String {
			return value.NullValue(), fmt.Errorf("open: expected string path")
		}
		path := args[0].Data.(string)
		f, err := os.Open(path)
		if err != nil {
			return value.NullValue(), fmt.Errorf("open: %v", err)
		}
		return value.Value{Type: value.Handle, Data: f}, nil
	})

	// create - create file for writing
	g.Register("create", func(args []value.Value) (value.Value, error) {
		if len(args) != 1 {
			return value.NullValue(), fmt.Errorf("create: want 1 argument")
		}
		if args[0].Type != value.String {
			return value.NullValue(), fmt.Errorf("create: expected string path")
		}
		path := args[0].Data.(string)
		f, err := os.Create(path)
		if err != nil {
			return value.NullValue(), fmt.Errorf("create: %v", err)
		}
		return value.Value{Type: value.Handle, Data: f}, nil
	})

	// fread - read from file handle
	g.Register("fread", func(args []value.Value) (value.Value, error) {
		if len(args) != 1 {
			return value.NullValue(), fmt.Errorf("fread: want 1 argument")
		}
		if args[0].Type != value.Handle {
			return value.NullValue(), fmt.Errorf("fread: expected handle")
		}
		f, ok := args[0].Data.(*os.File)
		if !ok {
			return value.NullValue(), fmt.Errorf("fread: invalid handle")
		}
		buf := make([]byte, 4096)
		n, err := f.Read(buf)
		if err != nil {
			if err == io.EOF {
				return value.NewList(nil), nil
			}
			return value.NullValue(), fmt.Errorf("fread: %v", err)
		}
		return value.NewString(string(buf[:n])), nil
	})

	// fwrite - write to file handle
	g.Register("fwrite", func(args []value.Value) (value.Value, error) {
		if len(args) != 2 {
			return value.NullValue(), fmt.Errorf("fwrite: want 2 arguments")
		}
		if args[0].Type != value.Handle {
			return value.NullValue(), fmt.Errorf("fwrite: expected handle")
		}
		f, ok := args[0].Data.(*os.File)
		if !ok {
			return value.NullValue(), fmt.Errorf("fwrite: invalid handle")
		}
		if args[1].Type != value.String {
			return value.NullValue(), fmt.Errorf("fwrite: expected string data")
		}
		data := args[1].Data.(string)
		_, err := f.WriteString(data)
		if err != nil {
			return value.NullValue(), fmt.Errorf("fwrite: %v", err)
		}
		return value.True(), nil
	})

	// fclose - close file handle
	g.Register("fclose", func(args []value.Value) (value.Value, error) {
		if len(args) != 1 {
			return value.NullValue(), fmt.Errorf("fclose: want 1 argument")
		}
		if args[0].Type != value.Handle {
			return value.NullValue(), fmt.Errorf("fclose: expected handle")
		}
		f, ok := args[0].Data.(*os.File)
		if !ok {
			return value.NullValue(), fmt.Errorf("fclose: invalid handle")
		}
		err := f.Close()
		if err != nil {
			return value.NullValue(), fmt.Errorf("fclose: %v", err)
		}
		return value.True(), nil
	})
}
