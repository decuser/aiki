package evaluator

import (
	"sort"
	"sync"
	"sync/atomic"

	"aiki/engine"
	"aiki/engine/semantics/value"
)

// Counters accumulates semantic operation counts at evaluator choke points.
// A nil *Counters means probes are disabled — the evaluator checks this with
// a single pointer-nil test per node, paying no other cost when inactive.
//
// When active, each choke point increments exactly one integer. No events,
// no allocation, no string formatting. The design follows DTrace: probes
// exist at fixed points, they are free when disabled, and when enabled they
// record into a preallocated structure.
type Counters struct {
	Arithmetic int64
	Comparison int64
	Call       int64
	Iteration  int64
	Index      int64
	Send       int64
	Recv       int64
	StoreRead  int64
	StoreWrite int64

	numberResultSmallInteger    int64
	numberResultCompactRational int64
	numberResultBinaryCarrier   int64
	numberResultBigRational     int64
	numberBinaryCertified       int64
	numberBinaryFallback        int64
	numberPromotedBigRational   int64

	numberCallSmallInteger    int64
	numberCallCompactRational int64
	numberCallBinaryCarrier   int64
	numberCallBigRational     int64

	callUserEntry    int64
	callSubstrate    int64
	callTailReuse    int64
	callTailEnvReuse int64

	listFrontierPromoted      int64
	listFrontierExtended      int64
	listFrontierGrown         int64
	listFrontierForked        int64
	listElementsCopied        int64
	listBackingSlotsAllocated int64

	envPhysicalRoot     int64
	envPhysicalEnclosed int64
	envPhysicalCall     int64
	envPhysicalIsolated int64
	envLogicalCall      int64

	envCallReached1     int64
	envCallReached2     int64
	envCallReached3     int64
	envCallReached5     int64
	envEnclosedReached1 int64
	envEnclosedReached2 int64
	envEnclosedReached3 int64
	envEnclosedReached5 int64

	envCallCompactAllocations     int64
	envEnclosedCompactAllocations int64
	envCallMapPromotions          int64
	envEnclosedMapPromotions      int64
	envCallLocalNew               int64
	envCallLocalUpdate            int64
	envEnclosedLocalNew           int64
	envEnclosedLocalUpdate        int64

	// Coverage maps source positions to hit counts.
	// Key is "file:line". Only populated when coverage is enabled.
	mu       sync.Mutex
	coverage map[string]int64
	sites    map[siteKey]*siteCounter
}

type siteKey struct {
	kind     engine.SemanticKind
	file     string
	line     int
	col      int
	function string
	detail   string
}

type siteCounter struct {
	site  engine.SemanticSite
	count int64
}

func (c *Counters) WantsSites() bool {
	return c != nil && c.sites != nil
}

// Hit records one semantic operation. Scalar fields remain plain int64 so
// existing consumers can inspect them directly after execution; increments
// are atomic so spawned computations can contribute safely.
func (c *Counters) Hit(kind engine.SemanticKind, site engine.SemanticSite) {
	switch kind {
	case engine.SemanticArithmetic:
		atomic.AddInt64(&c.Arithmetic, 1)
	case engine.SemanticComparison:
		atomic.AddInt64(&c.Comparison, 1)
	case engine.SemanticCall:
		atomic.AddInt64(&c.Call, 1)
	case engine.SemanticIteration:
		atomic.AddInt64(&c.Iteration, 1)
	case engine.SemanticIndex:
		atomic.AddInt64(&c.Index, 1)
	case engine.SemanticSend:
		atomic.AddInt64(&c.Send, 1)
	case engine.SemanticReceive:
		atomic.AddInt64(&c.Recv, 1)
	case engine.SemanticStoreRead:
		atomic.AddInt64(&c.StoreRead, 1)
	case engine.SemanticStoreWrite:
		atomic.AddInt64(&c.StoreWrite, 1)
	}
	if c.sites != nil && site.Line > 0 {
		key := siteKey{kind: kind, file: site.File, line: site.Line, col: site.Col, function: site.Function, detail: site.Detail}
		c.mu.Lock()
		entry := c.sites[key]
		if entry == nil {
			entry = &siteCounter{site: site}
			c.sites[key] = entry
		}
		entry.count++
		c.mu.Unlock()
	}
}

