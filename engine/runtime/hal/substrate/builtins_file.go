package substrate

import (
	"bufio"
	"io"
	"os"
	"path/filepath"

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
	case "read_write":
		f, err = os.OpenFile(path.Val, os.O_RDWR|os.O_CREATE, 0644)
		modeStr = "read_write"
	default:
		return value.NewShapedError("io", "invalid mode: %s (expected :read, :write, :append, or :read_write)", mode.Val)
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

// halFileReadLine reads the next line from file.
// read_line(file) -> string | [@end] | [@error, :io, "message"]
func (g *GoRuntime) halFileReadLine(args []value.Value, ctx *hal.EvalContext) value.Value {
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
	g.mu.Lock()
	reader, exists := g.fileReaders[file]
	if !exists {
		reader = bufio.NewReader(file.F)
		g.fileReaders[file] = reader
	}
	g.mu.Unlock()

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
func (g *GoRuntime) halFileClose(args []value.Value, ctx *hal.EvalContext) value.Value {
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
	g.mu.Lock()
	delete(g.fileReaders, file)
	g.mu.Unlock()

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

// halFileList returns the immediate entries in a directory, sorted by name.
// list(path) -> [string, ...] | [@error, :io, "message"]
func halFileList(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("_file_list: want 1 argument, got %d", len(args))
	}
	path, ok := args[0].(*value.String)
	if !ok {
		return value.NewFault("_file_list: expected string path, got %s", args[0].Type())
	}
	entries, err := os.ReadDir(path.Val)
	if err != nil {
		return value.NewShapedError("io", "%s", err.Error())
	}
	elems := make([]value.Value, len(entries))
	for i, entry := range entries {
		elems[i] = &value.String{Val: entry.Name()}
	}
	return &value.List{Elements: elems}
}

// halFileReadAt reads up to count bytes beginning at offset without changing
// the file's sequential cursor.
func halFileReadAt(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 3 {
		return value.NewFault("_file_read_at: want 3 arguments, got %d", len(args))
	}
	file, ok := args[0].(*value.File)
	if !ok {
		return value.NewFault("_file_read_at: expected file, got %s", args[0].Type())
	}
	if file.F == nil {
		return value.NewShapedError("io", "file is closed")
	}
	offset, ok := fileNonNegativeInt64(args[1])
	if !ok {
		return value.NewFault("_file_read_at: offset must be a non-negative integer")
	}
	count64, ok := fileNonNegativeInt64(args[2])
	if !ok {
		return value.NewFault("_file_read_at: count must be a non-negative integer")
	}
	if count64 > int64(^uint(0)>>1) {
		return value.NewShapedError("io", "read count too large")
	}
	buf := make([]byte, int(count64))
	n, err := file.F.ReadAt(buf, offset)
	if err != nil && err != io.EOF {
		return value.NewShapedError("io", "%s", err.Error())
	}
	return &value.Bytes{Val: buf[:n]}
}

// halFileWriteAt writes bytes beginning at offset without changing the file's
// sequential cursor.
func halFileWriteAt(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 3 {
		return value.NewFault("_file_write_at: want 3 arguments, got %d", len(args))
	}
	file, ok := args[0].(*value.File)
	if !ok {
		return value.NewFault("_file_write_at: expected file, got %s", args[0].Type())
	}
	if file.F == nil {
		return value.NewShapedError("io", "file is closed")
	}
	offset, ok := fileNonNegativeInt64(args[1])
	if !ok {
		return value.NewFault("_file_write_at: offset must be a non-negative integer")
	}
	data, ok := args[2].(*value.Bytes)
	if !ok {
		return value.NewFault("_file_write_at: expected bytes, got %s", args[2].Type())
	}
	n, err := file.F.WriteAt(data.Val, offset)
	if err != nil {
		return value.NewShapedError("io", "%s", err.Error())
	}
	if n != len(data.Val) {
		return value.NewShapedError("io", "short write: wrote %d of %d bytes", n, len(data.Val))
	}
	return value.TRUE
}

func fileNonNegativeInt64(v value.Value) (int64, bool) {
	n, ok := v.(*value.Number)
	if !ok || !n.Val.IsInt() || !n.Val.Num().IsInt64() {
		return 0, false
	}
	i := n.Val.Num().Int64()
	return i, i >= 0
}

// halFileStat returns basic portable metadata for path.
// stat(path) -> [@stat, size, modified_ms, is_dir] | [@error, :io, "message"]
func halFileStat(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("_file_stat: want 1 argument, got %d", len(args))
	}
	path, ok := args[0].(*value.String)
	if !ok {
		return value.NewFault("_file_stat: expected string path, got %s", args[0].Type())
	}
	info, err := os.Stat(path.Val)
	if err != nil {
		return value.NewShapedError("io", "%s", err.Error())
	}
	isDir := value.FALSE
	if info.IsDir() {
		isDir = value.TRUE
	}
	return &value.List{Shape: "stat", Elements: []value.Value{
		value.NewNumber(info.Size(), 1),
		value.NewNumber(info.ModTime().UnixMilli(), 1),
		isDir,
	}}
}

func halFileRename(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("_file_rename: want 2 arguments, got %d", len(args))
	}
	oldPath, ok := args[0].(*value.String)
	if !ok {
		return value.NewFault("_file_rename: expected string old path, got %s", args[0].Type())
	}
	newPath, ok := args[1].(*value.String)
	if !ok {
		return value.NewFault("_file_rename: expected string new path, got %s", args[1].Type())
	}
	if err := os.Rename(oldPath.Val, newPath.Val); err != nil {
		return value.NewShapedError("io", "%s", err.Error())
	}
	return value.TRUE
}

