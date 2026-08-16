package value

// Authority is the immutable set of executable grants a lexical definition
// may resolve. Canonical host grants use HAL identities; non-HAL primitives use
// their implementation names. Authority is deliberately distinct from Scope.
type Authority struct {
	grants map[string]struct{}
}

// NoAuthority returns an authority set with no executable grants.
func NoAuthority() Authority { return Authority{} }

// NewAuthority creates an authority set containing exactly the supplied grant
// identities. The set is copied and must be treated as immutable.
func NewAuthority(names ...string) Authority {
	if len(names) == 0 {
		return Authority{}
	}
	grants := make(map[string]struct{}, len(names))
	for _, name := range names {
		grants[name] = struct{}{}
	}
	return Authority{grants: grants}
}

// Allows reports whether this authority grants the named identity.
func (a Authority) Allows(name string) bool {
	_, ok := a.grants[name]
	return ok
}

// Names returns a copy of the granted identities.
func (a Authority) Names() []string {
	out := make([]string, 0, len(a.grants))
	for name := range a.grants {
		out = append(out, name)
	}
	return out
}
