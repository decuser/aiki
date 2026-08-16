package hal

// HostOperation describes one canonical irreducible crossing from Aiki into
// the host. It is architectural metadata: the Aiki-facing meaning, HAL
// contract identity, and substrate realization remain distinct names.
type HostOperation struct {
	Identity            string
	Primitive           string
	Authority           string
	Context             []string
	Effect              string
	Blocking            string
	Lifetime            string
	Optionality         string
	ErrorContract       string
	SemanticObservation string
	SubstrateProvenance string
	AikiBindings        []AikiBinding
}

// AikiBinding records one programmer-facing Aiki name implemented through a
// host operation and the source file that owns that meaning.
type AikiBinding struct {
	Name   string
	Source string
}

// HostOperationProvider is an optional runtime capability for inspecting the
// canonical host contracts currently bound by a substrate. It is intentionally
// separate from RuntimeContract during M1 so evaluator behavior does not change.
type HostOperationProvider interface {
	HostOperations() []HostOperation
}
