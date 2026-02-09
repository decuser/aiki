package value

import (
	"fmt"
	"math/big"
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
	ErrorType    Type = "error"
	NullType     Type = "null"
	ReturnType   Type = "return" // internal, for control flow
)

type Value interface {
	Type() Type
	Inspect() string
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
	Shape    string   // empty for raw lists
	Fields   []string // field names if shaped
}

func (l *List) Type() Type { return ListType }
func (l *List) Inspect() string {
	parts := make([]string, len(l.Elements))
	for i, e := range l.Elements {
		parts[i] = e.Inspect()
	}
	if l.Shape != "" {
		return fmt.Sprintf("[@%s %s]", l.Shape, strings.Join(parts, " "))
	}
	return fmt.Sprintf("[%s]", strings.Join(parts, " "))
}

// Function
type Function struct {
	Parameters []string
	Body       *ast.BlockStatement
	Env        *Env
}

func (f *Function) Type() Type { return FunctionType }
func (f *Function) Inspect() string {
	return fmt.Sprintf("(%s) { ... }", strings.Join(f.Parameters, " "))
}

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
}

func (e *Error) Type() Type      { return ErrorType }
func (e *Error) Inspect() string { return "error: " + e.Message }

func NewError(format string, a ...interface{}) *Error {
	return &Error{Message: fmt.Sprintf(format, a...)}
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
