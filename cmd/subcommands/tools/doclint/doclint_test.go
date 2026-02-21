package doclint_test

import (
	"os"
	"path/filepath"
	"testing"

	"aiki/cmd/subcommands/tools/doclint"
)

func write(path string, contents string) {
	os.MkdirAll(filepath.Dir(path), 0o755)
	os.WriteFile(path, []byte(contents), 0o644)
}

func TestDoclintBasic(t *testing.T) {
	root := t.TempDir()

	// Create repo structure
	write(filepath.Join(root, "doclint.ini"), `
[scope]
roots = doc, work
include = **/*.md
exclude = work/backlog.md

[contracts]
require_header = true

[tags]
valid = NOW, PLAN, HIST, WHY, PHIL, RULE
`)

	// Valid file
	write(filepath.Join(root, "doc/design.md"), `
<!-- contract
allowed: NOW HIST
-->
NOW valid-tag
HIST good
`)

	// Missing header → violation
	write(filepath.Join(root, "doc/bad1.md"), `
NOW fail-here
`)

	// Invalid tag → violation
	write(filepath.Join(root, "doc/bad2.md"), `
<!-- contract
allowed: HIST
-->
PLAN notallowed
`)

	// Excluded file → ignored even if malformed
	write(filepath.Join(root, "work/backlog.md"), `
HIST bogus
NOW bogus
`)

	cfg, err := doclint.LoadConfig(root)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	violations, err := doclint.Check(cfg)
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}

	// Expect exactly TWO violations:
	// 1. missing header
	// 2. tag not allowed
	if len(violations) != 2 {
		t.Fatalf("expected 2 violations, got %d: %+v", len(violations), violations)
	}

	paths := []string{violations[0].Path, violations[1].Path}
	want := map[string]bool{
		"doc/bad1.md": true,
		"doc/bad2.md": true,
	}

	for _, p := range paths {
		if !want[p] {
			t.Fatalf("unexpected violation at %s (want doc/bad1.md or doc/bad2.md)", p)
		}
	}
}

func TestDoclintFindConfigRoot(t *testing.T) {
	root := t.TempDir()

	// repo root config
	write(filepath.Join(root, "doclint.ini"), `
[scope]
roots = doc
include = **/*.md

[contracts]
require_header = true

[tags]
valid = NOW HIST PLAN WHY PHIL RULE
`)

	// scope root directory declared in config
	if err := os.MkdirAll(filepath.Join(root, "doc"), 0o755); err != nil {
		t.Fatalf("mkdir doc: %v", err)
	}

	// deep directory where we start the search
	deep := filepath.Join(root, "a", "b", "c")
	if err := os.MkdirAll(deep, 0o755); err != nil {
		t.Fatalf("mkdir deep: %v", err)
	}

	// file under the declared root
	write(filepath.Join(root, "doc", "test.md"), `
<!-- contract
allowed: HIST
-->
HIST ok
`)

	cfg, err := doclint.LoadConfig(deep)
	if err != nil {
		t.Fatalf("LoadConfig failed from deep: %v", err)
	}

	if cfg.Root != root {
		t.Fatalf("expected config root %s, got %s", root, cfg.Root)
	}

	violations, err := doclint.Check(cfg)
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}

	if len(violations) != 0 {
		t.Fatalf("unexpected violations: %+v", violations)
	}
}
