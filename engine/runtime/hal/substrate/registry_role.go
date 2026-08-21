package substrate

import "aiki/engine/runtime/primitives"

func (g *GoRuntime) registryFor(role primitives.Role) map[string]*Builtin {
	switch role {
	case primitives.RoleIntrinsic:
		return g.intrinsics
	case primitives.RoleRuntime:
		return g.runtimePrimitives
	case primitives.RoleProvider:
		return g.providers
	case primitives.RoleHost:
		return g.hostRegistry
	case primitives.RoleService:
		return g.services
	default:
		panic("unknown registry role: " + string(role))
	}
}

func (g *GoRuntime) registries() []map[string]*Builtin {
	return []map[string]*Builtin{
		g.intrinsics,
		g.runtimePrimitives,
		g.providers,
		g.hostRegistry,
		g.services,
	}
}

func (g *GoRuntime) lookupBuiltin(name string) (*Builtin, bool) {
	for _, registry := range g.registries() {
		if b, ok := registry[name]; ok {
			return b, true
		}
	}
	return nil, false
}

func (g *GoRuntime) registerPrimitive(name string, fn BuiltinFunc) {
	g.registerPrimitiveWithContextPolicy(name, fn, true)
}

// registerContextFreePrimitive is used only for audited builtins whose result
// cannot depend on evaluator context. Unknown/new registrations remain
// context-requiring by default, so omission costs performance rather than
// correctness.
func (g *GoRuntime) registerContextFreePrimitive(name string, fn BuiltinFunc) {
	g.registerPrimitiveWithContextPolicy(name, fn, false)
}

func (g *GoRuntime) registerPrimitiveWithContextPolicy(name string, fn BuiltinFunc, needsContext bool) {
	if fn == nil {
		panic("runtime primitive registration has nil function: " + name)
	}
	role, ok := primitives.RoleOf(name)
	if !ok {
		panic("runtime primitive has no architectural role: " + name)
	}
	if _, exists := g.lookupBuiltin(name); exists {
		panic("duplicate compatibility primitive registration: " + name)
	}
	g.registryFor(role)[name] = &Builtin{
		name:         name,
		fn:           fn,
		runtime:      g,
		needsContext: needsContext,
	}
}

// PrimitiveRegistrations returns the runtime's actual primitive registrations
// classified by architectural role. The returned map is a validation snapshot;
// callers cannot mutate runtime dispatch through it.
func (g *GoRuntime) PrimitiveRegistrations() map[string]primitives.Role {
	g.mu.RLock()
	defer g.mu.RUnlock()
	out := map[string]primitives.Role{}
	for _, role := range []primitives.Role{
		primitives.RoleIntrinsic,
		primitives.RoleRuntime,
		primitives.RoleProvider,
		primitives.RoleHost,
		primitives.RoleService,
	} {
		for name := range g.registryFor(role) {
			out[name] = role
		}
	}
	return out
}
