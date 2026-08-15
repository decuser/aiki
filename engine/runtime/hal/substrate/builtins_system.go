package substrate

import (
	"os"
	"sync"

	"aiki/engine/runtime/hal"
	"aiki/engine/semantics/value"
)

var (
	programArgsMu sync.RWMutex
	programArgs   []string
)

// SetProgramArgs sets the arguments visible to Aiki system.args(). The slice is
// copied so host-side mutation cannot alter a running Aiki program's view.
func SetProgramArgs(args []string) {
	programArgsMu.Lock()
	programArgs = append(programArgs[:0], args...)
	programArgsMu.Unlock()
}

func halSystemArgs(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 0 {
		return value.NewFault("system.args: want 0 arguments, got %d", len(args))
	}
	programArgsMu.RLock()
	copyArgs := append([]string(nil), programArgs...)
	programArgsMu.RUnlock()

	elements := make([]value.Value, len(copyArgs))
	for i, arg := range copyArgs {
		elements[i] = &value.String{Val: arg}
	}
	return &value.List{Elements: elements}
}

func halSystemEnv(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("system.env: want 1 argument, got %d", len(args))
	}
	name, ok := args[0].(*value.String)
	if !ok {
		return value.NewFault("system.env: expected string, got %s", args[0].Type())
	}
	v, ok := os.LookupEnv(name.Val)
	if !ok {
		return value.NewShapedError("environment", "environment variable not set: %s", name.Val)
	}
	return &value.String{Val: v}
}

func halModuleRoots(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 0 {
		return value.NewFault("module_roots: want 0 arguments, got %d", len(args))
	}
	if GlobalRegistry == nil {
		return value.NewFault("module_roots: module registry is not initialized")
	}
	roots := GlobalRegistry.Roots()
	elements := make([]value.Value, len(roots))
	for i, root := range roots {
		elements[i] = &value.String{Val: root}
	}
	return &value.List{Elements: elements}
}
