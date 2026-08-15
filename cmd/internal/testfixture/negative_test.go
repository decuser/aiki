package testfixture

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFixture(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "fixture_smoke.ai")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestNegativeKindOf(t *testing.T) {
	path := writeFixture(t, "# note\n# @negative parse\nlet x =\n")
	kind, err := NegativeKindOf(path)
	if err != nil {
		t.Fatal(err)
	}
	if kind != NegativeParse {
		t.Fatalf("got %q, want %q", kind, NegativeParse)
	}
}

func TestNegativeMarkerMustBeInHeader(t *testing.T) {
	path := writeFixture(t, "let x = 1\n# @negative parse\n")
	kind, err := NegativeKindOf(path)
	if err != nil {
		t.Fatal(err)
	}
	if kind != NegativeNone {
		t.Fatalf("late marker recognized as %q", kind)
	}
}

func TestNegativeUnknownKindFails(t *testing.T) {
	path := writeFixture(t, "# @negative runtime\nlet x = 1\n")
	if _, err := NegativeKindOf(path); err == nil {
		t.Fatal("expected unknown kind error")
	}
}

func TestNegativeMarkerRejectedOutsideSmokeFixture(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ordinary.ai")
	if err := os.WriteFile(path, []byte("# @negative parse\nlet x =\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := NegativeKindOf(path); err == nil {
		t.Fatal("expected marker outside *_smoke.ai to fail")
	}
}
