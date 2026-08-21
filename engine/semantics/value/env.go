package value

import "aiki/engine/observe"

// Scope represents lexical/tooling role. It does not confer raw runtime authority.
type Scope int

const (
	ScopeUser    Scope = iota // ordinary user lexical role
	ScopePrelude              // prelude/trusted-library lexical role
)

// StackFrame represents a call site in the stack trace.
type StackFrame struct {
	Name  string
	File  string
	Line  int
	Scope Scope
}

// Env holds variable bindings and scope chain.
type Env struct {
	store            map[string]Value
	shapes           map[string]*ShapeDef
	callParamNames   []string
	callParamValues  []Value
	callParamDeleted uint64
	outer            *Env
	file             string        // this env's file (not shared)
	source           string        // this env's source (not shared)
	sourceLines      []string      // cached source lines for diagnostics/profiling
	stack            *[]StackFrame // shared across enclosed envs
	stackLimit       *int          // shared recursion limit (non tail frames)
	scope            Scope
	authority        Authority
	exports          []string              // exported names for modules
	packageName      string                // package name declared by this module
	semanticProbe    observe.SemanticProbe // dynamic profiling context
}

// NewEnv creates a new environment with user scope.
func NewEnv() *Env {
	stack := make([]StackFrame, 0)
	limit := 10000
	return &Env{
		store:      make(map[string]Value),
		shapes:     make(map[string]*ShapeDef),
		scope:      ScopeUser,
		authority:  NoAuthority(),
		stack:      &stack,
		stackLimit: &limit,
	}
}

// NewEnvWithScope creates a new environment with explicit scope.
func NewEnvWithScope(scope Scope) *Env {
	env := NewEnv()
	env.scope = scope
	return env
}

// NewEnclosedEnv creates a child environment, inheriting scope and shared state.
func NewEnclosedEnv(outer *Env) *Env {
	return &Env{
		outer:         outer,
		scope:         outer.scope,
		authority:     outer.authority,
		stack:         outer.stack,
		stackLimit:    outer.stackLimit,
		semanticProbe: outer.semanticProbe,
	}
}

// NewEnclosedEnvWithScope creates a child environment with an explicit scope.
// Used to create user-scope envs enclosed by prelude-scope envs.
func NewEnclosedEnvWithScope(outer *Env, scope Scope) *Env {
	return &Env{
		outer:         outer,
		scope:         scope,
		authority:     outer.authority,
		stack:         outer.stack,
		stackLimit:    outer.stackLimit,
		semanticProbe: outer.semanticProbe,
	}
}

// NewIsolatedEnclosedEnv creates a child that can read through outer but owns
// its own dynamic call-stack state. This is used by spawn: spawned computations
// share the prelude vocabulary but must not share mutable execution metadata
// with the parent goroutine. The current stack-limit value is copied.
func NewIsolatedEnclosedEnv(outer *Env) *Env {
	return NewIsolatedEnclosedEnvWithAuthority(outer, outer.authority)
}

// NewIsolatedEnclosedEnvWithAuthority creates an isolated dynamic environment
// that can read lexical vocabulary through outer while carrying an explicit
// definition-bound authority set. Spawn uses this to see prelude names without
// inheriting prelude privilege.
func NewIsolatedEnclosedEnvWithAuthority(outer *Env, authority Authority) *Env {
	stack := make([]StackFrame, 0)
	limit := outer.GetStackLimit()
	return &Env{
		outer:         outer,
		scope:         outer.scope,
		authority:     authority,
		stack:         &stack,
		stackLimit:    &limit,
		semanticProbe: outer.semanticProbe,
	}
}

// NewCallEnv creates a function-call environment. Name lookup is lexical and
// therefore chains through lexicalOuter, while execution metadata is dynamic
// and therefore follows caller. Keeping those two relationships separate is
// essential for isolated spawn: calling a prelude function must not reconnect
// a spawned computation to the parent's call stack.
func NewCallEnv(lexicalOuter, caller *Env) *Env {
	return &Env{
		outer:         lexicalOuter,
		scope:         lexicalOuter.scope,
		authority:     lexicalOuter.authority,
		stack:         caller.stack,
		stackLimit:    caller.stackLimit,
		semanticProbe: caller.semanticProbe,
	}
}

