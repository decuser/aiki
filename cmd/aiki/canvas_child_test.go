package main

import (
	"bytes"
	"testing"
	"time"

	"aiki/engine/runtime/hal/substrate"
	"aiki/engine/semantics/value"
)

func TestCanvasStdinLoopClose(t *testing.T) {
	cvs := &value.Canvas{
		Width:    10,
		Height:   10,
		Commands: make(chan value.CanvasCmd, 10),
		Done:     make(chan struct{}),
		Ready:    make(chan struct{}),
	}

	var b bytes.Buffer
	_ = substrate.CanvasWriteFrame(&b, substrate.CanvasWireCmd{Op: "nope"})
	_ = substrate.CanvasWriteFrame(&b, substrate.CanvasWireClose{})
	go canvasStdinLoop(bytes.NewReader(b.Bytes()), cvs)

	select {
	case <-cvs.Done:
	case <-time.After(1 * time.Second):
		t.Fatalf("Done not closed")
	}
}

func TestCanvasStdinLoopSetBGEnqueuesClear(t *testing.T) {
	cvs := &value.Canvas{
		Width:    10,
		Height:   10,
		Commands: make(chan value.CanvasCmd, 10),
		Done:     make(chan struct{}),
		Ready:    make(chan struct{}),
	}

	var b bytes.Buffer
	_ = substrate.CanvasWriteFrame(&b, substrate.CanvasWireSetBG{RGBA: substrate.DefaultBG})
	_ = substrate.CanvasWriteFrame(&b, substrate.CanvasWireClose{})
	go canvasStdinLoop(bytes.NewReader(b.Bytes()), cvs)

	select {
	case cmd := <-cvs.Commands:
		if cmd.Op != "clear" {
			t.Fatalf("expected clear after set_bg, got %s", cmd.Op)
		}
	case <-time.After(1 * time.Second):
		t.Fatalf("expected a command to be enqueued")
	}
}

func TestCanvasStdinLoopEOFClosesDone(t *testing.T) {
	cvs := &value.Canvas{
		Width:    10,
		Height:   10,
		Commands: make(chan value.CanvasCmd, 10),
		Done:     make(chan struct{}),
		Ready:    make(chan struct{}),
	}

	go canvasStdinLoop(bytes.NewReader(nil), cvs)

	select {
	case <-cvs.Done:
	case <-time.After(1 * time.Second):
		t.Fatalf("Done not closed on EOF")
	}
}
