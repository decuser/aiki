package substrate

import (
	"os"
	"strings"
)

// snapshotHostEnvironment captures ambient host environment once, when a
// GoRuntime is constructed. Aiki-facing system operations use the runtime-owned
// copy thereafter and never consult or mutate process-global environment state.
func snapshotHostEnvironment() map[string]string {
	out := make(map[string]string)
	for _, entry := range os.Environ() {
		name, val, ok := strings.Cut(entry, "=")
		if ok {
			out[name] = val
		}
	}
	return out
}
