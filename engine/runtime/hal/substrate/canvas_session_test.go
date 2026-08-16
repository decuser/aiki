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

	cvs := &CanvasResource{
		Width:    100,
		Height:   100,
		BG:       DefaultBG,
		FG:       DefaultFG,
		PenSize:  2,
		Commands: make(chan CanvasCmd, 10),
		Done:     make(chan struct{}),
		Ready:    make(chan struct{}),
	}

	if err := startCanvasSession(cvs); err != nil {
		t.Fatalf("startCanvasSession: %v", err)
	}

	// Ensure session exists
	cvs.sessionMu.Lock()
	sess := cvs.session
	cvs.sessionMu.Unlock()
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
		cvs.sessionMu.Lock()
		alive := cvs.session != nil
		cvs.sessionMu.Unlock()
		if !alive {
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

	cvs := &CanvasResource{
		Width:    100,
		Height:   100,
		BG:       DefaultBG,
		FG:       DefaultFG,
		PenSize:  2,
		Commands: make(chan CanvasCmd, 10),
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

// Basic sanity that runtime-owned CloseAllCanvases closes and reaps multiple sessions.
func TestCloseAllCanvasesReapsAll(t *testing.T) {
	restore := withFakeCanvasExec(t)
	defer restore()

	r1 := &CanvasResource{Width: 10, Height: 10, BG: DefaultBG, FG: DefaultFG, PenSize: 1, Commands: make(chan CanvasCmd, 10), Done: make(chan struct{}), Ready: make(chan struct{})}
	r2 := &CanvasResource{Width: 10, Height: 10, BG: DefaultBG, FG: DefaultFG, PenSize: 1, Commands: make(chan CanvasCmd, 10), Done: make(chan struct{}), Ready: make(chan struct{})}

	if err := startCanvasSession(r1); err != nil {
		t.Fatalf("startCanvasSession c1: %v", err)
	}
	if err := startCanvasSession(r2); err != nil {
		t.Fatalf("startCanvasSession c2: %v", err)
	}

	// CloseAllCanvases only closes canvases tracked by this runtime.
	rt := NewGoRuntime()
	c1 := &value.Canvas{ID: 1}
	c2 := &value.Canvas{ID: 2}
	rt.trackCanvas(c1, r1)
	rt.trackCanvas(c2, r2)

	rt.CloseAllCanvases()

	// Confirm sessions map is empty (or at least these entries are gone).
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		r1.sessionMu.Lock()
		ok1 := r1.session != nil
		r1.sessionMu.Unlock()
		r2.sessionMu.Lock()
		ok2 := r2.session != nil
		r2.sessionMu.Unlock()
		if !ok1 && !ok2 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("canvases not fully reaped")
}
