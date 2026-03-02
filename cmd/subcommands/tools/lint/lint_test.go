package lint

import (
    "os"
    "path/filepath"
    "testing"
)

func TestLint_CheckFormatting(t *testing.T) {
    dir := t.TempDir()
    p1 := filepath.Join(dir, "a.ai")
    // Unformatted: missing spaces and brace on next line.
    if err := os.WriteFile(p1, []byte("let x=1\nif x\n{\nreturn x\n}\n"), 0644); err != nil {
        t.Fatal(err)
    }
    p2 := filepath.Join(dir, "b.ai")
    if err := os.WriteFile(p2, []byte("let y = 2\n"), 0644); err != nil {
        t.Fatal(err)
    }

    bad, err := CheckFormatting([]string{dir + "/..."})
    if err != nil {
        t.Fatalf("CheckFormatting: %v", err)
    }
    if len(bad) != 1 {
        t.Fatalf("expected 1 bad file, got %d: %v", len(bad), bad)
    }
    if bad[0] != p1 {
        t.Fatalf("expected bad file %s, got %s", p1, bad[0])
    }
}
