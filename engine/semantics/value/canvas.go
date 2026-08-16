package value

import "fmt"

// Canvas is an opaque Aiki-visible handle to a runtime-owned canvas resource.
// Concrete drawing/session state belongs to the active runtime substrate.
type Canvas struct {
	ID uint64
}

func (c *Canvas) Type() Type { return CanvasType }
func (c *Canvas) Inspect() string {
	if c == nil || c.ID == 0 {
		return "<canvas>"
	}
	return fmt.Sprintf("<canvas:%d>", c.ID)
}
