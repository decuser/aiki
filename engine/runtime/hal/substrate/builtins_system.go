package substrate

import (
	"aiki/engine/runtime/hal"
	"aiki/engine/semantics/value"
	"sort"
	"strings"
)

func (g *GoRuntime) halSystemExit(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("system.exit: want 1 argument, got %d", len(args))
	}
	n, ok := args[0].(*value.Number)
	if !ok {
		return value.NewFault("system.exit: expected number, got %s", args[0].Type())
	}
	if !n.IsInt() {
		return value.NewFault("system.exit: exit code must be an integer")
	}
	if !n.IsInt64() {
		return value.NewFault("system.exit: exit code must be from 0 through 255")
	}
	code := n.Int64Value()
	if code < 0 || code > 255 {
		return value.NewFault("system.exit: exit code must be from 0 through 255")
	}
	return &value.ProgramExitSignal{Code: int(code)}
}

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

func validEnvironmentName(name string) bool {
	return name != "" && !strings.ContainsRune(name, '=') && !strings.ContainsRune(name, '\x00')
}

func (g *GoRuntime) environmentSnapshot() map[string]string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	out := make(map[string]string, len(g.environment))
	for name, val := range g.environment {
		out[name] = val
	}
	return out
}

func environmentList(env map[string]string) []string {
	names := make([]string, 0, len(env))
	for name := range env {
		names = append(names, name)
	}
	sort.Strings(names)
	out := make([]string, 0, len(names))
	for _, name := range names {
		out = append(out, name+"="+env[name])
	}
	return out
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
	val, found := g.environment[name.Val]
	g.mu.RUnlock()
	if lookup != nil {
		val, found = lookup(name.Val)
	}
	if !found {
		return value.NewShapedError("environment", "environment variable not set: %s", name.Val)
	}
	return &value.String{Val: val}
}

func (g *GoRuntime) halSystemEnviron(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 0 {
		return value.NewFault("system.environ: want 0 arguments, got %d", len(args))
	}
	env := g.environmentSnapshot()
	names := make([]string, 0, len(env))
	for name := range env {
		names = append(names, name)
	}
	sort.Strings(names)
	entries := make([]value.Value, 0, len(names))
	for _, name := range names {
		entries = append(entries, &value.List{Shape: "env", Elements: []value.Value{&value.String{Val: name}, &value.String{Val: env[name]}}})
	}
	return &value.List{Elements: entries}
}

func (g *GoRuntime) halSystemSetEnv(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("system.set_env: want 2 arguments, got %d", len(args))
	}
	name, ok := args[0].(*value.String)
	if !ok {
		return value.NewFault("system.set_env: expected string name, got %s", args[0].Type())
	}
	val, ok := args[1].(*value.String)
	if !ok {
		return value.NewFault("system.set_env: expected string value, got %s", args[1].Type())
	}
	if !validEnvironmentName(name.Val) || strings.ContainsRune(val.Val, '\x00') {
		return value.NewShapedError("environment", "invalid environment variable")
	}
	g.mu.Lock()
	g.environment[name.Val] = val.Val
	g.mu.Unlock()
	return value.TRUE
}

func (g *GoRuntime) halSystemUnsetEnv(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("system.unset_env: want 1 argument, got %d", len(args))
	}
	name, ok := args[0].(*value.String)
	if !ok {
		return value.NewFault("system.unset_env: expected string, got %s", args[0].Type())
	}
	if !validEnvironmentName(name.Val) {
		return value.NewShapedError("environment", "invalid environment variable name")
	}
	g.mu.Lock()
	delete(g.environment, name.Val)
	g.mu.Unlock()
	return value.TRUE
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

func (g *GoRuntime) halSystemHas(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("system.has: want 1 argument, got %d", len(args))
	}
	name, ok := args[0].(*value.Symbol)
	if !ok {
		return value.NewFault("system.has: expected symbol, got %s", args[0].Type())
	}
	if g.HasCapability(name.Val) {
		return value.TRUE
	}
	return value.FALSE
}

func (g *GoRuntime) halSystemRequire(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("system.require: want 1 argument, got %d", len(args))
	}
	name, ok := args[0].(*value.Symbol)
	if !ok {
		return value.NewFault("system.require: expected symbol, got %s", args[0].Type())
	}
	if _, known := hal.CapabilityDefinition(name.Val); !known {
		return value.NewShapedError("unsupported", "unknown capability: :%s", name.Val)
	}
	if !g.HasCapability(name.Val) {
		return value.NewShapedError("unsupported", "unsupported capability: :%s", name.Val)
	}
	return value.EMPTY
}
