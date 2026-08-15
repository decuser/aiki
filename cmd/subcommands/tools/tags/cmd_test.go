package tags

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTagsWritesTopLevelDefinitions(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "sample.ai")
	if err := os.WriteFile(p, []byte("let outer = 1\nlet f = (x) { let y = x; return y }\n"), 0644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(d, "tags")
	if code := Run([]string{"-o", out, p}); code != 0 {
		t.Fatalf("code=%d", code)
	}
	b, _ := os.ReadFile(out)
	s := string(b)
	if !strings.Contains(s, "outer\t"+p+"\t1;\"") || !strings.Contains(s, "f\t"+p+"\t2;\"") || strings.Contains(s, "\ny\t") {
		t.Fatalf("tags=%q", s)
	}
}
