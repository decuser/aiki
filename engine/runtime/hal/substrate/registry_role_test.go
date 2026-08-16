package substrate

import "testing"

func TestCompatibilityRegistrySeparatedByRole(t *testing.T) {
	rt := NewGoRuntime()

	want := map[registryRole]int{
		roleIntrinsic: 9,
		roleNative:    40,
		roleProvider:  14,
		roleHost:      40,
		roleService:   17,
	}

	seen := make(map[string]registryRole)
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

	if got := len(seen); got != 120 {
		t.Fatalf("classified %d compatibility primitives, want 120", got)
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

	if got := len(rt.hostBindings); got != 38 {
		t.Errorf("canonical host bindings changed during M6: got %d, want 38", got)
	}
}
