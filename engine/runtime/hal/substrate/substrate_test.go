package substrate

import (
	"bytes"
	"testing"
	"time"

	"aiki/engine/semantics/value"
)

func TestFirst(t *testing.T) {
	rt := NewGoRuntime()
	list := &value.List{Elements: []value.Value{value.NewNumber(1, 1), value.NewNumber(2, 1)}}
	v, _ := rt.Execute("_first", []value.Value{list}, nil)
	if v.Inspect() != "1" {
		t.Errorf("got %s", v.Inspect())
	}
}

func TestRest(t *testing.T) {
	rt := NewGoRuntime()
	list := &value.List{Elements: []value.Value{value.NewNumber(1, 1), value.NewNumber(2, 1)}}
	v, _ := rt.Execute("_rest", []value.Value{list}, nil)
	if v.Inspect() != "[2]" {
		t.Errorf("got %s", v.Inspect())
	}
}

func TestLength(t *testing.T) {
	rt := NewGoRuntime()
	list := &value.List{Elements: []value.Value{value.NewNumber(1, 1), value.NewNumber(2, 1)}}
	v, _ := rt.Execute("_length", []value.Value{list}, nil)
	if v.Inspect() != "2" {
		t.Errorf("got %s", v.Inspect())
	}
}

func TestEmpty(t *testing.T) {
	rt := NewGoRuntime()
	v, _ := rt.Execute("_empty", []value.Value{&value.List{}}, nil)
	if v != value.TRUE {
		t.Error("expected true")
	}
}

func TestPrint(t *testing.T) {
	var buf bytes.Buffer
	Stdout = &buf
	defer func() { Stdout = nil }()

	rt := NewGoRuntime()
	rt.Execute("_print", []value.Value{&value.String{Val: "hello"}}, nil)
	if buf.String() != "hello" {
		t.Errorf("got %q", buf.String())
	}
}

func TestEqual(t *testing.T) {
	rt := NewGoRuntime()
	v, _ := rt.Execute("_equal", []value.Value{value.NewNumber(1, 1), value.NewNumber(1, 1)}, nil)
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

func TestRandom(t *testing.T) {
	// random returns value in range
	result := halRandom([]value.Value{value.NewNumber(10, 1)}, nil)
	n, ok := result.(*value.Number)
	if !ok {
		t.Fatalf("expected Number, got %T", result)
	}
	v := n.Val.Num().Int64()
	if v < 0 || v >= 10 {
		t.Errorf("random(10) returned %d, want 0-9", v)
	}
}

func TestSeedReproducible(t *testing.T) {
	// same seed = same sequence
	halSeed([]value.Value{value.NewNumber(42, 1)}, nil)
	r1 := halRandom([]value.Value{value.NewNumber(100, 1)}, nil).(*value.Number)

	halSeed([]value.Value{value.NewNumber(42, 1)}, nil)
	r2 := halRandom([]value.Value{value.NewNumber(100, 1)}, nil).(*value.Number)

	if r1.Val.Cmp(r2.Val) != 0 {
		t.Errorf("same seed gave different results: %s vs %s", r1.Inspect(), r2.Inspect())
	}
}

func TestRandomErrors(t *testing.T) {
	// zero max
	result := halRandom([]value.Value{value.NewNumber(0, 1)}, nil)
	if !value.IsFault(result) {
		t.Error("random(0) should fault")
	}

	// negative max
	result = halRandom([]value.Value{value.NewNumber(-5, 1)}, nil)
	if !value.IsFault(result) {
		t.Error("random(-5) should fault")
	}
}

func TestChannel(t *testing.T) {
	result := halChannel([]value.Value{}, nil)
	ch, ok := result.(*value.Channel)
	if !ok {
		t.Fatalf("expected Channel, got %T", result)
	}
	if ch.C == nil {
		t.Error("channel is nil")
	}
}

func TestSendRecv(t *testing.T) {
	ch := value.NewChannel()
	val := value.NewNumber(42, 1)

	// Send in goroutine (blocks until recv)
	go func() {
		halSend([]value.Value{ch, val}, nil)
	}()

	// Recv
	result := halRecv([]value.Value{ch}, nil)
	n, ok := result.(*value.Number)
	if !ok {
		t.Fatalf("expected Number, got %T", result)
	}
	if n.Inspect() != "42" {
		t.Errorf("got %s, want 42", n.Inspect())
	}
}

func TestSendRecvBlocks(t *testing.T) {
	ch := value.NewChannel()
	done := make(chan bool)

	go func() {
		halRecv([]value.Value{ch}, nil)
		done <- true
	}()

	// Should not complete yet
	select {
	case <-done:
		t.Error("recv should block")
	case <-time.After(10 * time.Millisecond):
		// good
	}

	// Now send
	halSend([]value.Value{ch, value.TRUE}, nil)

	// Should complete
	select {
	case <-done:
		// good
	case <-time.After(100 * time.Millisecond):
		t.Error("recv should have completed")
	}
}
