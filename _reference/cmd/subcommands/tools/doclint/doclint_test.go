package doclint

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckFile_Valid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "valid.md")

	content := `<!-- contract
allowed: NOW PLAN
-->

# Test

NOW this is current.
PLAN this is planned.
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	opts := Options{RequireHeader: true}
	violations, err := Check([]string{path}, opts)
	if err != nil {
		t.Fatal(err)
	}

	if len(violations) != 0 {
		t.Errorf("expected no violations, got %d: %v", len(violations), violations)
	}
}

func TestCheckFile_DisallowedTag(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "invalid.md")

	content := `<!-- contract
allowed: NOW
-->

# Test

NOW this is fine.
PLAN this is not allowed.
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	opts := Options{RequireHeader: true}
	violations, err := Check([]string{path}, opts)
	if err != nil {
		t.Fatal(err)
	}

	if len(violations) != 1 {
		t.Errorf("expected 1 violation, got %d", len(violations))
	}

	if len(violations) > 0 && violations[0].Message != "tag not allowed by contract: PLAN" {
		t.Errorf("unexpected message: %s", violations[0].Message)
	}
}

func TestCheckFile_MissingHeader(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "noheader.md")

	content := `# Test

Just some content.
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	opts := Options{RequireHeader: true}
	violations, err := Check([]string{path}, opts)
	if err != nil {
		t.Fatal(err)
	}

	// Files without headers are skipped, not violations
	if len(violations) != 0 {
		t.Errorf("expected no violations (skip), got %d: %v", len(violations), violations)
	}
}

func TestCheckFile_NoHeaderRequired(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "noheader.md")

	content := `# Test

Just some content.
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	opts := Options{RequireHeader: false}
	violations, err := Check([]string{path}, opts)
	if err != nil {
		t.Fatal(err)
	}

	if len(violations) != 0 {
		t.Errorf("expected no violations, got %d: %v", len(violations), violations)
	}
}

func TestCheckDir(t *testing.T) {
	dir := t.TempDir()

	if err := os.MkdirAll(filepath.Join(dir, "doc"), 0o755); err != nil {
		t.Fatal(err)
	}

	content := `<!-- contract
allowed: NOW
-->

NOW valid tag.
`
	if err := os.WriteFile(filepath.Join(dir, "doc", "test.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	opts := Options{RequireHeader: true}
	violations, err := Check([]string{filepath.Join(dir, "doc")}, opts)
	if err != nil {
		t.Fatal(err)
	}

	if len(violations) != 0 {
		t.Errorf("expected no violations, got %d: %v", len(violations), violations)
	}
}

func TestCheckRecursive(t *testing.T) {
	dir := t.TempDir()

	if err := os.MkdirAll(filepath.Join(dir, "sub", "deep"), 0o755); err != nil {
		t.Fatal(err)
	}

	content := `<!-- contract
allowed: NOW
-->

NOW valid.
`
	if err := os.WriteFile(filepath.Join(dir, "sub", "deep", "test.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	// Change to temp dir to test ./...
	oldWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldWd)

	opts := Options{RequireHeader: true}
	violations, err := Check([]string{"./..."}, opts)
	if err != nil {
		t.Fatal(err)
	}

	if len(violations) != 0 {
		t.Errorf("expected no violations, got %d: %v", len(violations), violations)
	}
}

func TestEmptyAllowedList(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.md")

	content := `<!-- contract
-->

NOW this will fail.
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	opts := Options{RequireHeader: true}
	violations, err := Check([]string{path}, opts)
	if err != nil {
		t.Fatal(err)
	}

	// Should get violation for empty allowed list
	if len(violations) < 1 {
		t.Errorf("expected at least 1 violation, got %d", len(violations))
	}
}