// Snapshot returns a race-safe scalar snapshot.
// NumberArithmeticResult records hidden Number realization only when profiling
// is already active. It is deliberately called from evaluator numeric choke
// points rather than from Number itself, so unprofiled arithmetic pays no probe
// branch or atomic-counter cost.
func (c *Counters) NumberArithmeticResult(left, right, result *value.Number) {
	if c == nil || result == nil {
		return
	}
	resultRep := result.ProfileRepresentation()
	switch resultRep {
	case value.NumberSmallInteger:
		atomic.AddInt64(&c.numberResultSmallInteger, 1)
	case value.NumberCompactRational:
		atomic.AddInt64(&c.numberResultCompactRational, 1)
	case value.NumberBinaryCarrier:
		atomic.AddInt64(&c.numberResultBinaryCarrier, 1)
	case value.NumberBigRational:
		atomic.AddInt64(&c.numberResultBigRational, 1)
	}
	if left == nil || right == nil {
		return
	}
	lrep := left.ProfileRepresentation()
	rrep := right.ProfileRepresentation()
	if lrep == value.NumberBinaryCarrier && rrep == value.NumberBinaryCarrier {
		if resultRep == value.NumberBinaryCarrier || resultRep == value.NumberSmallInteger {
			atomic.AddInt64(&c.numberBinaryCertified, 1)
		} else {
			atomic.AddInt64(&c.numberBinaryFallback, 1)
		}
	}
	if resultRep == value.NumberBigRational && lrep != value.NumberBigRational && rrep != value.NumberBigRational {
		atomic.AddInt64(&c.numberPromotedBigRational, 1)
	}
}

// NumberCallResult records the hidden representation of a Number returned
// across an Aiki call boundary. Call-return realization is kept separate from
// arithmetic-result realization so profiling does not conflate two populations.
func (c *Counters) NumberCallResult(result value.Value) {
	if c == nil {
		return
	}
	n, ok := result.(*value.Number)
	if !ok || n == nil {
		return
	}
	switch n.ProfileRepresentation() {
	case value.NumberSmallInteger:
		atomic.AddInt64(&c.numberCallSmallInteger, 1)
	case value.NumberCompactRational:
		atomic.AddInt64(&c.numberCallCompactRational, 1)
	case value.NumberBinaryCarrier:
		atomic.AddInt64(&c.numberCallBinaryCarrier, 1)
	case value.NumberBigRational:
		atomic.AddInt64(&c.numberCallBigRational, 1)
	}
}

func (c *Counters) UserCallEntry() {
	if c != nil {
		atomic.AddInt64(&c.callUserEntry, 1)
	}
}

func (c *Counters) SubstrateCall() {
	if c != nil {
		atomic.AddInt64(&c.callSubstrate, 1)
	}
}

func (c *Counters) TailCallReuse() {
	if c != nil {
		atomic.AddInt64(&c.callTailReuse, 1)
	}
}

func (c *Counters) TailEnvReuse() {
	if c != nil {
		atomic.AddInt64(&c.callTailEnvReuse, 1)
	}
}

func (c *Counters) RecordListAppend(promoted, extended, grown, forked bool, copied, allocated int) {
	if c == nil {
		return
	}
	if promoted {
		atomic.AddInt64(&c.listFrontierPromoted, 1)
	}
	if extended {
		atomic.AddInt64(&c.listFrontierExtended, 1)
	}
	if grown {
		atomic.AddInt64(&c.listFrontierGrown, 1)
	}
	if forked {
		atomic.AddInt64(&c.listFrontierForked, 1)
	}
	if copied > 0 {
		atomic.AddInt64(&c.listElementsCopied, int64(copied))
	}
	if allocated > 0 {
		atomic.AddInt64(&c.listBackingSlotsAllocated, int64(allocated))
	}
}

