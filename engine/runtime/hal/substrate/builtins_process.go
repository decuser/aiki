package substrate

import (
	"bufio"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"aiki/engine/runtime/hal"
	"aiki/engine/semantics/value"
)

type EndpointResource struct {
	Reader     io.Reader
	Writer     io.Writer
	Closer     io.Closer
	Buffered   *bufio.Reader
	LocalAddr  string
	RemoteAddr string
	mu         sync.Mutex
	closed     bool
}

func (r *EndpointResource) close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return nil
	}
	r.closed = true
	if r.Closer != nil {
		return r.Closer.Close()
	}
	return nil
}

func (r *EndpointResource) isClosed() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.closed
}

type ProcessResource struct {
	Cmd      *exec.Cmd
	Stdin    *value.Endpoint
	Stdout   *value.Endpoint
	Stderr   *value.Endpoint
	waitOnce sync.Once
	done     chan struct{}
	exitCode int64
	waitErr  error
}

func (r *ProcessResource) wait() (int64, error) {
	r.waitOnce.Do(func() {
		err := r.Cmd.Wait()
		if err == nil {
			r.exitCode = 0
		} else if ee, ok := err.(*exec.ExitError); ok {
			r.exitCode = int64(ee.ExitCode())
		} else {
			r.waitErr = err
		}
		close(r.done)
	})
	<-r.done
	return r.exitCode, r.waitErr
}

func processArgs(args []value.Value, opname string) (string, []string, *value.Fault) {
	if len(args) != 2 {
		return "", nil, value.NewFault("%s: want 2 arguments, got %d", opname, len(args))
	}
	name, ok := args[0].(*value.String)
	if !ok {
		return "", nil, value.NewFault("%s: expected string command, got %s", opname, args[0].Type())
	}
	list, ok := args[1].(*value.List)
	if !ok {
		return "", nil, value.NewFault("%s: expected list of string arguments, got %s", opname, args[1].Type())
	}
	argv := make([]string, len(list.Elements))
	for i, elem := range list.Elements {
		s, ok := elem.(*value.String)
		if !ok {
			return "", nil, value.NewFault("%s: argument %d must be string, got %s", opname, i, elem.Type())
		}
		argv[i] = s.Val
	}
	return name.Val, argv, nil
}

func (g *GoRuntime) newEndpoint(label string, reader io.Reader, writer io.Writer, closer io.Closer) *value.Endpoint {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.nextEndpointID++
	ep := &value.Endpoint{ID: g.nextEndpointID, Label: label}
	resource := &EndpointResource{Reader: reader, Writer: writer, Closer: closer}
	if reader != nil {
		resource.Buffered = bufio.NewReader(reader)
	}
	g.endpointResources[ep] = resource
	return ep
}

func (g *GoRuntime) endpointResource(ep *value.Endpoint) (*EndpointResource, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	r, ok := g.endpointResources[ep]
	return r, ok
}

func (g *GoRuntime) processResource(p *value.Process) (*ProcessResource, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	r, ok := g.processResources[p]
	return r, ok
}

func (g *GoRuntime) halProcessStart(args []value.Value, ctx *hal.EvalContext) value.Value {
	name, argv, fault := processArgs(args, "process.start")
	if fault != nil {
		return fault
	}
	g.mu.RLock()
	cwd := g.workingDir
	g.mu.RUnlock()
	command := name
	if !filepath.IsAbs(command) && strings.ContainsAny(command, `/\\`) {
		command = filepath.Join(cwd, command)
	}
	cmd := exec.Command(command, argv...)
	cmd.Dir = cwd
	env := g.environmentSnapshot()
	env["PWD"] = cwd
	cmd.Env = environmentList(env)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return value.NewShapedError("process", "%s", err.Error())
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		_ = stdin.Close()
		return value.NewShapedError("process", "%s", err.Error())
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		_ = stdin.Close()
		_ = stdout.Close()
		return value.NewShapedError("process", "%s", err.Error())
	}
	if err := cmd.Start(); err != nil {
		_ = stdin.Close()
		_ = stdout.Close()
		_ = stderr.Close()
		return value.NewShapedError("process", "%s", err.Error())
	}

	stdinEP := g.newEndpoint("process.stdin", nil, stdin, stdin)
	stdoutEP := g.newEndpoint("process.stdout", stdout, nil, stdout)
	stderrEP := g.newEndpoint("process.stderr", stderr, nil, stderr)
	g.mu.Lock()
	g.nextProcessID++
	ph := &value.Process{ID: g.nextProcessID}
	g.processResources[ph] = &ProcessResource{Cmd: cmd, Stdin: stdinEP, Stdout: stdoutEP, Stderr: stderrEP, done: make(chan struct{})}
	g.mu.Unlock()
	return ph
}

