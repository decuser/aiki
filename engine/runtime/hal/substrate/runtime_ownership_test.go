package substrate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"aiki/engine/runtime/modules"
	"aiki/engine/semantics/value"
)

func TestRuntimeProcessContextIsIsolated(t *testing.T) {
	a := NewGoRuntime()
	b := NewGoRuntime()
	a.SetProgramArgs([]string{"alpha"})
	b.SetProgramArgs([]string{"beta", "gamma"})
	a.SetEnvLookup(func(name string) (string, bool) { return "A:" + name, true })
	b.SetEnvLookup(func(name string) (string, bool) { return "B:" + name, true })

	av, _ := a.Execute("_system_args", nil, nil)
	bv, _ := b.Execute("_system_args", nil, nil)
	if av.Inspect() != `["alpha"]` || bv.Inspect() != `["beta", "gamma"]` {
		t.Fatalf("runtime argument snapshots leaked: a=%s b=%s", av.Inspect(), bv.Inspect())
	}

	name := &value.String{Val: "TOKEN"}
	ae, _ := a.Execute("_system_env", []value.Value{name}, nil)
	be, _ := b.Execute("_system_env", []value.Value{name}, nil)
	if ae.Inspect() != `"A:TOKEN"` || be.Inspect() != `"B:TOKEN"` {
		t.Fatalf("runtime environment views leaked: a=%s b=%s", ae.Inspect(), be.Inspect())
	}
}

func TestRuntimeRNGStateIsIsolated(t *testing.T) {
	a := NewGoRuntime()
	b := NewGoRuntime()
	seed := []value.Value{value.NewNumber(42, 1)}
	max := []value.Value{value.NewNumber(1000000, 1)}
	a.halSeed(seed, nil)
	b.halSeed(seed, nil)

	a1 := a.halRandom(max, nil).Inspect()
	b1 := b.halRandom(max, nil).Inspect()
	if a1 != b1 {
		t.Fatalf("equal explicit seeds should begin equally: a=%s b=%s", a1, b1)
	}
	_ = a.halRandom(max, nil)
	b2 := b.halRandom(max, nil).Inspect()

	c := NewGoRuntime()
	c.halSeed(seed, nil)
	_ = c.halRandom(max, nil)
	c2 := c.halRandom(max, nil).Inspect()
	if b2 != c2 {
		t.Fatalf("advancing runtime a disturbed runtime b: b=%s control=%s", b2, c2)
	}
}

func TestFileReaderAuxiliaryStateIsRuntimeOwned(t *testing.T) {
	path := filepath.Join(t.TempDir(), "lines.txt")
	if err := os.WriteFile(path, []byte("one\ntwo\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	file := &value.File{Path: path, F: f, Mode: "read"}

	a := NewGoRuntime()
	b := NewGoRuntime()
	if got := a.halFileReadLine([]value.Value{file}, nil).Inspect(); got != `"one"` {
		t.Fatalf("read line = %s, want one", got)
	}
	if len(a.fileReaders) != 1 || len(b.fileReaders) != 0 {
		t.Fatalf("file reader cache not runtime-owned: a=%d b=%d", len(a.fileReaders), len(b.fileReaders))
	}
}

func TestRuntimeIOEndpointsAreIsolated(t *testing.T) {
	var aOut, bOut strings.Builder
	a := NewGoRuntime()
	b := NewGoRuntime()
	a.SetIO(strings.NewReader("alpha\n"), &aOut)
	b.SetIO(strings.NewReader("beta\n"), &bOut)

	if got, _ := a.Execute("_read", nil, nil); got.Inspect() != `"alpha"` {
		t.Fatalf("runtime a read = %s", got.Inspect())
	}
	if got, _ := b.Execute("_read", nil, nil); got.Inspect() != `"beta"` {
		t.Fatalf("runtime b read = %s", got.Inspect())
	}

	_, _ = a.Execute("_print", []value.Value{&value.String{Val: "A"}}, nil)
	_, _ = b.Execute("_print", []value.Value{&value.String{Val: "B"}}, nil)
	if aOut.String() != "A" || bOut.String() != "B" {
		t.Fatalf("runtime output endpoints leaked: a=%q b=%q", aOut.String(), bOut.String())
	}
}

func TestRuntimeREPLUserEnvironmentIsIsolated(t *testing.T) {
	a := NewGoRuntime()
	b := NewGoRuntime()
	aEnv := value.NewEnv()
	bEnv := value.NewEnv()
	aEnv.Set("x", value.TRUE)
	bEnv.Set("x", value.TRUE)
	a.SetUserEnv(aEnv)
	b.SetUserEnv(bEnv)

	arg := []value.Value{&value.String{Val: "x"}}
	if got := a.halDelete(arg, nil); got != value.TRUE {
		t.Fatalf("runtime a delete = %s", got.Inspect())
	}
	if _, ok := aEnv.Get("x"); ok {
		t.Fatal("runtime a delete did not affect runtime a environment")
	}
	if _, ok := bEnv.Get("x"); !ok {
		t.Fatal("runtime a delete affected runtime b environment")
	}
}

func TestRuntimeModuleRegistryIsIsolated(t *testing.T) {
	a := NewGoRuntime()
	b := NewGoRuntime()
	a.SetModuleRegistry(modules.NewModuleRegistry([]string{"alpha", "shared"}))
	b.SetModuleRegistry(modules.NewModuleRegistry([]string{"beta"}))

	av, err := a.Execute("_module_roots", nil, nil)
	if err != nil {
		t.Fatalf("runtime a module roots: %v", err)
	}
	bv, err := b.Execute("_module_roots", nil, nil)
	if err != nil {
		t.Fatalf("runtime b module roots: %v", err)
	}
	if av.Inspect() != `["alpha", "shared"]` || bv.Inspect() != `["beta"]` {
		t.Fatalf("runtime module registries leaked: a=%s b=%s", av.Inspect(), bv.Inspect())
	}
}

func TestRuntimeTestStateIsIsolated(t *testing.T) {
	a := NewGoRuntime()
	b := NewGoRuntime()
	a.SetTestFile("a_test.ai")
	b.SetTestFile("b_test.ai")
	a.recordTestPass()
	a.recordTestFailure(nil, "boom")
	b.recordTestPass()

	ap, af, am := a.TestResults()
	bp, bf, bm := b.TestResults()
	if ap != 1 || af != 1 || len(am) != 1 {
		t.Fatalf("runtime a test state = passed %d failed %d messages %d", ap, af, len(am))
	}
	if bp != 1 || bf != 0 || len(bm) != 0 {
		t.Fatalf("runtime b test state leaked: passed %d failed %d messages %d", bp, bf, len(bm))
	}
}
