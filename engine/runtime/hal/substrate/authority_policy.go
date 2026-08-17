package substrate

import (
	"aiki/engine/runtime/hal"
	"aiki/engine/semantics/value"
)

// AuthorityForSource returns only the grants explicitly declared for a
// canonical trusted source. Host bindings are translated to canonical HAL
// identities; non-host primitives retain their implementation names. Merely
// residing beneath lib/ confers nothing.
func (g *GoRuntime) AuthorityForSource(path string) value.Authority {
	bindings, ok := hal.AuthorityPrimitivesForSource(path)
	if !ok {
		return value.NoAuthority()
	}
	grants := make([]string, 0, len(bindings))
	for _, name := range bindings {
		grants = append(grants, g.authorityKey(name))
	}
	return value.NewAuthority(grants...)
}
