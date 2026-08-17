package invariant

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"aiki/engine/runtime/hal/substrate"
	"aiki/engine/runtime/primitives"
)

func TestRuntimePrimitiveRegistrationsMatchArchitecture(t *testing.T) {
	rt := substrate.NewGoRuntime()
	if err := validatePrimitiveRegistrations(primitives.Definitions(), rt.PrimitiveRegistrations()); err != nil {
		t.Fatal(err)
	}
}

func TestRuntimeRegistrationInvariantRejectsMissingPrimitive(t *testing.T) {
	rt := substrate.NewGoRuntime()
	actual := rt.PrimitiveRegistrations()
	delete(actual, "_file_stat")
	err := validatePrimitiveRegistrations(primitives.Definitions(), actual)
	if err == nil || !strings.Contains(err.Error(), "missing runtime primitive _file_stat") {
		t.Fatalf("expected missing primitive failure, got %v", err)
	}
}

func TestRuntimeRegistrationInvariantRejectsWrongRole(t *testing.T) {
	rt := substrate.NewGoRuntime()
	actual := rt.PrimitiveRegistrations()
	actual["_file_stat"] = primitives.RoleNative
	err := validatePrimitiveRegistrations(primitives.Definitions(), actual)
	if err == nil || !strings.Contains(err.Error(), "runtime primitive _file_stat registered as native, architecture requires host") {
		t.Fatalf("expected wrong-role failure, got %v", err)
	}
}

func validatePrimitiveRegistrations(want, actual map[string]primitives.Role) error {
	var problems []string
	for name, role := range want {
		got, ok := actual[name]
		if !ok {
			problems = append(problems, "missing runtime primitive "+name)
			continue
		}
		if got != role {
			problems = append(problems, fmt.Sprintf("runtime primitive %s registered as %s, architecture requires %s", name, got, role))
		}
	}
	for name := range actual {
		if _, ok := want[name]; !ok {
			problems = append(problems, "runtime registers primitive with no architectural definition "+name)
		}
	}
	if len(problems) == 0 {
		return nil
	}
	sort.Strings(problems)
	return fmt.Errorf("runtime primitive-registration invariant failure:\n  %s", strings.Join(problems, "\n  "))
}
