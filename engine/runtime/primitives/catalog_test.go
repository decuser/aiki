package primitives

import "testing"

func TestDefinitionsAreFullyClassified(t *testing.T) {
	defs := Definitions()
	want := map[Role]int{
		RoleIntrinsic: 9,
		RoleNative:    40,
		RoleProvider:  14,
		RoleHost:      48,
		RoleService:   19,
	}
	got := map[Role]int{}
	for name, role := range defs {
		if name == "" {
			t.Fatal("empty primitive identity")
		}
		got[role]++
	}
	for role, n := range want {
		if got[role] != n {
			t.Errorf("%s definitions = %d, want %d", role, got[role], n)
		}
	}
	if len(defs) != 130 {
		t.Fatalf("definitions = %d, want 130", len(defs))
	}
}
