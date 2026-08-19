package primitives

import (
	"fmt"
	"sort"
	"strings"

	"aiki/engine/runtime/hal"
)

// ValidateDefinitions verifies the architectural primitive-role catalog against
// the canonical HAL operation set. Inputs are explicit so invariant assurance
// can mutate a copy and prove the same validator rejects drift.
func ValidateDefinitions(definitions map[string]Role, operations map[string]hal.HostOperation) error {
	validRoles := map[Role]bool{
		RoleIntrinsic: true,
		RoleRuntime:   true,
		RoleProvider:  true,
		RoleHost:      true,
		RoleService:   true,
	}
	var problems []string
	for name, role := range definitions {
		if name == "" {
			problems = append(problems, "empty primitive identity")
		}
		if !validRoles[role] {
			problems = append(problems, fmt.Sprintf("primitive %s has unknown role %s", name, role))
		}
	}
	for name, role := range definitions {
		if role == RoleHost {
			if _, ok := operations[name]; !ok {
				problems = append(problems, fmt.Sprintf("host primitive %s has no canonical HAL operation", name))
			}
		}
	}
	for primitive := range operations {
		role, ok := definitions[primitive]
		if !ok {
			problems = append(problems, fmt.Sprintf("HAL primitive %s has no architectural role", primitive))
			continue
		}
		if role != RoleHost {
			problems = append(problems, fmt.Sprintf("HAL primitive %s has role %s, want host", primitive, role))
		}
	}
	if len(problems) == 0 {
		return nil
	}
	sort.Strings(problems)
	return fmt.Errorf("primitive-role invariant failure:\n  %s", strings.Join(problems, "\n  "))
}

func ValidateCanonicalDefinitions() error {
	return ValidateDefinitions(Definitions(), hal.OperationDefinitions())
}
