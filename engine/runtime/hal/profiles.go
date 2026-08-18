package hal

import "fmt"

// RuntimeProfile declares the capabilities required before a runtime may begin
// execution. Any capability not required by a profile is optional.
type RuntimeProfile struct {
	Name         string
	Capabilities []string
}

const DefaultRuntimeProfile = "desktop"

var runtimeProfiles = map[string]RuntimeProfile{
	"desktop": {
		Name:         "desktop",
		Capabilities: []string{"basic-io", "filesystem", "file_lock", "process", "signals", "network", "terminal", "time", "random"},
	},
	// minimal is intentionally empty. It exists to exercise the profile gate and
	// to provide a foundation for restricted/embedded runtimes without defining
	// their policy prematurely.
	"minimal": {
		Name:         "minimal",
		Capabilities: []string{},
	},
}

// RuntimeProfileDefinition returns an independent copy of one profile.
func RuntimeProfileDefinition(name string) (RuntimeProfile, bool) {
	profile, ok := runtimeProfiles[name]
	if !ok {
		return RuntimeProfile{}, false
	}
	profile.Capabilities = append([]string(nil), profile.Capabilities...)
	return profile, true
}

// RuntimeProfiles returns an independent copy of all known profiles.
func RuntimeProfiles() map[string]RuntimeProfile {
	out := make(map[string]RuntimeProfile, len(runtimeProfiles))
	for name := range runtimeProfiles {
		profile, _ := RuntimeProfileDefinition(name)
		out[name] = profile
	}
	return out
}

// ValidateProfileAvailability checks a profile against the supplied capability
// availability query. It is the profile gate's architectural rule independent
// of any concrete substrate.
func ValidateProfileAvailability(name string, hasCapability func(string) bool) error {
	profile, ok := RuntimeProfileDefinition(name)
	if !ok {
		return fmt.Errorf("unknown runtime profile: %s", name)
	}
	for _, capability := range profile.Capabilities {
		if !hasCapability(capability) {
			return fmt.Errorf("runtime profile %s requires unsupported capability: :%s", name, capability)
		}
	}
	return nil
}
