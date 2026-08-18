package runner

import (
	"errors"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/gofrs/flock"
)

func TestProgramExitHaltsNestedEvaluationAndCarriesStatus(t *testing.T) {
	withRepoRoot(t)
	err := RunSource("exit.ai", `
let system = import("system")
let stop = () {
    system.exit(7)
    undefined_after_exit()
}
stop()
undefined_after_exit()
`)
	var exitErr *ExitStatusError
	if !errors.As(err, &exitErr) {
		t.Fatalf("RunSource error = %v, want ExitStatusError", err)
	}
	if exitErr.Code != 7 {
		t.Fatalf("exit status = %d, want 7", exitErr.Code)
	}
}

func TestProgramExitZeroHaltsSuccessfully(t *testing.T) {
	withRepoRoot(t)
	if err := RunSource("exit-zero.ai", `
let system = import("system")
system.exit(0)
undefined_after_exit()
`); err != nil {
		t.Fatalf("RunSource: %v", err)
	}
}

func TestProgramExitRunsRuntimeCleanupBeforeReturningStatus(t *testing.T) {
	withRepoRoot(t)
	path := filepath.Join(t.TempDir(), "exit-cleanup.lock")
	source := fmt.Sprintf(`
let file = import("file")
let system = import("system")
let held = file.lock(%q)
system.exit(7)
`, path)

	err := RunSource("exit-cleanup.ai", source)
	var exitErr *ExitStatusError
	if !errors.As(err, &exitErr) || exitErr.Code != 7 {
		t.Fatalf("RunSource error = %v, want exit status 7", err)
	}

	probe := flock.New(path)
	locked, err := probe.TryLock()
	if err != nil {
		t.Fatalf("probe lock after exit cleanup: %v", err)
	}
	if !locked {
		t.Fatal("file lock remained held after system.exit runtime cleanup")
	}
	if err := probe.Unlock(); err != nil {
		t.Fatalf("unlock probe: %v", err)
	}
}
