package value

import "image/color"

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
}

func (c *Canvas) Type() Type      { return CanvasType }
func (c *Canvas) Inspect() string { return "<canvas>" }

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
