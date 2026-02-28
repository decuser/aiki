package substrate

import (
	"bytes"
	"testing"

	"aiki/engine/semantics/value"
)

func TestFirst(t *testing.T) {
	rt := NewGoRuntime()
	list := &value.List{Elements: []value.Value{value.NewNumber(1, 1), value.NewNumber(2, 1)}}
	v, _ := rt.Execute("_first", []value.Value{list})
	if v.Inspect() != "1" {
		t.Errorf("got %s", v.Inspect())
	}
}

func TestRest(t *testing.T) {
	rt := NewGoRuntime()
	list := &value.List{Elements: []value.Value{value.NewNumber(1, 1), value.NewNumber(2, 1)}}
	v, _ := rt.Execute("_rest", []value.Value{list})
	if v.Inspect() != "[2]" {
		t.Errorf("got %s", v.Inspect())
	}
}

func TestLength(t *testing.T) {
	rt := NewGoRuntime()
	list := &value.List{Elements: []value.Value{value.NewNumber(1, 1), value.NewNumber(2, 1)}}
	v, _ := rt.Execute("_length", []value.Value{list})
	if v.Inspect() != "2" {
		t.Errorf("got %s", v.Inspect())
	}
}

func TestEmpty(t *testing.T) {
	rt := NewGoRuntime()
	v, _ := rt.Execute("_empty", []value.Value{&value.List{}})
	if v != value.TRUE {
		t.Error("expected true")
	}
}

func TestPrint(t *testing.T) {
	var buf bytes.Buffer
	Stdout = &buf
	defer func() { Stdout = nil }()

	rt := NewGoRuntime()
	rt.Execute("_print", []value.Value{&value.String{Val: "hello"}})
	if buf.String() != "hello" {
		t.Errorf("got %q", buf.String())
	}
}

func TestEqual(t *testing.T) {
	rt := NewGoRuntime()
	v, _ := rt.Execute("_equal", []value.Value{value.NewNumber(1, 1), value.NewNumber(1, 1)})
	if v != value.TRUE {
		t.Error("expected true")
	}
}

// TestUserScopeGetsNothing verifies user scope cannot access any builtins from runtime.
func TestUserScopeGetsNothing(t *testing.T) {
	rt := NewGoRuntime()

	// User cannot see anything - all access comes from prelude.ai bindings in Env
	if rt.HasBuiltin("first", value.ScopeUser) {
		t.Error("user should NOT see first from runtime")
	}
	if rt.HasBuiltin("_first", value.ScopeUser) {
		t.Error("user should NOT see _first")
	}

	_, ok := rt.GetBuiltin("first", value.ScopeUser)
	if ok {
		t.Error("user should NOT get first from runtime")
	}
	_, ok = rt.GetBuiltin("_first", value.ScopeUser)
	if ok {
		t.Error("user should NOT get _first")
	}
}

// TestPreludeScopeSeesOnlyPrefixed verifies prelude scope sees only _prefixed HAL primitives.
func TestPreludeScopeSeesOnlyPrefixed(t *testing.T) {
	rt := NewGoRuntime()

	// Prelude cannot see non-prefixed (those come from prelude.ai, not registry)
	if rt.HasBuiltin("first", value.ScopePrelude) {
		t.Error("prelude should NOT see non-prefixed 'first' from registry")
	}

	// Prelude can see _prefixed
	if !rt.HasBuiltin("_first", value.ScopePrelude) {
		t.Error("prelude should see _first")
	}

	// GetBuiltin matches HasBuiltin
	_, ok := rt.GetBuiltin("first", value.ScopePrelude)
	if ok {
		t.Error("prelude should NOT get non-prefixed 'first' from registry")
	}

	b, ok := rt.GetBuiltin("_first", value.ScopePrelude)
	if !ok || b == nil {
		t.Error("prelude GetBuiltin(_first) failed")
	}
}
