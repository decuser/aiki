package value

import (
	"image/color"
	"sync"
)

// Canvas represents a drawing surface.
type Canvas struct {
	Width    int
	Height   int
	BG       color.RGBA
	FG       color.RGBA
	PenSize  float32
	Commands chan CanvasCmd
	Done     chan struct{}
	Ready    chan struct{}

	// Turtle overlay state
	TurtleMu      sync.RWMutex
	TurtleX       float64
	TurtleY       float64
	TurtleHeading float64
	TurtleVisible bool
	TurtleColor   color.RGBA
}

func (c *Canvas) Type() Type      { return CanvasType }
func (c *Canvas) Inspect() string { return "<canvas>" }

// SetTurtle updates turtle overlay state.
func (c *Canvas) SetTurtle(x, y, heading float64, visible bool, clr color.RGBA) {
	c.TurtleMu.Lock()
	c.TurtleX = x
	c.TurtleY = y
	c.TurtleHeading = heading
	c.TurtleVisible = visible
	c.TurtleColor = clr
	c.TurtleMu.Unlock()
}

// CanvasCmd represents a drawing operation.
type CanvasCmd struct {
	Op      string
	Args    []int
	Color   color.RGBA
	PenSize float32
	Text    string
	Path    string
	Result  chan error
}
