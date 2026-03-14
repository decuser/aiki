package value

import "testing"

func TestNumber(t *testing.T) {
	n := NewNumber(3, 4)
	if n.Inspect() != "3/4" {
		t.Errorf("got %s", n.Inspect())
	}
	if n.Type() != NumberType {
		t.Error("wrong type")
	}
}

func TestNumberInt(t *testing.T) {
	n := NewNumber(42, 1)
	if n.Inspect() != "42" {
		t.Errorf("got %s", n.Inspect())
	}
}

func TestNumberFromString(t *testing.T) {
	n, err := NewNumberFromString("1/3")
	if err != nil || n.Inspect() != "1/3" {
		t.Errorf("got %v, %v", n, err)
	}
}

func TestBoolean(t *testing.T) {
	if TRUE.Inspect() != "true" || FALSE.Inspect() != "false" {
		t.Error("boolean inspect")
	}
}

func TestString(t *testing.T) {
	s := &String{Val: "hello"}
	if s.Inspect() != "hello" || s.Type() != StringType {
		t.Error("string")
	}
}

func TestList(t *testing.T) {
	l := &List{Elements: []Value{NewNumber(1, 1), NewNumber(2, 1)}}
	if l.Inspect() != "[1, 2]" {
		t.Errorf("got %s", l.Inspect())
	}
}

func TestIsTruthy(t *testing.T) {
	if !IsTruthy(TRUE) || IsTruthy(FALSE) || IsTruthy(EMPTY) || !IsTruthy(NewNumber(0, 1)) {
		t.Error("truthy")
	}
}

func TestShapedErrorFalsy(t *testing.T) {
	se := NewShapedError("io", "test error")
	if IsTruthy(se) {
		t.Error("shaped error should be falsy")
	}

	// Regular non-empty list should still be truthy
	list := &List{Elements: []Value{NewNumber(1, 1)}}
	if !IsTruthy(list) {
		t.Error("non-empty list should be truthy")
	}

	// Other shaped lists should still be truthy
	other := &List{Shape: "point", Elements: []Value{NewNumber(1, 1), NewNumber(2, 1)}}
	if !IsTruthy(other) {
		t.Error("non-error shaped list should be truthy")
	}
}

func TestEnv(t *testing.T) {
	e := NewEnv()
	e.Set("x", NewNumber(1, 1))
	v, ok := e.Get("x")
	if !ok || v.Inspect() != "1" {
		t.Error("env get")
	}
}

func TestEnvEnclosed(t *testing.T) {
	outer := NewEnv()
	outer.Set("x", NewNumber(1, 1))
	inner := NewEnclosedEnv(outer)
	v, ok := inner.Get("x")
	if !ok || v.Inspect() != "1" {
		t.Error("enclosed get")
	}
}

func TestFaultType(t *testing.T) {
	f := NewFault("test error")
	if f.Type() != FaultType {
		t.Errorf("expected FaultType, got %s", f.Type())
	}
}

func TestFaultMessage(t *testing.T) {
	f := NewFault("division by %s", "zero")
	if f.Message != "division by zero" {
		t.Errorf("expected 'division by zero', got %s", f.Message)
	}
}

func TestFaultAt(t *testing.T) {
	stack := []StackFrame{{Name: "foo", File: "test.aiki", Line: 5, Scope: ScopeUser}}
	f := NewFaultAt("test.aiki", 10, "let x = 1/0", stack, "division by zero")

	if f.File != "test.aiki" {
		t.Errorf("expected file 'test.aiki', got %s", f.File)
	}
	if f.Line != 10 {
		t.Errorf("expected line 10, got %d", f.Line)
	}
	if f.Source != "let x = 1/0" {
		t.Errorf("expected source 'let x = 1/0', got %s", f.Source)
	}
	if len(f.Stack) != 1 || f.Stack[0].Name != "foo" {
		t.Error("stack not preserved")
	}
}

