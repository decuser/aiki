package main

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"

	"aiki/engine/runtime/hal/substrate"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Game implements ebiten.Game for canvas rendering.
type Game struct {
	canvas  *substrate.CanvasResource
	buffer  *ebiten.Image
	overlay *ebiten.Image
	ops     []substrate.CanvasCmd
	redoOps []substrate.CanvasCmd
	maxOps  int
}

// NewGame creates a new game for the given canvas.
func NewGame(canvas *substrate.CanvasResource) *Game {
	return &Game{
		canvas:  canvas,
		buffer:  ebiten.NewImage(canvas.Width, canvas.Height),
		overlay: ebiten.NewImage(canvas.Width, canvas.Height),
		maxOps:  100,
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

func (g *Game) handleCmd(cmd substrate.CanvasCmd) {
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

func (g *Game) drawOp(cmd substrate.CanvasCmd) {
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

// drawTurtle draws the turtle triangle on the overlay.
func (g *Game) drawTurtle() {
	g.overlay.Clear()

	g.canvas.TurtleMu.RLock()
	visible := g.canvas.TurtleVisible
	x := g.canvas.TurtleX
	y := g.canvas.TurtleY
	heading := g.canvas.TurtleHeading
	clr := g.canvas.TurtleColor
	g.canvas.TurtleMu.RUnlock()

	if !visible {
		return
	}

	// Triangle size in pixels (fixed, not scaled)
	size := float64(12)

	// Convert heading to radians (0 = north/up, clockwise positive)
	// In screen coords, up is -Y, so we adjust
	rad := (heading - 90) * math.Pi / 180

	// Triangle points: tip at front, two back corners
	// Tip is at turtle position
	tipX := x
	tipY := y

	// Back corners at 135 and 225 degrees from heading
	backAngle := 140.0 * math.Pi / 180
	backDist := size

	leftRad := rad + backAngle
	rightRad := rad - backAngle

	leftX := x + backDist*math.Cos(leftRad)
	leftY := y + backDist*math.Sin(leftRad)

	rightX := x + backDist*math.Cos(rightRad)
	rightY := y + backDist*math.Sin(rightRad)

	// Draw filled triangle
	path := vector.Path{}
	path.MoveTo(float32(tipX), float32(tipY))
	path.LineTo(float32(leftX), float32(leftY))
	path.LineTo(float32(rightX), float32(rightY))
	path.Close()

	vs, is := path.AppendVerticesAndIndicesForFilling(nil, nil)
	for i := range vs {
		vs[i].SrcX = 1
		vs[i].SrcY = 1
		vs[i].ColorR = float32(clr.R) / 255
		vs[i].ColorG = float32(clr.G) / 255
		vs[i].ColorB = float32(clr.B) / 255
		vs[i].ColorA = float32(clr.A) / 255
	}

	op := &ebiten.DrawTrianglesOptions{}
	op.AntiAlias = true
	g.overlay.DrawTriangles(vs, is, emptyImage, op)
}

// emptyImage is a 1x1 white image used for solid color triangles.
var emptyImage = func() *ebiten.Image {
	img := ebiten.NewImage(3, 3)
	img.Fill(color.White)
	return img
}()

// Draw renders the buffer and overlay to the screen.
func (g *Game) Draw(screen *ebiten.Image) {
	screen.DrawImage(g.buffer, nil)
	g.drawTurtle()
	screen.DrawImage(g.overlay, nil)
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
func RunEbiten(canvas *substrate.CanvasResource) {
	ebiten.SetWindowSize(canvas.Width, canvas.Height)
	ebiten.SetWindowTitle("Aiki")
	game := NewGame(canvas)
	game.buffer.Fill(canvas.BG)
	close(canvas.Ready)
	ebiten.RunGame(game)
}
