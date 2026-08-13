// Package test implements the `aiki test` subcommand.
//
// Usage:
//
//	aiki test ./...
//	aiki test lib/math
//	aiki test path/to/foo_test.ai
package test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"aiki/engine/runner"
	"aiki/engine/runtime/hal/substrate"
)

// Run is the entry point for the `aiki test` subcommand.
func Run(args []string) int {
	targets := args
	if len(targets) == 0 {
		targets = []string{"."}
	}

	files, err := discoverTests(targets)
	if err != nil {
		fmt.Fprintln(os.Stderr, "aiki test:", err)
		return 1
	}
	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "aiki test: no *_test.ai files found")
		return 1
	}

	totalPassed := 0
	totalFailed := 0
	anyFailed := false

	for _, f := range files {
		substrate.ResetTestState()
		substrate.SetTestFile(f)

		err := runner.Run(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "FAIL %s\n    %v\n", f, err)
			anyFailed = true
			totalFailed++
			continue
		}

		passed, failed, failures := substrate.TestResults()
		totalPassed += passed
		totalFailed += failed

		if failed > 0 {
			fmt.Fprintf(os.Stdout, "FAIL %s\n", f)
			for _, msg := range failures {
				fmt.Fprintln(os.Stdout, msg)
			}
			anyFailed = true
		} else {
			fmt.Fprintf(os.Stdout, "PASS %s\n", f)
		}
	}

	fmt.Fprintf(os.Stdout, "\n%d tests, %d passed, %d failed\n", totalPassed+totalFailed, totalPassed, totalFailed)

	if anyFailed {
		return 1
	}
	return 0
}

func discoverTests(targets []string) ([]string, error) {
	var files []string
	for _, target := range targets {
		if target == "./..." {
			target = "."
		}
		// If it's a specific file
		if strings.HasSuffix(target, "_test.ai") {
			if _, err := os.Stat(target); err != nil {
				return nil, fmt.Errorf("cannot find %s: %w", target, err)
			}
			files = append(files, target)
			continue
		}
		// Walk directory
		err := filepath.WalkDir(target, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if strings.HasSuffix(path, "_test.ai") {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return files, nil
}
