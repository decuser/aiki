package substrate

import (
	"bufio"
	"io"
	"os"

	"aiki/engine/runtime/hal"
	"aiki/engine/semantics/value"
)

// halFileOpen opens a file with the specified mode.
// open(path, mode) -> file | [@error, :io, "message"]
func halFileOpen(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("_file_open: want 2 arguments, got %d", len(args))
	}
	path, ok := args[0].(*value.String)
	if !ok {
		return value.NewFault("_file_open: expected string path, got %s", args[0].Type())
	}
	mode, ok := args[1].(*value.Symbol)
	if !ok {
		return value.NewFault("_file_open: expected symbol mode, got %s", args[1].Type())
	}

	var f *os.File
	var err error
	var modeStr string

	switch mode.Val {
	case "read":
		f, err = os.Open(path.Val)
		modeStr = "read"
	case "write":
		f, err = os.Create(path.Val)
		modeStr = "write"
	case "append":
		f, err = os.OpenFile(path.Val, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		modeStr = "append"
	default:
		return value.NewShapedError("io", "invalid mode: %s (expected :read, :write, or :append)", mode.Val)
	}

	if err != nil {
		return value.NewShapedError("io", "%s", err.Error())
	}

	return &value.File{
		Path: path.Val,
		F:    f,
		Mode: modeStr,
	}
}

// halFileReadText reads entire remaining file content as string.
// read_text(file) -> string | [@error, :io, "message"]
func halFileReadText(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("_file_read_text: want 1 argument, got %d", len(args))
	}
	file, ok := args[0].(*value.File)
	if !ok {
		return value.NewFault("_file_read_text: expected file, got %s", args[0].Type())
	}
	if file.F == nil {
		return value.NewShapedError("io", "file is closed")
	}

	data, err := io.ReadAll(file.F)
	if err != nil {
		return value.NewShapedError("io", "%s", err.Error())
	}

	return &value.String{Val: string(data)}
}

// halFileReadBytes reads entire remaining file content as bytes.
// read_bytes(file) -> bytes | [@error, :io, "message"]
func halFileReadBytes(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("_file_read_bytes: want 1 argument, got %d", len(args))
	}
	file, ok := args[0].(*value.File)
	if !ok {
		return value.NewFault("_file_read_bytes: expected file, got %s", args[0].Type())
	}
	if file.F == nil {
		return value.NewShapedError("io", "file is closed")
	}

	data, err := io.ReadAll(file.F)
	if err != nil {
		return value.NewShapedError("io", "%s", err.Error())
	}

	return &value.Bytes{Val: data}
}

// fileReaders caches buffered readers for line-by-line reading.
var fileReaders = make(map[*value.File]*bufio.Reader)

// halFileReadLine reads the next line from file.
// read_line(file) -> string | [@end] | [@error, :io, "message"]
func halFileReadLine(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("_file_read_line: want 1 argument, got %d", len(args))
	}
	file, ok := args[0].(*value.File)
	if !ok {
		return value.NewFault("_file_read_line: expected file, got %s", args[0].Type())
	}
	if file.F == nil {
		return value.NewShapedError("io", "file is closed")
	}

	// Get or create buffered reader for this file
	reader, exists := fileReaders[file]
	if !exists {
		reader = bufio.NewReader(file.F)
		fileReaders[file] = reader
	}

	line, err := reader.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			// Return any partial line content, or [@end] if nothing left
			if len(line) > 0 {
				return &value.String{Val: line}
			}
			return &value.List{Shape: "end", Elements: []value.Value{}}
		}
		return value.NewShapedError("io", "%s", err.Error())
	}

	// Strip trailing newline
	if len(line) > 0 && line[len(line)-1] == '\n' {
		line = line[:len(line)-1]
	}
	// Strip trailing \r if present (Windows line endings)
	if len(line) > 0 && line[len(line)-1] == '\r' {
		line = line[:len(line)-1]
	}

	return &value.String{Val: line}
}

// halFileWriteText writes a string to file.
// write_text(file, s) -> true | [@error, :io, "message"]
func halFileWriteText(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("_file_write_text: want 2 arguments, got %d", len(args))
	}
	file, ok := args[0].(*value.File)
	if !ok {
		return value.NewFault("_file_write_text: expected file, got %s", args[0].Type())
	}
	if file.F == nil {
		return value.NewShapedError("io", "file is closed")
	}
	s, ok := args[1].(*value.String)
	if !ok {
		return value.NewFault("_file_write_text: expected string, got %s", args[1].Type())
	}

	_, err := file.F.WriteString(s.Val)
	if err != nil {
		return value.NewShapedError("io", "%s", err.Error())
	}

	return value.TRUE
}

// halFileWriteBytes writes bytes to file.
// write_bytes(file, b) -> true | [@error, :io, "message"]
func halFileWriteBytes(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("_file_write_bytes: want 2 arguments, got %d", len(args))
	}
	file, ok := args[0].(*value.File)
	if !ok {
		return value.NewFault("_file_write_bytes: expected file, got %s", args[0].Type())
	}
	if file.F == nil {
		return value.NewShapedError("io", "file is closed")
	}
	b, ok := args[1].(*value.Bytes)
	if !ok {
		return value.NewFault("_file_write_bytes: expected bytes, got %s", args[1].Type())
	}

	_, err := file.F.Write(b.Val)
	if err != nil {
		return value.NewShapedError("io", "%s", err.Error())
	}

	return value.TRUE
}

// halFileClose closes a file.
// close(file) -> true | [@error, :io, "message"]
func halFileClose(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("_file_close: want 1 argument, got %d", len(args))
	}
	file, ok := args[0].(*value.File)
	if !ok {
		return value.NewFault("_file_close: expected file, got %s", args[0].Type())
	}
	if file.F == nil {
		return value.NewShapedError("io", "file is already closed")
	}

	// Clean up buffered reader if exists
	delete(fileReaders, file)

	err := file.F.Close()
	file.F = nil // Mark as closed

	if err != nil {
		return value.NewShapedError("io", "%s", err.Error())
	}

	return value.TRUE
}

// halFileExists checks if a path exists.
// exists(path) -> boolean
func halFileExists(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("_file_exists: want 1 argument, got %d", len(args))
	}
	path, ok := args[0].(*value.String)
	if !ok {
		return value.NewFault("_file_exists: expected string path, got %s", args[0].Type())
	}

	_, err := os.Stat(path.Val)
	if err == nil {
		return value.TRUE
	}
	if os.IsNotExist(err) {
		return value.FALSE
	}
	// Other errors (permission, etc.) - treat as not exists
	return value.FALSE
}

// halFileDelete deletes a file.
// delete(path) -> true | [@error, :io, "message"]
func halFileDelete(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("_file_delete: want 1 argument, got %d", len(args))
	}
	path, ok := args[0].(*value.String)
	if !ok {
		return value.NewFault("_file_delete: expected string path, got %s", args[0].Type())
	}

	err := os.Remove(path.Val)
	if err != nil {
		return value.NewShapedError("io", "%s", err.Error())
	}

	return value.TRUE
}
