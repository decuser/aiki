package evaluator

import "sync"

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

	// Coverage maps source positions to hit counts.
	// Key is "file:line". Only populated when coverage is enabled.
	mu       sync.Mutex
	coverage map[string]int64
}

// NewCounters creates a counter block with all probes at zero.
func NewCounters() *Counters {
	return &Counters{}
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
