package experiment

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestCreateUsesDistributionSequenceNotDestination(t *testing.T) {
	dist := t.TempDir()
	if err := os.Mkdir(filepath.Join(dist, "experiments"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"001-first", "004-later"} {
		if err := os.Mkdir(filepath.Join(dist, "experiments", name), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	destination := t.TempDir()
	// A destination-local number must not influence the distribution sequence.
	if err := os.Mkdir(filepath.Join(destination, "099-local-only"), 0o755); err != nil {
		t.Fatal(err)
	}

	created, err := Create(dist, destination, "Profiler Calibration")
	if err != nil {
		t.Fatal(err)
	}
	if created != "005-profiler-calibration" {
		t.Fatalf("created %q, want 005-profiler-calibration", created)
	}

	target := filepath.Join(destination, created)
	for _, dir := range []string{"experiment", "results", "analyses"} {
		if info, err := os.Stat(filepath.Join(target, dir)); err != nil {
			t.Fatal(err)
		} else if !info.IsDir() {
			t.Fatalf("%s is not a directory", dir)
		}
	}
	if info, err := os.Stat(filepath.Join(target, "experiment", "run.sh")); err != nil {
		t.Fatal(err)
	} else if info.Mode().Perm()&0o111 == 0 {
		t.Fatalf("experiment/run.sh is not executable: mode %v", info.Mode())
	}
	if _, err := os.Stat(filepath.Join(target, "experiment", "PROCEDURE.md")); err != nil {
		t.Fatal(err)
	}

	readme, err := os.ReadFile(filepath.Join(target, "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(readme), "# Experiment 005 — Profiler Calibration") {
		t.Fatalf("README title missing:\n%s", readme)
	}
}

func TestNextNumberRejectsDuplicateDistributionNumbers(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"007-one", "007-two"} {
		if err := os.Mkdir(filepath.Join(dir, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := nextNumber(dir); err == nil || !strings.Contains(err.Error(), "duplicate experiment number 007") {
		t.Fatalf("nextNumber error = %v, want duplicate-number error", err)
	}
}

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"Profiler Calibration": "profiler-calibration",
		"  Foo -- Bar!  ":      "foo-bar",
		"Aiki_2 / Scale":       "aiki-2-scale",
		"!!!":                  "",
	}
	for in, want := range cases {
		if got := slugify(in); got != want {
			t.Errorf("slugify(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestCreateStartsAtOneAndRefusesExistingTarget(t *testing.T) {
	dist := t.TempDir()
	if err := os.Mkdir(filepath.Join(dist, "experiments"), 0o755); err != nil {
		t.Fatal(err)
	}
	destination := t.TempDir()

	created, err := Create(dist, destination, "First")
	if err != nil {
		t.Fatal(err)
	}
	if created != "001-first" {
		t.Fatalf("created %q, want 001-first", created)
	}

	if _, err := Create(dist, destination, "First"); err == nil {
		t.Fatal("second create unexpectedly succeeded")
	}
}

func TestDistributionRootFollowsExecutableSymlink(t *testing.T) {
	dist := t.TempDir()
	if err := os.Mkdir(filepath.Join(dist, "experiments"), 0o755); err != nil {
		t.Fatal(err)
	}
	realExe := filepath.Join(dist, "aiki")
	if err := os.WriteFile(realExe, []byte("binary"), 0o755); err != nil {
		t.Fatal(err)
	}
	bin := t.TempDir()
	link := filepath.Join(bin, "aiki")
	if err := os.Symlink(realExe, link); err != nil {
		t.Fatal(err)
	}

	got, err := distributionRootFromExecutable(link)
	if err != nil {
		t.Fatal(err)
	}
	if got != dist {
		t.Fatalf("distribution root = %q, want %q", got, dist)
	}
}

func TestGeneratedRunnerLogsToResults(t *testing.T) {
	dist := t.TempDir()
	if err := os.Mkdir(filepath.Join(dist, "experiments"), 0o755); err != nil {
		t.Fatal(err)
	}
	destination := t.TempDir()
	created, err := Create(dist, destination, "Logging")
	if err != nil {
		t.Fatal(err)
	}

	bin := t.TempDir()
	fake := filepath.Join(bin, "aiki")
	if err := os.WriteFile(fake, []byte("#!/bin/sh\nif [ \"$1\" = \"-v\" ]; then echo 'aiki test-version'; exit 0; fi\necho unexpected >&2\nexit 2\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	runnerDir := filepath.Join(destination, created, "experiment")
	cmd := exec.Command("./run.sh")
	cmd.Dir = runnerDir
	cmd.Env = append(os.Environ(), "PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("run.sh failed: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "aiki test-version") {
		t.Fatalf("runner output missing version:\n%s", out)
	}

	entries, err := os.ReadDir(filepath.Join(destination, created, "results"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("results entries = %#v, want one timestamped run transcript", entries)
	}
	matched, err := regexp.MatchString(`^run-[0-9]{4}-[0-9]{2}-[0-9]{2}-[0-9]{6}\.[0-9]{3}\.txt$`, entries[0].Name())
	if err != nil {
		t.Fatal(err)
	}
	if !matched {
		t.Fatalf("result name = %q, want millisecond timestamp", entries[0].Name())
	}
	body, err := os.ReadFile(filepath.Join(destination, created, "results", entries[0].Name()))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "aiki test-version") {
		t.Fatalf("result transcript missing version:\n%s", body)
	}
}
