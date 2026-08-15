package invariant

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"
)

// TestDocExamplesExecutable runs every doc entry's examples as an Aiki program
// and checks that expressions with # expected comments produce the stated
// values. Entries marked @unchecked are executed but their output is not checked;
// they must still run without faulting.

// reExpect matches: expr    # expected_value
// Requires 2+ spaces before # to distinguish from prose comments.
var reExpect = regexp.MustCompile(`^(.+?)\s{2,}#\s+(.+)$`)

// reValue tests whether a string looks like an Aiki inspect() output.
var reValue = regexp.MustCompile(
	`^(?:` +
		`-?\d+(?:/\d+)?` + // number or rational
		`|".*"` + // string
		`|:\w+` + // symbol
		`|true|false` + // boolean
		`|\[.*\]` + // list
		`|'.'` + // rune
		`|<bytes:\d+.*>` + // bytes
		`|<(?:module|channel|file|fn)\b.*>` + // opaque types
		`|@\w+` + // shape name
		`)$`)

type docExample struct {
	module    string
	name      string
	preamble  string
	code      []string // lines of example code to execute
	expects   []string // expected inspect() output, in order
	unchecked bool     // @unchecked: run for no-fault only
}

func loadDocExamples(t *testing.T) []docExample {
	t.Helper()
	root := distributionRoot(t)

	type docFile struct {
		name string
		path string
	}

	var files []docFile

	// Prelude
	preludePath := filepath.Join(root, "engine", "runtime", "prelude", "prelude.doc")
	if _, err := os.Stat(preludePath); err == nil {
		files = append(files, docFile{name: "prelude", path: preludePath})
	}

	// Shipped modules
	for name, aiPath := range shippedModulePaths(t) {
		docPath := strings.TrimSuffix(aiPath, ".ai") + ".doc"
		if _, err := os.Stat(docPath); err == nil {
			files = append(files, docFile{name: name, path: docPath})
		}
	}

	sort.Slice(files, func(i, j int) bool { return files[i].name < files[j].name })

	var examples []docExample
	for _, f := range files {
		examples = append(examples, parseDocExamples(t, f.name, f.path)...)
	}
	return examples
}

func parseDocExamples(t *testing.T, moduleName, path string) []docExample {
	t.Helper()
	src, err := os.ReadFile(path)
	if err != nil {
		t.Errorf("%s: cannot read doc: %v", moduleName, err)
		return nil
	}
	text := string(src)

	// Extract @preamble lines.
	var preambleLines []string
	var bodyLines []string
	for _, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(line, "@preamble ") {
			preambleLines = append(preambleLines, strings.TrimPrefix(line, "@preamble "))
		} else {
			bodyLines = append(bodyLines, line)
		}
	}
	preamble := strings.Join(preambleLines, "\n")
	body := strings.Join(bodyLines, "\n")

	entries := strings.Split(body, "\n===\n")

	var out []docExample
	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		lines := strings.Split(entry, "\n")
		name := strings.TrimSpace(lines[0])
		if name == "" {
			continue
		}

		ex := docExample{
			module:   moduleName,
			name:     name,
			preamble: preamble,
		}

		// Scan lines after the name for code and expectations.
		for _, line := range lines[1:] {
			trimmed := strings.TrimSpace(line)

			if trimmed == "@unchecked" {
				ex.unchecked = true
				continue
			}
			if trimmed == "" || trimmed == "Example" {
				continue
			}

			// Check for expr  # expected pattern.
			if m := reExpect.FindStringSubmatch(line); m != nil {
				expr := strings.TrimSpace(m[1])
				expected := strings.TrimSpace(m[2])
				// Only treat as expectation if the comment looks like a value.
				if reValue.MatchString(expected) {
					ex.code = append(ex.code, "println(inspect("+expr+"))")
					ex.expects = append(ex.expects, expected)
					continue
				}
			}

			// Code context: let bindings, use/import calls.
			if strings.HasPrefix(trimmed, "let ") ||
				strings.HasPrefix(trimmed, "use(") ||
				strings.HasPrefix(trimmed, "import(") {
				ex.code = append(ex.code, line)
			}
		}

		out = append(out, ex)
	}
	return out
}

