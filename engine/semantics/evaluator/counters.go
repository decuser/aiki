package evaluator

import (
	"sort"
	"sync"
	"sync/atomic"

	"aiki/engine"
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
	m := engine.SemanticMeasurement{Counts: c.Snapshot()}
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
