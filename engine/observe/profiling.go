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

// SemanticSiteCount records the number of observations at one semantic site.
type SemanticSiteCount struct {
	Kind  SemanticKind
	Site  SemanticSite
	Count int64
}

// SemanticMeasurement is the result of one measured Aiki computation.
type SemanticMeasurement struct {
	Counts SemanticCounts
	Sites  []SemanticSiteCount
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
