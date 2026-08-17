package hal

import (
	"fmt"
	"sort"
	"strings"
)

// ValidateMetadata verifies the internal HAL metadata graph. Inputs are explicit
// so invariant tests can exercise the same validator with deliberately malformed
// graphs rather than duplicating validation logic.
func ValidateMetadata(operations map[string]HostOperation, capabilities map[string]Capability, profiles map[string]RuntimeProfile) error {
	var problems []string

	identities := map[string]string{}
	for primitive, op := range operations {
		if primitive != op.Primitive {
			problems = append(problems, fmt.Sprintf("operation key %s disagrees with primitive %s", primitive, op.Primitive))
		}
		if op.Identity == "" {
			problems = append(problems, fmt.Sprintf("operation %s has empty HAL identity", primitive))
		} else if prior, ok := identities[op.Identity]; ok {
			problems = append(problems, fmt.Sprintf("HAL identity %s is shared by %s and %s", op.Identity, prior, primitive))
		} else {
			identities[op.Identity] = primitive
		}
		if op.Authority != op.Identity {
			problems = append(problems, fmt.Sprintf("operation %s authority %s does not match canonical identity %s", primitive, op.Authority, op.Identity))
		}
	}

	for name, capability := range capabilities {
		if name != capability.Name {
			problems = append(problems, fmt.Sprintf("capability key %s disagrees with name %s", name, capability.Name))
		}
		if len(capability.Operations) == 0 {
			problems = append(problems, fmt.Sprintf("capability %s has no HAL operations", name))
		}
		seen := map[string]bool{}
		for _, identity := range capability.Operations {
			if _, ok := identities[identity]; !ok {
				problems = append(problems, fmt.Sprintf("capability %s names unknown HAL operation %s", name, identity))
			}
			if seen[identity] {
				problems = append(problems, fmt.Sprintf("capability %s repeats HAL operation %s", name, identity))
			}
			seen[identity] = true
		}
	}

	for name, profile := range profiles {
		if name != profile.Name {
			problems = append(problems, fmt.Sprintf("profile key %s disagrees with name %s", name, profile.Name))
		}
		seen := map[string]bool{}
		for _, capability := range profile.Capabilities {
			if _, ok := capabilities[capability]; !ok {
				problems = append(problems, fmt.Sprintf("profile %s names unknown capability %s", name, capability))
			}
			if seen[capability] {
				problems = append(problems, fmt.Sprintf("profile %s repeats capability %s", name, capability))
			}
			seen[capability] = true
		}
	}

	if len(problems) == 0 {
		return nil
	}
	sort.Strings(problems)
	return fmt.Errorf("HAL metadata invariant failure:\n  %s", strings.Join(problems, "\n  "))
}

// ValidateCanonicalMetadata verifies the repository's authoritative HAL graph.
func ValidateCanonicalMetadata() error {
	return ValidateMetadata(OperationDefinitions(), Capabilities(), RuntimeProfiles())
}
