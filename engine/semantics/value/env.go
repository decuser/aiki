package value

// Env holds variable bindings and scope chain.
type Env struct {
	store  map[string]Value
	shapes map[string]*ShapeDef
	outer  *Env
	file   string
	source string
}

// NewEnv creates a new environment.
func NewEnv() *Env {
	return &Env{
		store:  make(map[string]Value),
		shapes: make(map[string]*ShapeDef),
	}
}

// NewEnclosedEnv creates a child environment.
func NewEnclosedEnv(outer *Env) *Env {
	env := NewEnv()
	env.outer = outer
	return env
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

// SetFile sets the current file name.
func (e *Env) SetFile(file string) {
	e.file = file
}

// GetFile returns the current file name.
func (e *Env) GetFile() string {
	if e.file != "" {
		return e.file
	}
	if e.outer != nil {
		return e.outer.GetFile()
	}
	return "<unknown>"
}

// SetSource sets the source code.
func (e *Env) SetSource(source string) {
	e.source = source
}

// GetSourceLine returns a specific line from source.
func (e *Env) GetSourceLine(line int) string {
	src := e.source
	if src == "" && e.outer != nil {
		src = e.outer.source
	}
	if src == "" {
		return ""
	}
	lines := splitLines(src)
	if line < 1 || line > len(lines) {
		return ""
	}
	return lines[line-1]
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
