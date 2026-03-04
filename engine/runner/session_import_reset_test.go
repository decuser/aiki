package runner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSessionImportAfterReset(t *testing.T) {
	tmp := t.TempDir()

	// Create a minimal module in the temp dir so registry scan finds it via root "."
	mod := `package "import_target"

export(:ANSWER, :add_two)

let ANSWER = 40

let add_two = (x) {
	return x + 2
}
`
	if err := os.WriteFile(filepath.Join(tmp, "import_target.ai"), []byte(mod), 0o644); err != nil {
		t.Fatalf("write module: %v", err)
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer func() { _ = os.Chdir(wd) }()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	s, err := NewSession()
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}

	// First import
	v := s.Eval(`import("import_target", :ANSWER, :add_two)`)
	if v != nil && v.Type() == "error" {
		t.Fatalf("import error: %s", v.Inspect())
	}
	v = s.Eval(`ANSWER`)
	if v.Inspect() != "40" {
		t.Fatalf("ANSWER mismatch after import: got %s", v.Inspect())
	}

	// Reset and import again
	if err := s.Reset(); err != nil {
		t.Fatalf("Reset: %v", err)
	}
	v = s.Eval(`import("import_target", :ANSWER, :add_two)`)
	if v != nil && v.Type() == "error" {
		t.Fatalf("import error after reset: %s", v.Inspect())
	}
	v = s.Eval(`add_two(40)`)
	if v.Inspect() != "42" {
		t.Fatalf("add_two mismatch after reset import: got %s", v.Inspect())
	}
}