func halFileMkdir(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("_file_mkdir: want 1 argument, got %d", len(args))
	}
	path, ok := args[0].(*value.String)
	if !ok {
		return value.NewFault("_file_mkdir: expected string path, got %s", args[0].Type())
	}
	if err := os.Mkdir(path.Val, 0755); err != nil {
		return value.NewShapedError("io", "%s", err.Error())
	}
	return value.TRUE
}

func halFileMkdirAll(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("_file_mkdir_all: want 1 argument, got %d", len(args))
	}
	path, ok := args[0].(*value.String)
	if !ok {
		return value.NewFault("_file_mkdir_all: expected string path, got %s", args[0].Type())
	}
	if err := os.MkdirAll(path.Val, 0755); err != nil {
		return value.NewShapedError("io", "%s", err.Error())
	}
	return value.TRUE
}

func halFileRemoveAll(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("_file_remove_all: want 1 argument, got %d", len(args))
	}
	path, ok := args[0].(*value.String)
	if !ok {
		return value.NewFault("_file_remove_all: expected string path, got %s", args[0].Type())
	}
	if err := os.RemoveAll(path.Val); err != nil {
		return value.NewShapedError("io", "%s", err.Error())
	}
	return value.TRUE
}

func halFileTemp(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 0 {
		return value.NewFault("_file_temp: want 0 arguments, got %d", len(args))
	}
	f, err := os.CreateTemp("", "aiki-")
	if err != nil {
		return value.NewShapedError("io", "%s", err.Error())
	}
	name := f.Name()
	if err := f.Close(); err != nil {
		os.Remove(name)
		return value.NewShapedError("io", "%s", err.Error())
	}
	return &value.String{Val: name}
}

func halFileTempDir(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 0 {
		return value.NewFault("_file_temp_dir: want 0 arguments, got %d", len(args))
	}
	name, err := os.MkdirTemp("", "aiki-")
	if err != nil {
		return value.NewShapedError("io", "%s", err.Error())
	}
	return &value.String{Val: name}
}

