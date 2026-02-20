// errors.go provides rich error construction with position context.
package evaluator

import (
	"aiki/engine"
	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
	"fmt"
	"strings"
)

// EvalError represents an evaluation error with context.
type EvalError struct {
	Message string
	File    string
	Line    int
	Column  int
	Source  string
	Stack   []StackFrame
}

func (e *EvalError) Error() string {
	return e.Message
}

// Inspect returns a formatted error string for display.
func (e *EvalError) Inspect() string {
	return e.InspectAtLayer(LayerUser)
}

// InspectAtLayer returns a formatted error filtered by layer visibility.
func (e *EvalError) InspectAtLayer(layer Layer) string {
	var b strings.Builder

	// Header: file:line: message
	if e.File != "" {
		fmt.Fprintf(&b, "%s:", e.File)
	}
	if e.Line > 0 {
		fmt.Fprintf(&b, "%d:", e.Line)
	}
	if e.File != "" || e.Line > 0 {
		b.WriteString(" ")
	}
	b.WriteString(e.Message)
	b.WriteString("\n")

	// Source line with pointer
	if e.Source != "" {
		b.WriteString("    ")
		b.WriteString(e.Source)
		b.WriteString("\n")
	}

	// Stack trace (filtered by layer)
	if len(e.Stack) > 0 {
		for i := len(e.Stack) - 1; i >= 0; i-- {
			frame := e.Stack[i]

			// Filter by layer visibility
			if frame.Layer > layer {
				continue
			}

			fmt.Fprintf(&b, "    from '%s'", frame.Name)
			if frame.File != "" {
				fmt.Fprintf(&b, " at %s", frame.File)
			}
			if frame.Line > 0 {
				fmt.Fprintf(&b, ":%d", frame.Line)
			}
			b.WriteString("\n")
		}
	}

	return b.String()
}

// ToValue converts an EvalError to a value.Value for error propagation.
func (e *EvalError) ToValue() value.Value {
	return value.Value{
		Type: value.Error,
		Data: e,
	}
}

// makeError creates a rich error with file, line, source context.
func makeError(scope *Scope, node *syntax.Node, format string, args ...interface{}) *EvalError {
	line := 0
	col := 0
	if node != nil {
		line = node.Pos.Line
		col = node.Pos.Col
		if line == 0 {
			line = nodePosition(node)
		}
	}

	return &EvalError{
		Message: fmt.Sprintf(format, args...),
		File:    scope.GetFile(),
		Line:    line,
		Column:  col,
		Source:  scope.GetSourceLine(line),
		Stack:   scope.CopyStack(),
	}
}

// makeErrorAt creates an error at a specific position.
func makeErrorAt(file string, pos engine.Position, format string, args ...interface{}) *EvalError {
	return &EvalError{
		Message: fmt.Sprintf(format, args...),
		File:    file,
		Line:    pos.Line,
		Column:  pos.Col,
	}
}

// isError checks if a value is an error.
func isError(val value.Value) bool {
	return val.Type == value.Error
}

// asError extracts the EvalError from a value, if present.
func asError(val value.Value) *EvalError {
	if val.Type == value.Error {
		if e, ok := val.Data.(*EvalError); ok {
			return e
		}
	}
	return nil
}

// wrapHalError annotates a HAL error with call site position.
func wrapHalError(err error, scope *Scope, node *syntax.Node) *EvalError {
	return &EvalError{
		Message: err.Error(),
		File:    scope.GetFile(),
		Line:    nodePosition(node),
		Source:  scope.GetSourceLine(nodePosition(node)),
		Stack:   scope.CopyStack(),
	}
}
