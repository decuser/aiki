package canary

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"
)

// TestSubcommandDebugExists verifies the debug subcommand runs and produces output.
func TestSubcommandDebugExists(t *testing.T) {
	// Running without args should print usage to stderr
	cmd := exec.Command("go", "run", "../../cmd/aiki", "debug")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Run() // ignore exit code, we just want to see it ran

	output := stderr.String()
	if !strings.Contains(output, "usage:") && !strings.Contains(output, "debug") {
		t.Errorf("debug subcommand produced unexpected output: %q", output)
	}
}

// TestSubcommandFmtExists verifies the fmt subcommand runs.
func TestSubcommandFmtExists(t *testing.T) {
	cmd := exec.Command("go", "run", "../../cmd/aiki", "fmt")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Run()

	output := stderr.String()
	if !strings.Contains(output, "usage:") && !strings.Contains(output, "fmt") {
		t.Errorf("fmt subcommand produced unexpected output: %q", output)
	}
}

// TestSubcommandLintExists verifies the lint subcommand runs.
func TestSubcommandLintExists(t *testing.T) {
	cmd := exec.Command("go", "run", "../../cmd/aiki", "lint")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Run()

	output := stderr.String()
	if !strings.Contains(output, "usage:") && !strings.Contains(output, "lint") {
		t.Errorf("lint subcommand produced unexpected output: %q", output)
	}
}

// TestSubcommandTreecheckExists verifies the treecheck subcommand runs.
func TestSubcommandTreecheckExists(t *testing.T) {
	cmd := exec.Command("go", "run", "../../cmd/aiki", "treecheck", "-root", "../..")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		t.Fatalf("treecheck subcommand failed: %v\nstdout: %s\nstderr: %s", err, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "treecheck ok") {
		t.Fatalf("treecheck subcommand produced unexpected output: %q", stdout.String())
	}
}

// TestSubcommandVersionExists verifies the version subcommand runs.
func TestSubcommandVersionExists(t *testing.T) {
	cmd := exec.Command("go", "run", "../../cmd/aiki", "version")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	err := cmd.Run()
	if err != nil {
		t.Fatalf("version subcommand failed: %v", err)
	}

	output := stdout.String()
	if output == "" {
		t.Error("version subcommand produced no output")
	}
}
