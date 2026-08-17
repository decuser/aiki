package substrate

import (
	"os"
	"testing"

	"aiki/engine/semantics/value"
)

func TestTerminalRejectsNonTerminalFile(t *testing.T) {
	rt := NewGoRuntime()
	defer rt.CloseAllResources()

	f, err := os.CreateTemp("", "aiki-term-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	v := &value.File{Path: f.Name(), F: f, Mode: "read_write"}
	if got := rt.halTermIs([]value.Value{v}, nil); got != value.FALSE {
		t.Fatalf("term.is regular file = %s, want false", got.Inspect())
	}
	if got := rt.halTermSize([]value.Value{v}, nil); !value.IsShapedError(got) {
		t.Fatalf("term.size regular file = %s, want shaped error", got.Inspect())
	}
	if got := rt.halTermRaw([]value.Value{v}, nil); !value.IsShapedError(got) {
		t.Fatalf("term.raw regular file = %s, want shaped error", got.Inspect())
	}
}
