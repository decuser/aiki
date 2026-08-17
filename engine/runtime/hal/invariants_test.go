package hal

import "testing"

func TestOperationDefinitionsHaveStableUniqueIdentities(t *testing.T) {
	seen := map[string]string{}
	for primitive, op := range OperationDefinitions() {
		if primitive != op.Primitive {
			t.Errorf("operation key %s disagrees with primitive %s", primitive, op.Primitive)
		}
		if op.Identity == "" {
			t.Errorf("%s has empty HAL identity", primitive)
		}
		if previous, exists := seen[op.Identity]; exists {
			t.Errorf("HAL identity %s is shared by %s and %s", op.Identity, previous, primitive)
		}
		seen[op.Identity] = primitive
		if op.Authority != op.Identity {
			t.Errorf("%s authority %s does not match canonical identity %s", primitive, op.Authority, op.Identity)
		}
	}
}

func TestCapabilitiesNameOnlyHALOperations(t *testing.T) {
	known := map[string]bool{}
	for _, op := range OperationDefinitions() {
		known[op.Identity] = true
	}
	for name, capability := range Capabilities() {
		if name != capability.Name {
			t.Errorf("capability key %s disagrees with name %s", name, capability.Name)
		}
		if len(capability.Operations) == 0 {
			t.Errorf("capability %s has no HAL operations", name)
		}
		seen := map[string]bool{}
		for _, identity := range capability.Operations {
			if !known[identity] {
				t.Errorf("capability %s names unknown HAL operation %s", name, identity)
			}
			if seen[identity] {
				t.Errorf("capability %s repeats HAL operation %s", name, identity)
			}
			seen[identity] = true
		}
	}
}

func TestProfilesNameOnlyCapabilities(t *testing.T) {
	capabilities := Capabilities()
	for name, profile := range RuntimeProfiles() {
		if name != profile.Name {
			t.Errorf("profile key %s disagrees with name %s", name, profile.Name)
		}
		seen := map[string]bool{}
		for _, capability := range profile.Capabilities {
			if _, ok := capabilities[capability]; !ok {
				t.Errorf("profile %s names unknown capability %s", name, capability)
			}
			if seen[capability] {
				t.Errorf("profile %s repeats capability %s", name, capability)
			}
			seen[capability] = true
		}
	}
}
