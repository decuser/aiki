package substrate

import (
	"bytes"
	"testing"

	"aiki/engine/semantics/value"
)

func TestFirst(t *testing.T) {
	rt := NewGoRuntime()
	list := &value.List{Elements: []value.Value{value.NewNumber(1, 1), value.NewNumber(2, 1)}}
	v, _ := rt.Execute("first", []value.Value{list})
	if v.Inspect() != "1" {
		t.Errorf("got %s", v.Inspect())
	}
}

func TestRest(t *testing.T) {
	rt := NewGoRuntime()
	list := &value.List{Elements: []value.Value{value.NewNumber(1, 1), value.NewNumber(2, 1)}}
	v, _ := rt.Execute("rest", []value.Value{list})
	if v.Inspect() != "[2]" {
		t.Errorf("got %s", v.Inspect())
	}
}

func TestLength(t *testing.T) {
	rt := NewGoRuntime()
	list := &value.List{Elements: []value.Value{value.NewNumber(1, 1), value.NewNumber(2, 1)}}
	v, _ := rt.Execute("length", []value.Value{list})
	if v.Inspect() != "2" {
		t.Errorf("got %s", v.Inspect())
	}
}

func TestEmpty(t *testing.T) {
	rt := NewGoRuntime()
	v, _ := rt.Execute("empty", []value.Value{&value.List{}})
	if v != value.TRUE {
		t.Error("expected true")
	}
}

func TestPrint(t *testing.T) {
	var buf bytes.Buffer
	Stdout = &buf
	defer func() { Stdout = nil }()

	rt := NewGoRuntime()
	rt.Execute("print", []value.Value{&value.String{Val: "hello"}})
	if buf.String() != "hello" {
		t.Errorf("got %q", buf.String())
	}
}

func TestEqual(t *testing.T) {
	rt := NewGoRuntime()
	v, _ := rt.Execute("equal", []value.Value{value.NewNumber(1, 1), value.NewNumber(1, 1)})
	if v != value.TRUE {
		t.Error("expected true")
	}
}

func TestHasBuiltin(t *testing.T) {
	rt := NewGoRuntime()
	if !rt.HasBuiltin("first") {
		t.Error("missing first")
	}
	if rt.HasBuiltin("nonexistent") {
		t.Error("unexpected builtin")
	}
}

func TestGetBuiltin(t *testing.T) {
	rt := NewGoRuntime()
	b, ok := rt.GetBuiltin("length")
	if !ok || b == nil {
		t.Error("GetBuiltin failed")
	}
}
