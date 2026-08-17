package substrate

import (
	"errors"
	"io"
	"os"
	"sync"

	"github.com/chzyer/readline"

	"aiki/engine/runtime/hal"
	"aiki/engine/semantics/value"
)

type TerminalResource struct {
	File  *os.File
	State *readline.State
	mu    sync.Mutex
	done  bool
}

func (r *TerminalResource) restore() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.done {
		return nil
	}
	r.done = true
	return readline.Restore(int(r.File.Fd()), r.State)
}

func fileFromIO(v any) (*os.File, bool) {
	f, ok := v.(*os.File)
	return f, ok
}

func (g *GoRuntime) terminalFile(v value.Value) (*os.File, *value.List) {
	switch x := v.(type) {
	case *value.Symbol:
		g.mu.RLock()
		defer g.mu.RUnlock()
		switch x.Val {
		case "stdin":
			if f, ok := fileFromIO(g.stdin); ok {
				return f, nil
			}
		case "stdout":
			if f, ok := fileFromIO(g.stdout); ok {
				return f, nil
			}
		case "stderr":
			if f, ok := fileFromIO(g.stderr); ok {
				return f, nil
			}
		default:
			return nil, value.NewShapedError("terminal", "unknown standard endpoint :%s", x.Val)
		}
		return nil, value.NewShapedError("terminal", ":%s is not backed by a terminal-capable file", x.Val)

	case *value.File:
		if x.F == nil {
			return nil, value.NewShapedError("terminal", "file is closed")
		}
		return x.F, nil

	case *value.Endpoint:
		r, ok := g.endpointResource(x)
		if !ok {
			return nil, value.NewShapedError("terminal", "endpoint does not belong to this runtime")
		}
		if r.isClosed() {
			return nil, value.NewShapedError("terminal", "endpoint is closed")
		}
		for _, candidate := range []any{r.Reader, r.Writer, r.Closer} {
			if f, ok := fileFromIO(candidate); ok {
				return f, nil
			}
		}
		return nil, value.NewShapedError("terminal", "%s is not backed by a terminal-capable file", x.Inspect())

	default:
		return nil, value.NewShapedError("terminal", "expected standard endpoint, file, or endpoint, got %s", v.Type())
	}
}

func (g *GoRuntime) halTermIs(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("term.is: want 1 argument, got %d", len(args))
	}
	f, errv := g.terminalFile(args[0])
	if errv != nil {
		return value.FALSE
	}
	if readline.IsTerminal(int(f.Fd())) {
		return value.TRUE
	}
	return value.FALSE
}

func (g *GoRuntime) halTermSize(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("term.size: want 1 argument, got %d", len(args))
	}
	f, errv := g.terminalFile(args[0])
	if errv != nil {
		return errv
	}
	if !readline.IsTerminal(int(f.Fd())) {
		return value.NewShapedError("terminal", "endpoint is not a terminal")
	}
	w, h, err := readline.GetSize(int(f.Fd()))
	if err != nil {
		return value.NewShapedError("terminal", "size: %v", err)
	}
	return &value.List{Shape: "terminal_size", Elements: []value.Value{
		value.NewNumber(int64(w), 1), value.NewNumber(int64(h), 1),
	}}
}

func (g *GoRuntime) halTermRaw(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("term.raw: want 1 argument, got %d", len(args))
	}
	f, errv := g.terminalFile(args[0])
	if errv != nil {
		return errv
	}
	if !readline.IsTerminal(int(f.Fd())) {
		return value.NewShapedError("terminal", "endpoint is not a terminal")
	}
	state, err := readline.MakeRaw(int(f.Fd()))
	if err != nil {
		return value.NewShapedError("terminal", "raw: %v", err)
	}
	g.mu.Lock()
	g.nextTerminalID++
	h := &value.TerminalState{ID: g.nextTerminalID}
	g.terminalResources[h] = &TerminalResource{File: f, State: state}
	g.mu.Unlock()
	return h
}

func (g *GoRuntime) halTermRestore(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("term.restore: want 1 argument, got %d", len(args))
	}
	h, ok := args[0].(*value.TerminalState)
	if !ok {
		return value.NewFault("term.restore: expected terminal state, got %s", args[0].Type())
	}
	g.mu.RLock()
	r, ok := g.terminalResources[h]
	g.mu.RUnlock()
	if !ok {
		return value.NewShapedError("terminal", "terminal state does not belong to this runtime")
	}
	if err := r.restore(); err != nil && !errors.Is(err, io.EOF) {
		return value.NewShapedError("terminal", "restore: %v", err)
	}
	return value.TRUE
}

func (g *GoRuntime) CloseAllTerminals() {
	g.mu.RLock()
	resources := make([]*TerminalResource, 0, len(g.terminalResources))
	for _, r := range g.terminalResources {
		resources = append(resources, r)
	}
	g.mu.RUnlock()
	for _, r := range resources {
		_ = r.restore()
	}
}
