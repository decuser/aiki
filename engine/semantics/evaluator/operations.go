// operations.go implements operator and function call application.
package evaluator

import (
	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
	"fmt"
	"strings"
)

// applyOp applies a binary operator to two values.
func applyOp(op string, left, right value.Value, scope *Scope, node *syntax.Node) (value.Value, error) {
	// Numeric operations
	if left.Type == value.Number && right.Type == value.Number {
		l, lok := left.Data.(float64)
		r, rok := right.Data.(float64)
		if !lok || !rok {
			return value.NullValue(), makeError(scope, node, "invalid number data")
		}

		switch op {
		case "+":
			return value.Value{Type: value.Number, Data: l + r}, nil
		case "-":
			return value.Value{Type: value.Number, Data: l - r}, nil
		case "*":
			return value.Value{Type: value.Number, Data: l * r}, nil
		case "/":
			if r == 0 {
				return value.NullValue(), makeError(scope, node, "division by zero")
			}
			return value.Value{Type: value.Number, Data: l / r}, nil
		case "<":
			return value.Value{Type: value.Boolean, Data: l < r}, nil
		case ">":
			return value.Value{Type: value.Boolean, Data: l > r}, nil
		case "<=":
			return value.Value{Type: value.Boolean, Data: l <= r}, nil
		case ">=":
			return value.Value{Type: value.Boolean, Data: l >= r}, nil
		}
	}

	// String concatenation
	if left.Type == value.String && right.Type == value.String && op == "+" {
		l, _ := left.Data.(string)
		r, _ := right.Data.(string)
		return value.Value{Type: value.String, Data: l + r}, nil
	}

	// Boolean operations
	if op == "and" {
		return value.Value{Type: value.Boolean, Data: isTruthy(left) && isTruthy(right)}, nil
	}
	if op == "or" {
		return value.Value{Type: value.Boolean, Data: isTruthy(left) || isTruthy(right)}, nil
	}

	return value.NullValue(), makeError(scope, node, "invalid operation: %s %s %s",
		value.TypeName(left.Type), op, value.TypeName(right.Type))
}

// applyUnaryOp applies a unary operator to a value.
func applyUnaryOp(op string, operand value.Value, scope *Scope, node *syntax.Node) (value.Value, error) {
	switch op {
	case "-":
		if operand.Type != value.Number {
			return value.NullValue(), makeError(scope, node, "invalid unary -: expected number")
		}
		if n, ok := operand.Data.(float64); ok {
			return value.Value{Type: value.Number, Data: -n}, nil
		}
	case "not":
		return value.Value{Type: value.Boolean, Data: !isTruthy(operand)}, nil
	}
	return value.NullValue(), makeError(scope, node, "unknown unary operator: %s", op)
}

// applyCall applies a function to arguments.
func applyCall(e *Evaluator, fn value.Value, args []value.Value, scope *Scope, node *syntax.Node) (value.Value, error) {
	if fn.Type != value.Function {
		return value.NullValue(), makeError(scope, node, "not a function: %s", value.TypeName(fn.Type))
	}

	// Native function (HAL builtin or intrinsic)
	if name, ok := fn.Data.(string); ok {
		// Check for intrinsic
		if strings.HasPrefix(name, "intrinsic:") {
			intrinsicName := strings.TrimPrefix(name, "intrinsic:")
			return evalIntrinsic(e, intrinsicName, args, scope, node)
		}

		if e.runtime == nil {
			return value.NullValue(), makeError(scope, node, "runtime not initialized")
		}

		// Notify observer of effect
		if e.observer != nil {
			e.observer.OnEffect(name, "hal", node.Pos)
		}

		result, err := e.runtime.Execute(name, args)
		if err != nil {
			return value.NullValue(), wrapHalError(err, scope, node)
		}
		return result, nil
	}

	// User-defined function
	if userFn, ok := fn.Data.(*Function); ok {
		return applyUserFunc(e, userFn, args, scope, node)
	}

	return value.NullValue(), makeError(scope, node, "invalid function data")
}

