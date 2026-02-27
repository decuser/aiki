// Package substrate implements the Go runtime substrate for HAL.
package substrate

import (
	"fmt"
	"io"
	"os"
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
type GoRuntime struct {
	registry map[string]*Builtin
	mu       sync.RWMutex
}

// Verify GoRuntime implements RuntimeContract
var _ hal.RuntimeContract = (*GoRuntime)(nil)

// NewGoRuntime creates a new Go runtime substrate.
func NewGoRuntime() *GoRuntime {
	rt := &GoRuntime{
		registry: make(map[string]*Builtin),
	}
	rt.registerIO()
	rt.registerList()
	rt.registerType()
	rt.registerMath()
	return rt
}

// Execute calls a registered primitive function.
func (g *GoRuntime) Execute(name string, args []value.Value) (value.Value, error) {
	g.mu.RLock()
	b, ok := g.registry[name]
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

// HasBuiltin checks if a name is registered.
func (g *GoRuntime) HasBuiltin(name string) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	_, ok := g.registry[name]
	return ok
}

// GetBuiltin returns a callable for the named builtin.
func (g *GoRuntime) GetBuiltin(name string) (value.Callable, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	b, ok := g.registry[name]
	if !ok {
		return nil, false
	}
	return b, true
}

// register adds a builtin function.
func (g *GoRuntime) register(name string, fn func(args []value.Value) value.Value) {
	g.mu.Lock()
	g.registry[name] = &Builtin{name: name, fn: fn}
	g.mu.Unlock()
}
