package engine

import "aiki/engine/observe"

// Profiling/observation types live in the leaf observe package. These aliases
// retain the existing engine-facing API while callers migrate independently.
type SemanticKind = observe.SemanticKind

const (
	SemanticArithmetic = observe.SemanticArithmetic
	SemanticComparison = observe.SemanticComparison
	SemanticCall       = observe.SemanticCall
	SemanticIteration  = observe.SemanticIteration
	SemanticIndex      = observe.SemanticIndex
	SemanticSend       = observe.SemanticSend
	SemanticReceive    = observe.SemanticReceive
	SemanticStoreRead  = observe.SemanticStoreRead
	SemanticStoreWrite = observe.SemanticStoreWrite
)

type SemanticSite = observe.SemanticSite
type SemanticProbe = observe.SemanticProbe
type AttributionProbe = observe.AttributionProbe
type SemanticCounts = observe.SemanticCounts
type NumberRealizationCounts = observe.NumberRealizationCounts
type CallRealizationCounts = observe.CallRealizationCounts
type ListRealizationCounts = observe.ListRealizationCounts
type ListRealizationProbe = observe.ListRealizationProbe
type EnvKind = observe.EnvKind

const (
	EnvKindRoot     = observe.EnvKindRoot
	EnvKindEnclosed = observe.EnvKindEnclosed
	EnvKindCall     = observe.EnvKindCall
	EnvKindIsolated = observe.EnvKindIsolated
)

type EnvRealizationCounts = observe.EnvRealizationCounts
type EnvRealizationProbe = observe.EnvRealizationProbe
type SemanticSiteCount = observe.SemanticSiteCount
type SemanticMeasurement = observe.SemanticMeasurement
type ProfileLabels = observe.ProfileLabels
