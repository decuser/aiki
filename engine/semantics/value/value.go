// Package value defines the core value types for Aiki.
package value

import (
	"fmt"
	"math/big"
	"strings"
)

// Type identifies value kinds.
type Type string

const (
	NumberType   Type = "number"
	BooleanType  Type = "boolean"
	RuneType     Type = "rune"
	StringType   Type = "string"
	BytesType    Type = "bytes"
	SymbolType   Type = "symbol"
	ListType     Type = "list"
	FunctionType Type = "function"
	ErrorType    Type = "error"
	FaultType    Type = "fault"
	ReturnType   Type = "return"
	HandleType   Type = "handle"
	ChannelType  Type = "channel"
	ModuleType   Type = "module"
	CanvasType   Type = "canvas"
)

// Value is the interface all Aiki values implement.
type Value interface {
	Type() Type
	Inspect() string
}

// Number is exact rational arithmetic.
type Number struct {
	Val *big.Rat
}

func (n *Number) Type() Type { return NumberType }
func (n *Number) Inspect() string {
	if n.Val.IsInt() {
		return n.Val.Num().String()
	}
	return n.Val.RatString()
}

func NewNumber(num, denom int64) *Number {
	return &Number{Val: big.NewRat(num, denom)}
}

func NewNumberFromString(s string) (*Number, error) {
	r := new(big.Rat)
	if _, ok := r.SetString(s); !ok {
		return nil, fmt.Errorf("invalid number: %s", s)
	}
	return &Number{Val: r}, nil
}

// Boolean
type Boolean struct {
	Val bool
}

func (b *Boolean) Type() Type { return BooleanType }
func (b *Boolean) Inspect() string {
	if b.Val {
		return "true"
	}
	return "false"
}

var (
	TRUE  = &Boolean{Val: true}
	FALSE = &Boolean{Val: false}
)

// Rune
type Rune struct {
	Val rune
}

func (r *Rune) Type() Type      { return RuneType }
func (r *Rune) Inspect() string { return fmt.Sprintf("'%c'", r.Val) }

// String
type String struct {
	Val string
}

func (s *String) Type() Type      { return StringType }
func (s *String) Inspect() string { return s.Val }

// Bytes is an immutable sequence of bytes (0-255).
type Bytes struct {
	Val []byte
}

func (b *Bytes) Type() Type { return BytesType }
func (b *Bytes) Inspect() string {
	if len(b.Val) <= 20 {
		return fmt.Sprintf("<bytes:%d %v>", len(b.Val), b.Val)
	}
	return fmt.Sprintf("<bytes:%d [%v...]>", len(b.Val), b.Val[:20])
}

// Symbol
type Symbol struct {
	Val string
}

func (s *Symbol) Type() Type      { return SymbolType }
func (s *Symbol) Inspect() string { return ":" + s.Val }

// List
type List struct {
	Elements []Value
	Shape    string // empty for raw list, shape name for shaped list
}

