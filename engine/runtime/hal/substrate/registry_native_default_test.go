package substrate

import (
	"os"
	"path/filepath"
	"testing"

	"aiki/engine/syntax"
	"aiki/engine/syntax/grammar"
)

func testRegistryGrammar(t *testing.T) *grammar.Grammar {
	t.Helper()
	g, err := grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
	if err != nil {
		t.Fatalf("loading grammar: %v", err)
	}
	return g
}

func writeTestModule(t *testing.T, root, rel, pkg string) string {
	t.Helper()
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	src := "package \"" + pkg + "\"\n\nlet value = 1\nexport(:value)\n"
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatalf("write module: %v", err)
	}
	return path
}

func TestRegistryNativeDefaultAlias(t *testing.T) {
	root := t.TempDir()
	nativePath := writeTestModule(t, root, "math/native.ai", "math/native")
	writeTestModule(t, root, "math/ffi.ai", "math/ffi")

	r := NewModuleRegistry([]string{root})
	if err := r.Scan(testRegistryGrammar(t)); err != nil {
		t.Fatalf("scan: %v", err)
	}

	path, canonical, ok := r.Resolve("math")
	if !ok {
		t.Fatal("expected bare math alias")
	}
	if path != nativePath {
		t.Fatalf("math path: got %q, want %q", path, nativePath)
	}
	if canonical != "math/native" {
		t.Fatalf("math canonical: got %q, want %q", canonical, "math/native")
	}

	if !r.HasPackage("math") || !r.HasPackage("math/native") || !r.HasPackage("math/ffi") {
		t.Fatalf("expected math, math/native, and math/ffi to be registered: %v", r.ListPackages())
	}
}

func TestRegistryExplicitBareWins(t *testing.T) {
	root := t.TempDir()
	barePath := writeTestModule(t, root, "foo.ai", "foo")
	writeTestModule(t, root, "foo/native.ai", "foo/native")

	r := NewModuleRegistry([]string{root})
	if err := r.Scan(testRegistryGrammar(t)); err != nil {
		t.Fatalf("scan: %v", err)
	}

	path, canonical, ok := r.Resolve("foo")
	if !ok {
		t.Fatal("expected explicit foo package")
	}
	if path != barePath || canonical != "foo" {
		t.Fatalf("explicit foo must win: path=%q canonical=%q", path, canonical)
	}
}

func TestRegistryFFIOnlyDoesNotBecomeDefault(t *testing.T) {
	root := t.TempDir()
	writeTestModule(t, root, "regex/ffi.ai", "regex/ffi")

	r := NewModuleRegistry([]string{root})
	if err := r.Scan(testRegistryGrammar(t)); err != nil {
		t.Fatalf("scan: %v", err)
	}

	if r.HasPackage("regex") {
		t.Fatal("ffi-only package must not become bare default")
	}
	if !r.HasPackage("regex/ffi") {
		t.Fatal("expected regex/ffi package")
	}
}

func TestRegistryCanonicalListExcludesAliases(t *testing.T) {
	root := t.TempDir()
	writeTestModule(t, root, "bytes/native.ai", "bytes/native")
	writeTestModule(t, root, "bytes/ffi.ai", "bytes/ffi")

	r := NewModuleRegistry([]string{root})
	if err := r.Scan(testRegistryGrammar(t)); err != nil {
		t.Fatalf("scan: %v", err)
	}

	public := r.ListPackages()
	canonical := r.ListCanonicalPackages()
	if len(public) != 3 {
		t.Fatalf("public packages: got %v", public)
	}
	if len(canonical) != 2 {
		t.Fatalf("canonical packages: got %v", canonical)
	}
	for _, name := range canonical {
		if name == "bytes" {
			t.Fatal("bare alias must not appear in canonical package list")
		}
	}
}

func TestModuleRootsUseExplicitDistributionAndDevelopmentRoots(t *testing.T) {
	home := filepath.Join("home", "user")
	exe := filepath.Join("opt", "aiki")
	work := filepath.Join("forge", "dev", "aiki")
	got := ModuleRoots(home, exe, work)
	want := []string{
		filepath.Join(exe, "lib"),
		filepath.Join(exe, "vendor"),
		filepath.Join(work, "lib"),
		filepath.Join(work, "vendor"),
		filepath.Join(home, ".aiki", "lib"),
	}
	if len(got) != len(want) {
		t.Fatalf("module roots: got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("module root %d: got %q, want %q", i, got[i], want[i])
		}
	}
}

func TestModuleRootsDoNotScanWorkingDirectoryItself(t *testing.T) {
	work := filepath.Join("forge", "dev")
	got := ModuleRoots("", "", work)
	for _, root := range got {
		if root == filepath.Clean(work) {
			t.Fatalf("working directory must not be a registry root: %v", got)
		}
	}
}

func TestModuleRegistryRootsReturnsCopyInLookupOrder(t *testing.T) {
	want := []string{"/dist/lib", "/dist/vendor", "/work/lib", "/home/user/.aiki/lib"}
	r := NewModuleRegistry(want)
	got := r.Roots()
	if len(got) != len(want) {
		t.Fatalf("roots length: got %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("root %d: got %q, want %q", i, got[i], want[i])
		}
	}
	got[0] = "/mutated"
	if reread := r.Roots(); reread[0] != want[0] {
		t.Fatalf("Roots exposed mutable registry state: got %q, want %q", reread[0], want[0])
	}
}
