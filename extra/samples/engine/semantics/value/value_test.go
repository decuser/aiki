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
