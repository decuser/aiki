package invariant

import (
	"strings"
	"testing"

	"aiki/engine/runtime/hal"
	"aiki/engine/runtime/primitives"
)

func TestPrimitiveRolesAreCoherent(t *testing.T) {
	if err := primitives.ValidateCanonicalDefinitions(); err != nil {
		t.Fatal(err)
	}
}

func TestPrimitiveRoleInvariantRejectsMissingHALPrimitive(t *testing.T) {
	defs := primitives.Definitions()
	delete(defs, "_file_stat")
	err := primitives.ValidateDefinitions(defs, hal.OperationDefinitions())
	if err == nil || !strings.Contains(err.Error(), "HAL primitive _file_stat has no architectural role") {
		t.Fatalf("expected missing HAL role failure, got %v", err)
	}
}

func TestPrimitiveRoleInvariantRejectsWrongHALRole(t *testing.T) {
	defs := primitives.Definitions()
	defs["_file_stat"] = primitives.RoleNative
	err := primitives.ValidateDefinitions(defs, hal.OperationDefinitions())
	if err == nil || !strings.Contains(err.Error(), "HAL primitive _file_stat has role native, want host") {
		t.Fatalf("expected wrong HAL role failure, got %v", err)
	}
}
