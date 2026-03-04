package substrate

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"testing"
	"time"

	"aiki/engine/semantics/value"
)

// Test helper child process. This is invoked by exec'ing the test binary itself.
func TestCanvasFakeChild(t *testing.T) {
	if os.Getenv("AIKI_CANVAS_FAKE_CHILD") != "1" {
		return
	}

	fmt.Fprintln(os.Stdout, "READY")

	if os.Getenv("AIKI_CANVAS_FAKE_EXIT_EARLY") == "1" {
		os.Exit(0)
	}

	for {
		cmd, err := CanvasReadCommand(os.Stdin)
		if err != nil {
			if err == io.EOF {
				os.Exit(0)
			}
			os.Exit(0)
		}
		if _, ok := cmd.(CanvasWireClose); ok {
			os.Exit(0)
		}
	}
	os.Exit(0)
}

func withFakeCanvasExec(t *testing.T) func() {
	t.Helper()

	prev := canvasExecCommand
	canvasExecCommand = func(name string, args ...string) *exec.Cmd {
		// Ignore name and args; run this test binary as the child.
		cmd := exec.Command(os.Args[0], "-test.run=TestCanvasFakeChild")
		cmd.Env = append(os.Environ(), "AIKI_CANVAS_FAKE_CHILD=1")
		return cmd
	}

	return func() {
		canvasExecCommand = prev
	}
}

func TestStartCanvasSessionAndDestroy(t *testing.T) {
	restore := withFakeCanvasExec(t)
	defer restore()

	cvs := &value.Canvas{
		Width:    100,
		Height:   100,
		BG:       DefaultBG,
		FG:       DefaultFG,
		PenSize:  2,
		Commands: make(chan value.CanvasCmd, 10),
		Done:     make(chan struct{}),
		Ready:    make(chan struct{}),
	}

	if err := startCanvasSession(cvs); err != nil {
		t.Fatalf("startCanvasSession: %v", err)
	}

	// Ensure session exists
	sessionsMu.Lock()
	sess := sessions[cvs]
	sessionsMu.Unlock()
	if sess == nil || sess.cmd == nil || sess.stdin == nil {
		t.Fatalf("expected session to be registered")
	}

	// Destroy should reap child
	select {
	case <-cvs.Done:
		t.Fatalf("Done closed too early")
	default:
	}

	close(cvs.Done)

	// Wait for session removal
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		sessionsMu.Lock()
		_, ok := sessions[cvs]
		sessionsMu.Unlock()
		if !ok {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("session not removed after destroy")
}

func TestChildExitEarlyDoesNotPanicOnSend(t *testing.T) {
	// Inject a fake child that exits immediately after READY.
	prev := canvasExecCommand
	canvasExecCommand = func(name string, args ...string) *exec.Cmd {
		cmd := exec.Command(os.Args[0], "-test.run=TestCanvasFakeChild")
		cmd.Env = append(os.Environ(), "AIKI_CANVAS_FAKE_CHILD=1", "AIKI_CANVAS_FAKE_EXIT_EARLY=1")
		return cmd
	}
	defer func() { canvasExecCommand = prev }()

	cvs := &value.Canvas{
		Width:    100,
		Height:   100,
		BG:       DefaultBG,
		FG:       DefaultFG,
		PenSize:  2,
		Commands: make(chan value.CanvasCmd, 10),
		Done:     make(chan struct{}),
		Ready:    make(chan struct{}),
	}

	if err := startCanvasSession(cvs); err != nil {
		t.Fatalf("startCanvasSession: %v", err)
	}

	// Allow child to exit
	time.Sleep(50 * time.Millisecond)

	// Sending should not block or panic even if process is gone.
	done := make(chan struct{})
	go func() {
		sendCanvasWire(cvs, CanvasWireCmd{Op: "dot", Args: []int32{1, 2}})
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatalf("sendCanvasWire blocked after child exit")
	}

	// Close session if still present to avoid leakage.
	select {
	case <-cvs.Done:
	default:
		close(cvs.Done)
	}
}

// Basic sanity that CloseAllCanvases closes and reaps multiple sessions.
func TestCloseAllCanvasesReapsAll(t *testing.T) {
	restore := withFakeCanvasExec(t)
	defer restore()

	c1 := &value.Canvas{Width: 10, Height: 10, BG: DefaultBG, FG: DefaultFG, PenSize: 1, Commands: make(chan value.CanvasCmd, 10), Done: make(chan struct{}), Ready: make(chan struct{})}
	c2 := &value.Canvas{Width: 10, Height: 10, BG: DefaultBG, FG: DefaultFG, PenSize: 1, Commands: make(chan value.CanvasCmd, 10), Done: make(chan struct{}), Ready: make(chan struct{})}

	if err := startCanvasSession(c1); err != nil {
		t.Fatalf("startCanvasSession c1: %v", err)
	}
	if err := startCanvasSession(c2); err != nil {
		t.Fatalf("startCanvasSession c2: %v", err)
	}

	// CloseAllCanvases only closes tracked canvases.
	trackCanvas(c1)
	trackCanvas(c2)

	CloseAllCanvases()

	// Confirm sessions map is empty (or at least these entries are gone).
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		sessionsMu.Lock()
		_, ok1 := sessions[c1]
		_, ok2 := sessions[c2]
		sessionsMu.Unlock()
		if !ok1 && !ok2 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("canvases not fully reaped")
}
