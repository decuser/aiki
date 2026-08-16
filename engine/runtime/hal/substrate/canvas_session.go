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

func canvasSessionAlive(cvs *CanvasResource) bool {
	if cvs == nil {
		return false
	}
	cvs.sessionMu.Lock()
	ok := cvs.session != nil
	cvs.sessionMu.Unlock()
	return ok
}

func waitCanvasSession(cvs *CanvasResource) {
	if cvs == nil {
		return
	}
	cvs.sessionMu.Lock()
	sess := cvs.session
	cvs.sessionMu.Unlock()
	if sess != nil {
		_ = sess.wait()
	}
}

func startCanvasSession(cvs *CanvasResource) error {
	cvs.sessionMu.Lock()
	if cvs.session != nil {
		cvs.sessionMu.Unlock()
		return nil
	}
	cvs.sessionMu.Unlock()

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

	go io.Copy(io.Discard, stdout)

	sess := &canvasSession{
		cmd: cmd, stdin: stdin, bw: bufio.NewWriterSize(stdin, 64*1024),
		waitCh: make(chan struct{}), sendCh: make(chan any, 4096), doneCh: make(chan struct{}),
	}
	go sess.writerLoop()
	cvs.sessionMu.Lock()
	cvs.session = sess
	cvs.sessionMu.Unlock()

	close(cvs.Ready)
	bch := make(chan struct{})
	cvs.sessionMu.Lock()
	cvs.bridgeDone = bch
	cvs.sessionMu.Unlock()
	go func() {
		bridgeCanvasCommands(cvs)
		close(bch)
	}()

	go func() {
		_ = sess.wait()
		cvs.sessionMu.Lock()
		owned := cvs.session == sess
		if owned {
			cvs.session = nil
		}
		cvs.sessionMu.Unlock()
		if owned {
			select {
			case <-cvs.Done:
			default:
				close(cvs.Done)
			}
		}
	}()

	return nil
}

func bridgeWait(cvs *CanvasResource) {
	if cvs == nil {
		return
	}
	cvs.sessionMu.Lock()
	ch := cvs.bridgeDone
	cvs.sessionMu.Unlock()
	if ch != nil {
		<-ch
	}
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

func bridgeCanvasCommands(cvs *CanvasResource) {
	for {
		select {
		case <-cvs.Done:
			// Deliver whatever is already queued before closing. A program
			// that draws and then destroys the canvas would otherwise lose
			// the drawing, since closing and draining are both triggered by
			// Done and would race.
			for {
				select {
				case cmd := <-cvs.Commands:
					sendCanvasCmd(cvs, cmd)
				default:
					sendCanvasClose(cvs)
					return
				}
			}
		case cmd := <-cvs.Commands:
			sendCanvasCmd(cvs, cmd)
		}
	}
}

func sendCanvasCmd(cvs *CanvasResource, cmd CanvasCmd) {
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

func sendCanvasWire(cvs *CanvasResource, msg any) {
	cvs.sessionMu.Lock()
	sess := cvs.session
	cvs.sessionMu.Unlock()
	if sess == nil {
		return
	}
	select {
	case sess.sendCh <- msg:
	case <-sess.doneCh:
	}
}

func sendCanvasSetBG(cvs *CanvasResource, rgba color.RGBA) {
	cvs.sessionMu.Lock()
	sess := cvs.session
	cvs.sessionMu.Unlock()
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

func sendCanvasSetFG(cvs *CanvasResource, rgba color.RGBA) {
	cvs.sessionMu.Lock()
	sess := cvs.session
	cvs.sessionMu.Unlock()
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

func SendCanvasTurtle(cvs *CanvasResource, x, y, heading float64, visible bool, rgba color.RGBA) {
	sendCanvasWire(cvs, CanvasWireTurtle{
		X:       float32(x),
		Y:       float32(y),
		Heading: float32(heading),
		Visible: visible,
		RGBA:    rgba,
	})
}

func sendCanvasClose(cvs *CanvasResource) {
	cvs.sessionMu.Lock()
	sess := cvs.session
	cvs.session = nil
	cvs.sessionMu.Unlock()
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