func (c *Counters) RecordEnvPhysical(kind engine.EnvKind) {
	if c == nil {
		return
	}
	switch kind {
	case engine.EnvKindRoot:
		atomic.AddInt64(&c.envPhysicalRoot, 1)
	case engine.EnvKindEnclosed:
		atomic.AddInt64(&c.envPhysicalEnclosed, 1)
	case engine.EnvKindCall:
		atomic.AddInt64(&c.envPhysicalCall, 1)
	case engine.EnvKindIsolated:
		atomic.AddInt64(&c.envPhysicalIsolated, 1)
	}
}

func (c *Counters) RecordEnvLogicalCall() {
	if c != nil {
		atomic.AddInt64(&c.envLogicalCall, 1)
	}
}

func (c *Counters) RecordEnvBindingThreshold(kind engine.EnvKind, threshold int) {
	if c == nil {
		return
	}
	var target *int64
	switch kind {
	case engine.EnvKindCall:
		switch threshold {
		case 1:
			target = &c.envCallReached1
		case 2:
			target = &c.envCallReached2
		case 3:
			target = &c.envCallReached3
		case 5:
			target = &c.envCallReached5
		}
	case engine.EnvKindEnclosed:
		switch threshold {
		case 1:
			target = &c.envEnclosedReached1
		case 2:
			target = &c.envEnclosedReached2
		case 3:
			target = &c.envEnclosedReached3
		case 5:
			target = &c.envEnclosedReached5
		}
	}
	if target != nil {
		atomic.AddInt64(target, 1)
	}
}

func (c *Counters) RecordEnvCompactAllocation(kind engine.EnvKind) {
	if c == nil {
		return
	}
	switch kind {
	case engine.EnvKindCall:
		atomic.AddInt64(&c.envCallCompactAllocations, 1)
	case engine.EnvKindEnclosed:
		atomic.AddInt64(&c.envEnclosedCompactAllocations, 1)
	}
}

func (c *Counters) RecordEnvMapPromotion(kind engine.EnvKind) {
	if c == nil {
		return
	}
	switch kind {
	case engine.EnvKindCall:
		atomic.AddInt64(&c.envCallMapPromotions, 1)
	case engine.EnvKindEnclosed:
		atomic.AddInt64(&c.envEnclosedMapPromotions, 1)
	}
}

func (c *Counters) RecordEnvLocalSet(kind engine.EnvKind, update bool) {
	if c == nil {
		return
	}
	switch kind {
	case engine.EnvKindCall:
		if update {
			atomic.AddInt64(&c.envCallLocalUpdate, 1)
		} else {
			atomic.AddInt64(&c.envCallLocalNew, 1)
		}
	case engine.EnvKindEnclosed:
		if update {
			atomic.AddInt64(&c.envEnclosedLocalUpdate, 1)
		} else {
			atomic.AddInt64(&c.envEnclosedLocalNew, 1)
		}
	}
}

