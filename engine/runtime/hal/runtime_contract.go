// Package hal defines the runtime contract that the evaluator uses
// to interact with the host platform.
package hal

import "aiki/engine/semantics/value"

// RuntimeContract defines the behavioral requirements of the evaluator.
// The evaluator delegates all host-level effects through this interface.
type RuntimeContract interface {
	// Primitives - pure value functions
	Execute(name string, args []value.Value) (value.Value, error)

	// HasBuiltin checks if a name is a registered builtin.
	HasBuiltin(name string) bool

	// Concurrency - used by spawn intrinsic
	Spawn(task func())

	// Channels - used by channel intrinsics
	MakeChannel() value.Value
	Send(ch value.Value, val value.Value) error
	Recv(ch value.Value) (value.Value, error)

	// File system - used by load/import intrinsics
	ReadFile(path string) ([]byte, error)
	ResolvePath(name string, relativeTo string) (string, error)

	// Error logging - for spawned goroutines
	LogError(err error)
}
