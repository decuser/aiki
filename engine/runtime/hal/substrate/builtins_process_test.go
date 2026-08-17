package substrate

import (
	"fmt"
	"io"
	"os"
	"testing"

	"aiki/engine/semantics/value"
)

func TestProcessHelper(t *testing.T) {
	mode := ""
	for i, arg := range os.Args {
		if arg == "--" && i+1 < len(os.Args) {
			mode = os.Args[i+1]
			break
		}
	}
	switch mode {
	case "echo":
		data, _ := io.ReadAll(os.Stdin)
		fmt.Fprint(os.Stdout, string(data))
		fmt.Fprint(os.Stderr, "helper-stderr\n")
	case "exit7":
		os.Exit(7)
	case "block":
		select {}
	case "env":
		fmt.Fprint(os.Stdout, os.Getenv("AIKI_CHILD_ENV_TEST"))
		os.Exit(0)
	default:
		return
	}
}

func processStartArgs(command string, argv ...string) []value.Value {
	args := make([]value.Value, len(argv))
	for i, arg := range argv {
		args[i] = &value.String{Val: arg}
	}
	return []value.Value{&value.String{Val: command}, &value.List{Elements: args}}
}

func TestProcessLifecycleAndIOEndpoints(t *testing.T) {
	rt := NewGoRuntime()
	defer rt.CloseAllResources()

	started := rt.halProcessStart(processStartArgs(os.Args[0], "-test.run=TestProcessHelper", "--", "echo"), nil)
	proc, ok := started.(*value.Process)
	if !ok {
		t.Fatalf("process.start = %v (%T), want process", started, started)
	}

	stdin, ok := rt.halProcessStdin([]value.Value{proc}, nil).(*value.Endpoint)
	if !ok {
		t.Fatal("process.stdin did not return endpoint")
	}
	stdout, ok := rt.halProcessStdout([]value.Value{proc}, nil).(*value.Endpoint)
	if !ok {
		t.Fatal("process.stdout did not return endpoint")
	}
	stderr, ok := rt.halProcessStderr([]value.Value{proc}, nil).(*value.Endpoint)
	if !ok {
		t.Fatal("process.stderr did not return endpoint")
	}

	if got := rt.halIOWrite([]value.Value{stdin, &value.String{Val: "hello\n"}}, nil); got != value.TRUE {
		t.Fatalf("io.write(stdin) = %v", got)
	}
	if got := rt.halIOClose([]value.Value{stdin}, nil); got != value.TRUE {
		t.Fatalf("io.close(stdin) = %v", got)
	}

	if got := rt.halIOReadLine([]value.Value{stdout}, nil); got.Inspect() != `"hello"` {
		t.Fatalf("stdout line = %s, want %q", got.Inspect(), "hello")
	}
	if got := rt.halIOReadLine([]value.Value{stderr}, nil); got.Inspect() != `"helper-stderr"` {
		t.Fatalf("stderr line = %s", got.Inspect())
	}
	if got := rt.halProcessWait([]value.Value{proc}, nil); got.Inspect() != "0" {
		t.Fatalf("process.wait = %s, want 0", got.Inspect())
	}
	if got := rt.halProcessWait([]value.Value{proc}, nil); got.Inspect() != "0" {
		t.Fatalf("second process.wait = %s, want 0", got.Inspect())
	}
}

func TestProcessWaitReturnsNonzeroExitCode(t *testing.T) {
	rt := NewGoRuntime()
	defer rt.CloseAllResources()
	started := rt.halProcessStart(processStartArgs(os.Args[0], "-test.run=TestProcessHelper", "--", "exit7"), nil)
	proc, ok := started.(*value.Process)
	if !ok {
		t.Fatalf("process.start = %v", started)
	}
	if got := rt.halProcessWait([]value.Value{proc}, nil); got.Inspect() != "7" {
		t.Fatalf("process.wait = %s, want 7", got.Inspect())
	}
}

func TestProcessHandleIsRuntimeOwned(t *testing.T) {
	rt1 := NewGoRuntime()
	rt2 := NewGoRuntime()
	defer rt1.CloseAllResources()
	defer rt2.CloseAllResources()
	started := rt1.halProcessStart(processStartArgs(os.Args[0], "-test.run=TestProcessHelper", "--", "exit7"), nil)
	proc := started.(*value.Process)
	got := rt2.halProcessStdout([]value.Value{proc}, nil)
	if !value.IsShapedError(got) {
		t.Fatalf("cross-runtime process handle = %v, want shaped error", got)
	}
}

func TestChildProcessesInheritRuntimeEnvironmentSnapshot(t *testing.T) {
	rt := NewGoRuntime()
	defer rt.CloseAllResources()
	name := &value.String{Val: "AIKI_CHILD_ENV_TEST"}
	val := &value.String{Val: "runtime-owned"}
	if got := rt.halSystemSetEnv([]value.Value{name, val}, nil); got != value.TRUE {
		t.Fatalf("set env = %s", got.Inspect())
	}

	started := rt.halProcessStart(processStartArgs(os.Args[0], "-test.run=TestProcessHelper", "--", "env"), nil)
	proc, ok := started.(*value.Process)
	if !ok {
		t.Fatalf("process.start = %v (%T), want process", started, started)
	}
	stdout, ok := rt.halProcessStdout([]value.Value{proc}, nil).(*value.Endpoint)
	if !ok {
		t.Fatal("process.stdout did not return endpoint")
	}
	gotBytes := rt.halIORead([]value.Value{stdout, value.NewNumber(64, 1)}, nil)
	b, ok := gotBytes.(*value.Bytes)
	if !ok || string(b.Val) != "runtime-owned" {
		t.Fatalf("streaming child env = %s", gotBytes.Inspect())
	}
	if got := rt.halProcessWait([]value.Value{proc}, nil); got.Inspect() != "0" {
		t.Fatalf("process.wait = %s", got.Inspect())
	}

	result := rt.halSystemExec(processStartArgs(os.Args[0], "-test.run=TestProcessHelper", "--", "env"), nil)
	list, ok := result.(*value.List)
	if !ok || list.Shape != "ok" || len(list.Elements) != 3 {
		t.Fatalf("system.exec = %s, want @ok result", result.Inspect())
	}
	if got := list.Elements[0].Inspect(); got != `"runtime-owned"` {
		t.Fatalf("captured child env = %s", got)
	}
}
