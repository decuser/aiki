package substrate

import (
	"os"
	"testing"
	"time"

	"aiki/engine/semantics/value"
)

func TestSignalWatchReturnsReceiveOnlyChannelAndStops(t *testing.T) {
	rt := NewGoRuntime()
	defer rt.CloseAllResources()
	got := rt.halSignalWatch([]value.Value{&value.List{Elements: []value.Value{&value.Symbol{Val: "interrupt"}}}}, nil)
	ch, ok := got.(*value.Channel)
	if !ok {
		t.Fatalf("signal.watch = %v (%T), want channel", got, got)
	}
	if ch.CanSend() {
		t.Fatal("signal watch channel must be receive-only")
	}
	if got := rt.halSignalStop([]value.Value{ch}, nil); got != value.TRUE {
		t.Fatalf("signal.stop = %v", got)
	}
}

func TestSignalWatchRejectsUnknownSignal(t *testing.T) {
	rt := NewGoRuntime()
	defer rt.CloseAllResources()
	got := rt.halSignalWatch([]value.Value{&value.List{Elements: []value.Value{&value.Symbol{Val: "bogus"}}}}, nil)
	if !value.IsShapedError(got) {
		t.Fatalf("signal.watch bogus = %v, want shaped error", got)
	}
}

func TestSignalSourceIsRuntimeOwned(t *testing.T) {
	rt1 := NewGoRuntime()
	rt2 := NewGoRuntime()
	defer rt1.CloseAllResources()
	defer rt2.CloseAllResources()
	got := rt1.halSignalWatch([]value.Value{&value.List{Elements: []value.Value{&value.Symbol{Val: "interrupt"}}}}, nil)
	ch := got.(*value.Channel)
	if got := rt2.halSignalStop([]value.Value{ch}, nil); !value.IsShapedError(got) {
		t.Fatalf("cross-runtime signal.stop = %v, want shaped error", got)
	}
}

func TestSignalSendTerminateStopsChild(t *testing.T) {
	rt := NewGoRuntime()
	defer rt.CloseAllResources()
	started := rt.halProcessStart(processStartArgs(os.Args[0], "-test.run=TestProcessHelper", "--", "block"), nil)
	proc, ok := started.(*value.Process)
	if !ok {
		t.Fatalf("process.start = %v", started)
	}
	time.Sleep(20 * time.Millisecond)
	if got := rt.halSignalSend([]value.Value{proc, &value.Symbol{Val: "terminate"}}, nil); got != value.TRUE {
		t.Fatalf("signal.send terminate = %v", got)
	}
	if got := rt.halProcessWait([]value.Value{proc}, nil); value.IsShapedError(got) {
		t.Fatalf("process.wait after terminate = %v", got)
	}
}
