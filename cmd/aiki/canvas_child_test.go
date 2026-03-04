package main

import (
	"strings"
	"testing"
	"time"

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

	in := `{"op":"nope"}
{"op":"close"}
`
	go canvasStdinLoop(strings.NewReader(in), cvs)

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

	in := `{"op":"set_bg","rgba":[0,0,0,255]}
{"op":"close"}
`
	go canvasStdinLoop(strings.NewReader(in), cvs)

	select {
	case cmd := <-cvs.Commands:
		if cmd.Op != "clear" {
			t.Fatalf("expected clear after set_bg, got %s", cmd.Op)
		}
	case <-time.After(1 * time.Second):
		t.Fatalf("expected a command to be enqueued")
	}
}