// ResetCallEnv reinitializes a call environment for a proven non-escaping tail
// frame. The caller must ensure that no closure or other value can retain this
// Env after the current invocation.
func (e *Env) ResetCallEnv(lexicalOuter, caller *Env) {
	if e.store != nil {
		clear(e.store)
	}
	if e.shapes != nil {
		clear(e.shapes)
	}
	e.callParamNames = nil
	e.callParamValues = nil
	e.callParamDeleted = 0
	e.outer = lexicalOuter
	e.file = ""
	e.source = ""
	e.sourceLines = nil
	e.stack = caller.stack
	e.stackLimit = caller.stackLimit
	e.scope = lexicalOuter.scope
	e.authority = lexicalOuter.authority
	e.exports = nil
	e.packageName = ""
	e.semanticProbe = caller.semanticProbe
}

// BindCallParams binds ordinary function parameters without constructing a
// per-call map. Parameter names belong to the immutable function object and
// values belong to the argument vector for this invocation. Calls with more
// than 64 fixed parameters fall back to ordinary map bindings so deletion can
// retain exact lexical semantics without auxiliary allocation.
func (e *Env) BindCallParams(names []string, values []Value) {
	if len(names) == 0 {
		return
	}
	if len(names) > 64 {
		if e.store == nil {
			e.store = make(map[string]Value, len(names))
		}
		for i, name := range names {
			e.store[name] = values[i]
		}
		return
	}
	e.callParamNames = names
	e.callParamValues = values[:len(names)]
	e.callParamDeleted = 0
}

func (e *Env) callParamIndex(name string) (int, bool) {
	for i, candidate := range e.callParamNames {
		if candidate == name {
			return i, true
		}
	}
	return 0, false
}

// SetSemanticProbe sets the dynamic semantic profiling context for this env.
// A nil probe explicitly disables profiling for this dynamic branch.
func (e *Env) SetSemanticProbe(probe observe.SemanticProbe) {
	e.semanticProbe = probe
}

// GetSemanticProbe returns the dynamic semantic profiling context.
func (e *Env) GetSemanticProbe() observe.SemanticProbe {
	return e.semanticProbe
}

// GetScope returns the scope.
func (e *Env) GetScope() Scope {
	return e.scope
}

// SetAuthority assigns the immutable raw-primitive authority for definitions
// created in this environment.
func (e *Env) SetAuthority(authority Authority) {
	e.authority = authority
}

// GetAuthority returns the definition-bound raw-primitive authority.
func (e *Env) GetAuthority() Authority {
	return e.authority
}

// Outer returns the enclosing environment, or nil if none.
func (e *Env) Outer() *Env {
	return e.outer
}

// PushFrame adds a stack frame.
func (e *Env) PushFrame(name string, line int, scope Scope) {
	*e.stack = append(*e.stack, StackFrame{
		Name:  name,
		File:  e.GetFile(),
		Line:  line,
		Scope: scope,
	})
}

// PopFrame removes the top stack frame.
func (e *Env) PopFrame() {
	if len(*e.stack) > 0 {
		*e.stack = (*e.stack)[:len(*e.stack)-1]
	}
}

// CurrentFrame returns the active Aiki call frame without copying the stack.
func (e *Env) CurrentFrame() (StackFrame, bool) {
	if e == nil || e.stack == nil || len(*e.stack) == 0 {
		return StackFrame{}, false
	}
	return (*e.stack)[len(*e.stack)-1], true
}

// CopyStack returns a copy of the current call stack.
func (e *Env) CopyStack() []StackFrame {
	cp := make([]StackFrame, len(*e.stack))
	copy(cp, *e.stack)
	return cp
}

// GetStackLimit returns the current non tail call stack limit.
func (e *Env) GetStackLimit() int {
	if e.stackLimit == nil {
		return 0
	}
	return *e.stackLimit
}

// SetStackLimit sets the non tail call stack limit. Must be >= 1.
func (e *Env) SetStackLimit(n int) {
	if e.stackLimit == nil {
		e.stackLimit = new(int)
	}
	*e.stackLimit = n
}

// StackDepth returns current call stack depth.
func (e *Env) StackDepth() int {
	return len(*e.stack)
}

// ReplaceTopFrame replaces the top stack frame. No effect if stack is empty.
func (e *Env) ReplaceTopFrame(name string, line int, scope Scope) {
	if len(*e.stack) == 0 {
		return
	}
	(*e.stack)[len(*e.stack)-1] = StackFrame{
		Name:  name,
		File:  e.GetFile(),
		Line:  line,
		Scope: scope,
	}
}

