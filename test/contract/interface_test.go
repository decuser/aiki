package contract

import (
	"testing"

	"aiki/engine/runtime/hal"
	"aiki/engine/runtime/hal/substrate"
	"aiki/engine/semantics/value"
)

// TestGoRuntimeImplementsRuntimeContract verifies GoRuntime satisfies RuntimeContract.
func TestGoRuntimeImplementsRuntimeContract(t *testing.T) {
	var _ hal.RuntimeContract = (*substrate.GoRuntime)(nil)

	rt := substrate.NewGoRuntime()

	// HasBuiltin returns sensible values
	if !rt.HasBuiltin("_print", value.ScopePrelude) {
		t.Error("HasBuiltin should return true for _print in prelude scope")
	}
	if rt.HasBuiltin("_print", value.ScopeUser) {
		t.Error("HasBuiltin should return false for _print in user scope")
	}
	if rt.HasBuiltin("nonexistent", value.ScopePrelude) {
		t.Error("HasBuiltin should return false for nonexistent builtin")
	}

	// GetBuiltin returns callable
	callable, ok := rt.GetBuiltin("_print", value.ScopePrelude)
	if !ok {
		t.Error("GetBuiltin should return ok=true for _print")
	}
	if callable == nil {
		t.Error("GetBuiltin should return non-nil callable")
	}

	// GetBuiltin respects scope
	_, ok = rt.GetBuiltin("_print", value.ScopeUser)
	if ok {
		t.Error("GetBuiltin should return ok=false for user scope")
	}
}

// TestAllValueTypesImplementValue verifies all concrete value types implement Value.
func TestAllValueTypesImplementValue(t *testing.T) {
	tests := []struct {
		name     string
		value    value.Value
		wantType value.Type
	}{
		{"Number", value.NewNumber(42, 1), value.NumberType},
		{"Boolean true", value.TRUE, value.BooleanType},
		{"Boolean false", value.FALSE, value.BooleanType},
		{"String", &value.String{Val: "hello"}, value.StringType},
		{"Symbol", &value.Symbol{Val: "foo"}, value.SymbolType},
		{"Rune", &value.Rune{Val: 'A'}, value.RuneType},
		{"List empty", &value.List{Elements: []value.Value{}}, value.ListType},
		{"List with elements", &value.List{Elements: []value.Value{value.NewNumber(1, 1)}}, value.ListType},
		{"Shaped list", &value.List{Shape: "point", Elements: []value.Value{}}, value.ListType},
		{"Fault", value.NewFault("test error"), value.FaultType},
		{"Bytes", &value.Bytes{Val: []byte{1, 2, 3}}, value.BytesType},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Type() returns expected type
			if got := tt.value.Type(); got != tt.wantType {
				t.Errorf("Type() = %v, want %v", got, tt.wantType)
			}

			// Inspect() returns non-empty string
			inspect := tt.value.Inspect()
			if inspect == "" {
				t.Error("Inspect() returned empty string")
			}
		})
	}
}

// TestCallableInterface verifies Callable implementations can be called.
func TestCallableInterface(t *testing.T) {
	rt := substrate.NewGoRuntime()

	// Get a builtin callable
	callable, ok := rt.GetBuiltin("_length", value.ScopePrelude)
	if !ok {
		t.Fatal("could not get _length builtin")
	}

	// Verify it implements Callable
	if callable.Type() != value.FunctionType {
		t.Errorf("callable.Type() = %v, want %v", callable.Type(), value.FunctionType)
	}

	// Verify Inspect returns something
	if callable.Inspect() == "" {
		t.Error("callable.Inspect() returned empty string")
	}

	// Verify Call works
	args := []value.Value{&value.List{Elements: []value.Value{
		value.NewNumber(1, 1),
		value.NewNumber(2, 1),
		value.NewNumber(3, 1),
	}}}
	result := callable.Call(args)
	if result == nil {
		t.Fatal("Call returned nil")
	}

	num, ok := result.(*value.Number)
	if !ok {
		t.Fatalf("Call returned %T, want *value.Number", result)
	}
	if num.Inspect() != "3" {
		t.Errorf("_length([1,2,3]) = %s, want 3", num.Inspect())
	}
}

// TestFunctionImplementsCallable verifies user-defined functions implement Callable.
func TestFunctionImplementsCallable(t *testing.T) {
	fn := &value.Function{
		Name:   "test",
		Params: []string{"x"},
		Body:   nil, // Body is *syntax.Node, not needed for interface test
		Env:    nil,
	}

	// Verify it's a Value
	if fn.Type() != value.FunctionType {
		t.Errorf("Function.Type() = %v, want %v", fn.Type(), value.FunctionType)
	}

	if fn.Inspect() == "" {
		t.Error("Function.Inspect() returned empty string")
	}
}