func halFileCopy(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("_file_copy: want 2 arguments, got %d", len(args))
	}
	src, ok := args[0].(*value.String)
	if !ok {
		return value.NewFault("_file_copy: expected string source, got %s", args[0].Type())
	}
	dst, ok := args[1].(*value.String)
	if !ok {
		return value.NewFault("_file_copy: expected string destination, got %s", args[1].Type())
	}
	in, err := os.Open(src.Val)
	if err != nil {
		return value.NewShapedError("io", "%s", err.Error())
	}
	defer in.Close()
	out, err := os.Create(dst.Val)
	if err != nil {
		return value.NewShapedError("io", "%s", err.Error())
	}
	_, copyErr := io.Copy(out, in)
	closeErr := out.Close()
	if copyErr != nil {
		return value.NewShapedError("io", "%s", copyErr.Error())
	}
	if closeErr != nil {
		return value.NewShapedError("io", "%s", closeErr.Error())
	}
	return value.TRUE
}

func halFileSize(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("_file_size: want 1 argument, got %d", len(args))
	}
	path, ok := args[0].(*value.String)
	if !ok {
		return value.NewFault("_file_size: expected string path, got %s", args[0].Type())
	}
	info, err := os.Stat(path.Val)
	if err != nil {
		return value.NewShapedError("io", "%s", err.Error())
	}
	return value.NewNumber(info.Size(), 1)
}

// halFileWalk returns root and its descendants without following directory
// symlinks. Paths use the host representation returned by filepath.WalkDir.
func halFileWalk(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("_file_walk: want 1 argument, got %d", len(args))
	}
	root, ok := args[0].(*value.String)
	if !ok {
		return value.NewFault("_file_walk: expected string path, got %s", args[0].Type())
	}
	paths := []value.Value{}
	err := filepath.WalkDir(root.Val, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		paths = append(paths, &value.String{Val: path})
		return nil
	})
	if err != nil {
		return value.NewShapedError("io", "%s", err.Error())
	}
	return &value.List{Elements: paths}
}

func halFileSymlink(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("_file_symlink: want 2 arguments, got %d", len(args))
	}
	target, ok := args[0].(*value.String)
	if !ok {
		return value.NewFault("_file_symlink: expected string target, got %s", args[0].Type())
	}
	link, ok := args[1].(*value.String)
	if !ok {
		return value.NewFault("_file_symlink: expected string link path, got %s", args[1].Type())
	}
	if err := os.Symlink(target.Val, link.Val); err != nil {
		return value.NewShapedError("io", "%s", err.Error())
	}
	return value.TRUE
}

func halFileReadLink(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("_file_read_link: want 1 argument, got %d", len(args))
	}
	path, ok := args[0].(*value.String)
	if !ok {
		return value.NewFault("_file_read_link: expected string path, got %s", args[0].Type())
	}
	target, err := os.Readlink(path.Val)
	if err != nil {
		return value.NewShapedError("io", "%s", err.Error())
	}
	return &value.String{Val: target}
}

// halFilePermissions returns the host permission bits represented by Go's
// portable FileMode.Perm vocabulary.
func halFilePermissions(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("_file_permissions: want 1 argument, got %d", len(args))
	}
	path, ok := args[0].(*value.String)
	if !ok {
		return value.NewFault("_file_permissions: expected string path, got %s", args[0].Type())
	}
	info, err := os.Stat(path.Val)
	if err != nil {
		return value.NewShapedError("io", "%s", err.Error())
	}
	return value.NewNumber(int64(info.Mode().Perm()), 1)
}

func halFileChmod(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("_file_chmod: want 2 arguments, got %d", len(args))
	}
	path, ok := args[0].(*value.String)
	if !ok {
		return value.NewFault("_file_chmod: expected string path, got %s", args[0].Type())
	}
	mode, ok := args[1].(*value.Number)
	if !ok || !mode.Val.IsInt() {
		return value.NewFault("_file_chmod: mode must be an integer")
	}
	m := mode.Val.Num().Int64()
	if m < 0 || m > 0777 {
		return value.NewShapedError("io", "permission mode out of range: %d", m)
	}
	if err := os.Chmod(path.Val, os.FileMode(m)); err != nil {
		return value.NewShapedError("io", "%s", err.Error())
	}
	return value.TRUE
}
