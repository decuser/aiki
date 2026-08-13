package main

import (
	"fmt"
	"io"
	"os"

	"aiki/engine/runtime/hal/substrate"
	"aiki/engine/semantics/value"
)

func runCanvasChild(opts Options) {
	// The same child records instead of rendering when asked to, so a program
	// that draws can run without a display. See canvas_record.go.
	if canvasRecordRequested() {
		runCanvasRecordChild(opts)
		return
	}

	if opts.CanvasW <= 0 || opts.CanvasH <= 0 {
		fmt.Fprintln(os.Stderr, "canvas: width and height must be positive")
		os.Exit(2)
	}

	cvs := &value.Canvas{
		Width:    opts.CanvasW,
		Height:   opts.CanvasH,
		BG:       substrate.DefaultBG,
		FG:       substrate.DefaultFG,
		PenSize:  2,
		Commands: make(chan value.CanvasCmd, 256),
		Done:     make(chan struct{}),
		Ready:    make(chan struct{}),
	}

	go canvasStdinLoop(os.Stdin, cvs)
	go func() {
		<-cvs.Ready
		fmt.Fprintln(os.Stdout, "READY")
	}()

	substrate.RunEbiten(cvs)
}

func canvasStdinLoop(r io.Reader, cvs *value.Canvas) {
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

func handleCanvasWire(cmd any, cvs *value.Canvas) {
	switch m := cmd.(type) {
	case substrate.CanvasWireBatch:
		for _, one := range m.Cmds {
			handleCanvasWire(one, cvs)
		}
		return
	case substrate.CanvasWireClose:
		select {
		case <-cvs.Done:
		default:
			close(cvs.Done)
		}
		return
	case substrate.CanvasWireSetBG:
		cvs.BG = m.RGBA
		cvs.Commands <- value.CanvasCmd{Op: "clear"}
		return
	case substrate.CanvasWireSetFG:
		cvs.FG = m.RGBA
		return
	case substrate.CanvasWireTurtle:
		cvs.SetTurtle(float64(m.X), float64(m.Y), float64(m.Heading), m.Visible, m.RGBA)
		return
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
		cvs.Commands <- value.CanvasCmd{Op: m.Op, Args: args, Color: clr, PenSize: pen}
		return
	default:
		return
	}
}
