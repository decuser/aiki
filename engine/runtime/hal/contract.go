// Package hal defines the runtime contract that the evaluator uses
// to interact with the host platform.
package hal

import "aiki/engine/semantics/value"

// Scope represents the visibility level for builtins.
type Scope int

const (
	ScopeUser    Scope = iota // User code - sees prelude exports only
	ScopePrelude              // Prelude - sees HAL primitives (_prefixed)
)

// RuntimeContract defines the behavioral requirements of the evaluator.
// The evaluator delegates all host-level effects through this interface.
type RuntimeContract interface {
	// Execute calls a registered primitive function by name.
	Execute(name string, args []value.Value) (value.Value, error)

	// HasBuiltin checks if a name is visible at the given scope.
	HasBuiltin(name string, scope Scope) bool

	// GetBuiltin returns a callable for the named builtin at the given scope.
	GetBuiltin(name string, scope Scope) (value.Callable, bool)
}
