package substrate

import (
	"bufio"
	"errors"
	"fmt"
	"image/color"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"aiki/engine/semantics/value"
)

// canvasExecCommand allows tests to inject a fake child process.
var canvasExecCommand = exec.Command

type canvasSession struct {
	cmd   *exec.Cmd
	stdin io.WriteCloser
	mu    sync.Mutex

	waitOnce sync.Once
	waitErr  error
	waitCh   chan struct{}
}

func (s *canvasSession) wait() error {
	s.waitOnce.Do(func() {
		s.waitErr = s.cmd.Wait()
		close(s.waitCh)
	})
	<-s.waitCh
	return s.waitErr
}

var (
	sessionsMu sync.Mutex
	sessions   = map[*value.Canvas]*canvasSession{}
)

func canvasSessionAlive(cvs *value.Canvas) bool {
	sessionsMu.Lock()
	_, ok := sessions[cvs]
	sessionsMu.Unlock()
	return ok
}

func waitCanvasSession(cvs *value.Canvas) {
	sessionsMu.Lock()
	sess := sessions[cvs]
	sessionsMu.Unlock()
	if sess == nil {
		return
	}
	_ = sess.wait()
}

func startCanvasSession(cvs *value.Canvas) error {
	sessionsMu.Lock()
	if _, ok := sessions[cvs]; ok {
		sessionsMu.Unlock()
		return nil
	}
	sessionsMu.Unlock()

	exe, err := os.Executable()
	if err != nil {
		return err
	}

	args := []string{
		"-canvas",
		fmt.Sprintf("-canvasw=%d", cvs.Width),
		fmt.Sprintf("-canvash=%d", cvs.Height),
	}

	cmd := canvasExecCommand(exe, args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return err
	}

	if err := waitForCanvasReady(stdout); err != nil {
		_ = stdin.Close()
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
		return err
	}

	// Drain any remaining stdout so the pipe does not fill.
	go io.Copy(io.Discard, stdout)

	sess := &canvasSession{cmd: cmd, stdin: stdin, waitCh: make(chan struct{})}
	sessionsMu.Lock()
	sessions[cvs] = sess
	sessionsMu.Unlock()

	close(cvs.Ready)

	go bridgeCanvasCommands(cvs)

	// Parent requested close.
	go func() {
		<-cvs.Done
		sendCanvasClose(cvs)
	}()

	// Child exited on its own (window closed). Mark canvas as done.
	go func() {
		_ = sess.wait()
		sessionsMu.Lock()
		_, ok := sessions[cvs]
		delete(sessions, cvs)
		sessionsMu.Unlock()
		if ok {
			select {
			case <-cvs.Done:
			default:
				close(cvs.Done)
			}
		}
	}()

	return nil
}

func waitForCanvasReady(r io.Reader) error {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "READY" {
			return nil
		}
	}
	if err := sc.Err(); err != nil {
		return err
	}
	return errors.New("canvas: child exited before READY")
}

func bridgeCanvasCommands(cvs *value.Canvas) {
	for {
		select {
		case <-cvs.Done:
			return
		case cmd := <-cvs.Commands:
			sendCanvasCmd(cvs, cmd)
		}
	}
}

func sendCanvasCmd(cvs *value.Canvas, cmd value.CanvasCmd) {
	args := make([]int32, len(cmd.Args))
	for i, a := range cmd.Args {
		args[i] = int32(a)
	}
	w := CanvasWireCmd{Op: cmd.Op, Args: args, Pen: cmd.PenSize}
	// CanvasCmd always carries a color; preserve it.
	w.HasRGBA = true
	w.RGBA = cmd.Color
	sendCanvasWire(cvs, w)
}

func sendCanvasWire(cvs *value.Canvas, msg any) {
	sessionsMu.Lock()
	sess := sessions[cvs]
	sessionsMu.Unlock()
	if sess == nil {
		return
	}
	sess.mu.Lock()
	_ = CanvasWriteFrame(sess.stdin, msg)
	sess.mu.Unlock()
}

func sendCanvasSetBG(cvs *value.Canvas, rgba color.RGBA) {
	sendCanvasWire(cvs, CanvasWireSetBG{RGBA: rgba})
}

func sendCanvasSetFG(cvs *value.Canvas, rgba color.RGBA) {
	sendCanvasWire(cvs, CanvasWireSetFG{RGBA: rgba})
}

func sendCanvasClose(cvs *value.Canvas) {
	sessionsMu.Lock()
	sess := sessions[cvs]
	delete(sessions, cvs)
	sessionsMu.Unlock()
	if sess == nil {
		return
	}

	func() {
		sess.mu.Lock()
		defer sess.mu.Unlock()
		_ = CanvasWriteFrame(sess.stdin, CanvasWireClose{})
	}()

	_ = sess.stdin.Close()
	_ = sess.wait()
}
