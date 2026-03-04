package runner

import (
	"testing"

	"aiki/engine/runtime/hal/substrate"
)

func TestSessionResetSignalAndRegistry(t *testing.T) {
	s, err := NewSession()
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}

	v := s.Eval(`reset()`)
	if v == nil || v.Type() != "reset" {
		if v == nil {
			t.Fatalf("expected reset value, got nil")
		}
		t.Fatalf("expected reset value, got %s", v.Type())
	}

	// Seed registry pointer so we can detect rebuild.
	old := substrate.GlobalRegistry
	if old == nil {
		// Ensure non nil for comparison by initializing once.
		if err := s.Reset(); err != nil {
			t.Fatalf("Reset: %v", err)
		}
		old = substrate.GlobalRegistry
		if old == nil {
			t.Fatalf("expected registry after Reset")
		}
	}

	if err := s.Reset(); err != nil {
		t.Fatalf("Reset: %v", err)
	}
	if substrate.GlobalRegistry == nil {
		t.Fatalf("expected GlobalRegistry non nil after Reset")
	}
	if substrate.GlobalRegistry == old {
		t.Fatalf("expected registry to be rebuilt on Reset")
	}
}
