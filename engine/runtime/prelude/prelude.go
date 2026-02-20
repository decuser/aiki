// Package prelude loads the standard library into the runtime.
package prelude

import (
	"aiki/engine/runtime/hal/substrate"
	"aiki/engine/semantics/evaluator"
)

// Load registers standard library functions in the scope and runtime.
// Note: The HAL primitives are already registered in go_runtime.go registerDefaults().
// This function can be used to add scope-level bindings or Aiki-defined stdlib functions.
func Load(scope *evaluator.Scope, rt *substrate.GoRuntime) {
	// HAL primitives are registered in go_runtime.go
	// This is a placeholder for any additional prelude initialization
	_ = scope
	_ = rt
}
