package substrate

import (
	"runtime/pprof"
	"sync"
	"testing"

	"aiki/engine"
	"aiki/engine/runtime/hal"
	"aiki/engine/semantics/value"
)

type recordingProbe struct {
	mu   sync.Mutex
	hits map[engine.SemanticKind]int
}

func (p *recordingProbe) Hit(kind engine.SemanticKind, site engine.SemanticSite) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.hits == nil {
		p.hits = make(map[engine.SemanticKind]int)
	}
	p.hits[kind]++
}

func TestStoreSemanticProbe(t *testing.T) {
	probe := &recordingProbe{}
	ctx := &hal.EvalContext{Probe: probe}
	s := halStoreNew([]value.Value{value.NewNumber(4, 1)}, ctx)
	if _, ok := s.(*value.Store); !ok {
		t.Fatalf("store.new: got %T", s)
	}
	if got := halStoreSet([]value.Value{s, value.NewNumber(2, 1), value.NewNumber(7094, 1)}, ctx); got != value.EMPTY {
		t.Fatalf("store.set: got %s", got.Inspect())
	}
	got := halStoreGet([]value.Value{s, value.NewNumber(2, 1)}, ctx)
	if got.Inspect() != "7094" {
		t.Fatalf("store.get: got %s", got.Inspect())
	}
	if probe.hits[engine.SemanticStoreRead] != 1 || probe.hits[engine.SemanticStoreWrite] != 1 {
		t.Fatalf("store read/write: expected 1/1, got %d/%d", probe.hits[engine.SemanticStoreRead], probe.hits[engine.SemanticStoreWrite])
	}
}

func TestFailedStoreAccessDoesNotCountSemanticWork(t *testing.T) {
	probe := &recordingProbe{}
	ctx := &hal.EvalContext{Probe: probe}
	s := halStoreNew([]value.Value{value.NewNumber(2, 1)}, ctx)
	got := halStoreGet([]value.Value{s, value.NewNumber(99, 1)}, ctx)
	if _, ok := got.(*value.Fault); !ok {
		t.Fatalf("store.get out of bounds: got %T (%s)", got, got.Inspect())
	}
	if probe.hits[engine.SemanticStoreRead] != 0 {
		t.Fatalf("failed store.get should not count a read, got %d", probe.hits[engine.SemanticStoreRead])
	}
}

func TestProfileContextCarriesAikiCorrelationLabels(t *testing.T) {
	rt := NewGoRuntime()
	labels := engine.ProfileLabels{
		Layer:     "substrate",
		Function:  "append",
		File:      "<prelude>",
		Line:      "71",
		Primitive: "_append",
	}
	ctx := rt.profileContext(labels)
	for key, want := range map[string]string{
		"aiki_layer":     labels.Layer,
		"aiki_function":  labels.Function,
		"aiki_file":      labels.File,
		"aiki_line":      labels.Line,
		"aiki_primitive": labels.Primitive,
	} {
		got, ok := pprof.Label(ctx, key)
		if !ok || got != want {
			t.Fatalf("%s: got %q/%v, want %q/true", key, got, ok, want)
		}
	}
}
