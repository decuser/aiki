package smoke

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func Run(args []string) int {
	targets, err := collectSmokeFiles(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, "smoke:", err)
		return 1
	}
	if len(targets) == 0 {
		fmt.Fprintln(os.Stderr, "smoke: no *_smoke.ai files found")
		return 1
	}

	for _, aiPath := range targets {
		if err := runPair(aiPath); err != nil {
			fmt.Fprintln(os.Stderr, "smoke FAIL:", aiPath)
			fmt.Fprintln(os.Stderr, "  ", err)
			return 1
		}
	}

	fmt.Fprintf(os.Stdout, "smoke ok (%d tests)\n", len(targets))
	return 0
}

func collectSmokeFiles(args []string) ([]string, error) {
	if len(args) == 0 {
		args = []string{"."}
	}

	var out []string

	for _, root := range args {
		if root == "./..." {
			root = "."
		}
		err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if strings.HasSuffix(path, "_smoke.ai") {
				out = append(out, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return out, nil
}

func runPair(aiPath string) error {
	goldPath := strings.TrimSuffix(aiPath, ".ai") + ".gold"

	stdin, expOut, expErr, expExit, expCanvas, displays, err := loadTranscript(goldPath)
	if err != nil {
		return fmt.Errorf("load transcript: %w", err)
	}

	// When the gold contains CANVAS: lines, run the program with the
	// recorder so the canvas command stream can be compared rather than
	// requiring a human to watch it.
	var canvasRecordPath string
	if len(expCanvas) > 0 {
		f, err := os.CreateTemp("", "aiki-canvas-*.transcript")
		if err != nil {
			return fmt.Errorf("create canvas transcript: %w", err)
		}
		canvasRecordPath = f.Name()
		f.Close()
		defer os.Remove(canvasRecordPath)
	}

	stdout, stderr, exitCode, err := runAikiFile(aiPath, stdin, canvasRecordPath)
	if err != nil {
		return fmt.Errorf("run error: %w", err)
	}

	// exit code
	if exitCode != expExit {
		return fmt.Errorf("exit mismatch: expected %d, got %d\nstderr:\n%s",
			expExit, exitCode, string(stderr))
	}

	// stderr
	stdErrTrim := bytes.TrimRight(stderr, "\n")
	expErrTrim := bytes.TrimRight(expErr, "\n")
	if !bytes.Equal(stdErrTrim, expErrTrim) {
		return fmt.Errorf("stderr mismatch\n--- expected ---\n%q\n--- got ---\n%q\n",
			string(expErr), string(stderr))
	}

	// stdout
	stdTrim := bytes.TrimRight(stdout, "\n")
	expTrim := bytes.TrimRight(expOut, "\n")
	if !bytes.Equal(stdTrim, expTrim) {
		return fmt.Errorf("output mismatch\n--- expected ---\n%q\n--- got ---\n%q\n",
			string(expOut), string(stdout))
	}

	// canvas transcript
	if len(expCanvas) > 0 {
		gotCanvas, err := os.ReadFile(canvasRecordPath)
		if err != nil {
			return fmt.Errorf("read canvas transcript: %w", err)
		}
		gotLines := strings.Split(strings.TrimRight(string(gotCanvas), "\n"), "\n")
		if len(gotLines) == 1 && gotLines[0] == "" {
			gotLines = nil
		}
		if !linesEqual(expCanvas, gotLines) {
			return fmt.Errorf("canvas transcript mismatch\n--- expected %d lines ---\n%s\n--- got %d lines ---\n%s\n",
				len(expCanvas), strings.Join(expCanvas, "\n"),
				len(gotLines), strings.Join(gotLines, "\n"))
		}
	}

	// DISPLAY prompts (interactive, for make visual only)
	for _, msg := range displays {
		fmt.Printf("DISPLAY: %s [y/N]: ", strings.TrimSpace(msg))
		var resp string
		fmt.Scanln(&resp)
		if strings.ToLower(strings.TrimSpace(resp)) != "y" {
			return fmt.Errorf("display check failed")
		}
	}

	return nil
}

func linesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func unescape(s string) string {
	u, err := strconv.Unquote(`"` + s + `"`)
	if err != nil {
		// fall back to raw if something is weird
		return s
	}
	return u
}

func loadTranscript(path string) (stdin, stdout, stderr []byte, expExit int, canvas []string, displays []string, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, nil, 0, nil, nil, err
	}
	var inBuf, outBuf, errBuf bytes.Buffer
	expExit = 0

	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "IN:"):
			inBuf.WriteString(unescape(strings.TrimPrefix(line, "IN:")))
		case strings.HasPrefix(line, "OUT:"):
			outBuf.WriteString(unescape(strings.TrimPrefix(line, "OUT:")))
		case strings.HasPrefix(line, "ERR:"):
			errBuf.WriteString(unescape(strings.TrimPrefix(line, "ERR:")))
		case strings.HasPrefix(line, "CANVAS:"):
			canvas = append(canvas, strings.TrimPrefix(line, "CANVAS:"))
		case strings.HasPrefix(line, "EXIT:"):
			n, e := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(line, "EXIT:")))
			if e != nil {
				return nil, nil, nil, 0, nil, nil, fmt.Errorf("bad EXIT line %q: %w", line, e)
			}
			expExit = n
		case strings.HasPrefix(line, "DISPLAY:"):
			displays = append(displays, strings.TrimPrefix(line, "DISPLAY:"))
		case strings.TrimSpace(line) == "":
			continue
		default:
			return nil, nil, nil, 0, nil, nil, fmt.Errorf("invalid transcript line: %q", line)
		}
	}

	return inBuf.Bytes(), outBuf.Bytes(), errBuf.Bytes(), expExit, canvas, displays, nil
}

func runAikiFile(path string, stdin []byte, canvasRecordPath string) (stdout []byte, stderr []byte, exitCode int, err error) {
	exe, err := os.Executable()
	if err != nil {
		return nil, nil, 1, fmt.Errorf("locate executable: %w", err)
	}

	cmd := exec.Command(exe, path)

	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}

	cmd.Env = os.Environ()
	if canvasRecordPath != "" {
		cmd.Env = append(cmd.Env, "AIKI_CANVAS=record", "AIKI_CANVAS_RECORD="+canvasRecordPath)
	}

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	runErr := cmd.Run()

	stdout = outBuf.Bytes()
	stderr = errBuf.Bytes()

	if runErr != nil {
		if ee, ok := runErr.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
			return stdout, stderr, exitCode, nil
		}
		return stdout, stderr, 1, runErr
	}

	return stdout, stderr, 0, nil
}
