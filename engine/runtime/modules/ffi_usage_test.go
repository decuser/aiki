package modules

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestFFIUsageFollowsNativeAliasesAndTransitiveFFI(t *testing.T) {
	root := t.TempDir()
	lib := filepath.Join(root, "lib")
	mustWriteUsageFile(t, filepath.Join(lib, "thing", "native.ai"), `package "thing/native"
export()
`)
	mustWriteUsageFile(t, filepath.Join(lib, "thing", "native.help"), ``)
	mustWriteUsageFile(t, filepath.Join(lib, "thing", "native.doc"), ``)
	mustWriteUsageFile(t, filepath.Join(lib, "fast", "ffi.ai"), `package "fast/ffi"
export()
`)
	mustWriteUsageFile(t, filepath.Join(lib, "fast", "ffi.help"), ``)
	mustWriteUsageFile(t, filepath.Join(lib, "fast", "ffi.doc"), ``)
	mustWriteUsageFile(t, filepath.Join(lib, "mixed", "mixed.ai"), `package "mixed"
let fast = import("fast/ffi")
export()
`)
	mustWriteUsageFile(t, filepath.Join(lib, "mixed", "mixed.help"), ``)
	mustWriteUsageFile(t, filepath.Join(lib, "mixed", "mixed.doc"), ``)
	entry := filepath.Join(root, "main.ai")
	mustWriteUsageFile(t, entry, `let native = import("thing")
let mixed = import("mixed")
`)

	got, err := FFIUsage(testGrammar(t), entry, []string{lib})
	if err != nil {
		t.Fatalf("FFIUsage: %v", err)
	}
	if want := []string{"fast/ffi"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("FFIUsage = %v, want %v", got, want)
	}
}

func TestFFIUsageFollowsRelativeImports(t *testing.T) {
	root := t.TempDir()
	lib := filepath.Join(root, "lib")
	mustWriteUsageFile(t, filepath.Join(lib, "fast", "ffi.ai"), `package "fast/ffi"
export()
`)
	mustWriteUsageFile(t, filepath.Join(lib, "fast", "ffi.help"), ``)
	mustWriteUsageFile(t, filepath.Join(lib, "fast", "ffi.doc"), ``)
	entry := filepath.Join(root, "main.ai")
	mustWriteUsageFile(t, entry, `let local = import("./local")
`)
	mustWriteUsageFile(t, filepath.Join(root, "local.ai"), `let fast = import("fast/ffi")
`)

	got, err := FFIUsage(testGrammar(t), entry, []string{lib})
	if err != nil {
		t.Fatalf("FFIUsage: %v", err)
	}
	if want := []string{"fast/ffi"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("FFIUsage = %v, want %v", got, want)
	}
}

func mustWriteUsageFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}
