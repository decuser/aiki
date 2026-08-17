package substrate

import (
	"sync"

	"github.com/gofrs/flock"

	"aiki/engine/runtime/hal"
	"aiki/engine/semantics/value"
)

type FileLockResource struct {
	Lock *flock.Flock
	mu   sync.Mutex
	done bool
}

func (r *FileLockResource) unlock() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.done {
		return nil
	}
	r.done = true
	return r.Lock.Unlock()
}

func (g *GoRuntime) newFileLock(path string, blocking bool) value.Value {
	l := flock.New(path)
	if blocking {
		if err := l.Lock(); err != nil {
			return value.NewShapedError("lock", "%v", err)
		}
	} else {
		ok, err := l.TryLock()
		if err != nil {
			return value.NewShapedError("lock", "%v", err)
		}
		if !ok {
			return value.FALSE
		}
	}
	g.mu.Lock()
	g.nextFileLockID++
	h := &value.FileLock{ID: g.nextFileLockID}
	g.fileLockResources[h] = &FileLockResource{Lock: l}
	g.mu.Unlock()
	return h
}

func (g *GoRuntime) halFileLock(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("_file_lock: want 1 argument, got %d", len(args))
	}
	path, ok := args[0].(*value.String)
	if !ok {
		return value.NewFault("_file_lock: expected string path, got %s", args[0].Type())
	}
	return g.newFileLock(g.resolveHostPath(path.Val), true)
}

func (g *GoRuntime) halFileTryLock(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("_file_try_lock: want 1 argument, got %d", len(args))
	}
	path, ok := args[0].(*value.String)
	if !ok {
		return value.NewFault("_file_try_lock: expected string path, got %s", args[0].Type())
	}
	return g.newFileLock(g.resolveHostPath(path.Val), false)
}

func (g *GoRuntime) halFileUnlock(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("_file_unlock: want 1 argument, got %d", len(args))
	}
	h, ok := args[0].(*value.FileLock)
	if !ok {
		return value.NewFault("_file_unlock: expected file lock, got %s", args[0].Type())
	}
	g.mu.RLock()
	r, ok := g.fileLockResources[h]
	g.mu.RUnlock()
	if !ok {
		return value.NewShapedError("lock", "file lock does not belong to this runtime")
	}
	if err := r.unlock(); err != nil {
		return value.NewShapedError("lock", "%v", err)
	}
	return value.TRUE
}

func (g *GoRuntime) CloseAllFileLocks() {
	g.mu.RLock()
	resources := make([]*FileLockResource, 0, len(g.fileLockResources))
	for _, r := range g.fileLockResources {
		resources = append(resources, r)
	}
	g.mu.RUnlock()
	for _, r := range resources {
		_ = r.unlock()
	}
}
