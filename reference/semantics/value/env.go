package value

import "strings"

// Env holds variable bindings with lexical scoping.
type Env struct {
	store    map[string]Value
	outer    *Env
	shapes   map[string]*ShapeDef
	snapshot map[string]Value
	stack    *[]StackFrame
	file     *string   // current filename (shared)
	source   *[]string // source lines (shared)
	exports  []string
}

// ShapeDef holds a shape definition.
type ShapeDef struct {
	Name   string
	Fields []string
	Embeds []string
}

// NewEnv creates a new environment.
func NewEnv(outer *Env) *Env {
	e := &Env{
		store:  make(map[string]Value),
		outer:  outer,
		shapes: make(map[string]*ShapeDef),
	}
	if outer != nil {
		e.stack = outer.stack
		e.file = outer.file
		e.source = outer.source
	} else {
		s := make([]StackFrame, 0)
		e.stack = &s
		empty := ""
		e.file = &empty
		lines := make([]string, 0)
		e.source = &lines
	}
	return e
}

// SetFile sets the current filename for error reporting.
func (e *Env) SetFile(filename string) {
	*e.file = filename
}

// GetFile returns the current filename.
func (e *Env) GetFile() string {
	return *e.file
}

// SetSource sets the source code for line lookup.
func (e *Env) SetSource(code string) {
	*e.source = strings.Split(code, "\n")
}

// GetSourceLine returns the source line at the given line number (1-indexed).
func (e *Env) GetSourceLine(line int) string {
	if line < 1 || line > len(*e.source) {
		return ""
	}
	return (*e.source)[line-1]
}

// PushFrame adds a stack frame with layer information.
func (e *Env) PushFrame(name string, line int, layer Layer) {
	*e.stack = append(*e.stack, StackFrame{
		Name:  name,
		File:  *e.file,
		Line:  line,
		Layer: layer,
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

// SnapshotPrelude saves current bindings for later restore.
func (e *Env) SnapshotPrelude() {
	e.snapshot = make(map[string]Value)
	for name, val := range e.store {
		e.snapshot[name] = val
	}
}

// GetSnapshot returns the strict snapshot, walking up scope chain.
func (e *Env) GetSnapshot() map[string]Value {
	if e.snapshot != nil {
		return e.snapshot
	}
	if e.outer != nil {
		return e.outer.GetSnapshot()
	}
	return nil
}

// SetExports records the list of exported names.
func (e *Env) SetExports(names []string) {
	e.exports = names
}

// GetExports returns the export list, walking up scope chain.
func (e *Env) GetExports() []string {
	if e.exports != nil {
		return e.exports
	}
	if e.outer != nil {
		return e.outer.GetExports()
	}
	return nil
}

// Get retrieves a value, walking up the scope chain.
func (e *Env) Get(name string) (Value, bool) {
	val, ok := e.store[name]
	if !ok && e.outer != nil {
		return e.outer.Get(name)
	}
	return val, ok
}

// Set binds a new name in the current scope.
func (e *Env) Set(name string, val Value) string {
	shadowed := ""
	if e.outer != nil {
		if _, ok := e.outer.Get(name); ok {
			shadowed = "strict"
		}
	}
	e.store[name] = val
	return shadowed
}

// Update mutates an existing binding, walking up the scope chain.
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

// GetShape retrieves a shape definition, walking up the scope chain.
func (e *Env) GetShape(name string) (*ShapeDef, bool) {
	def, ok := e.shapes[name]
	if !ok && e.outer != nil {
		return e.outer.GetShape(name)
	}
	return def, ok
}

// ResolveFields returns all fields for a shape, including embedded ones.
func (e *Env) ResolveFields(name string) ([]string, bool) {
	def, ok := e.GetShape(name)
	if !ok {
		return nil, false
	}

	var fields []string
	for _, embed := range def.Embeds {
		embedded, ok := e.ResolveFields(embed)
		if !ok {
			return nil, false
		}
		fields = append(fields, embedded...)
	}
	fields = append(fields, def.Fields...)
	return fields, true
}

// Symbols returns all names bound in this scope (not outer).
func (e *Env) Symbols() []string {
	names := make([]string, 0, len(e.store))
	for name := range e.store {
		names = append(names, name)
	}
	return names
}

// Delete removes a binding from the current scope.
func (e *Env) Delete(name string) bool {
	if _, ok := e.store[name]; ok {
		delete(e.store, name)
		return true
	}
	return false
}

// Restore restores a name from the strict snapshot.
func (e *Env) Restore(name string) bool {
	snapshot := e.GetSnapshot()
	if snapshot == nil {
		return false
	}
	val, ok := snapshot[name]
	if !ok {
		return false
	}
	e.store[name] = val
	return true
}
