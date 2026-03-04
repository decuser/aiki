package substrate

import "io"

// SetStdout sets the writer used by IO related builtins.
func SetStdout(w io.Writer) {
	Stdout = w
}

// ResetModuleRegistry clears the global module registry so it will be rebuilt on next init.
func ResetModuleRegistry() {
	GlobalRegistry = nil
}