// Get retrieves a value by name.
func (e *Env) Get(name string) (Value, bool) {
	if val, ok := e.store[name]; ok {
		return val, true
	}
	if i, ok := e.callParamIndex(name); ok {
		if e.callParamDeleted&(uint64(1)<<uint(i)) == 0 {
			return e.callParamValues[i], true
		}
	}
	if e.outer != nil {
		return e.outer.Get(name)
	}
	return nil, false
}

// Set binds a value to a name in current scope.
func (e *Env) Set(name string, val Value) {
	if i, ok := e.callParamIndex(name); ok {
		e.callParamValues[i] = val
		e.callParamDeleted &^= uint64(1) << uint(i)
		return
	}
	if e.store == nil {
		e.store = make(map[string]Value)
	}
	e.store[name] = val
}

// Update updates an existing binding, searching outer scopes.
func (e *Env) Update(name string, val Value) bool {
	if _, ok := e.store[name]; ok {
		e.store[name] = val
		return true
	}
	if i, ok := e.callParamIndex(name); ok && e.callParamDeleted&(uint64(1)<<uint(i)) == 0 {
		e.callParamValues[i] = val
		return true
	}
	if e.outer != nil {
		return e.outer.Update(name, val)
	}
	return false
}

// Delete removes a binding from the current scope only.
// Returns true if the binding existed and was deleted.
// This allows shadowed outer bindings (e.g., prelude) to show through again.
func (e *Env) Delete(name string) bool {
	if _, ok := e.store[name]; ok {
		delete(e.store, name)
		return true
	}
	if i, ok := e.callParamIndex(name); ok {
		bit := uint64(1) << uint(i)
		if e.callParamDeleted&bit != 0 {
			return false
		}
		e.callParamDeleted |= bit
		return true
	}
	return false
}

// HasOwn checks if a binding exists in this scope only (not outer scopes).
func (e *Env) HasOwn(name string) bool {
	_, ok := e.store[name]
	return ok
}

// GetPreludeEnv walks up the outer chain to find the prelude env.
// This is the root prelude env, not function call envs that inherit ScopePrelude.
func (e *Env) GetPreludeEnv() *Env {
	curr := e
	for curr != nil {
		if curr.scope == ScopePrelude {
			// Check if this is the actual prelude root (outer is nil or non-prelude)
			if curr.outer == nil || curr.outer.scope != ScopePrelude {
				return curr
			}
		}
		curr = curr.outer
	}
	return nil
}

// DefineShape registers a shape definition.
func (e *Env) DefineShape(def *ShapeDef) {
	if e.shapes == nil {
		e.shapes = make(map[string]*ShapeDef)
	}
	e.shapes[def.Name] = def
}

// CollectShapes returns the shape definitions visible from this environment,
// with nearer definitions shadowing those in enclosing environments.
// Only shape definitions are copied; no value bindings are included.
func (e *Env) CollectShapes() map[string]*ShapeDef {
	if e == nil {
		return nil
	}
	out := e.outer.CollectShapes()
	if out == nil {
		out = make(map[string]*ShapeDef, len(e.shapes))
	}
	for name, def := range e.shapes {
		out[name] = def
	}
	return out
}

// GetShape retrieves a shape definition.
func (e *Env) GetShape(name string) (*ShapeDef, bool) {
	def, ok := e.shapes[name]
	if !ok && e.outer != nil {
		return e.outer.GetShape(name)
	}
	return def, ok
}

// SetFile sets the current file name for this env.
func (e *Env) SetFile(file string) {
	e.file = file
}

// GetFile returns the file name, chaining up if not set.
func (e *Env) GetFile() string {
	if e.file != "" {
		return e.file
	}
	if e.outer != nil {
		return e.outer.GetFile()
	}
	return "<unknown>"
}

// SetSource sets the source code for this env.
func (e *Env) SetSource(source string) {
	e.source = source
	e.sourceLines = splitLines(source)
}

// GetSourceLine returns a specific line from source, chaining up if needed.
func (e *Env) GetSourceLine(line int) string {
	if len(e.sourceLines) > 0 {
		if line >= 1 && line <= len(e.sourceLines) {
			return e.sourceLines[line-1]
		}
	}
	if e.outer != nil {
		return e.outer.GetSourceLine(line)
	}
	return ""
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

// SetExports sets the list of exported names for this module.
func (e *Env) SetExports(names []string) {
	e.exports = names
}

// GetExports returns the list of exported names.
func (e *Env) GetExports() []string {
	return e.exports
}

// SetPackageName sets the package name for this module.
func (e *Env) SetPackageName(name string) {
	e.packageName = name
}

// GetPackageName returns the package name declared by this module.
func (e *Env) GetPackageName() string {
	return e.packageName
}
