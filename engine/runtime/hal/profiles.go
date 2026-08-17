package hal

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
		Capabilities: []string{"basic-io", "filesystem", "process", "time"},
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
