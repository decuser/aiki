package hal

// Capability names a host-dependent Aiki facility as a set of HAL operation
// identities. Capabilities are metadata; they never participate in dispatch.
type Capability struct {
	Name       string
	Operations []string
}

var capabilityRegistry = map[string]Capability{
	"basic-io": {
		Name:       "basic-io",
		Operations: []string{"HAL.io.print", "HAL.io.read", "HAL.io.read_bytes", "HAL.io.read_line", "HAL.io.write"},
	},
	"filesystem": {
		Name: "filesystem",
		Operations: []string{
			"HAL.file.open", "HAL.file.read_text", "HAL.file.read_bytes", "HAL.file.read_line",
			"HAL.file.write_text", "HAL.file.write_bytes", "HAL.file.close", "HAL.file.exists",
			"HAL.file.delete", "HAL.file.list", "HAL.file.read_at", "HAL.file.write_at",
			"HAL.file.stat", "HAL.file.rename", "HAL.file.mkdir", "HAL.file.mkdir_all",
			"HAL.file.remove_all", "HAL.file.temp", "HAL.file.temp_dir", "HAL.file.copy", "HAL.file.size", "HAL.file.walk",
		},
	},

	"symlink": {
		Name:       "symlink",
		Operations: []string{"HAL.file.symlink", "HAL.file.read_link"},
	},
	"permissions": {
		Name:       "permissions",
		Operations: []string{"HAL.file.permissions", "HAL.file.chmod"},
	},
	"process": {
		Name: "process",
		Operations: []string{
			"HAL.system.args", "HAL.system.env", "HAL.system.cwd", "HAL.system.chdir", "HAL.process.exec",
		},
	},
	"time": {
		Name:       "time",
		Operations: []string{"HAL.time.sleep", "HAL.time.after", "HAL.time.now"},
	},
	"canvas": {
		Name: "canvas",
		Operations: []string{
			"HAL.canvas.open", "HAL.canvas.command", "HAL.canvas.close",
			"HAL.canvas.width", "HAL.canvas.height", "HAL.canvas.alive",
		},
	},
}

// CapabilityDefinition returns an independent copy of one capability.
func CapabilityDefinition(name string) (Capability, bool) {
	capability, ok := capabilityRegistry[name]
	if !ok {
		return Capability{}, false
	}
	capability.Operations = append([]string(nil), capability.Operations...)
	return capability, true
}

// Capabilities returns an independent copy of the capability registry.
func Capabilities() map[string]Capability {
	out := make(map[string]Capability, len(capabilityRegistry))
	for name := range capabilityRegistry {
		capability, _ := CapabilityDefinition(name)
		out[name] = capability
	}
	return out
}

// CapabilityAvailable reports whether every HAL operation required by a
// capability is present in the supplied bound-identity set.
func CapabilityAvailable(name string, boundIdentities map[string]bool) bool {
	capability, ok := CapabilityDefinition(name)
	if !ok {
		return false
	}
	for _, identity := range capability.Operations {
		if !boundIdentities[identity] {
			return false
		}
	}
	return true
}
