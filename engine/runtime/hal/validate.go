package hal

import (
	"fmt"
	"sort"
	"strings"
)

var validContexts = map[string]bool{
	"runtime.clock":      true,
	"runtime.filesystem": true,
	"runtime.graphics":   true,
	"runtime.io":         true,
	"runtime.network":    true,
	"runtime.platform":   true,
	"runtime.process":    true,
	"runtime.random":     true,
	"runtime.resources":  true,
	"runtime.signals":    true,
	"runtime.terminal":   true,
}

var validEffects = map[string]bool{
	"async source":         true,
	"bounded call":         true,
	"resource acquisition": true,
	"state mutation":       true,
}

var validBlocking = map[string]bool{
	"may block":        true,
	"nonblocking":      true,
	"waits externally": true,
}

var validLifetimes = map[string]bool{
	"operates on resource": true,
	"returns resource":     true,
	"runtime-owned state":  true,
	"stateless":            true,
}

var validOptionality = map[string]bool{
	"constitutive": true,
	"optional":     true,
}

var validErrorContracts = map[string]bool{
	"fault":                                       true,
	"fault or shaped :canvas result":              true,
	"fault or shaped :environment result":         true,
	"fault or shaped :io result":                  true,
	"fault or shaped :io/:end result":             true,
	"fault or shaped :network result":             true,
	"fault or shaped :process result":             true,
	"fault or shaped :signal result":              true,
	"fault or shaped :signal/:unsupported result": true,
	"fault or shaped :terminal result":            true,
	"fault, false, or shaped :lock result":        true,
	"fault or shaped :lock result":                true,
}

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
		if len(op.Context) == 0 {
			problems = append(problems, fmt.Sprintf("operation %s has no context", primitive))
		}
		seenContext := map[string]bool{}
		for _, context := range op.Context {
			if !validContexts[context] {
				problems = append(problems, fmt.Sprintf("operation %s has unknown context %s", primitive, context))
			}
			if seenContext[context] {
				problems = append(problems, fmt.Sprintf("operation %s repeats context %s", primitive, context))
			}
			seenContext[context] = true
		}
		if !validEffects[op.Effect] {
			problems = append(problems, fmt.Sprintf("operation %s has unknown effect %q", primitive, op.Effect))
		}
		if !validBlocking[op.Blocking] {
			problems = append(problems, fmt.Sprintf("operation %s has unknown blocking class %q", primitive, op.Blocking))
		}
		if !validLifetimes[op.Lifetime] {
			problems = append(problems, fmt.Sprintf("operation %s has unknown lifetime %q", primitive, op.Lifetime))
		}
		if !validOptionality[op.Optionality] {
			problems = append(problems, fmt.Sprintf("operation %s has unknown optionality %q", primitive, op.Optionality))
		}
		if !validErrorContracts[op.ErrorContract] {
			problems = append(problems, fmt.Sprintf("operation %s has unknown error contract %q", primitive, op.ErrorContract))
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
