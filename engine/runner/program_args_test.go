package runner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunExposesProgramArguments(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "args.ai")
	out := filepath.Join(dir, "out.txt")

	source := `let system = import("system")
let file = import("file")
let args = system.args()
let f = file.open(args[0], :write)
file.write_text(f, args[1] + "|" + args[2])
file.close(f)
`
	if err := os.WriteFile(script, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Run(script, out, "alpha", "beta"); err != nil {
		t.Fatalf("Run: %v", err)
	}

	got, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "alpha|beta" {
		t.Fatalf("system.args output = %q, want %q", string(got), "alpha|beta")
	}
}
