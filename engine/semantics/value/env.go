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

const envCompactBindingCapacity = 4

// envBindings owns ordinary local bindings for one Env. The object itself is
// allocated lazily on first ordinary local binding, so environments with no
// locals pay no storage footprint. Up to four bindings remain compact; the
// fifth promotes the block to a Go map.
type envBindings struct {
	n      uint8
	names  [envCompactBindingCapacity]string
	values [envCompactBindingCapacity]Value
	spill  map[string]Value
}

func (b *envBindings) get(name string) (Value, bool) {
	if b == nil {
		return nil, false
	}
	if b.spill != nil {
		v, ok := b.spill[name]
		return v, ok
	}
	for i := 0; i < int(b.n); i++ {
		if b.names[i] == name {
			return b.values[i], true
		}
	}
	return nil, false
}

func (b *envBindings) has(name string) bool {
	_, ok := b.get(name)
	return ok
}

func (b *envBindings) set(name string, val Value) (existed bool, count int, promoted bool) {
	if b.spill != nil {
		_, existed = b.spill[name]
		b.spill[name] = val
		return existed, len(b.spill), false
	}

	for i := 0; i < int(b.n); i++ {
		if b.names[i] == name {
			b.values[i] = val
			return true, int(b.n), false
		}
	}

	if int(b.n) < envCompactBindingCapacity {
		i := int(b.n)
		b.names[i] = name
		b.values[i] = val
		b.n++
		return false, int(b.n), false
	}

	m := make(map[string]Value, envCompactBindingCapacity+1)
	for i := 0; i < int(b.n); i++ {
		m[b.names[i]] = b.values[i]
		b.names[i] = ""
		b.values[i] = nil
	}
	b.n = 0
	m[name] = val
	b.spill = m
	return false, len(m), true
}

func (b *envBindings) update(name string, val Value) bool {
	if b == nil {
		return false
	}
	if b.spill != nil {
		if _, ok := b.spill[name]; !ok {
			return false
		}
		b.spill[name] = val
		return true
	}
	for i := 0; i < int(b.n); i++ {
		if b.names[i] == name {
			b.values[i] = val
			return true
		}
	}
	return false
}

func (b *envBindings) delete(name string) bool {
	if b == nil {
		return false
	}
	if b.spill != nil {
		if _, ok := b.spill[name]; !ok {
			return false
		}
		delete(b.spill, name)
		return true
	}
	for i := 0; i < int(b.n); i++ {
		if b.names[i] != name {
			continue
		}
		last := int(b.n) - 1
		for j := i; j < last; j++ {
			b.names[j] = b.names[j+1]
			b.values[j] = b.values[j+1]
		}
		b.names[last] = ""
		b.values[last] = nil
		b.n--
		return true
	}
	return false
}

func (b *envBindings) length() int {
	if b == nil {
		return 0
	}
	if b.spill != nil {
		return len(b.spill)
	}
	return int(b.n)
}

// reset retains the external block allocation for a physically reused tail
// frame but returns it to compact-empty realization, even if the prior logical
// invocation had promoted to a map.
func (b *envBindings) reset() {
	if b == nil {
		return
	}
	if b.spill != nil {
		clear(b.spill)
		b.spill = nil
	}
	for i := 0; i < int(b.n); i++ {
		b.names[i] = ""
		b.values[i] = nil
	}
	b.n = 0
}

func (b *envBindings) forceMap(capacity int) map[string]Value {
	if b.spill == nil {
		b.spill = make(map[string]Value, capacity)
		for i := 0; i < int(b.n); i++ {
			b.spill[b.names[i]] = b.values[i]
			b.names[i] = ""
			b.values[i] = nil
		}
		b.n = 0
	}
	return b.spill
}

// Env holds variable bindings and scope chain.
type envCold struct {
	shapes      map[string]*ShapeDef
	file        string
	sourceLines []string
	exports     []string
	packageName string
}

// Env holds variable bindings and scope chain.
//
// The hot object deliberately contains only execution/binding state. Module,
// source, diagnostic, and shape metadata live behind cold and are allocated
// only for environments that actually own such metadata.
type Env struct {
	bindings         *envBindings
	callParamNames   []string
	callParamValues  []Value
	callParamDeleted uint64
	outer            *Env
	stack            *[]StackFrame // shared across enclosed envs
	stackLimit       *int          // shared recursion limit (non tail frames)
	scope            Scope
	authority        Authority
	semanticProbe    observe.SemanticProbe // dynamic profiling context
	cold             *envCold

	envKind          observe.EnvKind
	bindingHighwater uint8
}

