//go:build rigorous

package conformance

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// TestSelfhostBehaviorAcceptanceCorpus runs a compact, language-owned behavior
// corpus through the independent interpreter. The selected cases span core
// evaluation, recoverable errors/patterns, qualified pipelines, self-hosted
// modules (including HAL-backed modules), relative imports, Unicode positions,
// and file effects. Concurrency/debug/graphics remain explicitly out of scope
// for the self-host proof.
func TestSelfhostBehaviorAcceptanceCorpus(t *testing.T) {
	root := distributionRoot(t)
	exe := buildAiki(t)
	cases := []string{
		"math_smoke.ai",
		"match_smoke.ai",
		"pipeline_smoke.ai",
		"string_smoke.ai",
		"regex_rune_positions_smoke.ai",
		"import_smoke.ai",
		"file_smoke.ai",
		"bytes_pure_smoke.ai",
		"hash_native_smoke.ai",
	}

	var harness strings.Builder
	harness.WriteString("let bootstrap = import(\"selfhost/bootstrap\")\n")
	for i, name := range cases {
		sourcePath := filepath.Join(root, "test", "behavior", name)
		source, err := os.ReadFile(sourcePath)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		begin := "<<<AIKI-SELFHOST:BEGIN:" + name + ">>>"
		end := "<<<AIKI-SELFHOST:END:" + name + ">>>"
		fmt.Fprintf(&harness, "println(%s)\n", strconv.Quote(begin))
		fmt.Fprintf(&harness, "let result%d = bootstrap.run(%s, %s)\n", i, strconv.Quote(string(source)), strconv.Quote(filepath.ToSlash(filepath.Join("test", "behavior", name))))
		fmt.Fprintf(&harness, "if equal(shape(result%d), :self_fault) { println(\"<<<AIKI-SELFHOST:FAULT>>>\" + inspect(result%d)) }\n", i, i)
		fmt.Fprintf(&harness, "println(%s)\n", strconv.Quote(end))
	}

	stdout, stderr, exitCode, err := runSelfhostAcceptance(exe, harness.String(), root)
	if err != nil {
		t.Fatalf("run acceptance harness: %v", err)
	}
	if exitCode != 0 {
		t.Fatalf("acceptance harness exited %d\nstderr: %s", exitCode, stderr)
	}

	for _, name := range cases {
		begin := "<<<AIKI-SELFHOST:BEGIN:" + name + ">>>\n"
		end := "<<<AIKI-SELFHOST:END:" + name + ">>>\n"
		start := strings.Index(stdout, begin)
		if start < 0 {
			t.Errorf("%s: begin marker missing", name)
			continue
		}
		start += len(begin)
		finish := strings.Index(stdout[start:], end)
		if finish < 0 {
			t.Errorf("%s: end marker missing", name)
			continue
		}
		got := stdout[start : start+finish]
		want := behaviorGoldStdout(t, filepath.Join(root, "test", "behavior", strings.TrimSuffix(name, ".ai")+".gold"))
		if got != want {
			t.Errorf("%s: output differs\nwant:\n%s\n\ngot:\n%s", name, want, got)
		}
	}
}

func behaviorGoldStdout(t *testing.T, path string) string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open gold %s: %v", filepath.Base(path), err)
	}
	defer f.Close()
	var out strings.Builder
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "OUT:") {
			continue
		}
		decoded, err := strconv.Unquote("\"" + strings.TrimPrefix(line, "OUT:") + "\"")
		if err != nil {
			t.Fatalf("decode gold %s line %q: %v", filepath.Base(path), line, err)
		}
		out.WriteString(decoded)
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan gold %s: %v", filepath.Base(path), err)
	}
	return out.String()
}

func runSelfhostAcceptance(exe, source, workDir string) (stdout, stderr string, exitCode int, err error) {
	tmp, err := os.CreateTemp("", "aiki-selfhost-acceptance-*.ai")
	if err != nil {
		return "", "", 1, err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.WriteString(source); err != nil {
		tmp.Close()
		return "", "", 1, err
	}
	tmp.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, exe, tmp.Name())
	cmd.Dir = workDir
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	runErr := cmd.Run()
	stdout, stderr = outBuf.String(), errBuf.String()
	if ctx.Err() == context.DeadlineExceeded {
		return stdout, stderr, 1, fmt.Errorf("self-host acceptance timed out")
	}
	if runErr != nil {
		if ee, ok := runErr.(*exec.ExitError); ok {
			return stdout, stderr, ee.ExitCode(), nil
		}
		return stdout, stderr, 1, runErr
	}
	return stdout, stderr, 0, nil
}
