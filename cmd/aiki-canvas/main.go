package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"aiki/engine/runtime/hal/substrate"
)

func main() {
	width := flag.Int("canvasw", 0, "canvas width")
	height := flag.Int("canvash", 0, "canvas height")
	flag.Parse()

	if *width <= 0 || *height <= 0 {
		fmt.Fprintln(os.Stderr, "canvas: width and height must be positive")
		os.Exit(2)
	}

	cvs := &substrate.CanvasResource{
		Width:    *width,
		Height:   *height,
		BG:       substrate.DefaultBG,
		FG:       substrate.DefaultFG,
		PenSize:  2,
		Commands: make(chan substrate.CanvasCmd, 256),
		Done:     make(chan struct{}),
		Ready:    make(chan struct{}),
	}

	go canvasStdinLoop(os.Stdin, cvs)
	go func() {
		<-cvs.Ready
		fmt.Fprintln(os.Stdout, "READY")
	}()

	RunEbiten(cvs)
}

func canvasStdinLoop(r io.Reader, cvs *substrate.CanvasResource) {
	dec := substrate.NewCanvasDecoder(r)
	for {
		cmd, err := dec.Read()
		if err != nil {
			break
		}
		handleCanvasWire(cmd, cvs)
	}
	select {
	case <-cvs.Done:
	default:
		close(cvs.Done)
	}
}

func handleCanvasWire(cmd any, cvs *substrate.CanvasResource) {
	switch m := cmd.(type) {
	case substrate.CanvasWireBatch:
		for _, one := range m.Cmds {
			handleCanvasWire(one, cvs)
		}
	case substrate.CanvasWireClose:
		select {
		case <-cvs.Done:
		default:
			close(cvs.Done)
		}
	case substrate.CanvasWireSetBG:
		cvs.BG = m.RGBA
		cvs.Commands <- substrate.CanvasCmd{Op: "clear"}
	case substrate.CanvasWireSetFG:
		cvs.FG = m.RGBA
	case substrate.CanvasWireTurtle:
		cvs.SetTurtle(float64(m.X), float64(m.Y), float64(m.Heading), m.Visible, m.RGBA)
	case substrate.CanvasWireCmd:
		clr := cvs.FG
		if m.HasRGBA {
			clr = m.RGBA
		}
		pen := m.Pen
		if pen <= 0 {
			pen = cvs.PenSize
		}
		args := make([]int, len(m.Args))
		for i, a := range m.Args {
			args[i] = int(a)
		}
		cvs.Commands <- substrate.CanvasCmd{Op: m.Op, Args: args, Color: clr, PenSize: pen}
	}
}