func (c *Counters) EnvSnapshot() engine.EnvRealizationCounts {
	if c == nil {
		return engine.EnvRealizationCounts{}
	}
	return engine.EnvRealizationCounts{
		PhysicalRoot:          atomic.LoadInt64(&c.envPhysicalRoot),
		PhysicalEnclosed:      atomic.LoadInt64(&c.envPhysicalEnclosed),
		PhysicalCall:          atomic.LoadInt64(&c.envPhysicalCall),
		PhysicalIsolated:      atomic.LoadInt64(&c.envPhysicalIsolated),
		LogicalCall:           atomic.LoadInt64(&c.envLogicalCall),
		CallReached1:          atomic.LoadInt64(&c.envCallReached1),
		CallReached2:          atomic.LoadInt64(&c.envCallReached2),
		CallReached3:          atomic.LoadInt64(&c.envCallReached3),
		CallReached5:          atomic.LoadInt64(&c.envCallReached5),
		EnclosedReached1:      atomic.LoadInt64(&c.envEnclosedReached1),
		EnclosedReached2:      atomic.LoadInt64(&c.envEnclosedReached2),
		EnclosedReached3:      atomic.LoadInt64(&c.envEnclosedReached3),
		EnclosedReached5:            atomic.LoadInt64(&c.envEnclosedReached5),
		CallCompactAllocations:       atomic.LoadInt64(&c.envCallCompactAllocations),
		EnclosedCompactAllocations:   atomic.LoadInt64(&c.envEnclosedCompactAllocations),
		CallMapPromotions:            atomic.LoadInt64(&c.envCallMapPromotions),
		EnclosedMapPromotions:        atomic.LoadInt64(&c.envEnclosedMapPromotions),
		CallLocalNew:          atomic.LoadInt64(&c.envCallLocalNew),
		CallLocalUpdate:       atomic.LoadInt64(&c.envCallLocalUpdate),
		EnclosedLocalNew:      atomic.LoadInt64(&c.envEnclosedLocalNew),
		EnclosedLocalUpdate:   atomic.LoadInt64(&c.envEnclosedLocalUpdate),
	}
}

func (c *Counters) ListSnapshot() engine.ListRealizationCounts {
	if c == nil {
		return engine.ListRealizationCounts{}
	}
	return engine.ListRealizationCounts{
		FrontierPromoted:      atomic.LoadInt64(&c.listFrontierPromoted),
		FrontierExtended:      atomic.LoadInt64(&c.listFrontierExtended),
		FrontierGrown:         atomic.LoadInt64(&c.listFrontierGrown),
		FrontierForked:        atomic.LoadInt64(&c.listFrontierForked),
		ElementsCopied:        atomic.LoadInt64(&c.listElementsCopied),
		BackingSlotsAllocated: atomic.LoadInt64(&c.listBackingSlotsAllocated),
	}
}

func (c *Counters) CallSnapshot() engine.CallRealizationCounts {
	if c == nil {
		return engine.CallRealizationCounts{}
	}
	return engine.CallRealizationCounts{
		UserEntry:    atomic.LoadInt64(&c.callUserEntry),
		Substrate:    atomic.LoadInt64(&c.callSubstrate),
		TailReuse:    atomic.LoadInt64(&c.callTailReuse),
		TailEnvReuse: atomic.LoadInt64(&c.callTailEnvReuse),
	}
}

func (c *Counters) NumberCallSnapshot() engine.NumberRealizationCounts {
	if c == nil {
		return engine.NumberRealizationCounts{}
	}
	return engine.NumberRealizationCounts{
		ResultSmallInteger:    atomic.LoadInt64(&c.numberCallSmallInteger),
		ResultCompactRational: atomic.LoadInt64(&c.numberCallCompactRational),
		ResultBinaryCarrier:   atomic.LoadInt64(&c.numberCallBinaryCarrier),
		ResultBigRational:     atomic.LoadInt64(&c.numberCallBigRational),
	}
}

func (c *Counters) NumberSnapshot() engine.NumberRealizationCounts {
	if c == nil {
		return engine.NumberRealizationCounts{}
	}
	return engine.NumberRealizationCounts{
		ResultSmallInteger:    atomic.LoadInt64(&c.numberResultSmallInteger),
		ResultCompactRational: atomic.LoadInt64(&c.numberResultCompactRational),
		ResultBinaryCarrier:   atomic.LoadInt64(&c.numberResultBinaryCarrier),
		ResultBigRational:     atomic.LoadInt64(&c.numberResultBigRational),
		BinaryCertified:       atomic.LoadInt64(&c.numberBinaryCertified),
		BinaryFallback:        atomic.LoadInt64(&c.numberBinaryFallback),
		PromotedBigRational:   atomic.LoadInt64(&c.numberPromotedBigRational),
	}
}

