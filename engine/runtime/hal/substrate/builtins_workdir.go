package substrate

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"aiki/engine/runtime/hal"
	"aiki/engine/semantics/value"
)

// resolveHostPath interprets relative host paths in the runtime-owned working
// directory rather than the Go process working directory.
func (g *GoRuntime) resolveHostPath(path string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	g.mu.RLock()
	cwd := g.workingDir
	g.mu.RUnlock()
	return filepath.Clean(filepath.Join(cwd, path))
}

func (g *GoRuntime) withResolvedPath(args []value.Value, indexes ...int) []value.Value {
	out := append([]value.Value(nil), args...)
	for _, i := range indexes {
		if i >= 0 && i < len(out) {
			if s, ok := out[i].(*value.String); ok {
				out[i] = &value.String{Val: g.resolveHostPath(s.Val)}
			}
		}
	}
	return out
}

func (g *GoRuntime) halFileOpenPath(args []value.Value, ctx *hal.EvalContext) value.Value {
	return halFileOpen(g.withResolvedPath(args, 0), ctx)
}
func (g *GoRuntime) halFileExistsPath(args []value.Value, ctx *hal.EvalContext) value.Value {
	return halFileExists(g.withResolvedPath(args, 0), ctx)
}
func (g *GoRuntime) halFileDeletePath(args []value.Value, ctx *hal.EvalContext) value.Value {
	return halFileDelete(g.withResolvedPath(args, 0), ctx)
}
func (g *GoRuntime) halFileListPath(args []value.Value, ctx *hal.EvalContext) value.Value {
	return halFileList(g.withResolvedPath(args, 0), ctx)
}
func (g *GoRuntime) halFileStatPath(args []value.Value, ctx *hal.EvalContext) value.Value {
	return halFileStat(g.withResolvedPath(args, 0), ctx)
}
func (g *GoRuntime) halFileRenamePath(args []value.Value, ctx *hal.EvalContext) value.Value {
	return halFileRename(g.withResolvedPath(args, 0, 1), ctx)
}
func (g *GoRuntime) halFileMkdirPath(args []value.Value, ctx *hal.EvalContext) value.Value {
	return halFileMkdir(g.withResolvedPath(args, 0), ctx)
}
func (g *GoRuntime) halFileMkdirAllPath(args []value.Value, ctx *hal.EvalContext) value.Value {
	return halFileMkdirAll(g.withResolvedPath(args, 0), ctx)
}
func (g *GoRuntime) halFileRemoveAllPath(args []value.Value, ctx *hal.EvalContext) value.Value {
	return halFileRemoveAll(g.withResolvedPath(args, 0), ctx)
}
func (g *GoRuntime) halFileCopyPath(args []value.Value, ctx *hal.EvalContext) value.Value {
	return halFileCopy(g.withResolvedPath(args, 0, 1), ctx)
}
func (g *GoRuntime) halFileSizePath(args []value.Value, ctx *hal.EvalContext) value.Value {
	return halFileSize(g.withResolvedPath(args, 0), ctx)
}

func (g *GoRuntime) halSystemCwd(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 0 {
		return value.NewFault("_system_cwd: want 0 arguments, got %d", len(args))
	}
	g.mu.RLock()
	cwd := g.workingDir
	g.mu.RUnlock()
	return &value.String{Val: cwd}
}

func (g *GoRuntime) halSystemChdir(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("_system_chdir: want 1 argument, got %d", len(args))
	}
	path, ok := args[0].(*value.String)
	if !ok {
		return value.NewFault("_system_chdir: expected string path, got %s", args[0].Type())
	}
	resolved := g.resolveHostPath(path.Val)
	info, err := os.Stat(resolved)
	if err != nil {
		return value.NewShapedError("io", "%s", err.Error())
	}
	if !info.IsDir() {
		return value.NewShapedError("io", "not a directory: %s", path.Val)
	}
	g.mu.Lock()
	g.workingDir = resolved
	g.mu.Unlock()
	return value.TRUE
}

func (g *GoRuntime) halSystemExec(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("_system_exec: want 2 arguments, got %d", len(args))
	}
	name, ok := args[0].(*value.String)
	if !ok {
		return value.NewFault("_system_exec: expected string command, got %s", args[0].Type())
	}
	list, ok := args[1].(*value.List)
	if !ok {
		return value.NewFault("_system_exec: expected list of string arguments, got %s", args[1].Type())
	}
	argv := make([]string, len(list.Elements))
	for i, elem := range list.Elements {
		s, ok := elem.(*value.String)
		if !ok {
			return value.NewFault("_system_exec: argument %d must be string, got %s", i, elem.Type())
		}
		argv[i] = s.Val
	}

	g.mu.RLock()
	cwd := g.workingDir
	g.mu.RUnlock()
	command := name.Val
	if !filepath.IsAbs(command) && strings.ContainsAny(command, `/\\`) {
		command = filepath.Join(cwd, command)
	}
	cmd := exec.Command(command, argv...)
	cmd.Dir = cwd
	env := g.environmentSnapshot()
	env["PWD"] = cwd
	cmd.Env = environmentList(env)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	exitCode := int64(0)
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = int64(exitErr.ExitCode())
		} else {
			return value.NewShapedError("process", "%s", err.Error())
		}
	}
	return &value.List{Shape: "ok", Elements: []value.Value{
		&value.String{Val: stdout.String()},
		&value.String{Val: stderr.String()},
		value.NewNumber(exitCode, 1),
	}}
}

func halPathSeparator(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 0 {
		return value.NewFault("_path_separator: want 0 arguments, got %d", len(args))
	}
	return &value.String{Val: string(filepath.Separator)}
}

func (g *GoRuntime) halFileWalkPath(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return halFileWalk(args, ctx)
	}
	original, ok := args[0].(*value.String)
	if !ok {
		return halFileWalk(args, ctx)
	}
	resolved := g.resolveHostPath(original.Val)
	result := halFileWalk([]value.Value{&value.String{Val: resolved}}, ctx)
	list, ok := result.(*value.List)
	if !ok || filepath.IsAbs(original.Val) {
		return result
	}
	g.mu.RLock()
	cwd := g.workingDir
	g.mu.RUnlock()
	elements := make([]value.Value, len(list.Elements))
	copy(elements, list.Elements)
	for i, elem := range elements {
		path, ok := elem.(*value.String)
		if !ok {
			continue
		}
		rel, err := filepath.Rel(cwd, path.Val)
		if err == nil {
			elements[i] = &value.String{Val: rel}
		}
	}
	return &value.List{Elements: elements, Shape: list.Shape}
}

func (g *GoRuntime) halFileSymlinkPath(args []value.Value, ctx *hal.EvalContext) value.Value {
	// Preserve the target exactly: a relative symlink target is interpreted
	// relative to the link's directory by the host filesystem.
	return halFileSymlink(g.withResolvedPath(args, 1), ctx)
}

func (g *GoRuntime) halFileReadLinkPath(args []value.Value, ctx *hal.EvalContext) value.Value {
	return halFileReadLink(g.withResolvedPath(args, 0), ctx)
}

func (g *GoRuntime) halFilePermissionsPath(args []value.Value, ctx *hal.EvalContext) value.Value {
	return halFilePermissions(g.withResolvedPath(args, 0), ctx)
}

func (g *GoRuntime) halFileChmodPath(args []value.Value, ctx *hal.EvalContext) value.Value {
	return halFileChmod(g.withResolvedPath(args, 0), ctx)
}