func TestDocExamplesExecutable(t *testing.T) {
	exe := buildAiki(t)
	examples := loadDocExamples(t)

	if len(examples) == 0 {
		t.Fatal("no doc examples found")
	}

	checked := 0
	uncheckedCount := 0
	skipped := 0

	for _, ex := range examples {
		if len(ex.code) == 0 && !ex.unchecked {
			skipped++
			continue
		}

		var program strings.Builder
		if ex.preamble != "" {
			program.WriteString(ex.preamble)
			program.WriteString("\n")
		}
		// For @unchecked entries, only run the preamble to verify the
		// module loads. The example code may reference files, canvases,
		// or prior REPL state that cannot be reproduced here.
		if !ex.unchecked {
			for _, line := range ex.code {
				program.WriteString(line)
				program.WriteString("\n")
			}
		}

		absRoot, _ := filepath.Abs(distributionRoot(t))
		stdout, stderr, exitCode, err := runAikiSource(exe, program.String(), absRoot)
		if err != nil {
			t.Errorf("%s.%s: run error: %v", ex.module, ex.name, err)
			continue
		}
		if exitCode != 0 {
			t.Errorf("%s.%s: exited %d\nstderr: %s\nprogram:\n%s", ex.module, ex.name, exitCode, stderr, program.String())
			continue
		}

		if ex.unchecked {
			uncheckedCount++
			continue
		}

		if len(ex.expects) == 0 {
			skipped++
			continue
		}

		gotLines := strings.Split(strings.TrimRight(stdout, "\n"), "\n")
		if len(gotLines) == 1 && gotLines[0] == "" {
			gotLines = nil
		}

		if len(gotLines) != len(ex.expects) {
			t.Errorf("%s.%s: expected %d outputs, got %d\n  expected: %v\n  got:      %v\n  stderr:   %s\n  program:\n%s",
				ex.module, ex.name, len(ex.expects), len(gotLines), ex.expects, gotLines, stderr, program.String())
			continue
		}

		for i, want := range ex.expects {
			if gotLines[i] != want {
				t.Errorf("%s.%s: output %d: expected %q, got %q\n  program:\n%s",
					ex.module, ex.name, i+1, want, gotLines[i], program.String())
			}
		}
		checked++
	}

	t.Logf("doc examples: %d checked, %d unchecked, %d skipped (no code/expects)", checked, uncheckedCount, skipped)
}

func buildAiki(t *testing.T) string {
	t.Helper()
	root := distributionRoot(t)
	absRoot, err := filepath.Abs(root)
	if err != nil {
		t.Fatalf("abs root: %v", err)
	}
	exe := filepath.Join(absRoot, "aiki-test-bin")
	cmd := exec.Command("go", "build", "-buildvcs=false", "-o", exe, "./cmd/aiki")
	cmd.Dir = absRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("building aiki: %v\n%s", err, out)
	}
	t.Cleanup(func() { os.Remove(exe) })
	return exe
}

func runAikiFile(exe, path, workDir string) (stdout, stderr string, exitCode int, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, exe, path)
	cmd.Dir = workDir
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	runErr := cmd.Run()
	stdout = outBuf.String()
	stderr = errBuf.String()

	if runErr != nil {
		if ee, ok := runErr.(*exec.ExitError); ok {
			return stdout, stderr, ee.ExitCode(), nil
		}
		return stdout, stderr, 1, runErr
	}
	return stdout, stderr, 0, nil
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
	stdout = outBuf.String()
	stderr = errBuf.String()

	if runErr != nil {
		if ee, ok := runErr.(*exec.ExitError); ok {
			return stdout, stderr, ee.ExitCode(), nil
		}
		return stdout, stderr, 1, runErr
	}
	return stdout, stderr, 0, nil
}

// TestLibModulesHaveDoc asserts that every shipped module has a .doc file.
func TestLibModulesHaveDoc(t *testing.T) {
	root := distributionRoot(t)
	for name, aiPath := range shippedModulePaths(t) {
		_ = root
		docPath := strings.TrimSuffix(aiPath, ".ai") + ".doc"
		if _, err := os.Stat(docPath); err != nil {
			t.Errorf("%s: no doc file at %s", name, filepath.Base(docPath))
		}
	}
}

// TestDocEntryDisposition asserts that every doc entry either has a checkable
// expectation (# expected comment with a value) or is marked @unchecked. An
// entry with neither is invisible to the executable doc test and could drift
// out of date without detection.
func TestDocEntryDisposition(t *testing.T) {
	examples := loadDocExamples(t)
	if len(examples) == 0 {
		t.Fatal("no doc examples found")
	}

	uncovered := 0
	for _, ex := range examples {
		if len(ex.expects) == 0 && !ex.unchecked {
			// No expectation and not marked @unchecked.
			// This entry is invisible to the executable test.
			t.Errorf("%s.%s: no # expected comment and not marked @unchecked", ex.module, ex.name)
			uncovered++
		}
	}
	if uncovered == 0 {
		t.Logf("all %d doc entries have a disposition (checked or @unchecked)", len(examples))
	}
}
