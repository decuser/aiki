// Package hal defines the runtime contract that the evaluator uses
// to interact with the host platform.
package hal

import (
	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
	"aiki/engine/syntax/grammar"
)

// EvalContext provides evaluation context to builtins that need it.
// Most builtins ignore this; intrinsics (import, export, apply, load, spawn)
// use it to access the evaluator, grammar, and environment.
type EvalContext struct {
	Env     *value.Env
	Node    *syntax.Node
	Grammar *grammar.Grammar
	// Eval is a callback to evaluate AST nodes. Used by apply, spawn.
	Eval func(*syntax.Node, *value.Env) value.Value
}

// RuntimeContract defines the behavioral requirements of the evaluator.
// The evaluator delegates all host-level effects through this interface.
type RuntimeContract interface {
	// Execute calls a registered primitive function by name.
	// Context provides evaluation state for intrinsics that need it.
	Execute(name string, args []value.Value, ctx *EvalContext) (value.Value, error)

	// HasBuiltin checks if a name is visible at the given scope.
	HasBuiltin(name string, scope value.Scope) bool

	// GetBuiltin returns a callable for the named builtin at the given scope.
	GetBuiltin(name string, scope value.Scope) (value.Callable, bool)

	// SetContext sets the evaluation context for subsequent builtin calls.
	// Called by the evaluator before invoking builtins.
	SetContext(ctx *EvalContext)
}
