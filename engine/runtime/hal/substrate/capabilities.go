package substrate

import "aiki/engine/runtime/hal"

// HasCapability reports whether every HAL operation required by the named
// capability is bound by this runtime substrate.
func (g *GoRuntime) HasCapability(name string) bool {
	g.mu.RLock()
	bound := make(map[string]bool, len(g.hostBindings))
	for _, op := range g.hostBindings {
		bound[op.Identity] = true
	}
	g.mu.RUnlock()
	return hal.CapabilityAvailable(name, bound)
}

// ValidateProfile checks all capabilities required by a runtime profile before
// execution begins. Capabilities not named by the profile are optional.
func (g *GoRuntime) ValidateProfile(name string) error {
	return hal.ValidateProfileAvailability(name, g.HasCapability)
}
