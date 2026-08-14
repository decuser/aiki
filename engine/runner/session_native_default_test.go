package runner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSessionImportsNativeDefaultByBareName(t *testing.T) {
	tmp := t.TempDir()
	modDir := filepath.Join(tmp, "widget")
	if err := os.MkdirAll(modDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	mod := `package "widget/native"

let answer = 42
export(:answer)
`
	if err := os.WriteFile(filepath.Join(modDir, "native.ai"), []byte(mod), 0o644); err != nil {
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

	v := s.Eval(`let widget = import("widget")`)
	if v != nil && v.Type() == "error" {
		t.Fatalf("bare import error: %s", v.Inspect())
	}
	v = s.Eval(`widget`)
	if got := v.Inspect(); got != "<module widget>" {
		t.Fatalf("bare module identity: got %s", got)
	}
	v = s.Eval(`widget.answer`)
	if got := v.Inspect(); got != "42" {
		t.Fatalf("bare module export: got %s", got)
	}

	v = s.Eval(`let native = import("widget/native")`)
	if v != nil && v.Type() == "error" {
		t.Fatalf("native import error: %s", v.Inspect())
	}
	v = s.Eval(`native`)
	if got := v.Inspect(); got != "<module widget/native>" {
		t.Fatalf("native module identity: got %s", got)
	}
}
