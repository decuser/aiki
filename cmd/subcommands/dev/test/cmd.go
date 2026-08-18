// Package test implements the `aiki test` subcommand.
//
// Usage:
//
//	aiki test ./...
//	aiki test lib/math
//	aiki test path/to/foo_test.ai
//	aiki test -cover ./...
//	aiki test -coverprofile=coverage.out ./...
package test

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"aiki/engine/runner"
	"aiki/engine/runtime/hal/substrate"
	"aiki/engine/semantics/evaluator"
)

// Run is the entry point for the `aiki test` subcommand.
func Run(args []string) int {
	fs := flag.NewFlagSet("test", flag.ExitOnError)
	cover := fs.Bool("cover", false, "enable coverage analysis")
	coverprofile := fs.String("coverprofile", "", "write coverage profile to file")
	fs.Parse(args)

	if *coverprofile != "" {
		*cover = true
	}

	targets := fs.Args()
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

	// Merged coverage across all test files.
	var mergedCoverage map[string]int64

	for _, f := range files {
		rt := substrate.NewGoRuntime()
		rt.ResetTestState()
		rt.SetTestFile(f)

		var runErr error
		if *cover {
			counters := evaluator.NewCoverageCounters()
			counters, runErr = runner.RunWithCountersRuntime(f, counters, rt)
			cov := counters.Coverage()
			if mergedCoverage == nil {
				mergedCoverage = cov
			} else {
				for k, v := range cov {
					mergedCoverage[k] += v
				}
			}
		} else {
			runErr = runner.RunWithRuntime(f, rt)
		}

		if runErr != nil {
			rt.CloseAllResources()
			fmt.Fprintf(os.Stderr, "FAIL %s\n    %v\n", f, runErr)
			anyFailed = true
			totalFailed++
			continue
		}

		passed, failed, failures := rt.TestResults()
		rt.CloseAllResources()
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

	if *cover && mergedCoverage != nil {
		printCoverageSummary(mergedCoverage)
	}

	if *coverprofile != "" && mergedCoverage != nil {
		if err := writeCoverageProfile(*coverprofile, mergedCoverage); err != nil {
			fmt.Fprintf(os.Stderr, "aiki test: writing coverage profile: %v\n", err)
		}
	}

	if anyFailed {
		return 1
	}
	return 0
}

// printCoverageSummary groups coverage hits by file and prints a percentage
// based on the ratio of hit lines to total source lines.
func printCoverageSummary(coverage map[string]int64) {
	// Group by file.
	fileHits := map[string]map[int]bool{}
	for key := range coverage {
		file, line := parseCoverKey(key)
		if file == "" || file == "<prelude>" {
			continue
		}
		if fileHits[file] == nil {
			fileHits[file] = map[int]bool{}
		}
		fileHits[file][line] = true
	}

	// Sort files for stable output.
	var files []string
	for f := range fileHits {
		files = append(files, f)
	}
	sort.Strings(files)

	fmt.Fprintln(os.Stdout, "")
	totalHit := 0
	totalLines := 0
	for _, f := range files {
		// Skip prelude and test infrastructure.
		if strings.Contains(f, "prelude") || strings.HasSuffix(f, "_test.ai") {
			continue
		}
		lines := countSourceLines(f)
		if lines == 0 {
			continue
		}
		hit := len(fileHits[f])
		pct := float64(hit) / float64(lines) * 100
		fmt.Fprintf(os.Stdout, "coverage: %s\t%.1f%% (%d/%d lines)\n", f, pct, hit, lines)
		totalHit += hit
		totalLines += lines
	}
	if totalLines > 0 {
		pct := float64(totalHit) / float64(totalLines) * 100
		fmt.Fprintf(os.Stdout, "coverage: total\t\t%.1f%% (%d/%d lines)\n", pct, totalHit, totalLines)
	}
}

func parseCoverKey(key string) (string, int) {
	idx := strings.LastIndex(key, ":")
	if idx < 0 {
		return "", 0
	}
	file := key[:idx]
	line := 0
	for _, c := range key[idx+1:] {
		if c >= '0' && c <= '9' {
			line = line*10 + int(c-'0')
		}
	}
	return file, line
}

func countSourceLines(path string) int {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	lines := 0
	for _, b := range data {
		if b == '\n' {
			lines++
		}
	}
	if len(data) > 0 && data[len(data)-1] != '\n' {
		lines++
	}
	return lines
}

func writeCoverageProfile(path string, coverage map[string]int64) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintln(f, "mode: count")

	// Sort keys for stable output.
	var keys []string
	for k := range coverage {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		file, line := parseCoverKey(key)
		if file == "" || file == "<prelude>" || strings.HasSuffix(file, "_test.ai") {
			continue
		}
		count := coverage[key]
		// Go coverprofile format: file:startline.col,endline.col statements count
		// We approximate with line-level granularity.
		fmt.Fprintf(f, "%s:%d.1,%d.1 1 %d\n", file, line, line, count)
	}

	return nil
}

func discoverTests(targets []string) ([]string, error) {
	var files []string
	for _, target := range targets {
		if target == "./..." {
			target = "."
		}
		if strings.HasSuffix(target, "_test.ai") {
			if _, err := os.Stat(target); err != nil {
				return nil, fmt.Errorf("cannot find %s: %w", target, err)
			}
			files = append(files, target)
			continue
		}
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
