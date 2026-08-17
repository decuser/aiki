package substrate

import (
	"fmt"

	"aiki/engine/runtime/hal"
)

// HasCapability reports whether every HAL operation required by the named
// capability is bound by this runtime substrate.
func (g *GoRuntime) HasCapability(name string) bool {
	capability, ok := hal.CapabilityDefinition(name)
	if !ok {
		return false
	}

	g.mu.RLock()
	defer g.mu.RUnlock()

	bound := make(map[string]bool, len(g.hostBindings))
	for _, op := range g.hostBindings {
		bound[op.Identity] = true
	}
	for _, identity := range capability.Operations {
		if !bound[identity] {
			return false
		}
	}
	return true
}

// ValidateProfile checks all capabilities required by a runtime profile before
// execution begins. Capabilities not named by the profile are optional.
func (g *GoRuntime) ValidateProfile(name string) error {
	profile, ok := hal.RuntimeProfileDefinition(name)
	if !ok {
		return fmt.Errorf("unknown runtime profile: %s", name)
	}
	for _, capability := range profile.Capabilities {
		if !g.HasCapability(capability) {
			return fmt.Errorf("runtime profile %s requires unsupported capability: :%s", name, capability)
		}
	}
	return nil
}
