package conformance

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// TestSelfhostSelfInterpretation is the final Phase III bootstrap proof.
//
// The Go interpreter runs the blessed selfhost/bootstrap module. That Aiki
// interpreter then self-host-loads and executes a second copy of the Aiki
// bootstrap/interpreter, which evaluates a third-level Aiki program. The
// expected result exercises Aiki's left-to-right arithmetic semantics:
// 1 + 2 * 3 == 9.
func TestSelfhostSelfInterpretation(t *testing.T) {
	exe := buildAiki(t)

	source := `let bootstrap = import("selfhost/bootstrap")
let inner = "let bootstrap = import(\"selfhost/bootstrap\")\nbootstrap.run(\"1 + 2 * 3\\n\", \"third.ai\")\n"
let result = bootstrap.run(inner, "inner.ai")
println(inspect(result))
`

	tmp, err := os.CreateTemp("", "aiki-self-interpret-*.ai")
	if err != nil {
		t.Fatalf("create proof source: %v", err)
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.WriteString(source); err != nil {
		tmp.Close()
		t.Fatalf("write proof source: %v", err)
	}
	if err := tmp.Close(); err != nil {
		t.Fatalf("close proof source: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, exe, tmp.Name())
	// Run from outside the distribution tree. Named distribution imports must
	// resolve from the executable, while selfhost-internal path imports must
	// resolve from the importing file rather than accidentally from process CWD.
	cmd.Dir = t.TempDir()
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			t.Fatalf("self-interpretation exceeded 90s\nstdout:\n%s\nstderr:\n%s", stdout.String(), stderr.String())
		}
		t.Fatalf("self-interpretation failed: %v\nstdout:\n%s\nstderr:\n%s", err, stdout.String(), stderr.String())
	}
	if got := strings.TrimSpace(stdout.String()); got != "9" {
		t.Fatalf("self-interpretation result: want 9, got %q\nstderr:\n%s", got, stderr.String())
	}
}