// NewEnv creates a new environment with user scope.
func NewEnv() *Env {
	stack := make([]StackFrame, 0)
	limit := 10000
	return &Env{
		scope:      ScopeUser,
		authority:  NoAuthority(),
		stack:      &stack,
		stackLimit: &limit,
		envKind:    observe.EnvKindRoot,
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
	env := &Env{
		outer:         outer,
		scope:         outer.scope,
		authority:     outer.authority,
		stack:         outer.stack,
		stackLimit:    outer.stackLimit,
		semanticProbe: outer.semanticProbe,
		envKind:       observe.EnvKindEnclosed,
	}
	env.recordPhysical()
	return env
}

// NewEnclosedEnvWithScope creates a child environment with an explicit scope.
// Used to create user-scope envs enclosed by prelude-scope envs.
func NewEnclosedEnvWithScope(outer *Env, scope Scope) *Env {
	env := &Env{
		outer:         outer,
		scope:         scope,
		authority:     outer.authority,
		stack:         outer.stack,
		stackLimit:    outer.stackLimit,
		semanticProbe: outer.semanticProbe,
		envKind:       observe.EnvKindEnclosed,
	}
	env.recordPhysical()
	return env
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
	env := &Env{
		outer:         outer,
		scope:         outer.scope,
		authority:     authority,
		stack:         &stack,
		stackLimit:    &limit,
		semanticProbe: outer.semanticProbe,
		envKind:       observe.EnvKindIsolated,
	}
	env.recordPhysical()
	return env
}

// NewCallEnv creates a function-call environment. Name lookup is lexical and
// therefore chains through lexicalOuter, while execution metadata is dynamic
// and therefore follows caller. Keeping those two relationships separate is
// essential for isolated spawn: calling a prelude function must not reconnect
// a spawned computation to the parent's call stack.
func NewCallEnv(lexicalOuter, caller *Env) *Env {
	env := &Env{
		outer:         lexicalOuter,
		scope:         lexicalOuter.scope,
		authority:     lexicalOuter.authority,
		stack:         caller.stack,
		stackLimit:    caller.stackLimit,
		semanticProbe: caller.semanticProbe,
		envKind:       observe.EnvKindCall,
	}
	env.recordPhysical()
	env.recordLogicalCall()
	return env
}

// ResetCallEnv reinitializes a call environment for a proven non-escaping tail
// frame. The caller must ensure that no closure or other value can retain this
// Env after the current invocation.
func (e *Env) ResetCallEnv(lexicalOuter, caller *Env) {
	if e.bindings != nil {
		e.bindings.reset()
	}
	e.callParamNames = nil
	e.callParamValues = nil
	e.callParamDeleted = 0
	e.outer = lexicalOuter
	e.stack = caller.stack
	e.stackLimit = caller.stackLimit
	e.scope = lexicalOuter.scope
	e.authority = lexicalOuter.authority
	e.semanticProbe = caller.semanticProbe
	e.cold = nil
	e.envKind = observe.EnvKindCall
	e.bindingHighwater = 0
	e.recordLogicalCall()
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
		if e.bindings == nil {
			e.bindings = &envBindings{}
		}
		store := e.bindings.forceMap(len(names))
		for i, name := range names {
			store[name] = values[i]
		}
		return
	}
	e.callParamNames = names
	e.callParamValues = values[:len(names)]
	e.callParamDeleted = 0
}

