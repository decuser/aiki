package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"image/color"
	"io"
	"os"
	"strings"

	"aiki/engine/runtime/hal/substrate"
	"aiki/engine/semantics/value"
)

func runCanvasChild(opts Options) {
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
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var msg substrate.CanvasIPCMsg
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			continue
		}
		handleCanvasIPC(msg, cvs)
	}
	select {
	case <-cvs.Done:
	default:
		close(cvs.Done)
	}
}

func handleCanvasIPC(msg substrate.CanvasIPCMsg, cvs *value.Canvas) {
	if msg.Proto != 0 && msg.Proto != 1 {
		return
	}
	switch msg.Op {
	case "close":
		select {
		case <-cvs.Done:
		default:
			close(cvs.Done)
		}
		return
	case "set_bg":
		if c, ok := rgbaFromMsg(msg); ok {
			cvs.BG = c
			cvs.Commands <- value.CanvasCmd{Op: "clear"}
		}
		return
	case "set_fg":
		if c, ok := rgbaFromMsg(msg); ok {
			cvs.FG = c
		}
		return
	}

	clr := cvs.FG
	if c, ok := rgbaFromMsg(msg); ok {
		clr = c
	}

	pen := float32(msg.Pen)
	if pen <= 0 {
		pen = cvs.PenSize
	}

	cvs.Commands <- value.CanvasCmd{Op: msg.Op, Args: msg.Args, Color: clr, PenSize: pen}
}

func rgbaFromMsg(msg substrate.CanvasIPCMsg) (color.RGBA, bool) {
	if len(msg.RGBA) != 4 {
		return color.RGBA{}, false
	}
	r := msg.RGBA[0]
	g := msg.RGBA[1]
	b := msg.RGBA[2]
	a := msg.RGBA[3]
	if r < 0 || r > 255 || g < 0 || g > 255 || b < 0 || b > 255 || a < 0 || a > 255 {
		return color.RGBA{}, false
	}
	return color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)}, true
}
