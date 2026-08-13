package main

import (
	"bufio"
	"fmt"
	"image/color"
	"io"
	"os"
	"strings"

	"aiki/engine/runtime/hal/substrate"
)

// The canvas child normally opens a window and renders. It can instead record
// the command stream it receives and draw nothing, which lets programs that
// draw run without a display and lets their behavior be compared against a
// transcript rather than confirmed by a person watching the screen.
//
// Recording is requested through the environment rather than a flag, because
// the child inherits the environment of the process that spawned it and the
// spawning code therefore needs to know nothing about this mode:
//
//	AIKI_CANVAS=record             select the recording child
//	AIKI_CANVAS_RECORD=<path>      where to write the transcript
//
// Without a path the transcript goes to standard error, since standard output
// carries the readiness handshake.
//
// The recorder shares the decoder and the wire vocabulary with the rendering
// child. It is a third implementation of the canvas protocol, so an opcode
// that the two halves of the rendering path might agree to misread has to be
// read correctly here as well.

const (
	canvasModeEnv   = "AIKI_CANVAS"
	canvasRecordEnv = "AIKI_CANVAS_RECORD"
	canvasModeValue = "record"
)

// canvasRecordRequested reports whether this child should record rather than render.
func canvasRecordRequested() bool {
	return strings.EqualFold(os.Getenv(canvasModeEnv), canvasModeValue)
}

func runCanvasRecordChild(opts Options) {
	if opts.CanvasW <= 0 || opts.CanvasH <= 0 {
		fmt.Fprintln(os.Stderr, "canvas: width and height must be positive")
		os.Exit(2)
	}

	var sink io.Writer = os.Stderr
	if path := os.Getenv(canvasRecordEnv); path != "" {
		// Appended rather than truncated: a program may open several canvases
		// in turn, and each session adds its own header to one transcript.
		f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "canvas: cannot open %s: %v\n", path, err)
			os.Exit(2)
		}
		defer f.Close()
		sink = f
	}

	w := bufio.NewWriter(sink)
	defer w.Flush()

	fmt.Fprintf(w, "canvas %d %d\n", opts.CanvasW, opts.CanvasH)

	// Nothing has to be prepared before commands can be accepted, so the
	// parent's readiness handshake is answered immediately.
	fmt.Fprintln(os.Stdout, "READY")

	dec := substrate.NewCanvasDecoder(os.Stdin)
	for {
		cmd, err := dec.Read()
		if err != nil {
			break
		}
		done := recordCanvasWire(w, cmd)
		w.Flush()
		if done {
			break
		}
	}
}

// recordCanvasWire writes one line per command and reports whether the stream
// has been closed. Batches are flattened, so the transcript does not depend on
// how the sender chose to group commands into frames.
func recordCanvasWire(w io.Writer, cmd any) bool {
	switch m := cmd.(type) {
	case substrate.CanvasWireBatch:
		for _, one := range m.Cmds {
			if recordCanvasWire(w, one) {
				return true
			}
		}
		return false
	case substrate.CanvasWireClose:
		fmt.Fprintln(w, "close")
		return true
	case substrate.CanvasWireSetBG:
		fmt.Fprintf(w, "setbg %s\n", canvasHexColor(m.RGBA))
		return false
	case substrate.CanvasWireSetFG:
		fmt.Fprintf(w, "setfg %s\n", canvasHexColor(m.RGBA))
		return false
	case substrate.CanvasWireTurtle:
		fmt.Fprintf(w, "turtle %.3f %.3f %.3f %t %s\n",
			m.X, m.Y, m.Heading, m.Visible, canvasHexColor(m.RGBA))
		return false
	case substrate.CanvasWireCmd:
		var b strings.Builder
		b.WriteString(m.Op)
		for _, a := range m.Args {
			fmt.Fprintf(&b, " %d", a)
		}
		if m.HasRGBA {
			fmt.Fprintf(&b, " %s", canvasHexColor(m.RGBA))
		} else {
			b.WriteString(" fg")
		}
		fmt.Fprintf(&b, " pen %.2f", m.Pen)
		fmt.Fprintln(w, b.String())
		return false
	default:
		fmt.Fprintf(w, "unknown %T\n", cmd)
		return false
	}
}

func canvasHexColor(c color.RGBA) string {
	return fmt.Sprintf("#%02X%02X%02X%02X", c.R, c.G, c.B, c.A)
}