// ClearCallParams releases this environment's borrowed argument-vector view.
// It is used only after a proven non-escaping invocation has finished with a
// reusable argument frame; ordinary call semantics are unchanged.
func (e *Env) ClearCallParams() {
	e.callParamNames = nil
	e.callParamValues = nil
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

func (e *Env) ensureCold() *envCold {
	if e.cold == nil {
		e.cold = &envCold{}
	}
	return e.cold
}

func (e *Env) ensureBindings() *envBindings {
	if e.bindings == nil {
		e.bindings = &envBindings{}
		if probe := e.envProbe(); probe != nil {
			probe.RecordEnvCompactAllocation(e.envKind)
		}
	}
	return e.bindings
}

func (e *Env) envProbe() observe.EnvRealizationProbe {
	if e == nil || e.semanticProbe == nil {
		return nil
	}
	probe, _ := e.semanticProbe.(observe.EnvRealizationProbe)
	return probe
}

func (e *Env) recordPhysical() {
	if probe := e.envProbe(); probe != nil {
		probe.RecordEnvPhysical(e.envKind)
	}
}

func (e *Env) recordLogicalCall() {
	if probe := e.envProbe(); probe != nil {
		probe.RecordEnvLogicalCall()
	}
}

func (e *Env) recordBindingHighwater(n int) {
	if n <= int(e.bindingHighwater) {
		return
	}
	old := int(e.bindingHighwater)
	e.bindingHighwater = uint8(n)
	probe := e.envProbe()
	if probe == nil {
		return
	}
	for _, threshold := range []int{1, 2, 3, 5} {
		if old < threshold && n >= threshold {
			probe.RecordEnvBindingThreshold(e.envKind, threshold)
		}
	}
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
	if val, ok := e.bindings.get(name); ok {
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

	bindings := e.ensureBindings()
	existed, count, promoted := bindings.set(name, val)
	if promoted {
		if probe := e.envProbe(); probe != nil {
			probe.RecordEnvMapPromotion(e.envKind)
		}
	}
	if probe := e.envProbe(); probe != nil {
		probe.RecordEnvLocalSet(e.envKind, existed)
	}
	if !existed {
		e.recordBindingHighwater(count)
	}
}

// Update updates an existing binding, searching outer scopes.
func (e *Env) Update(name string, val Value) bool {
	if e.bindings.update(name, val) {
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
	if e.bindings.delete(name) {
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
	return e.bindings.has(name)
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
	cold := e.ensureCold()
	if cold.shapes == nil {
		cold.shapes = make(map[string]*ShapeDef)
	}
	cold.shapes[def.Name] = def
}

// CollectShapes returns the shape definitions visible from this environment,
// with nearer definitions shadowing those in enclosing environments.
// Only shape definitions are copied; no value bindings are included.
func (e *Env) CollectShapes() map[string]*ShapeDef {
	if e == nil {
		return nil
	}
	out := e.outer.CollectShapes()
	var shapes map[string]*ShapeDef
	if e.cold != nil {
		shapes = e.cold.shapes
	}
	if out == nil {
		out = make(map[string]*ShapeDef, len(shapes))
	}
	for name, def := range shapes {
		out[name] = def
	}
	return out
}

// GetShape retrieves a shape definition.
func (e *Env) GetShape(name string) (*ShapeDef, bool) {
	if e.cold != nil {
		if def, ok := e.cold.shapes[name]; ok {
			return def, true
		}
	}
	if e.outer != nil {
		return e.outer.GetShape(name)
	}
	return nil, false
}

// SetFile sets the current file name for this env.
func (e *Env) SetFile(file string) {
	e.ensureCold().file = file
}

// GetFile returns the file name, chaining up if not set.
func (e *Env) GetFile() string {
	if e.cold != nil && e.cold.file != "" {
		return e.cold.file
	}
	if e.outer != nil {
		return e.outer.GetFile()
	}
	return "<unknown>"
}

// SetSource sets the source code for this env.
func (e *Env) SetSource(source string) {
	e.ensureCold().sourceLines = splitLines(source)
}

// GetSourceLine returns a specific line from source, chaining up if needed.
func (e *Env) GetSourceLine(line int) string {
	if e.cold != nil && len(e.cold.sourceLines) > 0 {
		if line >= 1 && line <= len(e.cold.sourceLines) {
			return e.cold.sourceLines[line-1]
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
	e.ensureCold().exports = names
}

// GetExports returns the list of exported names.
func (e *Env) GetExports() []string {
	if e.cold == nil {
		return nil
	}
	return e.cold.exports
}

// SetPackageName sets the package name for this module.
func (e *Env) SetPackageName(name string) {
	e.ensureCold().packageName = name
}

// GetPackageName returns the package name declared by this module.
func (e *Env) GetPackageName() string {
	if e.cold == nil {
		return ""
	}
	return e.cold.packageName
}
