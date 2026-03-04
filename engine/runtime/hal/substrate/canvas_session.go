package substrate

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"aiki/engine/semantics/value"
)

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

	cmd := exec.Command(exe, args...)
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
	msg := CanvasIPCMsg{Op: cmd.Op, Args: cmd.Args, Pen: int(cmd.PenSize)}
	msg.RGBA = []int{int(cmd.Color.R), int(cmd.Color.G), int(cmd.Color.B), int(cmd.Color.A)}
	sendCanvasMsg(cvs, msg)
}

func sendCanvasMsg(cvs *value.Canvas, msg CanvasIPCMsg) {
	sessionsMu.Lock()
	sess := sessions[cvs]
	sessionsMu.Unlock()
	if sess == nil {
		return
	}
	line, err := msg.encodeLine()
	if err != nil {
		return
	}
	sess.mu.Lock()
	_, _ = sess.stdin.Write(line)
	sess.mu.Unlock()
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
		line, err := (CanvasIPCMsg{Op: "close"}).encodeLine()
		if err != nil {
			return
		}
		_, _ = sess.stdin.Write(line)
	}()

	_ = sess.stdin.Close()
	_ = sess.wait()
}
