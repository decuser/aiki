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
	// Keep output usable: cap frames and compress repeated frames (common in recursion overflows).
	const maxFrames = 20
	formatFrame := func(frame StackFrame) string {
		if frame.File != "" {
			return fmt.Sprintf("\n        from %s:%d:in '%s'", frame.File, frame.Line, frame.Name)
		}
		return fmt.Sprintf("\n        from line %d:in '%s'", frame.Line, frame.Name)
	}

	// Frames in the order we print them: caller, then its caller, outward.
	frames := make([]StackFrame, 0, func() int {
		if len(e.Stack) > 1 {
			return len(e.Stack) - 1
		}
		return 0
	}())
	for i := len(e.Stack) - 2; i >= 0; i-- {
		frames = append(frames, e.Stack[i])
	}

	printedFrames := 0
	frameKey := func(f StackFrame) string {
		return fmt.Sprintf("%s|%d|%s", f.File, f.Line, f.Name)
	}

	for i := 0; i < len(frames); {
		// Find run of identical frames.
		k := frameKey(frames[i])
		j := i + 1
		for j < len(frames) && frameKey(frames[j]) == k {
			j++
		}
		runLen := j - i

		// Helper to emit a frame, respecting cap.
		emitFrame := func(f StackFrame) bool {
			if printedFrames >= maxFrames {
				return false
			}
			sb.WriteString(formatFrame(f))
			printedFrames++
			return true
		}

		if runLen <= 3 {
			for x := i; x < j; x++ {
				if !emitFrame(frames[x]) {
					break
				}
			}
		} else {
			// Print first two, compress middle, print last one.
			if emitFrame(frames[i]) {
				_ = emitFrame(frames[i+1])
			}
			repeated := runLen - 3
			if repeated > 0 && printedFrames < maxFrames {
				sb.WriteString(fmt.Sprintf("\n        ... repeated %d times", repeated))
			}
			_ = emitFrame(frames[j-1])
		}

		if printedFrames >= maxFrames {
			remaining := len(frames) - printedFrames
			if remaining > 0 {
				sb.WriteString(fmt.Sprintf("\n        ... %d more", remaining))
			}
			break
		}

		i = j
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
