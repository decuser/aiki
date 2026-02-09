package value

// Env holds variable bindings with lexical scoping.
type Env struct {
	store    map[string]Value
	outer    *Env
	shapes   map[string]*ShapeDef
	snapshot map[string]Value // snapshot of prelude bindings for restore
}

// ShapeDef holds a shape definition.
type ShapeDef struct {
	Name   string
	Fields []string
	Embeds []string // names of embedded shapes
}

// NewEnv creates a new environment.
func NewEnv(outer *Env) *Env {
	return &Env{
		store:  make(map[string]Value),
		outer:  outer,
		shapes: make(map[string]*ShapeDef),
	}
}

// SnapshotPrelude saves current bindings for later restore.
func (e *Env) SnapshotPrelude() {
	e.snapshot = make(map[string]Value)
	for name, val := range e.store {
		e.snapshot[name] = val
	}
}

// GetSnapshot returns the prelude snapshot, walking up scope chain.
func (e *Env) GetSnapshot() map[string]Value {
	if e.snapshot != nil {
		return e.snapshot
	}
	if e.outer != nil {
		return e.outer.GetSnapshot()
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
// Returns what was shadowed: "prelude", "builtin", or "" if nothing.
func (e *Env) Set(name string, val Value) string {
	shadowed := ""

	// Check if this shadows something in outer scope
	if e.outer != nil {
		if _, ok := e.outer.Get(name); ok {
			shadowed = "prelude"
		}
	}

	e.store[name] = val
	return shadowed
}

// SetWithBuiltinCheck binds a name and checks for builtin shadowing.
// builtinNames should be passed from the eval package.
// Returns "builtin", "prelude", or "" if nothing shadowed.
func (e *Env) SetWithBuiltinCheck(name string, val Value, builtinNames map[string]bool) string {
	// Check builtin first
	if builtinNames != nil && builtinNames[name] {
		e.store[name] = val
		return "builtin"
	}

	// Check outer scope (prelude)
	if e.outer != nil {
		if _, ok := e.outer.Get(name); ok {
			e.store[name] = val
			return "prelude"
		}
	}

	e.store[name] = val
	return ""
}

// Update mutates an existing binding, walking up the scope chain.
// Returns false if the name was not found.
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
// Returns true if the name was found and deleted.
func (e *Env) Delete(name string) bool {
	if _, ok := e.store[name]; ok {
		delete(e.store, name)
		return true
	}
	return false
}

// Restore restores a name from the prelude snapshot.
// Returns true if the name was found in snapshot and restored.
func (e *Env) Restore(name string) bool {
	snapshot := e.GetSnapshot()
	if snapshot == nil {
		return false
	}

	// Get value from snapshot
	val, ok := snapshot[name]
	if !ok {
		return false
	}

	// Restore the original value
	e.store[name] = val
	return true
}
