// Package observe defines neutral observation contracts shared by the Aiki
// evaluator, runtime substrate, and semantic values. It intentionally imports
// no higher-level engine packages so low-level semantic state can participate
// in observation without depending on the engine facade.
package observe

// SemanticKind names an Aiki-level operation. These names describe language
// work, not incidental work performed by a particular substrate.
type SemanticKind string

const (
	SemanticArithmetic SemanticKind = "arithmetic"
	SemanticComparison SemanticKind = "comparison"
	SemanticCall       SemanticKind = "call"
	SemanticIteration  SemanticKind = "iteration"
	SemanticIndex      SemanticKind = "index"
	SemanticSend       SemanticKind = "send"
	SemanticReceive    SemanticKind = "receive"
	SemanticStoreRead  SemanticKind = "store_read"
	SemanticStoreWrite SemanticKind = "store_write"
)

// SemanticSite identifies the Aiki source site responsible for semantic work.
type SemanticSite struct {
	File     string
	Line     int
	Col      int
	Function string
	Detail   string
	Source   string
}

// SemanticProbe receives Aiki-level work events at semantic choke points.
// Implementations must be safe for concurrent use because spawn can execute
// Aiki code concurrently.
type SemanticProbe interface {
	Hit(kind SemanticKind, site SemanticSite)
}

// AttributionProbe is implemented by probes that want source-level identity.
// Summary-only counters do not request sites, keeping their hot path small.
type AttributionProbe interface {
	SemanticProbe
	WantsSites() bool
}

// SemanticCounts is a stable snapshot of Aiki-level work.
type SemanticCounts struct {
	Arithmetic int64
	Comparison int64
	Call       int64
	Iteration  int64
	Index      int64
	Send       int64
	Receive    int64
	StoreRead  int64
	StoreWrite int64
}

// NumberRealizationCounts describes hidden runtime representations observed
// while evaluating numeric operations. These are realization facts, not Aiki
// semantic units; representation remains unobservable to Aiki programs.
type NumberRealizationCounts struct {
	ResultSmallInteger    int64
	ResultCompactRational int64
	ResultBinaryCarrier   int64
	ResultBigRational     int64
	BinaryCertified       int64
	BinaryFallback        int64
	PromotedBigRational   int64
}

// CallRealizationCounts describes how semantic call events are realized by the
// evaluator. These counters are profiling facts, not language semantics.
type CallRealizationCounts struct {
	UserEntry    int64
	Substrate    int64
	TailReuse    int64
	TailEnvReuse int64

	ArgArity0     int64
	ArgArity1     int64
	ArgArity2     int64
	ArgArity3     int64
	ArgArity4     int64
	ArgArity5Plus int64
	ArgsEvaluated int64

	ArgFrameNew      int64
	ArgFrameReused   int64
	ArgFramePromoted int64
	ArgDurable       int64
	ArgTailTransfer  int64
}

// ListRealizationCounts describes hidden persistent-list append realization.
// These are profiling facts, not Aiki semantic units or user-visible kinds.
type ListRealizationCounts struct {
	FrontierPromoted      int64
	FrontierExtended      int64
	FrontierGrown         int64
	FrontierForked        int64
	ElementsCopied        int64
	BackingSlotsAllocated int64
}

// ListRealizationProbe is an optional extension implemented by profilers that
// want hidden list-representation facts.
type ListRealizationProbe interface {
	SemanticProbe
	RecordListAppend(promoted, extended, grown, forked bool, copied, allocated int)
}

// EnvKind classifies hidden environment realizations for profiling.
// These are runtime/storage categories, not Aiki scope kinds.
type EnvKind uint8

const (
	EnvKindRoot EnvKind = iota
	EnvKindEnclosed
	EnvKindCall
	EnvKindIsolated
)

// EnvRealizationCounts describes physical environment allocation and logical
// local-binding pressure. Binding thresholds record the maximum number of
// ordinary local store bindings reached during one logical environment
// lifetime; borrowed call parameters are not included.
type EnvRealizationCounts struct {
	PhysicalRoot     int64
	PhysicalEnclosed int64
	PhysicalCall     int64
	PhysicalIsolated int64
	LogicalCall      int64

	CallReached1 int64
	CallReached2 int64
	CallReached3 int64
	CallReached5 int64

	EnclosedReached1 int64
	EnclosedReached2 int64
	EnclosedReached3 int64
	EnclosedReached5 int64

	CallCompactAllocations     int64
	EnclosedCompactAllocations int64
	CallMapPromotions          int64
	EnclosedMapPromotions      int64
	CallLocalNew               int64
	CallLocalUpdate            int64
	EnclosedLocalNew           int64
	EnclosedLocalUpdate        int64
}

// EnvRealizationProbe is an optional profiling extension used by the semantic
// value layer. Implementations must be safe for concurrent use.
type EnvRealizationProbe interface {
	SemanticProbe
	RecordEnvPhysical(kind EnvKind)
	RecordEnvLogicalCall()
	RecordEnvBindingThreshold(kind EnvKind, threshold int)
	RecordEnvCompactAllocation(kind EnvKind)
	RecordEnvMapPromotion(kind EnvKind)
	RecordEnvLocalSet(kind EnvKind, update bool)
}

// SemanticSiteCount records the number of observations at one semantic site.
type SemanticSiteCount struct {
	Kind  SemanticKind
	Site  SemanticSite
	Count int64
}

// SemanticMeasurement is the result of one measured Aiki computation.
type SemanticMeasurement struct {
	Counts      SemanticCounts
	Numbers     NumberRealizationCounts
	CallNumbers NumberRealizationCounts
	Calls       CallRealizationCounts
	Lists       ListRealizationCounts
	Envs        EnvRealizationCounts
	Sites       []SemanticSiteCount
}

// ProfileLabels are the stable correlation dimensions written into Go CPU
// profiles. The fixed struct is intentionally comparable so the Go substrate
// can cache pprof label contexts without allocating on every Aiki call.
type ProfileLabels struct {
	Layer     string
	Function  string
	File      string
	Line      string
	Primitive string
}
