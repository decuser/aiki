package substrate

// registryRole records why a runtime binding exists. The old single native
// registry is permanently separated by architectural responsibility; roleHost
// does not itself imply that a binding has a canonical HAL contract.
type registryRole string

const (
	roleIntrinsic registryRole = "intrinsic"
	roleNative    registryRole = "native"
	roleProvider  registryRole = "provider"
	roleHost      registryRole = "host"
	roleService   registryRole = "service"
)

func (g *GoRuntime) registryFor(role registryRole) map[string]*Builtin {
	switch role {
	case roleIntrinsic:
		return g.intrinsics
	case roleNative:
		return g.natives
	case roleProvider:
		return g.providers
	case roleHost:
		return g.hostRegistry
	case roleService:
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

func (g *GoRuntime) registerRole(role registryRole, name string, fn BuiltinFunc) {
	if fn == nil {
		panic("HAL registration has nil function: " + name)
	}
	if _, exists := g.lookupBuiltin(name); exists {
		panic("duplicate compatibility primitive registration: " + name)
	}
	g.registryFor(role)[name] = &Builtin{name: name, fn: fn, runtime: g}
}
