package substrate

import (
	"image/color"
	"sync"
)

// CanvasResource is the Go substrate realization of an Aiki canvas handle.
// It is owned by one GoRuntime (or by the dedicated canvas child process).
type CanvasResource struct {
	Width    int
	Height   int
	BG       color.RGBA
	FG       color.RGBA
	PenSize  float32
	Commands chan CanvasCmd
	Done     chan struct{}
	Ready    chan struct{}

	sessionMu  sync.Mutex
	session    *canvasSession
	bridgeDone chan struct{}

	TurtleMu      sync.RWMutex
	TurtleX       float64
	TurtleY       float64
	TurtleHeading float64
	TurtleVisible bool
	TurtleColor   color.RGBA
}

// CanvasCmd is an internal substrate drawing operation.
type CanvasCmd struct {
	Op      string
	Args    []int
	Color   color.RGBA
	PenSize float32
	Text    string
	Path    string
	Result  chan error
}

func (c *CanvasResource) SetTurtle(x, y, heading float64, visible bool, clr color.RGBA) {
	c.TurtleMu.Lock()
	c.TurtleX = x
	c.TurtleY = y
	c.TurtleHeading = heading
	c.TurtleVisible = visible
	c.TurtleColor = clr
	c.TurtleMu.Unlock()
}
