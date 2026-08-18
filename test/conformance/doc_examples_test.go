package conformance

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"aiki/test/support/docexamples"
)

func TestDocExamplesExecutable(t *testing.T) {
	exe := buildAiki(t)
	examples, err := docexamples.Load(distributionRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(examples) == 0 {
		t.Fatal("no doc examples found")
	}

	checked, uncheckedCount, skipped := 0, 0, 0
	for _, ex := range examples {
		if len(ex.Code) == 0 && !ex.Unchecked {
			skipped++
			continue
		}
		var program strings.Builder
		if ex.Preamble != "" {
			program.WriteString(ex.Preamble)
			program.WriteString("\n")
		}
		if !ex.Unchecked {
			for _, line := range ex.Code {
				program.WriteString(line)
				program.WriteString("\n")
			}
		}

		absRoot, _ := filepath.Abs(distributionRoot(t))
		stdout, stderr, exitCode, err := runAikiSource(exe, program.String(), absRoot)
		if err != nil {
			t.Errorf("%s.%s: run error: %v", ex.Module, ex.Name, err)
			continue
		}
		if exitCode != 0 {
			t.Errorf("%s.%s: exited %d\nstderr: %s\nprogram:\n%s", ex.Module, ex.Name, exitCode, stderr, program.String())
			continue
		}
		if ex.Unchecked {
			uncheckedCount++
			continue
		}
		if len(ex.Expects) == 0 {
			skipped++
			continue
		}
		gotLines := strings.Split(strings.TrimRight(stdout, "\n"), "\n")
		if len(gotLines) == 1 && gotLines[0] == "" {
			gotLines = nil
		}
		if len(gotLines) != len(ex.Expects) {
			t.Errorf("%s.%s: expected %d outputs, got %d\n  expected: %v\n  got:      %v\n  stderr:   %s\n  program:\n%s",
				ex.Module, ex.Name, len(ex.Expects), len(gotLines), ex.Expects, gotLines, stderr, program.String())
			continue
		}
		for i, want := range ex.Expects {
			if gotLines[i] != want {
				t.Errorf("%s.%s: output %d: expected %q, got %q\n  program:\n%s", ex.Module, ex.Name, i+1, want, gotLines[i], program.String())
			}
		}
		checked++
	}
	t.Logf("doc examples: %d checked, %d unchecked, %d skipped (no code/expects)", checked, uncheckedCount, skipped)
}

func buildAiki(t *testing.T) string {
	t.Helper()
	absRoot, err := filepath.Abs(distributionRoot(t))
	if err != nil {
		t.Fatalf("abs root: %v", err)
	}
	exe := filepath.Join(t.TempDir(), "aiki-test-bin")
	cmd := exec.Command("go", "build", "-buildvcs=false", "-o", exe, "./cmd/aiki")
	cmd.Dir = absRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("building aiki: %v\n%s", err, out)
	}
	return exe
}

func runAikiSource(exe, source, workDir string) (stdout, stderr string, exitCode int, err error) {
	tmp, err := os.CreateTemp("", "aiki-doc-*.ai")
	if err != nil {
		return "", "", 1, err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.WriteString(source); err != nil {
		tmp.Close()
		return "", "", 1, err
	}
	tmp.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, exe, tmp.Name())
	cmd.Dir = workDir
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	runErr := cmd.Run()
	stdout, stderr = outBuf.String(), errBuf.String()
	if runErr != nil {
		if ee, ok := runErr.(*exec.ExitError); ok {
			return stdout, stderr, ee.ExitCode(), nil
		}
		return stdout, stderr, 1, runErr
	}
	return stdout, stderr, 0, nil
}
