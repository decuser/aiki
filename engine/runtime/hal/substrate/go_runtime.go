// Package substrate implements the Go runtime substrate for HAL.
package substrate

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"runtime/pprof"
	"sync"
	"sync/atomic"
	"time"

	"aiki/engine"
	"aiki/engine/runtime/hal"
	"aiki/engine/runtime/help"
	"aiki/engine/runtime/modules"
	"aiki/engine/runtime/primitives"
	"aiki/engine/semantics/value"
)

// BuiltinFunc is a HAL-level function that may use evaluation context.
// Simple builtins ignore ctx; intrinsics (import, export, apply, etc.) use it.
type BuiltinFunc func(args []value.Value, ctx *hal.EvalContext) value.Value
type ProbeBuiltinFunc func(args []value.Value, probe engine.SemanticProbe) value.Value

// Builtin is a HAL-level function implementing value.Callable.
type Builtin struct {
	name         string
	fn           BuiltinFunc
	runtime      *GoRuntime // back-reference for context access
	needsContext bool
	probeFn      ProbeBuiltinFunc
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

func (b *Builtin) NeedsEvalContext() bool      { return b.needsContext }
func (b *Builtin) NeedsRealizationProbe() bool { return b.probeFn != nil }
func (b *Builtin) CallWithProbe(args []value.Value, probe engine.SemanticProbe) value.Value {
	if b.probeFn != nil {
		return b.probeFn(args, probe)
	}
	return b.Call(args)
}

// Verify Builtin implements Callable
var _ value.Callable = (*Builtin)(nil)
var _ hal.ContextCallable = (*Builtin)(nil)
var _ hal.EvalContextRequired = (*Builtin)(nil)
var _ hal.ProbeCallable = (*Builtin)(nil)
var _ hal.RealizationProbeRequired = (*Builtin)(nil)

// GoRuntime implements hal.RuntimeContract using Go substrate bindings. Native
// machinery is separated by architectural role, and only canonical host
// operations populate hostBindings. User-visible names are defined in Aiki.
type GoRuntime struct {
	intrinsics        map[string]*Builtin
	runtimePrimitives map[string]*Builtin
	providers         map[string]*Builtin
	hostRegistry      map[string]*Builtin
	services          map[string]*Builtin
	hostBindings      map[string]hal.HostOperation
	stdin             io.Reader
	stdinReader       *bufio.Reader
	stdout            io.Writer
	stderr            io.Writer
	pageOutput        func(string) bool
	userEnv           *value.Env
	moduleRegistry    *modules.ModuleRegistry
	helpRegistry      *help.Registry
	openCanvases      []*value.Canvas
	canvasResources   map[*value.Canvas]*CanvasResource
	nextCanvasID      uint64
	programArgs       []string
	workingDir        string
	environment       map[string]string
	envLookup         func(string) (string, bool)
	rng               *rand.Rand
	fileReaders       map[*value.File]*bufio.Reader
	processResources  map[*value.Process]*ProcessResource
	endpointResources map[*value.Endpoint]*EndpointResource
	signalResources   map[*value.Channel]*SignalResource
	listenerResources map[*value.Listener]*ListenerResource
	datagramResources map[*value.Datagram]*DatagramResource
	terminalResources map[*value.TerminalState]*TerminalResource
	fileLockResources map[*value.FileLock]*FileLockResource
	nextProcessID     uint64
	nextEndpointID    uint64
	nextListenerID    uint64
	nextDatagramID    uint64
	nextTerminalID    uint64
	nextFileLockID    uint64
	testState         runtimeTestState
	mu                sync.RWMutex
	profileLabels     atomic.Bool
	labelContexts     sync.Map // map[engine.ProfileLabels]context.Context
	asyncFaults       chan *value.Fault
}

// Verify GoRuntime implements RuntimeContract
var _ hal.RuntimeContract = (*GoRuntime)(nil)
var _ hal.HostOperationProvider = (*GoRuntime)(nil)
var _ hal.ProfileLabeler = (*GoRuntime)(nil)

// NewGoRuntime creates a new Go runtime substrate.
func NewGoRuntime() *GoRuntime {
	workingDir, err := os.Getwd()
	if err != nil {
		workingDir = "."
	}
	rt := &GoRuntime{
		intrinsics:        make(map[string]*Builtin),
		runtimePrimitives: make(map[string]*Builtin),
		providers:         make(map[string]*Builtin),
		hostRegistry:      make(map[string]*Builtin),
		services:          make(map[string]*Builtin),
		hostBindings:      make(map[string]hal.HostOperation),
		canvasResources:   make(map[*value.Canvas]*CanvasResource),
		stdin:             os.Stdin,
		stdinReader:       bufio.NewReader(os.Stdin),
		stdout:            os.Stdout,
		stderr:            os.Stderr,
		workingDir:        workingDir,
		environment:       snapshotHostEnvironment(),
		rng:               rand.New(rand.NewSource(time.Now().UnixNano())),
		fileReaders:       make(map[*value.File]*bufio.Reader),
		processResources:  make(map[*value.Process]*ProcessResource),
		endpointResources: make(map[*value.Endpoint]*EndpointResource),
		signalResources:   make(map[*value.Channel]*SignalResource),
		listenerResources: make(map[*value.Listener]*ListenerResource),
		datagramResources: make(map[*value.Datagram]*DatagramResource),
		terminalResources: make(map[*value.TerminalState]*TerminalResource),
		fileLockResources: make(map[*value.FileLock]*FileLockResource),
		asyncFaults:       make(chan *value.Fault, 1),
	}
	rt.registerHAL()
	if err := rt.ValidateProfile(hal.DefaultRuntimeProfile); err != nil {
		panic(err)
	}
	return rt
}

// SetIO replaces the runtime-owned standard input/output endpoints. Nil values
// restore the corresponding process defaults.
func (g *GoRuntime) SetIO(in io.Reader, out io.Writer) {
	g.mu.Lock()
	if in == nil {
		in = os.Stdin
	}
	if out == nil {
		out = os.Stdout
	}
	g.stdin = in
	g.stdinReader = bufio.NewReader(in)
	g.stdout = out
	g.mu.Unlock()
}

// SetErrorOutput replaces the runtime-owned standard error endpoint. Nil
// restores the process default.
func (g *GoRuntime) SetErrorOutput(out io.Writer) {
	g.mu.Lock()
	if out == nil {
		out = os.Stderr
	}
	g.stderr = out
	g.mu.Unlock()
}

// SetPageOutput installs the optional runtime-owned pageable-text presenter.
func (g *GoRuntime) SetPageOutput(fn func(string) bool) {
	g.mu.Lock()
	g.pageOutput = fn
	g.mu.Unlock()
}

// SetUserEnv sets the session user environment used by REPL-only services such as delete.
func (g *GoRuntime) SetUserEnv(env *value.Env) {
	g.mu.Lock()
	g.userEnv = env
	g.mu.Unlock()
}

// SetModuleRegistry installs the runtime-owned module registry/cache.
func (g *GoRuntime) SetModuleRegistry(registry *modules.ModuleRegistry) {
	g.mu.Lock()
	g.moduleRegistry = registry
	g.mu.Unlock()
}

// SetHelpRegistry installs the runtime-owned prelude help/documentation registry.
func (g *GoRuntime) SetHelpRegistry(registry *help.Registry) {
	g.mu.Lock()
	g.helpRegistry = registry
	g.mu.Unlock()
}

// SetProgramArgs replaces the runtime-owned argument snapshot visible to system.args.
func (g *GoRuntime) SetProgramArgs(args []string) {
	g.mu.Lock()
	g.programArgs = append(g.programArgs[:0], args...)
	g.mu.Unlock()
}

// SetEnvLookup installs an optional environment lookup override for embedding/tests.
// A nil lookup restores the runtime-owned environment view. Child processes
// always inherit the runtime-owned environment snapshot.
func (g *GoRuntime) SetEnvLookup(lookup func(string) (string, bool)) {
	g.mu.Lock()
	g.envLookup = lookup
	g.mu.Unlock()
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
	b, ok := g.lookupBuiltin(name)
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

// authorityKey returns the architectural grant checked for a runtime binding.
// Canonical host operations are authorized by HAL identity; non-HAL language,
// provider, and service primitives retain their implementation primitive name.
func (g *GoRuntime) authorityKey(name string) string {
	g.mu.RLock()
	op, ok := g.hostBindings[name]
	g.mu.RUnlock()
	if ok {
		return op.Authority
	}
	return name
}

// HasBuiltin checks whether a runtime binding is executable under the supplied
// lexical authority. Language intrinsics import/use/export remain constitutive
// and do not consume a grant. Scope is intentionally not consulted.
func (g *GoRuntime) HasBuiltin(name string, authority value.Authority) bool {
	if name == "export" || name == "import" || name == "use" {
		g.mu.RLock()
		_, ok := g.intrinsics["_"+name]
		g.mu.RUnlock()
		return ok
	}
	if !authority.Allows(g.authorityKey(name)) {
		return false
	}
	g.mu.RLock()
	_, ok := g.lookupBuiltin(name)
	g.mu.RUnlock()
	return ok
}

// GetBuiltin resolves a runtime binding only when lexical authority grants its
// architectural key. Host operations use canonical HAL identities; other
// primitives use their implementation primitive names.
func (g *GoRuntime) GetBuiltin(name string, authority value.Authority) (value.Callable, bool) {
	if name == "export" || name == "import" || name == "use" {
		g.mu.RLock()
		b, ok := g.intrinsics["_"+name]
		g.mu.RUnlock()
		return b, ok
	}
	if !authority.Allows(g.authorityKey(name)) {
		return nil, false
	}
	g.mu.RLock()
	b, ok := g.lookupBuiltin(name)
	g.mu.RUnlock()
	return b, ok
}

// BuiltinNames returns the architectural runtime names visible to tooling.
// Primitive classification and scope visibility are engine-owned; the Go
// substrate binds implementations to that vocabulary.
func (g *GoRuntime) BuiltinNames(scope value.Scope) []string {
	return primitives.NamesForScope(scope)
}

// registerHost registers an existing host-role compatibility primitive and
// attaches the canonical host-operation contract that it realizes. Not every
// host-role primitive has canonical metadata yet; canonical descriptors are a
// stricter subset of this compatibility registry.
func (g *GoRuntime) registerHost(op hal.HostOperation, fn BuiltinFunc) {
	if op.Identity == "" || op.Primitive == "" || op.SubstrateProvenance == "" {
		panic("host operation registration requires identity, primitive, and substrate provenance")
	}
	if _, exists := g.hostBindings[op.Primitive]; exists {
		panic(fmt.Sprintf("duplicate host operation registration: %s", op.Primitive))
	}
	g.registerPrimitive(op.Primitive, fn)
	g.hostBindings[op.Primitive] = op
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

func (g *GoRuntime) ProfileLabelsEnabled() bool {
	return g.profileLabels.Load()
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
