package substrate

import (
	"bufio"
	"encoding/binary"
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
	cmd     *exec.Cmd
	stdin   io.WriteCloser
	bw      *bufio.Writer
	stateMu sync.Mutex

	sendCh chan any
	doneCh chan struct{}

	lastBG color.RGBA
	lastFG color.RGBA
	hasBG  bool
	hasFG  bool

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

	sess := &canvasSession{
		cmd:    cmd,
		stdin:  stdin,
		bw:     bufio.NewWriterSize(stdin, 64*1024),
		waitCh: make(chan struct{}),
		sendCh: make(chan any, 4096),
		doneCh: make(chan struct{}),
	}
	go sess.writerLoop()
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
	select {
	case sess.sendCh <- msg:
	case <-sess.doneCh:
	}
}

func sendCanvasSetBG(cvs *value.Canvas, rgba color.RGBA) {
	sessionsMu.Lock()
	sess := sessions[cvs]
	sessionsMu.Unlock()
	if sess == nil {
		return
	}
	sess.stateMu.Lock()
	if sess.hasBG && sess.lastBG == rgba {
		sess.stateMu.Unlock()
		return
	}
	sess.hasBG = true
	sess.lastBG = rgba
	sess.stateMu.Unlock()
	sendCanvasWire(cvs, CanvasWireSetBG{RGBA: rgba})
}

func sendCanvasSetFG(cvs *value.Canvas, rgba color.RGBA) {
	sessionsMu.Lock()
	sess := sessions[cvs]
	sessionsMu.Unlock()
	if sess == nil {
		return
	}
	sess.stateMu.Lock()
	if sess.hasFG && sess.lastFG == rgba {
		sess.stateMu.Unlock()
		return
	}
	sess.hasFG = true
	sess.lastFG = rgba
	sess.stateMu.Unlock()
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

	// Ask writer to flush close and exit.
	select {
	case sess.sendCh <- CanvasWireClose{}:
	case <-sess.doneCh:
	}
	close(sess.sendCh)
	<-sess.doneCh
	_ = sess.stdin.Close()
	_ = sess.wait()
}

func (s *canvasSession) writerLoop() {
	defer close(s.doneCh)
	enc := canvasNewEncoder()
	batch := make([]any, 0, 4096)

	writeOne := func(msg any) {
		payload, err := enc.encodePayload(msg)
		if err != nil {
			return
		}
		var hdr [4]byte
		binary.LittleEndian.PutUint32(hdr[:], uint32(len(payload)))
		_, _ = s.bw.Write(hdr[:])
		_, _ = s.bw.Write(payload)
		_ = s.bw.Flush()
	}

	flushBatch := func() {
		if len(batch) == 0 {
			return
		}
		msg := batch[0]
		if len(batch) > 1 {
			msg = CanvasWireBatch{Cmds: batch}
		}
		writeOne(msg)
		batch = batch[:0]
	}

	const maxBatch = 4096

	for {
		m, ok := <-s.sendCh
		if !ok {
			flushBatch()
			return
		}
		if _, isClose := m.(CanvasWireClose); isClose {
			flushBatch()
			writeOne(m)
			return
		}
		batch = append(batch, m)

		for len(batch) < maxBatch {
			select {
			case m2, ok2 := <-s.sendCh:
				if !ok2 {
					flushBatch()
					return
				}
				if _, isClose := m2.(CanvasWireClose); isClose {
					flushBatch()
					writeOne(m2)
					return
				}
				batch = append(batch, m2)
			default:
				flushBatch()
				goto next
			}
		}
		flushBatch()

		// Continue and block for the next item.

	next:
	}
}
