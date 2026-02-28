// Package substrate implements the Go runtime substrate for HAL.
package substrate

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"aiki/engine/runtime/hal"
	"aiki/engine/semantics/value"
)

// Stdout and Stdin for I/O redirection.
var (
	Stdout io.Writer = os.Stdout
	Stdin  io.Reader = os.Stdin
)

// Builtin is a HAL-level function implementing value.Callable.
type Builtin struct {
	name string
	fn   func(args []value.Value) value.Value
}

func (b *Builtin) Type() value.Type                    { return value.FunctionType }
func (b *Builtin) Inspect() string                     { return fmt.Sprintf("<builtin: %s>", b.name) }
func (b *Builtin) Call(args []value.Value) value.Value { return b.fn(args) }

// Verify Builtin implements Callable
var _ value.Callable = (*Builtin)(nil)

// GoRuntime implements hal.RuntimeContract using Go primitives.
// It maintains three registries for scope-based visibility:
//   - halRegistry: _prefixed primitives, only visible to prelude
//   - preludeRegistry: prelude-defined wrappers, visible to user
//   - userRegistry: reserved for future user-defined builtins
type GoRuntime struct {
	halRegistry     map[string]*Builtin // _print, _read, etc - prelude only
	preludeRegistry map[string]*Builtin // print, read, etc - user visible
	mu              sync.RWMutex
}

// Verify GoRuntime implements RuntimeContract
var _ hal.RuntimeContract = (*GoRuntime)(nil)

// NewGoRuntime creates a new Go runtime substrate.
func NewGoRuntime() *GoRuntime {
	rt := &GoRuntime{
		halRegistry:     make(map[string]*Builtin),
		preludeRegistry: make(map[string]*Builtin),
	}
	rt.registerHAL()
	rt.registerPrelude()
	return rt
}

// Execute calls a registered primitive function.
func (g *GoRuntime) Execute(name string, args []value.Value) (value.Value, error) {
	g.mu.RLock()
	b, ok := g.halRegistry[name]
	if !ok {
		b, ok = g.preludeRegistry[name]
	}
	g.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("builtin %s not found", name)
	}

	result := b.Call(args)
	if err, ok := result.(*value.Error); ok {
		return nil, fmt.Errorf("%s", err.Message)
	}
	return result, nil
}

// HasBuiltin checks if a name is visible at the given scope.
func (g *GoRuntime) HasBuiltin(name string, scope hal.Scope) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()

	switch scope {
	case hal.ScopePrelude:
		// Prelude sees both HAL (_prefixed) and prelude registry
		if _, ok := g.halRegistry[name]; ok {
			return true
		}
		if _, ok := g.preludeRegistry[name]; ok {
			return true
		}
	case hal.ScopeUser:
		// User only sees prelude registry, not _prefixed
		if strings.HasPrefix(name, "_") {
			return false
		}
		if _, ok := g.preludeRegistry[name]; ok {
			return true
		}
	}
	return false
}

// GetBuiltin returns a callable for the named builtin at the given scope.
func (g *GoRuntime) GetBuiltin(name string, scope hal.Scope) (value.Callable, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	switch scope {
	case hal.ScopePrelude:
		// Prelude sees HAL first, then prelude registry
		if b, ok := g.halRegistry[name]; ok {
			return b, true
		}
		if b, ok := g.preludeRegistry[name]; ok {
			return b, true
		}
	case hal.ScopeUser:
		// User only sees prelude registry, not _prefixed
		if strings.HasPrefix(name, "_") {
			return nil, false
		}
		if b, ok := g.preludeRegistry[name]; ok {
			return b, true
		}
	}
	return nil, false
}

// registerPreludeBuiltin adds a prelude-visible builtin.
func (g *GoRuntime) registerPreludeBuiltin(name string, fn func(args []value.Value) value.Value) {
	g.mu.Lock()
	g.preludeRegistry[name] = &Builtin{name: name, fn: fn}
	g.mu.Unlock()
}
