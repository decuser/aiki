package runner

import (
	"testing"
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

	// Clearing this runtime's registry must be repaired by Reset without
	// relying on package-global module state.
	s.Runtime.SetModuleRegistry(nil)
	if _, err := s.Runtime.Execute("_module_roots", nil, nil); err == nil {
		t.Fatal("expected module_roots to fail with cleared runtime registry")
	}

	if err := s.Reset(); err != nil {
		t.Fatalf("Reset: %v", err)
	}
	if _, err := s.Runtime.Execute("_module_roots", nil, nil); err != nil {
		t.Fatalf("module registry not restored by Reset: %v", err)
	}
}
