package substrate

import (
	"bytes"
	"testing"

	"aiki/engine/runtime/hal"
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

func TestHasBuiltinUser(t *testing.T) {
	rt := NewGoRuntime()
	// User can see "first" but not "_first"
	if !rt.HasBuiltin("first", hal.ScopeUser) {
		t.Error("user should see first")
	}
	if rt.HasBuiltin("_first", hal.ScopeUser) {
		t.Error("user should NOT see _first")
	}
}

func TestHasBuiltinPrelude(t *testing.T) {
	rt := NewGoRuntime()
	// Prelude can see both "first" and "_first"
	if !rt.HasBuiltin("first", hal.ScopePrelude) {
		t.Error("prelude should see first")
	}
	if !rt.HasBuiltin("_first", hal.ScopePrelude) {
		t.Error("prelude should see _first")
	}
}

func TestGetBuiltinUser(t *testing.T) {
	rt := NewGoRuntime()
	// User can get "length" but not "_length"
	b, ok := rt.GetBuiltin("length", hal.ScopeUser)
	if !ok || b == nil {
		t.Error("user GetBuiltin(length) failed")
	}
	_, ok = rt.GetBuiltin("_length", hal.ScopeUser)
	if ok {
		t.Error("user should NOT get _length")
	}
}

func TestGetBuiltinPrelude(t *testing.T) {
	rt := NewGoRuntime()
	// Prelude can get both
	b, ok := rt.GetBuiltin("length", hal.ScopePrelude)
	if !ok || b == nil {
		t.Error("prelude GetBuiltin(length) failed")
	}
	b, ok = rt.GetBuiltin("_length", hal.ScopePrelude)
	if !ok || b == nil {
		t.Error("prelude GetBuiltin(_length) failed")
	}
}