func (l *List) Type() Type { return ListType }
func (l *List) Inspect() string {
	var parts []string
	if l.Shape != "" {
		parts = append(parts, "@"+l.Shape)
	}
	for _, e := range l.Elements {
		parts = append(parts, e.Inspect())
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

// Function
type Function struct {
	Name   string
	Params []string
	Rest   string      // rest parameter name, empty if none
	Body   interface{} // *syntax.Node, but we don't import syntax here
	Env    interface{} // *Env, set at runtime
}

func (f *Function) Type() Type { return FunctionType }
func (f *Function) Inspect() string {
	if f.Name != "" {
		return fmt.Sprintf("<fn %s>", f.Name)
	}
	return "<fn>"
}

// Callable is implemented by values that can be called as functions.
type Callable interface {
	Value
	Call(args []Value) Value
}

// Error
type Error struct {
	Message string
	File    string
	Line    int
	Source  string // source line for context
	Stack   []StackFrame
}

func (e *Error) Type() Type { return ErrorType }

// Inspect returns Ruby-style error format:
// file:line:in 'func': message
//
//	source line
//	    from file:line:in 'func'
func (e *Error) Inspect() string {
	var sb strings.Builder

	// Find the innermost visible frame for the header
	funcName := "<main>"
	if len(e.Stack) > 0 {
		// Use the innermost (last) frame's name
		funcName = e.Stack[len(e.Stack)-1].Name
	}

	// First line: file:line:in 'func': message
	if e.File != "" && e.Line > 0 {
		sb.WriteString(fmt.Sprintf("%s:%d:in '%s': %s", e.File, e.Line, funcName, e.Message))
	} else if e.Line > 0 {
		sb.WriteString(fmt.Sprintf("line %d:in '%s': %s", e.Line, funcName, e.Message))
	} else {
		sb.WriteString(e.Message)
	}

	// Second line: source code (indented)
	if e.Source != "" {
		sb.WriteString(fmt.Sprintf("\n    %s", strings.TrimSpace(e.Source)))
	}

	// Stack trace: from file:line:in 'func' (skip innermost, shown above)
	for i := len(e.Stack) - 2; i >= 0; i-- {
		frame := e.Stack[i]
		if frame.File != "" {
			sb.WriteString(fmt.Sprintf("\n        from %s:%d:in '%s'", frame.File, frame.Line, frame.Name))
		} else {
			sb.WriteString(fmt.Sprintf("\n        from line %d:in '%s'", frame.Line, frame.Name))
		}
	}

	return sb.String()
}

func NewError(format string, args ...interface{}) *Error {
	return &Error{Message: fmt.Sprintf(format, args...)}
}

func NewErrorAt(file string, line int, source string, stack []StackFrame, format string, args ...interface{}) *Error {
	return &Error{
		Message: fmt.Sprintf(format, args...),
		File:    file,
		Line:    line,
		Source:  source,
		Stack:   stack,
	}
}

// EMPTY is the empty list singleton, used as "no value".
var EMPTY = &List{Elements: []Value{}}

// Return wraps a return value (internal use).
type Return struct {
	Val Value
}

func (r *Return) Type() Type      { return ReturnType }
func (r *Return) Inspect() string { return r.Val.Inspect() }

// Handle wraps an OS resource.
type Handle struct {
	Name string
	Res  interface{}
}

func (h *Handle) Type() Type      { return HandleType }
func (h *Handle) Inspect() string { return fmt.Sprintf("<handle %s>", h.Name) }

// Channel for concurrency.
type Channel struct {
	C chan Value
}

func (c *Channel) Type() Type      { return ChannelType }
func (c *Channel) Inspect() string { return "<channel>" }

// NewChannel creates a new unbuffered channel.
func NewChannel() *Channel {
	return &Channel{C: make(chan Value)}
}

// Module holds a loaded package's exports.
type Module struct {
	Name    string           // package name
	Exports map[string]Value // exported bindings
}

func (m *Module) Type() Type      { return ModuleType }
func (m *Module) Inspect() string { return "<module " + m.Name + ">" }

// Get retrieves an exported value by name.
func (m *Module) Get(name string) (Value, bool) {
	v, ok := m.Exports[name]
	return v, ok
}

// NewModule creates a module with the given name and exports.
func NewModule(name string, exports map[string]Value) *Module {
	return &Module{Name: name, Exports: exports}
}

// ExitSignal signals REPL to exit.
type ExitSignal struct{}

func (e *ExitSignal) Type() Type      { return "exit" }
func (e *ExitSignal) Inspect() string { return "" }

var EXIT = &ExitSignal{}

// ResetSignal signals REPL to reset environment.
type ResetSignal struct{}

func (r *ResetSignal) Type() Type      { return "reset" }
func (r *ResetSignal) Inspect() string { return "<reset>" }

var RESET = &ResetSignal{}

// ShapeDef defines a shaped list structure.
type ShapeDef struct {
	Name   string
	Fields []string
}

// Helper functions

func IsTruthy(v Value) bool {
	switch val := v.(type) {
	case *Boolean:
		return val.Val
	case *List:
		return len(val.Elements) > 0
	default:
		return true
	}
}

func IsError(v Value) bool {
	return v != nil && v.Type() == ErrorType
}

// Fault represents an internal evaluation failure that halts execution.
// Unlike Error (which is a shaped recoverable value), Fault is not
// an ordinary Aiki value - evaluation halts immediately when one occurs.
type Fault struct {
	Message string
	File    string
	Line    int
	Source  string // source line for context
	Stack   []StackFrame
}

func (f *Fault) Type() Type { return FaultType }

// Inspect returns Ruby-style error format matching Error.Inspect.
func (f *Fault) Inspect() string {
	var sb strings.Builder

	// Find the innermost visible frame for the header
	funcName := "<main>"
	if len(f.Stack) > 0 {
		funcName = f.Stack[len(f.Stack)-1].Name
	}

	// First line: file:line:in 'func': message
	if f.File != "" && f.Line > 0 {
		sb.WriteString(fmt.Sprintf("%s:%d:in '%s': %s", f.File, f.Line, funcName, f.Message))
	} else if f.Line > 0 {
		sb.WriteString(fmt.Sprintf("line %d:in '%s': %s", f.Line, funcName, f.Message))
	} else {
		sb.WriteString(f.Message)
	}

	// Second line: source code (indented)
	if f.Source != "" {
		sb.WriteString(fmt.Sprintf("\n    %s", strings.TrimSpace(f.Source)))
	}

	// Stack trace: from file:line:in 'func' (skip innermost, shown above)
	for i := len(f.Stack) - 2; i >= 0; i-- {
		frame := f.Stack[i]
		if frame.File != "" {
			sb.WriteString(fmt.Sprintf("\n        from %s:%d:in '%s'", frame.File, frame.Line, frame.Name))
		} else {
			sb.WriteString(fmt.Sprintf("\n        from line %d:in '%s'", frame.Line, frame.Name))
		}
	}

	return sb.String()
}

func NewFault(format string, args ...interface{}) *Fault {
	return &Fault{Message: fmt.Sprintf(format, args...)}
}

func NewFaultAt(file string, line int, source string, stack []StackFrame, format string, args ...interface{}) *Fault {
	return &Fault{
		Message: fmt.Sprintf(format, args...),
		File:    file,
		Line:    line,
		Source:  source,
		Stack:   stack,
	}
}

// IsFault returns true only for internal halting failures.
func IsFault(v Value) bool {
	_, ok := v.(*Fault)
	return ok
}
