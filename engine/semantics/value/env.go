package value

// Scope represents the visibility level for builtins.
type Scope int

const (
	ScopeUser    Scope = iota // User code - cannot access HAL primitives
	ScopePrelude              // Prelude - can access HAL primitives (_prefixed)
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
	store   map[string]Value
	shapes  map[string]*ShapeDef
	outer   *Env
	file    string         // this env's file (not shared)
	source  string         // this env's source (not shared)
	stack   *[]StackFrame  // shared across enclosed envs
	scope   Scope
	exports []string       // exported names for modules
}

// NewEnv creates a new environment with user scope.
func NewEnv() *Env {
	stack := make([]StackFrame, 0)
	return &Env{
		store:  make(map[string]Value),
		shapes: make(map[string]*ShapeDef),
		scope:  ScopeUser,
		stack:  &stack,
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
		store:  make(map[string]Value),
		shapes: make(map[string]*ShapeDef),
		outer:  outer,
		scope:  outer.scope,
		stack:  outer.stack,
	}
}

// GetScope returns the scope.
func (e *Env) GetScope() Scope {
	return e.scope
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

// CopyStack returns a copy of the current call stack.
func (e *Env) CopyStack() []StackFrame {
	cp := make([]StackFrame, len(*e.stack))
	copy(cp, *e.stack)
	return cp
}

// Get retrieves a value by name.
func (e *Env) Get(name string) (Value, bool) {
	val, ok := e.store[name]
	if !ok && e.outer != nil {
		return e.outer.Get(name)
	}
	return val, ok
}

// Set binds a value to a name in current scope.
func (e *Env) Set(name string, val Value) {
	e.store[name] = val
}

// Update updates an existing binding, searching outer scopes.
func (e *Env) Update(name string, val Value) bool {
	if _, ok := e.store[name]; ok {
		e.store[name] = val
		return true
	}
	if e.outer != nil {
		return e.outer.Update(name, val)
	}
	return false
}

// DefineShape registers a shape definition.
func (e *Env) DefineShape(def *ShapeDef) {
	e.shapes[def.Name] = def
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
}

// GetSourceLine returns a specific line from source, chaining up if needed.
func (e *Env) GetSourceLine(line int) string {
	if e.source != "" {
		lines := splitLines(e.source)
		if line >= 1 && line <= len(lines) {
			return lines[line-1]
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
