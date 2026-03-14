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
	err := NewError("test")
	num := NewNumber(1, 1)

	if !IsFault(fault) {
		t.Error("IsFault should return true for Fault")
	}
	if IsFault(err) {
		t.Error("IsFault should return false for Error")
	}
	if IsFault(num) {
		t.Error("IsFault should return false for Number")
	}
	if IsFault(nil) {
		t.Error("IsFault should return false for nil")
	}
}

func TestIsErrorNotFault(t *testing.T) {
	fault := NewFault("test")
	err := NewError("test")

	if IsError(fault) {
		t.Error("IsError should return false for Fault")
	}
	if !IsError(err) {
		t.Error("IsError should return true for Error")
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