func TestFaultInspect(t *testing.T) {
	stack := []StackFrame{
		{Name: "outer", File: "test.aiki", Line: 1, Scope: ScopeUser},
		{Name: "inner", File: "test.aiki", Line: 10, Scope: ScopeUser},
	}
	f := NewFaultAt("test.aiki", 10, "let x = 1/0", stack, "division by zero")
	inspect := f.Inspect()

	// Should contain file:line:in 'func': message
	if !contains(inspect, "test.aiki:10:in 'inner': division by zero") {
		t.Errorf("inspect missing header, got: %s", inspect)
	}
	// Should contain source line
	if !contains(inspect, "let x = 1/0") {
		t.Errorf("inspect missing source, got: %s", inspect)
	}
	// Should contain stack trace
	if !contains(inspect, "from test.aiki:1:in 'outer'") {
		t.Errorf("inspect missing stack trace, got: %s", inspect)
	}
}

func TestIsFault(t *testing.T) {
	fault := NewFault("test")
	num := NewNumber(1, 1)
	shapedErr := NewShapedError("test", "test error")

	if !IsFault(fault) {
		t.Error("IsFault should return true for Fault")
	}
	if IsFault(shapedErr) {
		t.Error("IsFault should return false for shaped error")
	}
	if IsFault(num) {
		t.Error("IsFault should return false for Number")
	}
	if IsFault(nil) {
		t.Error("IsFault should return false for nil")
	}
}

func TestNewShapedError(t *testing.T) {
	se := NewShapedError("io", "file not found: %s", "test.txt")

	if se.Shape != "error" {
		t.Errorf("expected shape 'error', got %s", se.Shape)
	}
	if len(se.Elements) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(se.Elements))
	}

	kind, ok := se.Elements[0].(*Symbol)
	if !ok || kind.Val != "io" {
		t.Errorf("expected :io, got %v", se.Elements[0])
	}

	msg, ok := se.Elements[1].(*String)
	if !ok || msg.Val != "file not found: test.txt" {
		t.Errorf("expected message, got %v", se.Elements[1])
	}
}

func TestIsShapedError(t *testing.T) {
	se := NewShapedError("io", "test")

	if !IsShapedError(se) {
		t.Error("IsShapedError should return true for shaped error")
	}

	// Not a shaped error: regular list
	list := &List{Elements: []Value{NewNumber(1, 1)}}
	if IsShapedError(list) {
		t.Error("IsShapedError should return false for regular list")
	}

	// Not a shaped error: wrong shape name
	wrongShape := &List{Shape: "other", Elements: []Value{&Symbol{Val: "x"}, &String{Val: "y"}}}
	if IsShapedError(wrongShape) {
		t.Error("IsShapedError should return false for wrong shape")
	}

	// Not a shaped error: wrong element types
	wrongTypes := &List{Shape: "error", Elements: []Value{&String{Val: "x"}, &String{Val: "y"}}}
	if IsShapedError(wrongTypes) {
		t.Error("IsShapedError should return false for wrong element types")
	}

	// Not a shaped error: wrong element count
	wrongCount := &List{Shape: "error", Elements: []Value{&Symbol{Val: "x"}}}
	if IsShapedError(wrongCount) {
		t.Error("IsShapedError should return false for wrong element count")
	}

	// Not a shaped error: Fault
	fault := NewFault("test")
	if IsShapedError(fault) {
		t.Error("IsShapedError should return false for Fault")
	}

	// Not a shaped error: plain number
	num := NewNumber(1, 1)
	if IsShapedError(num) {
		t.Error("IsShapedError should return false for Number")
	}
}

func TestShapedErrorInspect(t *testing.T) {
	se := NewShapedError("hal", "resource closed")
	expected := "[@error, :hal, resource closed]"
	if se.Inspect() != expected {
		t.Errorf("expected %s, got %s", expected, se.Inspect())
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
