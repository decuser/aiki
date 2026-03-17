package invariant

import (
	"testing"

	"aiki/engine/runtime/hal/substrate"
	"aiki/engine/semantics/value"
)

// TestHALAllRegistrationsNonNil verifies all registered HAL functions are non-nil.
// The register() method panics if fn is nil, so if NewGoRuntime() succeeds,
// all registrations are valid.
func TestHALAllRegistrationsNonNil(t *testing.T) {
	// Creating a GoRuntime runs registerHAL() which would panic if any are nil.
	// If we get here without panic, all registrations are valid.
	rt := substrate.NewGoRuntime()

	// Verify we have a reasonable number of registrations
	names := rt.BuiltinNames(value.ScopePrelude)
	if len(names) < 50 {
		t.Errorf("expected at least 50 HAL registrations, got %d", len(names))
	}
}
