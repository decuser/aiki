package substrate

import "io"

// PageOutput, when set, may present long help/documentation text through an
// interactive pager. It returns true when it handled the text. Non-interactive
// runners leave it nil, so output falls back to Stdout.
var PageOutput func(string) bool

// SetStdout sets the writer used by IO related builtins.
func SetStdout(w io.Writer) {
	Stdout = w
}

// SetPageOutput installs the optional pageable-text presenter.
func SetPageOutput(fn func(string) bool) {
	PageOutput = fn
}

// ResetModuleRegistry clears the global module registry so it will be rebuilt on next init.
func ResetModuleRegistry() {
	GlobalRegistry = nil
}
