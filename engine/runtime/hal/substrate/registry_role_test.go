package substrate

import (
	"testing"

	"aiki/engine/runtime/primitives"
)

func TestCompatibilityRegistrySeparatedByRole(t *testing.T) {
	rt := NewGoRuntime()

	want := map[primitives.Role]int{
		primitives.RoleIntrinsic: 9,
		primitives.RoleNative:    40,
		primitives.RoleProvider:  14,
		primitives.RoleHost:      48,
		primitives.RoleService:   19,
	}

	seen := make(map[string]primitives.Role)
	for role, count := range want {
		registry := rt.registryFor(role)
		if got := len(registry); got != count {
			t.Errorf("%s registry has %d entries, want %d", role, got, count)
		}
		for name := range registry {
			if prior, exists := seen[name]; exists {
				t.Errorf("primitive %s appears in both %s and %s registries", name, prior, role)
			}
			seen[name] = role
		}
	}

	definitions := primitives.Definitions()
	if got := len(definitions); got != 130 {
		t.Fatalf("architectural primitive definitions = %d, want 130", got)
	}
	if got := len(seen); got != len(definitions) {
		t.Fatalf("registered %d compatibility primitives, architecture defines %d", got, len(definitions))
	}
	for name, role := range definitions {
		got, ok := seen[name]
		if !ok {
			t.Errorf("architectural primitive %s (%s) has no substrate binding", name, role)
			continue
		}
		if got != role {
			t.Errorf("primitive %s registered as %s, architecture defines %s", name, got, role)
		}
	}
	for name, role := range seen {
		if _, ok := definitions[name]; !ok {
			t.Errorf("substrate primitive %s (%s) has no architectural definition", name, role)
		}
	}

	for _, name := range []string{"_apply", "_first", "_regex_match", "_file_open", "_canvas", "_profile_counts"} {
		if _, ok := rt.lookupBuiltin(name); !ok {
			t.Errorf("compatibility lookup lost %s", name)
		}
	}

	for _, name := range []string{"_dot", "_line", "_rect", "_fill_rect", "_circle", "_fill_circle", "_arc", "_clear", "_set_bg", "_set_fg", "_pen_size", "_set_turtle"} {
		if _, ok := rt.lookupBuiltin(name); ok {
			t.Errorf("obsolete Canvas compatibility primitive still registered: %s", name)
		}
	}

	if got := len(rt.hostBindings); got != 46 {
		t.Errorf("canonical host bindings = %d, want 46", got)
	}
}
