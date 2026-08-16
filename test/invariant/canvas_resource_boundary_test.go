package invariant

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCanvasSemanticValueIsOpaqueHandle(t *testing.T) {
	root := distributionRoot(t)
	path := filepath.Join(root, "engine", "semantics", "value", "canvas.go")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	s := string(b)
	for _, forbidden := range []string{
		"image/color", "sync.", "chan ", "CanvasCmd", "PenSize", "Turtle", "Commands", "Done", "Ready",
	} {
		if strings.Contains(s, forbidden) {
			t.Errorf("semantic Canvas value leaks substrate state %q", forbidden)
		}
	}
	if !strings.Contains(s, "ID uint64") {
		t.Errorf("semantic Canvas value must retain only opaque resource identity")
	}
}
