package experiment

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const usage = "usage: aiki experiment new <name>"

var experimentDirRE = regexp.MustCompile(`^([0-9]{3})-(.+)$`)

// Run implements the experiment scaffolding command.
func Run(args []string) int {
	if len(args) < 2 || args[0] != "new" {
		fmt.Fprintln(os.Stderr, usage)
		return 2
	}

	name := strings.TrimSpace(strings.Join(args[1:], " "))
	if name == "" {
		fmt.Fprintln(os.Stderr, "experiment: name must not be empty")
		return 2
	}

	root, err := distributionRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "experiment: %v\n", err)
		return 1
	}
	destination, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "experiment: current directory: %v\n", err)
		return 1
	}

	created, err := Create(root, destination, name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "experiment: %v\n", err)
		return 1
	}
	fmt.Printf("created %s\n", created)
	fmt.Println("  README.md")
	fmt.Println("  experiment/PROCEDURE.md")
	fmt.Println("  experiment/run.sh")
	fmt.Println("  results/")
	fmt.Println("  analyses/")
	return 0
}

func distributionRoot() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("locating running executable: %w", err)
	}
	return distributionRootFromExecutable(exe)
}

func distributionRootFromExecutable(exe string) (string, error) {
	if resolved, resolveErr := filepath.EvalSymlinks(exe); resolveErr == nil {
		exe = resolved
	}
	root := filepath.Dir(exe)
	info, err := os.Stat(filepath.Join(root, "experiments"))
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("distribution experiments directory not found beside %s", exe)
		}
		return "", fmt.Errorf("checking distribution experiments directory: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("distribution experiments path is not a directory: %s", filepath.Join(root, "experiments"))
	}
	return root, nil
}

// Create creates the next numbered experiment in destination. Numbering is
// derived only from distributionRoot/experiments; destination is never used as
// sequence authority.
func Create(distributionRoot, destination, name string) (string, error) {
	slug := slugify(name)
	if slug == "" {
		return "", fmt.Errorf("name %q does not contain any usable ASCII letters or digits", name)
	}

	number, err := nextNumber(filepath.Join(distributionRoot, "experiments"))
	if err != nil {
		return "", err
	}
	base := fmt.Sprintf("%03d-%s", number, slug)
	target := filepath.Join(destination, base)
	if _, err := os.Stat(target); err == nil {
		return "", fmt.Errorf("target already exists: %s", target)
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("checking target %s: %w", target, err)
	}

	if err := os.Mkdir(target, 0o755); err != nil {
		return "", fmt.Errorf("creating %s: %w", target, err)
	}
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.RemoveAll(target)
		}
	}()

	title := titleFromName(name)
	readme := renderREADME(number, title, base)
	if err := os.WriteFile(filepath.Join(target, "README.md"), []byte(readme), 0o644); err != nil {
		return "", fmt.Errorf("writing README.md: %w", err)
	}
	for _, dir := range []string{"experiment", "results", "analyses"} {
		if err := os.Mkdir(filepath.Join(target, dir), 0o755); err != nil {
			return "", fmt.Errorf("creating %s/: %w", dir, err)
		}
	}
	procedure := renderProcedure(number, title)
	if err := os.WriteFile(filepath.Join(target, "experiment", "PROCEDURE.md"), []byte(procedure), 0o644); err != nil {
		return "", fmt.Errorf("writing experiment/PROCEDURE.md: %w", err)
	}
	if err := os.WriteFile(filepath.Join(target, "experiment", "run.sh"), []byte(runScript), 0o755); err != nil {
		return "", fmt.Errorf("writing experiment/run.sh: %w", err)
	}

	cleanup = false
	return base, nil
}

func nextNumber(experimentsDir string) (int, error) {
	entries, err := os.ReadDir(experimentsDir)
	if err != nil {
		return 0, fmt.Errorf("reading distribution experiments %s: %w", experimentsDir, err)
	}

	seen := make(map[int]string)
	var numbers []int
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		match := experimentDirRE.FindStringSubmatch(entry.Name())
		if match == nil {
			continue
		}
		n, err := strconv.Atoi(match[1])
		if err != nil || n == 0 {
			continue
		}
		if prior, ok := seen[n]; ok {
			return 0, fmt.Errorf("duplicate experiment number %03d: %s and %s", n, prior, entry.Name())
		}
		seen[n] = entry.Name()
		numbers = append(numbers, n)
	}
	if len(numbers) == 0 {
		return 1, nil
	}
	sort.Ints(numbers)
	if numbers[len(numbers)-1] >= 999 {
		return 0, fmt.Errorf("experiment sequence exhausted at 999")
	}
	return numbers[len(numbers)-1] + 1, nil
}

func slugify(name string) string {
	var b strings.Builder
	separator := false
	for _, r := range strings.ToLower(strings.TrimSpace(name)) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			if separator && b.Len() > 0 {
				b.WriteByte('-')
			}
			separator = false
			b.WriteRune(r)
		case r == ' ', r == '\t', r == '\n', r == '_', r == '-':
			separator = true
		default:
			// Punctuation is discarded; if it separates two alphanumeric runs,
			// preserve a single word boundary in the slug.
			separator = true
		}
	}
	return strings.Trim(b.String(), "-")
}

func titleFromName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "Untitled"
	}
	return name
}

func renderREADME(number int, title, base string) string {
	return fmt.Sprintf(`# Experiment %03d — %s

This directory is a reproducible Aiki experiment.

## Layout

- `+"`experiment/`"+` contains the procedure, runner, and experimental materials.
- `+"`results/`"+` contains raw observations produced by runs.
- `+"`analyses/`"+` contains interpretations and subsequent analyses of those observations.

Start with `+"`experiment/PROCEDURE.md`"+`, then run:

`+"```sh"+`
cd experiment
./run.sh
`+"```"+`

The runner records its transcript in `+"`../results/`"+` while also displaying it on standard output.
`, number, title)
}

func renderProcedure(number int, title string) string {
	return fmt.Sprintf(`# Procedure — Experiment %03d: %s

## Question

What does this experiment investigate?

## Rationale

Why is the question worth examining?

## Materials

Describe the Aiki programs and other inputs kept in this directory.

## Procedure

Run from this directory:

`+"```sh"+`
./run.sh
`+"```"+`

The runner uses `+"`aiki`"+` from `+"`PATH`"+`, prints the executable and version under examination, and records the complete transcript in `+"`../results/`"+`.

## Measurements

Record what is measured and how.

## Expected relationships

State exact or qualitative expectations before observing the results.

## Caveats

Record machine-dependent, implementation-dependent, and methodological limitations.
`, number, title)
}

const runScript = `#!/bin/sh
set -eu

HERE=$(CDPATH= cd -- "$(dirname "$0")" && pwd)
RESULTS="$HERE/../results"
mkdir -p "$RESULTS"
cd "$HERE"

if ! command -v aiki >/dev/null 2>&1; then
    echo "experiment: aiki not found on PATH" >&2
    exit 1
fi

STAMP=$(date '+%Y-%m-%d-%H%M%S.%3N')
RESULT="$RESULTS/run-$STAMP.txt"

run_experiment() {
    printf 'Aiki executable: %s\n' "$(command -v aiki)"
    printf 'Aiki version: '
    aiki -v
    printf '\n'

    # Add experiment commands below.
}

run_experiment 2>&1 | tee "$RESULT"
printf '\nresult: %s\n' "$RESULT"
`
