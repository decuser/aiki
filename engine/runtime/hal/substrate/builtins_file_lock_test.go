package substrate

import (
	"path/filepath"
	"testing"

	"aiki/engine/semantics/value"
)

func TestFileLockExclusiveAndReusable(t *testing.T) {
	rt1 := NewGoRuntime()
	defer rt1.CloseAllResources()
	rt2 := NewGoRuntime()
	defer rt2.CloseAllResources()

	path := filepath.Join(t.TempDir(), "coord.lock")
	first := rt1.halFileTryLock([]value.Value{&value.String{Val: path}}, nil)
	lock1, ok := first.(*value.FileLock)
	if !ok {
		t.Fatalf("first try_lock = %s, want file lock", first.Inspect())
	}
	if second := rt2.halFileTryLock([]value.Value{&value.String{Val: path}}, nil); second != value.FALSE {
		t.Fatalf("contended try_lock = %s, want false", second.Inspect())
	}
	if got := rt1.halFileUnlock([]value.Value{lock1}, nil); got != value.TRUE {
		t.Fatalf("unlock = %s, want true", got.Inspect())
	}
	third := rt2.halFileTryLock([]value.Value{&value.String{Val: path}}, nil)
	lock2, ok := third.(*value.FileLock)
	if !ok {
		t.Fatalf("try_lock after release = %s, want file lock", third.Inspect())
	}
	if got := rt2.halFileUnlock([]value.Value{lock2}, nil); got != value.TRUE {
		t.Fatalf("second unlock = %s, want true", got.Inspect())
	}
}

func TestFileUnlockRejectsForeignRuntime(t *testing.T) {
	rt1 := NewGoRuntime()
	defer rt1.CloseAllResources()
	rt2 := NewGoRuntime()
	defer rt2.CloseAllResources()

	path := filepath.Join(t.TempDir(), "coord.lock")
	got := rt1.halFileTryLock([]value.Value{&value.String{Val: path}}, nil)
	lock, ok := got.(*value.FileLock)
	if !ok {
		t.Fatalf("try_lock = %s, want file lock", got.Inspect())
	}
	foreign := rt2.halFileUnlock([]value.Value{lock}, nil)
	if foreign.Type() != value.ListType {
		t.Fatalf("foreign unlock = %s, want shaped error", foreign.Inspect())
	}
}
