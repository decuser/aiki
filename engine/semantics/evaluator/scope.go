// scope.go implements the environment chain for variable binding.
package evaluator

import (
	"aiki/engine/semantics/value"
	"strings"
)

// Layer indicates where a binding or stack frame originates.
type Layer int

const (
	LayerUser Layer = iota
	LayerPrelude
	LayerHal
)

// StackFrame represents a call frame for error traces.
type StackFrame struct {
	Name  string
	File  string
	Line  int
	Layer Layer
}

// ShapeDef defines a shaped list structure.
type ShapeDef struct {
	Name   string
	Fields []string
}

// Scope represents a lexical environment with variable bindings.
type Scope struct {
	parent    *Scope
	bindings  map[string]value.Value
	shapes    map[string]*ShapeDef
	exports   []string
	file      string
	source    string
	stack     []StackFrame
	preludeAt int // index where prelude bindings end
}

// NewScope creates a new scope with an optional parent.
func NewScope(parent *Scope) *Scope {
	s := &Scope{
		parent:   parent,
		bindings: make(map[string]value.Value),
		shapes:   make(map[string]*ShapeDef),
	}
	if parent != nil {
		s.file = parent.file
		s.source = parent.source
		s.stack = parent.stack
	}
	return s
}

// Define binds a name to a value in this scope.
func (s *Scope) Define(name string, val value.Value) {
	s.bindings[name] = val
}

// Get retrieves a value by name, searching up the scope chain.
func (s *Scope) Get(name string) (value.Value, bool) {
	if val, ok := s.bindings[name]; ok {
		return val, true
	}
	if s.parent != nil {
		return s.parent.Get(name)
	}
	return value.Value{}, false
}

// Update modifies an existing binding, searching up the scope chain.
// Returns false if the name is not found.
func (s *Scope) Update(name string, val value.Value) bool {
	if _, ok := s.bindings[name]; ok {
		s.bindings[name] = val
		return true
	}
	if s.parent != nil {
		return s.parent.Update(name, val)
	}
	return false
}

// Has checks if a name exists in this scope or any parent.
func (s *Scope) Has(name string) bool {
	_, ok := s.Get(name)
	return ok
}

// DefineShape registers a shape definition.
func (s *Scope) DefineShape(def *ShapeDef) {
	if s.shapes == nil {
		s.shapes = make(map[string]*ShapeDef)
	}
	s.shapes[def.Name] = def
}

// GetShape retrieves a shape definition by name.
func (s *Scope) GetShape(name string) (*ShapeDef, bool) {
	// Strip @ prefix if present
	name = strings.TrimPrefix(name, "@")
	if def, ok := s.shapes[name]; ok {
		return def, true
	}
	if s.parent != nil {
		return s.parent.GetShape(name)
	}
	return nil, false
}

// SetFile sets the current file name for error reporting.
func (s *Scope) SetFile(file string) {
	s.file = file
}

// GetFile returns the current file name.
func (s *Scope) GetFile() string {
	if s.file != "" {
		return s.file
	}
	if s.parent != nil {
		return s.parent.GetFile()
	}
	return ""
}

// SetSource sets the source code for error reporting.
func (s *Scope) SetSource(source string) {
	s.source = source
}

// GetSourceLine returns the source line at the given line number.
func (s *Scope) GetSourceLine(line int) string {
	src := s.source
	if src == "" && s.parent != nil {
		src = s.parent.source
	}
	if src == "" || line < 1 {
		return ""
	}
	lines := strings.Split(src, "\n")
	if line <= len(lines) {
		return lines[line-1]
	}
	return ""
}

// SetExports sets the exported names for this module.
func (s *Scope) SetExports(names []string) {
	s.exports = names
}

// GetExports returns the exported names for this module.
func (s *Scope) GetExports() []string {
	return s.exports
}

// PushFrame adds a stack frame for call tracing.
func (s *Scope) PushFrame(name, file string, line int, layer Layer) {
	s.stack = append(s.stack, StackFrame{
		Name:  name,
		File:  file,
		Line:  line,
		Layer: layer,
	})
}

// PopFrame removes the top stack frame.
func (s *Scope) PopFrame() {
	if len(s.stack) > 0 {
		s.stack = s.stack[:len(s.stack)-1]
	}
}

// CopyStack returns a copy of the current stack for error reporting.
func (s *Scope) CopyStack() []StackFrame {
	if len(s.stack) == 0 {
		return nil
	}
	cp := make([]StackFrame, len(s.stack))
	copy(cp, s.stack)
	return cp
}

// SnapshotPrelude marks the current bindings as prelude.
// New bindings after this are user-level.
func (s *Scope) SnapshotPrelude() {
	s.preludeAt = len(s.bindings)
}

// IsPrelude checks if a binding is from prelude (vs user code).
func (s *Scope) IsPrelude(name string) bool {
	// Simple approximation - check root scope
	root := s
	for root.parent != nil {
		root = root.parent
	}
	_, inRoot := root.bindings[name]
	return inRoot
}
