// Package substrate implements the Go runtime substrate for HAL.
package substrate

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/pprof"
	"sort"
	"sync"
	"sync/atomic"

	"aiki/engine"
	"aiki/engine/runtime/hal"
	"aiki/engine/semantics/value"
)

// Stdout and Stdin for I/O redirection.
var (
	Stdout  io.Writer  = os.Stdout
	Stdin   io.Reader  = os.Stdin
	UserEnv *value.Env // Set by REPL session for delete() to access
)

// BuiltinFunc is a HAL-level function that may use evaluation context.
// Simple builtins ignore ctx; intrinsics (import, export, apply, etc.) use it.
type BuiltinFunc func(args []value.Value, ctx *hal.EvalContext) value.Value

// Builtin is a HAL-level function implementing value.Callable.
type Builtin struct {
	name    string
	fn      BuiltinFunc
	runtime *GoRuntime // back-reference for context access
}

func (b *Builtin) Type() value.Type { return value.FunctionType }
func (b *Builtin) Inspect() string  { return fmt.Sprintf("<builtin: %s>", b.name) }

// Call invokes the builtin. Context is retrieved from the runtime's current context.
func (b *Builtin) Call(args []value.Value) value.Value {
	return b.fn(args, nil)
}

func (b *Builtin) CallWithContext(args []value.Value, ctx *hal.EvalContext) value.Value {
	return b.fn(args, ctx)
}

// Verify Builtin implements Callable
var _ value.Callable = (*Builtin)(nil)
var _ hal.ContextCallable = (*Builtin)(nil)

// GoRuntime implements hal.RuntimeContract using Go primitives.
// It maintains a single registry of _prefixed HAL primitives.
// User-visible names are defined in prelude.ai, not here.
type GoRuntime struct {
	registry      map[string]*Builtin // _print, _read, etc - prelude only
	mu            sync.RWMutex
	profileLabels atomic.Bool
	labelContexts sync.Map // map[engine.ProfileLabels]context.Context
	asyncFaults   chan *value.Fault
}

// Verify GoRuntime implements RuntimeContract
var _ hal.RuntimeContract = (*GoRuntime)(nil)
var _ hal.ProfileLabeler = (*GoRuntime)(nil)

// NewGoRuntime creates a new Go runtime substrate.
func NewGoRuntime() *GoRuntime {
	rt := &GoRuntime{
		registry:    make(map[string]*Builtin),
		asyncFaults: make(chan *value.Fault, 1),
	}
	rt.registerHAL()
	return rt
}

// Execute calls a registered primitive function.
// AsyncFaults returns the runtime's first-pending spawned fault channel.
func (g *GoRuntime) AsyncFaults() <-chan *value.Fault { return g.asyncFaults }

// ReportAsyncFault records a spawned fault without blocking the failing worker.
// A buffer of one preserves the first pending fault until a blocking operation
// observes it.
func (g *GoRuntime) ReportAsyncFault(fault *value.Fault) {
	if fault == nil {
		return
	}
	select {
	case g.asyncFaults <- fault:
	default:
	}
}

func (g *GoRuntime) Execute(name string, args []value.Value, ctx *hal.EvalContext) (value.Value, error) {
	g.mu.RLock()
	b, ok := g.registry[name]
	g.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("builtin %s not found", name)
	}

	result := b.fn(args, ctx)
	if fault, ok := result.(*value.Fault); ok {
		return nil, fmt.Errorf("%s", fault.Message)
	}
	return result, nil
}

// HasBuiltin checks if a name is visible at the given scope.
func (g *GoRuntime) HasBuiltin(name string, scope value.Scope) bool {
	// export, import, and use are available in all scopes (they're language primitives)
	if name == "export" || name == "import" || name == "use" {
		g.mu.RLock()
		_, ok := g.registry["_"+name]
		g.mu.RUnlock()
		return ok
	}

	// User scope cannot access any other builtins directly.
	// All user-visible functions come from prelude.ai bindings in Env.
	if scope == value.ScopeUser {
		return false
	}

	// Prelude scope can access _prefixed HAL primitives.
	g.mu.RLock()
	_, ok := g.registry[name]
	g.mu.RUnlock()
	return ok
}