func (g *GoRuntime) processEndpoint(args []value.Value, which string) value.Value {
	if len(args) != 1 {
		return value.NewFault("process.%s: want 1 argument, got %d", which, len(args))
	}
	ph, ok := args[0].(*value.Process)
	if !ok {
		return value.NewFault("process.%s: expected process, got %s", which, args[0].Type())
	}
	r, ok := g.processResource(ph)
	if !ok {
		return value.NewShapedError("process", "process does not belong to this runtime")
	}
	switch which {
	case "stdin":
		return r.Stdin
	case "stdout":
		return r.Stdout
	default:
		return r.Stderr
	}
}
func (g *GoRuntime) halProcessStdin(args []value.Value, ctx *hal.EvalContext) value.Value {
	return g.processEndpoint(args, "stdin")
}
func (g *GoRuntime) halProcessStdout(args []value.Value, ctx *hal.EvalContext) value.Value {
	return g.processEndpoint(args, "stdout")
}
func (g *GoRuntime) halProcessStderr(args []value.Value, ctx *hal.EvalContext) value.Value {
	return g.processEndpoint(args, "stderr")
}

func (g *GoRuntime) halProcessWait(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("process.wait: want 1 argument, got %d", len(args))
	}
	ph, ok := args[0].(*value.Process)
	if !ok {
		return value.NewFault("process.wait: expected process, got %s", args[0].Type())
	}
	r, ok := g.processResource(ph)
	if !ok {
		return value.NewShapedError("process", "process does not belong to this runtime")
	}
	code, err := r.wait()
	if err != nil {
		return value.NewShapedError("process", "%s", err.Error())
	}
	return value.NewNumber(code, 1)
}

func (g *GoRuntime) halProcessTerminate(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("process.terminate: want 1 argument, got %d", len(args))
	}
	ph, ok := args[0].(*value.Process)
	if !ok {
		return value.NewFault("process.terminate: expected process, got %s", args[0].Type())
	}
	r, ok := g.processResource(ph)
	if !ok {
		return value.NewShapedError("process", "process does not belong to this runtime")
	}
	select {
	case <-r.done:
		return value.TRUE
	default:
	}
	if r.Cmd.Process == nil {
		return value.NewShapedError("process", "process has not started")
	}
	if err := r.Cmd.Process.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
		return value.NewShapedError("process", "%s", err.Error())
	}
	return value.TRUE
}

func (g *GoRuntime) CloseAllEndpoints() {
	g.mu.RLock()
	resources := make([]*EndpointResource, 0, len(g.endpointResources))
	for _, r := range g.endpointResources {
		resources = append(resources, r)
	}
	g.mu.RUnlock()
	for _, r := range resources {
		_ = r.close()
	}
}

func (g *GoRuntime) CloseAllProcesses() {
	g.mu.RLock()
	resources := make([]*ProcessResource, 0, len(g.processResources))
	for _, r := range g.processResources {
		resources = append(resources, r)
	}
	g.mu.RUnlock()
	for _, r := range resources {
		select {
		case <-r.done:
		default:
			if r.Cmd.Process != nil {
				_ = r.Cmd.Process.Kill()
			}
		}
		_, _ = r.wait()
	}
}

// CloseAllResources releases all runtime-owned external resources.
func (g *GoRuntime) CloseAllResources() {
	g.CloseAllFileLocks()
	g.CloseAllTerminals()
	g.CloseAllSignals()
	g.CloseAllNetwork()
	g.CloseAllEndpoints()
	g.CloseAllProcesses()
	g.CloseAllCanvases()
}
