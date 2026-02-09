package value

// Env holds variable bindings with lexical scoping.
type Env struct {
	store  map[string]Value
	outer  *Env
	shapes map[string]*ShapeDef
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

// Get retrieves a value, walking up the scope chain.
func (e *Env) Get(name string) (Value, bool) {
	val, ok := e.store[name]
	if !ok && e.outer != nil {
		return e.outer.Get(name)
	}
	return val, ok
}

// Set binds a new name in the current scope.
func (e *Env) Set(name string, val Value) {
	e.store[name] = val
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