// GetBuiltin returns a callable for the named builtin at the given scope.
func (g *GoRuntime) GetBuiltin(name string, scope value.Scope) (value.Callable, bool) {
	// export, import, and use are available in all scopes (they're language primitives)
	if name == "export" || name == "import" || name == "use" {
		g.mu.RLock()
		b, ok := g.registry["_"+name]
		g.mu.RUnlock()
		return b, ok
	}

	// User scope cannot access any other builtins directly.
	// All user-visible functions come from prelude.ai bindings in Env.
	if scope == value.ScopeUser {
		return nil, false
	}

	// Prelude scope can access _prefixed HAL primitives.
	g.mu.RLock()
	b, ok := g.registry[name]
	g.mu.RUnlock()
	return b, ok
}

// BuiltinNames returns the names that are visible to a given scope.
// This is intended for tooling like lint and fmt.
func (g *GoRuntime) BuiltinNames(scope value.Scope) []string {
	set := make(map[string]bool)

	// Language primitives that are always available.
	set["import"] = true
	set["use"] = true
	set["export"] = true

	// User scope cannot access _prefixed primitives directly.
	if scope == value.ScopeUser {
		out := make([]string, 0, len(set))
		for k := range set {
			out = append(out, k)
		}
		sort.Strings(out)
		return out
	}

	// Prelude scope can access _prefixed HAL primitives.
	g.mu.RLock()
	for name := range g.registry {
		set[name] = true
	}
	g.mu.RUnlock()

	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// register adds a builtin to the registry.
// Panics if fn is nil - catches stub registrations at startup.
func (g *GoRuntime) register(name string, fn BuiltinFunc) {
	if fn == nil {
		panic(fmt.Sprintf("HAL registration has nil function: %s", name))
	}
	g.registry[name] = &Builtin{name: name, fn: fn, runtime: g}
}

// resolveModulePath finds the .ai file for a module name.
func resolveModulePath(name string, env *value.Env) string {
	// Try relative to current file
	currentFile := env.GetFile()
	if currentFile != "" && currentFile != "<unknown>" {
		dir := filepath.Dir(currentFile)
		candidate := filepath.Join(dir, name+".ai")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	// Try as-is with .ai extension
	candidate := name + ".ai"
	if _, err := os.Stat(candidate); err == nil {
		return candidate
	}

	// Try lib/ directory relative to cwd
	libCandidate := filepath.Join("lib", name+".ai")
	if _, err := os.Stat(libCandidate); err == nil {
		return libCandidate
	}

	// Try lib/ directory relative to executable
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		libCandidate := filepath.Join(exeDir, "lib", name+".ai")
		if _, err := os.Stat(libCandidate); err == nil {
			return libCandidate
		}
	}

	// Try without extension
	if _, err := os.Stat(name); err == nil {
		return name
	}

	return ""
}

// SetProfileLabels enables or disables pprof label regions for this runtime.
func (g *GoRuntime) SetProfileLabels(enabled bool) {
	g.profileLabels.Store(enabled)
}

// WithProfileLabels executes fn under a cached pprof label context. It uses
// SetGoroutineLabels rather than pprof.Do so repeated Aiki calls do not rebuild
// context label maps on every invocation.
func (g *GoRuntime) WithProfileLabels(labels engine.ProfileLabels, restore engine.ProfileLabels, fn func()) {
	if !g.profileLabels.Load() {
		fn()
		return
	}
	pprof.SetGoroutineLabels(g.profileContext(labels))
	defer pprof.SetGoroutineLabels(g.profileContext(restore))
	fn()
}

func (g *GoRuntime) profileContext(labels engine.ProfileLabels) context.Context {
	if cached, ok := g.labelContexts.Load(labels); ok {
		return cached.(context.Context)
	}
	pairs := []string{
		"aiki_layer", labels.Layer,
		"aiki_function", labels.Function,
		"aiki_file", labels.File,
		"aiki_line", labels.Line,
		"aiki_primitive", labels.Primitive,
	}
	ctx := pprof.WithLabels(context.Background(), pprof.Labels(pairs...))
	actual, _ := g.labelContexts.LoadOrStore(labels, ctx)
	return actual.(context.Context)
}

func (b *Builtin) ProfileName() string { return b.name }
