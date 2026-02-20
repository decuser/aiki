// Package value defines the core value types for Aiki.
package value

import (
	"fmt"
	"strconv"
	"strings"
)

// Kind represents the type of a value.
type Kind int

const (
	Invalid Kind = iota
	Null
	Number
	Boolean
	String
	Rune
	Bytes
	Symbol
	List
	Function
	Return // Internal: return value wrapper
	Error  // Internal: error wrapper
	Handle
	Channel
	Canvas
)

// Value represents an Aiki value.
type Value struct {
	Type Kind
	Data any
}

// NullValue returns a null value.
func NullValue() Value {
	return Value{Type: Null}
}

// True returns a true boolean value.
func True() Value {
	return Value{Type: Boolean, Data: true}
}

// False returns a false boolean value.
func False() Value {
	return Value{Type: Boolean, Data: false}
}

// NewNumber creates a number value.
func NewNumber(n float64) Value {
	return Value{Type: Number, Data: n}
}

// NewString creates a string value.
func NewString(s string) Value {
	return Value{Type: String, Data: s}
}

// NewList creates a list value.
func NewList(elements []Value) Value {
	return Value{Type: List, Data: elements}
}

// NewSymbol creates a symbol value.
func NewSymbol(s string) Value {
	return Value{Type: Symbol, Data: s}
}

// String returns a string representation of the value.
func (v Value) String() string {
	return v.Inspect()
}

// Inspect returns a human-readable string representation.
func (v Value) Inspect() string {
	switch v.Type {
	case Invalid:
		return "<invalid>"
	case Null:
		return "null"
	case Number:
		if f, ok := v.Data.(float64); ok {
			// Check if it's a whole number
			if f == float64(int64(f)) {
				return fmt.Sprintf("%d", int64(f))
			}
			return fmt.Sprintf("%g", f)
		}
		return fmt.Sprintf("%v", v.Data)
	case Boolean:
		if b, ok := v.Data.(bool); ok {
			if b {
				return "true"
			}
			return "false"
		}
		return fmt.Sprintf("%v", v.Data)
	case String:
		if s, ok := v.Data.(string); ok {
			return s
		}
		return fmt.Sprintf("%v", v.Data)
	case Rune:
		if r, ok := v.Data.(rune); ok {
			return fmt.Sprintf("'%c'", r)
		}
		return fmt.Sprintf("%v", v.Data)
	case Bytes:
		if b, ok := v.Data.([]byte); ok {
			return fmt.Sprintf("b\"%s\"", string(b))
		}
		return fmt.Sprintf("%v", v.Data)
	case Symbol:
		if s, ok := v.Data.(string); ok {
			return ":" + s
		}
		return fmt.Sprintf(":%v", v.Data)
	case List:
		return inspectList(v)
	case Function:
		if name, ok := v.Data.(string); ok {
			return fmt.Sprintf("<builtin %s>", name)
		}
		return "<function>"
	case Return:
		return "<return>"
	case Error:
		if err, ok := v.Data.(error); ok {
			return err.Error()
		}
		return "<error>"
	case Handle:
		return "<handle>"
	case Channel:
		return "<channel>"
	case Canvas:
		return "<canvas>"
	default:
		return fmt.Sprintf("<%d: %v>", v.Type, v.Data)
	}
}

func inspectList(v Value) string {
	var elements []Value

	switch data := v.Data.(type) {
	case []Value:
		elements = data
	default:
		return "[]"
	}

	var parts []string
	for _, elem := range elements {
		parts = append(parts, elem.Inspect())
	}

	return "[" + strings.Join(parts, ", ") + "]"
}

// TypeName returns the type name for a Kind.
func TypeName(k Kind) string {
	switch k {
	case Invalid:
		return "invalid"
	case Null:
		return "null"
	case Number:
		return "number"
	case Boolean:
		return "boolean"
	case String:
		return "string"
	case Rune:
		return "rune"
	case Bytes:
		return "bytes"
	case Symbol:
		return "symbol"
	case List:
		return "list"
	case Function:
		return "function"
	case Return:
		return "return"
	case Error:
		return "error"
	case Handle:
		return "handle"
	case Channel:
		return "channel"
	case Canvas:
		return "canvas"
	default:
		return fmt.Sprintf("kind(%d)", k)
	}
}

// ParseNumber parses a number string (int, float, or rational).
func ParseNumber(s string) (float64, error) {
	// Check for rational (e.g., "3/4")
	if idx := strings.Index(s, "/"); idx > 0 {
		num, err := strconv.ParseFloat(s[:idx], 64)
		if err != nil {
			return 0, err
		}
		den, err := strconv.ParseFloat(s[idx+1:], 64)
		if err != nil {
			return 0, err
		}
		if den == 0 {
			return 0, fmt.Errorf("division by zero in rational")
		}
		return num / den, nil
	}

	return strconv.ParseFloat(s, 64)
}

// IsTruthy returns whether the value is truthy.
func (v Value) IsTruthy() bool {
	switch v.Type {
	case Boolean:
		if b, ok := v.Data.(bool); ok {
			return b
		}
		return false
	case Null:
		return false
	default:
		return true
	}
}

// AsNumber extracts a float64 from a number value.
func (v Value) AsNumber() (float64, bool) {
	if v.Type == Number {
		if f, ok := v.Data.(float64); ok {
			return f, true
		}
	}
	return 0, false
}

// AsString extracts a string from a string value.
func (v Value) AsString() (string, bool) {
	if v.Type == String {
		if s, ok := v.Data.(string); ok {
			return s, true
		}
	}
	return "", false
}

// AsBool extracts a bool from a boolean value.
func (v Value) AsBool() (bool, bool) {
	if v.Type == Boolean {
		if b, ok := v.Data.(bool); ok {
			return b, true
		}
	}
	return false, false
}

// AsList extracts elements from a list value.
func (v Value) AsList() ([]Value, bool) {
	if v.Type == List {
		if elems, ok := v.Data.([]Value); ok {
			return elems, true
		}
	}
	return nil, false
}
