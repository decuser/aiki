package substrate

import (
	"testing"

	"aiki/engine/runtime/hal"
	"aiki/engine/semantics/value"
)

func TestGoProvenanceCoversHALOperationDefinitions(t *testing.T) {
	definitions := hal.OperationDefinitions()
	for primitive := range definitions {
		if goHostProvenance[primitive] == "" {
			t.Errorf("HAL operation %s has no Go substrate provenance", primitive)
		}
	}
	for primitive := range goHostProvenance {
		if _, ok := definitions[primitive]; !ok {
			t.Errorf("Go substrate provenance names unknown HAL operation %s", primitive)
		}
	}
}

func TestCapabilityAvailabilityReflectsActualBindings(t *testing.T) {
	rt := NewGoRuntime()
	if !rt.HasCapability("filesystem") {
		t.Fatal("default Go runtime should provide :filesystem")
	}
	if rt.HasCapability("does-not-exist") {
		t.Fatal("unknown capability must not report available")
	}

	rt.mu.Lock()
	delete(rt.hostBindings, "_file_stat")
	rt.mu.Unlock()
	if rt.HasCapability("filesystem") {
		t.Fatal(":filesystem must become unavailable when a required HAL binding is absent")
	}
}

func TestProfileGateRejectsMissingRequiredCapability(t *testing.T) {
	rt := NewGoRuntime()
	rt.mu.Lock()
	delete(rt.hostBindings, "_file_stat")
	rt.mu.Unlock()
	if err := rt.ValidateProfile(hal.DefaultRuntimeProfile); err == nil {
		t.Fatal("desktop profile should reject missing :filesystem")
	}
	if err := rt.ValidateProfile("minimal"); err != nil {
		t.Fatalf("minimal profile should not require host capabilities: %v", err)
	}
}

func TestSystemCapabilityQueriesUseAikiNames(t *testing.T) {
	rt := NewGoRuntime()
	if got := rt.halSystemHas([]value.Value{&value.Symbol{Val: "filesystem"}}, nil); got != value.TRUE {
		t.Fatalf("system.has(:filesystem) = %v, want true", got)
	}
	if got := rt.halSystemRequire([]value.Value{&value.Symbol{Val: "filesystem"}}, nil); got != value.EMPTY {
		t.Fatalf("system.require(:filesystem) = %v, want empty", got)
	}
	if got := rt.halSystemRequire([]value.Value{&value.Symbol{Val: "unknown"}}, nil); !value.IsShapedError(got) {
		t.Fatalf("unknown capability should return shaped :unsupported error, got %v", got)
	}
}
