package substrate

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"aiki/engine/semantics/value"
)

// Game implements ebiten.Game for canvas rendering.
type Game struct {
	canvas  *value.Canvas
	buffer  *ebiten.Image
	ops     []value.CanvasCmd
	redoOps []value.CanvasCmd
	maxOps  int
}

// NewGame creates a new game for the given canvas.
func NewGame(canvas *value.Canvas) *Game {
	return &Game{
		canvas: canvas,
		buffer: ebiten.NewImage(canvas.Width, canvas.Height),
		maxOps: 100,
	}
}

// Update processes commands from the canvas channel.
func (g *Game) Update() error {
	for {
		select {
		case <-g.canvas.Done:
			return ebiten.Termination
		case cmd := <-g.canvas.Commands:
			g.handleCmd(cmd)
		default:
			return nil
		}
	}
}

func (g *Game) handleCmd(cmd value.CanvasCmd) {
	switch cmd.Op {
	case "clear":
		g.buffer.Fill(g.canvas.BG)
		g.ops = nil
		g.redoOps = nil
	case "undo":
		if len(g.ops) > 0 {
			op := g.ops[len(g.ops)-1]
			g.ops = g.ops[:len(g.ops)-1]
			if len(g.redoOps) < g.maxOps {
				g.redoOps = append(g.redoOps, op)
			}
			g.redraw()
		}
	case "redo":
		if len(g.redoOps) > 0 {
			op := g.redoOps[len(g.redoOps)-1]
			g.redoOps = g.redoOps[:len(g.redoOps)-1]
			g.ops = append(g.ops, op)
			g.drawOp(op)
		}
	case "save":
		g.save(cmd.Path, cmd.Result)
		return
	default:
		g.drawOp(cmd)
		if len(g.ops) >= g.maxOps {
			g.ops = g.ops[1:]
		}
		g.ops = append(g.ops, cmd)
		g.redoOps = nil
	}
	if cmd.Result != nil {
		cmd.Result <- nil
	}
}

func (g *Game) drawOp(cmd value.CanvasCmd) {
	clr := cmd.Color
	args := cmd.Args
	pen := cmd.PenSize
	if pen < 1 {
		pen = 2
	}

	switch cmd.Op {
	case "dot":
		if pen <= 1 {
			g.buffer.Set(args[0], args[1], clr)
		} else {
			vector.DrawFilledCircle(g.buffer, float32(args[0]), float32(args[1]), pen/2, clr, false)
		}
	case "line":
		vector.StrokeLine(g.buffer, float32(args[0]), float32(args[1]), float32(args[2]), float32(args[3]), pen, clr, false)
	case "rect":
		vector.StrokeRect(g.buffer, float32(args[0]), float32(args[1]), float32(args[2]), float32(args[3]), pen, clr, false)
	case "fill_rect":
		vector.DrawFilledRect(g.buffer, float32(args[0]), float32(args[1]), float32(args[2]), float32(args[3]), clr, false)
	case "circle":
		vector.StrokeCircle(g.buffer, float32(args[0]), float32(args[1]), float32(args[2]), pen, clr, false)
	case "fill_circle":
		vector.DrawFilledCircle(g.buffer, float32(args[0]), float32(args[1]), float32(args[2]), clr, false)
	case "arc":
		drawArc(g.buffer, args[0], args[1], args[2], args[3], args[4], clr, pen)
	case "oval":
		drawOval(g.buffer, args[0], args[1], args[2], args[3], clr, false, pen)
	case "fill_oval":
		drawOval(g.buffer, args[0], args[1], args[2], args[3], clr, true, pen)
	case "text":
		ebitenutil.DebugPrintAt(g.buffer, cmd.Text, args[0], args[1])
	}
}

func (g *Game) redraw() {
	g.buffer.Fill(g.canvas.BG)
	for _, op := range g.ops {
		g.drawOp(op)
	}
}

func (g *Game) save(path string, result chan error) {
	bounds := g.buffer.Bounds()
	img := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			img.Set(x, y, g.buffer.At(x, y))
		}
	}
	f, err := os.Create(path)
	if err != nil {
		result <- err
		return
	}
	defer f.Close()
	result <- png.Encode(f, img)
}

// Draw renders the buffer to the screen.
func (g *Game) Draw(screen *ebiten.Image) {
	screen.DrawImage(g.buffer, nil)
}

// Layout returns the canvas dimensions.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return g.canvas.Width, g.canvas.Height
}

func drawArc(img *ebiten.Image, cx, cy, r, startDeg, endDeg int, clr color.RGBA, pen float32) {
	start := float64(startDeg)
	end := float64(endDeg)
	if end < start {
		end += 360
	}
	for angle := start; angle <= end; angle += 0.5 {
		rad := angle * math.Pi / 180
		x := cx + int(float64(r)*math.Cos(rad))
		y := cy + int(float64(r)*math.Sin(rad))
		if pen <= 1 {
			img.Set(x, y, clr)
		} else {
			vector.DrawFilledCircle(img, float32(x), float32(y), pen/2, clr, false)
		}
	}
}

func drawOval(img *ebiten.Image, cx, cy, rx, ry int, clr color.RGBA, fill bool, pen float32) {
	for angle := 0.0; angle < 360; angle += 0.5 {
		rad := angle * math.Pi / 180
		x := cx + int(float64(rx)*math.Cos(rad))
		y := cy + int(float64(ry)*math.Sin(rad))
		if fill {
			vector.StrokeLine(img, float32(cx), float32(y), float32(x), float32(y), 1, clr, false)
		} else {
			if pen <= 1 {
				img.Set(x, y, clr)
			} else {
				vector.DrawFilledCircle(img, float32(x), float32(y), pen/2, clr, false)
			}
		}
	}
}

// RunEbiten starts the ebiten game loop for a canvas.
func RunEbiten(canvas *value.Canvas) {
	ebiten.SetWindowSize(canvas.Width, canvas.Height)
	ebiten.SetWindowTitle("Aiki")
	game := NewGame(canvas)
	game.buffer.Fill(canvas.BG)
	close(canvas.Ready)
	ebiten.RunGame(game)
}
