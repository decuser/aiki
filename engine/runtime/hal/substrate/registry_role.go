package substrate

import "aiki/engine/runtime/primitives"

func (g *GoRuntime) registryFor(role primitives.Role) map[string]*Builtin {
	switch role {
	case primitives.RoleIntrinsic:
		return g.intrinsics
	case primitives.RoleNative:
		return g.natives
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
		g.natives,
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
	g.registryFor(role)[name] = &Builtin{name: name, fn: fn, runtime: g}
}