func (c *Counters) Snapshot() engine.SemanticCounts {
	if c == nil {
		return engine.SemanticCounts{}
	}
	return engine.SemanticCounts{
		Arithmetic: atomic.LoadInt64(&c.Arithmetic),
		Comparison: atomic.LoadInt64(&c.Comparison),
		Call:       atomic.LoadInt64(&c.Call),
		Iteration:  atomic.LoadInt64(&c.Iteration),
		Index:      atomic.LoadInt64(&c.Index),
		Send:       atomic.LoadInt64(&c.Send),
		Receive:    atomic.LoadInt64(&c.Recv),
		StoreRead:  atomic.LoadInt64(&c.StoreRead),
		StoreWrite: atomic.LoadInt64(&c.StoreWrite),
	}
}

// Measurement returns scalar counts and a deterministic site snapshot.
func (c *Counters) Measurement() engine.SemanticMeasurement {
	m := engine.SemanticMeasurement{
		Counts:      c.Snapshot(),
		Numbers:     c.NumberSnapshot(),
		CallNumbers: c.NumberCallSnapshot(),
		Calls:       c.CallSnapshot(),
		Lists:       c.ListSnapshot(),
		Envs:        c.EnvSnapshot(),
	}
	if c == nil || c.sites == nil {
		return m
	}
	c.mu.Lock()
	for key, entry := range c.sites {
		m.Sites = append(m.Sites, engine.SemanticSiteCount{Kind: key.kind, Site: entry.site, Count: entry.count})
	}
	c.mu.Unlock()
	sort.Slice(m.Sites, func(i, j int) bool {
		a, b := m.Sites[i], m.Sites[j]
		if a.Count != b.Count {
			return a.Count > b.Count
		}
		if a.Kind != b.Kind {
			return a.Kind < b.Kind
		}
		if a.Site.File != b.Site.File {
			return a.Site.File < b.Site.File
		}
		if a.Site.Line != b.Site.Line {
			return a.Site.Line < b.Site.Line
		}
		if a.Site.Col != b.Site.Col {
			return a.Site.Col < b.Site.Col
		}
		if a.Site.Function != b.Site.Function {
			return a.Site.Function < b.Site.Function
		}
		return a.Site.Detail < b.Site.Detail
	})
	return m
}

// NewCounters creates a counter block with all probes at zero.
func NewCounters() *Counters {
	return &Counters{}
}

// NewAttributedCounters enables source-site aggregation in addition to totals.
func NewAttributedCounters() *Counters {
	return &Counters{sites: make(map[siteKey]*siteCounter)}
}

// NewCoverageCounters creates a counter block with coverage tracking enabled.
func NewCoverageCounters() *Counters {
	return &Counters{
		coverage: make(map[string]int64),
	}
}

// CoverHit records a coverage hit at the given file and line.
func (c *Counters) CoverHit(file string, line int) {
	if c.coverage == nil {
		return
	}
	// Fast path: most execution is single-goroutine. The mutex is here
	// for spawn, which runs in a separate goroutine with its own evaluator,
	// but coverage may merge results later.
	c.mu.Lock()
	c.coverage[coverKey(file, line)]++
	c.mu.Unlock()
}

// Coverage returns a snapshot of the coverage map.
func (c *Counters) Coverage() map[string]int64 {
	if c.coverage == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make(map[string]int64, len(c.coverage))
	for k, v := range c.coverage {
		out[k] = v
	}
	return out
}

func coverKey(file string, line int) string {
	// Avoid fmt.Sprintf in the hot path.
	buf := make([]byte, 0, len(file)+12)
	buf = append(buf, file...)
	buf = append(buf, ':')
	buf = appendInt(buf, line)
	return string(buf)
}

func appendInt(buf []byte, n int) []byte {
	if n == 0 {
		return append(buf, '0')
	}
	if n < 0 {
		buf = append(buf, '-')
		n = -n
	}
	var digits [20]byte
	i := len(digits)
	for n > 0 {
		i--
		digits[i] = byte('0' + n%10)
		n /= 10
	}
	return append(buf, digits[i:]...)
}
