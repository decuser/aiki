package substrate

import (
	"aiki/engine/runtime/hal"
	"aiki/engine/semantics/value"
)

func (g *GoRuntime) halSystemArgs(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 0 {
		return value.NewFault("system.args: want 0 arguments, got %d", len(args))
	}
	g.mu.RLock()
	copyArgs := append([]string(nil), g.programArgs...)
	g.mu.RUnlock()
	elements := make([]value.Value, len(copyArgs))
	for i, arg := range copyArgs {
		elements[i] = &value.String{Val: arg}
	}
	return &value.List{Elements: elements}
}

func (g *GoRuntime) halSystemEnv(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("system.env: want 1 argument, got %d", len(args))
	}
	name, ok := args[0].(*value.String)
	if !ok {
		return value.NewFault("system.env: expected string, got %s", args[0].Type())
	}
	g.mu.RLock()
	lookup := g.envLookup
	g.mu.RUnlock()
	v, ok := lookup(name.Val)
	if !ok {
		return value.NewShapedError("environment", "environment variable not set: %s", name.Val)
	}
	return &value.String{Val: v}
}

func (g *GoRuntime) halModuleRoots(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 0 {
		return value.NewFault("module_roots: want 0 arguments, got %d", len(args))
	}
	if g.moduleRegistry == nil {
		return value.NewFault("module_roots: module registry is not initialized")
	}
	roots := g.moduleRegistry.Roots()
	elements := make([]value.Value, len(roots))
	for i, root := range roots {
		elements[i] = &value.String{Val: root}
	}
	return &value.List{Elements: elements}
}