// applyUserFunc applies a user-defined function to arguments.
func applyUserFunc(e *Evaluator, fn *Function, args []value.Value, scope *Scope, node *syntax.Node) (value.Value, error) {
	// Create call scope from function's closure environment
	callScope := NewScope(fn.Env)

	// Bind parameters
	params := fn.Params
	restIdx := -1

	for i, param := range params {
		if strings.HasPrefix(param, "...") {
			restIdx = i
			break
		}
	}

	if restIdx >= 0 {
		// Has rest parameter
		restName := strings.TrimPrefix(params[restIdx], "...")

		// Bind regular params
		for i := 0; i < restIdx && i < len(args); i++ {
			callScope.Define(params[i], args[i])
		}

		// Bind rest as list
		var restArgs []value.Value
		if len(args) > restIdx {
			restArgs = args[restIdx:]
		}
		callScope.Define(restName, value.Value{Type: value.List, Data: restArgs})
	} else {
		// No rest parameter - check arity
		if len(args) != len(params) {
			return value.NullValue(), makeError(scope, node, "arity mismatch: %s expects %d arguments, got %d",
				fn.Name, len(params), len(args))
		}

		for i, param := range params {
			callScope.Define(param, args[i])
		}
	}

	// Push stack frame
	file := scope.GetFile()
	line := nodePosition(node)
	name := fn.Name
	if name == "" {
		name = "<anonymous>"
	}
	callScope.PushFrame(name, file, line, fn.Layer)

	// Evaluate body
	result, err := e.eval(fn.Body, callScope)

	// Pop stack frame
	callScope.PopFrame()

	if err != nil {
		return result, err
	}

	// Unwrap return value
	if result.Type == value.Return {
		if ret, ok := result.Data.(*ReturnValue); ok {
			return ret.Value, nil
		}
	}

	return result, nil
}

// applyIndex applies index access to a value.
func applyIndex(val, idx value.Value, scope *Scope, node *syntax.Node) (value.Value, error) {
	// List indexing
	if val.Type == value.List {
		if idx.Type != value.Number {
			return value.NullValue(), makeError(scope, node, "index must be a number")
		}

		i := int(idx.Data.(float64))

		// Handle regular list
		if list, ok := val.Data.([]value.Value); ok {
			if i < 0 || i >= len(list) {
				return value.NullValue(), makeError(scope, node, "index out of bounds: %d", i)
			}
			return list[i], nil
		}

		// Handle shaped list
		if shaped, ok := val.Data.(*ShapedList); ok {
			if i < 0 || i >= len(shaped.Elements) {
				return value.NullValue(), makeError(scope, node, "index out of bounds: %d", i)
			}
			return shaped.Elements[i], nil
		}
	}

	// String indexing
	if val.Type == value.String {
		if idx.Type != value.Number {
			return value.NullValue(), makeError(scope, node, "index must be a number")
		}

		s, _ := val.Data.(string)
		i := int(idx.Data.(float64))
		runes := []rune(s)

		if i < 0 || i >= len(runes) {
			return value.NullValue(), makeError(scope, node, "index out of bounds: %d", i)
		}

		return value.Value{Type: value.Rune, Data: runes[i]}, nil
	}

	return value.NullValue(), makeError(scope, node, "cannot index %s", value.TypeName(val.Type))
}

// applyAccess applies field access to a value.
func applyAccess(val value.Value, field string, scope *Scope, node *syntax.Node) (value.Value, error) {
	// Shaped list field access
	if val.Type == value.List {
		if shaped, ok := val.Data.(*ShapedList); ok {
			// Get shape definition
			shapeDef, ok := scope.GetShape(shaped.Shape)
			if !ok {
				return value.NullValue(), makeError(scope, node, "unknown shape: %s", shaped.Shape)
			}

			// Find field index
			for i, f := range shapeDef.Fields {
				if f == field {
					if i < len(shaped.Elements) {
						return shaped.Elements[i], nil
					}
					return value.NullValue(), makeError(scope, node, "field %s not set", field)
				}
			}

			return value.NullValue(), makeError(scope, node, "shape %s has no field %s", shaped.Shape, field)
		}

		// Numeric field access on plain list (e.g., list.0)
		if idx, err := value.ParseNumber(field); err == nil {
			return applyIndex(val, value.Value{Type: value.Number, Data: idx}, scope, node)
		}
	}

	return value.NullValue(), makeError(scope, node, "cannot access field %s on %s", field, value.TypeName(val.Type))
}

// String returns a string representation for error messages.
func (f *Function) String() string {
	if f.Name != "" {
		return fmt.Sprintf("<function %s>", f.Name)
	}
	return "<function>"
}
