package value

import "fmt"

// Process is an opaque Aiki-visible handle to a runtime-owned child process.
type Process struct{ ID uint64 }

func (p *Process) Type() Type { return ProcessType }
func (p *Process) Inspect() string {
	if p == nil || p.ID == 0 {
		return "<process>"
	}
	return fmt.Sprintf("<process:%d>", p.ID)
}

// Endpoint is an opaque Aiki-visible sequential I/O endpoint. Concrete reader,
// writer, and lifecycle state belongs to the active runtime substrate.
type Endpoint struct {
	ID    uint64
	Label string
}

func (e *Endpoint) Type() Type { return EndpointType }
func (e *Endpoint) Inspect() string {
	if e == nil || e.ID == 0 {
		return "<endpoint>"
	}
	if e.Label == "" {
		return fmt.Sprintf("<endpoint:%d>", e.ID)
	}
	return fmt.Sprintf("<endpoint:%s:%d>", e.Label, e.ID)
}
