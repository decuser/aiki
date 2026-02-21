// Package substrate implements the Go runtime substrate for HAL.
package substrate

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"aiki/engine/semantics/value"
)

// Stdout is the writer used by print. The REPL sets this to track output.
var Stdout io.Writer = os.Stdout

// GoRuntime implements hal.RuntimeContract using Go primitives.
type GoRuntime struct {
	registry map[string]func([]value.Value) (value.Value, error)
	mu       sync.RWMutex
}

// NewGoRuntime creates a new Go runtime substrate.
func NewGoRuntime() *GoRuntime {
	rt := &GoRuntime{
		registry: make(map[string]func([]value.Value) (value.Value, error)),
	}
	rt.registerDefaults()
	return rt
}

// Execute calls a registered primitive function.
func (g *GoRuntime) Execute(name string, args []value.Value) (value.Value, error) {
	g.mu.RLock()
	fn, ok := g.registry[name]
	g.mu.RUnlock()

	if !ok {
		return value.NullValue(), fmt.Errorf("native function %s not found in substrate", name)
	}

	return fn(args)
}

// HasBuiltin checks if a name is registered.
func (g *GoRuntime) HasBuiltin(name string) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	_, ok := g.registry[name]
	return ok
}

// Register adds a primitive function to the registry.
func (g *GoRuntime) Register(name string, fn func([]value.Value) (value.Value, error)) {
	g.mu.Lock()
	g.registry[name] = fn
	g.mu.Unlock()
}

// Spawn launches a goroutine.
func (g *GoRuntime) Spawn(task func()) {
	go task()
}

// MakeChannel creates a new channel value.
func (g *GoRuntime) MakeChannel() value.Value {
	ch := make(chan value.Value)
	return value.Value{Type: value.Channel, Data: ch}
}

// Send sends a value on a channel.
func (g *GoRuntime) Send(ch value.Value, val value.Value) error {
	if ch.Type != value.Channel {
		return fmt.Errorf("send: expected channel, got %s", value.TypeName(ch.Type))
	}
	c, ok := ch.Data.(chan value.Value)
	if !ok {
		return fmt.Errorf("send: invalid channel data")
	}
	c <- val
	return nil
}

// Recv receives a value from a channel.
func (g *GoRuntime) Recv(ch value.Value) (value.Value, error) {
	if ch.Type != value.Channel {
		return value.NullValue(), fmt.Errorf("recv: expected channel, got %s", value.TypeName(ch.Type))
	}
	c, ok := ch.Data.(chan value.Value)
	if !ok {
		return value.NullValue(), fmt.Errorf("recv: invalid channel data")
	}
	val := <-c
	return val, nil
}

// ReadFile reads a file from the filesystem.
func (g *GoRuntime) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// ResolvePath resolves a module name to a file path.
func (g *GoRuntime) ResolvePath(name string, relativeTo string) (string, error) {
	if relativeTo != "" {
		dir := filepath.Dir(relativeTo)
		candidate := filepath.Join(dir, name+".ai")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	candidate := name + ".ai"
	if _, err := os.Stat(candidate); err == nil {
		return candidate, nil
	}

	if _, err := os.Stat(name); err == nil {
		return name, nil
	}

	return "", fmt.Errorf("module not found: %s", name)
}

// LogError logs an error (from spawned goroutines).
func (g *GoRuntime) LogError(err error) {
	fmt.Fprintf(os.Stderr, "spawn error: %v\n", err)
}

// registerDefaults registers the default primitive functions.
func (g *GoRuntime) registerDefaults() {
	g.registerIO()
	g.registerList()
	g.registerMath()
	g.registerConvert()
	g.registerREPL()
}
