package value

import (
	"fmt"
	"math/big"
	"os"
	"strings"

	"aiki/ast"
)

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
	HandleType   Type = "handle"
	ErrorType    Type = "error"
	NullType     Type = "null"
	ReturnType   Type = "return"
)

type Value interface {
	Type() Type
	Inspect() string
}

// StackFrame represents a call site in the stack trace.
type StackFrame struct {
	Name string
	File string
	Line int
}

// Number is a rational (exact arithmetic).
type Number struct {
	Value *big.Rat
}

func (n *Number) Type() Type { return NumberType }
func (n *Number) Inspect() string {
	if n.Value.IsInt() {
		return n.Value.Num().String()
	}
	return n.Value.RatString()
}

func NewNumber(num, denom int64) *Number {
	return &Number{Value: big.NewRat(num, denom)}
}

func NewNumberFromString(s string) (*Number, error) {
	r := new(big.Rat)
	if _, ok := r.SetString(s); !ok {
		return nil, fmt.Errorf("invalid number: %s", s)
	}
	return &Number{Value: r}, nil
}

// Boolean
type Boolean struct {
	Value bool
}

func (b *Boolean) Type() Type { return BooleanType }
func (b *Boolean) Inspect() string {
	if b.Value {
		return "true"
	}
	return "false"
}

var (
	True  = &Boolean{Value: true}
	False = &Boolean{Value: false}
)

// Rune
type Rune struct {
	Value rune
}

func (r *Rune) Type() Type      { return RuneType }
func (r *Rune) Inspect() string { return fmt.Sprintf("'%c'", r.Value) }

// String
type String struct {
	Value string
}

func (s *String) Type() Type      { return StringType }
func (s *String) Inspect() string { return fmt.Sprintf("%q", s.Value) }

// Bytes
type Bytes struct {
	Value []byte
}

func (b *Bytes) Type() Type { return BytesType }
func (b *Bytes) Inspect() string {
	parts := make([]string, len(b.Value))
	for i, v := range b.Value {
		parts[i] = fmt.Sprintf("%d", v)
	}
	return fmt.Sprintf("[%s]", strings.Join(parts, " "))
}

// Symbol
type Symbol struct {
	Value string
}

func (s *Symbol) Type() Type      { return SymbolType }
func (s *Symbol) Inspect() string { return ":" + s.Value }

// List (raw or shaped)
type List struct {
	Elements []Value
	Shape    string
	Fields   []string
}

func (l *List) Type() Type { return ListType }
func (l *List) Inspect() string {
	strs := make([]string, len(l.Elements))
	for i, e := range l.Elements {
		strs[i] = e.Inspect()
	}
	if l.Shape != "" {
		return fmt.Sprintf("[@%s %s]", l.Shape, strings.Join(strs, " "))
	}
	return fmt.Sprintf("[%s]", strings.Join(strs, " "))
}

// Function
type Function struct {
	Name       string
	Parameters []string
	Body       *ast.BlockStatement
	Env        *Env
}

func (f *Function) Type() Type { return FunctionType }
func (f *Function) Inspect() string {
	return fmt.Sprintf("(%s) { ... }", strings.Join(f.Parameters, " "))
}

// Handle wraps an OS file handle
type Handle struct {
	File *os.File
	Path string
}

func (h *Handle) Type() Type      { return HandleType }
func (h *Handle) Inspect() string { return fmt.Sprintf("<handle: %s>", h.Path) }

// Builtin is a primitive function implemented in Go.
type Builtin struct {
	Name string
	Fn   func(args ...Value) Value
}

func (b *Builtin) Type() Type      { return FunctionType }
func (b *Builtin) Inspect() string { return fmt.Sprintf("<builtin: %s>", b.Name) }

// Error
type Error struct {
	Message string
	File    string
	Line    int
	Source  string // the source line where error occurred
	Stack   []StackFrame
}

func (e *Error) Type() Type { return ErrorType }

// Inspect returns Ruby-style error format:
// file:line:in 'func': message
//     source line
//         from file:line:in 'func'
//         from file:line:in 'func'
func (e *Error) Inspect() string {
	var sb strings.Builder

	// Determine function name for error location
	// Stack is ordered oldest->newest, so last element is innermost (where error occurred)
	funcName := "<main>"
	if len(e.Stack) > 0 {
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
	// Iterate from second-to-last down to first (newest to oldest callers)
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

func NewError(format string, a ...interface{}) *Error {
	return &Error{Message: fmt.Sprintf(format, a...)}
}

func NewErrorAt(file string, line int, source string, stack []StackFrame, format string, a ...interface{}) *Error {
	return &Error{
		Message: fmt.Sprintf(format, a...),
		File:    file,
		Line:    line,
		Source:  source,
		Stack:   stack,
	}
}

// Null
type Null struct{}

func (n *Null) Type() Type      { return NullType }
func (n *Null) Inspect() string { return "null" }

var NULL = &Null{}

// Return wraps a value for return statement control flow.
type Return struct {
	Value Value
}

func (r *Return) Type() Type      { return ReturnType }
func (r *Return) Inspect() string { return r.Value.Inspect() }
