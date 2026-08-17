package substrate

import (
	"bufio"
	"io"

	"aiki/engine/semantics/value"
)

func (g *GoRuntime) ioReader(endpoint value.Value) (io.Reader, string, *value.List) {
	switch v := endpoint.(type) {
	case *value.Symbol:
		if v.Val != "stdin" {
			return nil, "", value.NewShapedError("io", ":%s is not readable", v.Val)
		}
		g.mu.RLock()
		reader := g.stdinReader
		g.mu.RUnlock()
		return reader, "stdin", nil
	case *value.File:
		if v.F == nil {
			return nil, "", value.NewShapedError("io", "file is closed")
		}
		g.mu.Lock()
		reader, ok := g.fileReaders[v]
		if !ok {
			reader = bufio.NewReader(v.F)
			g.fileReaders[v] = reader
		}
		g.mu.Unlock()
		return reader, v.Path, nil
	default:
		return nil, "", value.NewShapedError("io", "expected :stdin or file, got %s", endpoint.Type())
	}
}

func (g *GoRuntime) ioWriter(endpoint value.Value) (io.Writer, string, *value.List) {
	switch v := endpoint.(type) {
	case *value.Symbol:
		g.mu.RLock()
		defer g.mu.RUnlock()
		switch v.Val {
		case "stdout":
			return g.stdout, "stdout", nil
		case "stderr":
			return g.stderr, "stderr", nil
		default:
			return nil, "", value.NewShapedError("io", ":%s is not writable", v.Val)
		}
	case *value.File:
		if v.F == nil {
			return nil, "", value.NewShapedError("io", "file is closed")
		}
		return v.F, v.Path, nil
	default:
		return nil, "", value.NewShapedError("io", "expected :stdout, :stderr, or file, got %s", endpoint.Type())
	}
}
