package runner

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"aiki/engine/runtime/hal"
	"aiki/engine/runtime/hal/substrate"
	"aiki/engine/semantics/evaluator"
	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
	"aiki/engine/syntax/grammar"
)

// Smoke runs smoke tests on the given paths.
func Smoke(args []string) int {
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
			fmt.Fprintln(os.Stderr, " ", err)
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

	stdin, expOut, expErr, expExit, _, err := loadTranscript(goldPath)
	if err != nil {
		return fmt.Errorf("load transcript: %w", err)
	}

	stdout, stderr, exitCode := runEngineFile(aiPath, stdin)

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

	return nil
}

func unescape(s string) string {
	u, err := strconv.Unquote(`"` + s + `"`)
	if err != nil {
		return s
	}
	return u
}

func loadTranscript(path string) (stdin, stdout, stderr []byte, expExit int, displays []string, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, nil, 0, nil, err
	}
	var inBuf, outBuf, errBuf bytes.Buffer
	expExit = 0
	displays = nil

	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "IN:"):
			inBuf.WriteString(unescape(strings.TrimPrefix(line, "IN:")))
		case strings.HasPrefix(line, "OUT:"):
			outBuf.WriteString(unescape(strings.TrimPrefix(line, "OUT:")))
		case strings.HasPrefix(line, "ERR:"):
			errBuf.WriteString(unescape(strings.TrimPrefix(line, "ERR:")))
		case strings.HasPrefix(line, "EXIT:"):
			n, e := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(line, "EXIT:")))
			if e != nil {
				return nil, nil, nil, 0, nil, fmt.Errorf("bad EXIT line %q: %w", line, e)
			}
			expExit = n
		case strings.HasPrefix(line, "DISPLAY:"):
			displays = append(displays, strings.TrimPrefix(line, "DISPLAY:"))
		case strings.TrimSpace(line) == "":
			continue
		default:
			return nil, nil, nil, 0, nil, fmt.Errorf("invalid transcript line: %q", line)
		}
	}

	return inBuf.Bytes(), outBuf.Bytes(), errBuf.Bytes(), expExit, displays, nil
}

func runEngineFile(path string, stdin []byte) (stdout []byte, stderr []byte, exitCode int) {
	source, err := os.ReadFile(path)
	if err != nil {
		return nil, []byte(err.Error()), 1
	}

	// Capture stdout/stderr, redirect stdin
	var outBuf, errBuf bytes.Buffer
	oldStdout := substrate.Stdout
	oldStdin := substrate.Stdin
	substrate.Stdout = &outBuf
	substrate.Stdin = bytes.NewReader(stdin) // stdin may be empty, that's fine - returns EOF
	defer func() {
		substrate.Stdout = oldStdout
		substrate.Stdin = oldStdin
	}()

	// Load grammar
	g, err := grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
	if err != nil {
		return nil, []byte(err.Error()), 1
	}

	// Create runtime and environment
	rt := substrate.NewGoRuntime()
	env := value.NewEnv()

	// Load prelude with ScopePrelude
	if err := loadPrelude(g, rt, env); err != nil {
		errBuf.WriteString(err.Error())
		return outBuf.Bytes(), errBuf.Bytes(), 1
	}

	// Lex user code
	lexer := syntax.NewLexer(g, path, string(source), nil)
	tokens, err := lexer.Tokenize()
	if err != nil {
		return nil, []byte(err.Error()), 1
	}

	// Parse user code
	parser := syntax.NewParser(g, tokens, string(source), nil)
	ast, err := parser.Parse()
	if err != nil {
		return nil, []byte(err.Error()), 1
	}

	// Eval user code with ScopeUser
	env.SetFile(path)
	env.SetSource(string(source))
	ev := evaluator.NewWithScope(rt, nil, hal.ScopeUser)
	result := ev.Eval(ast, env)
	if errVal, ok := result.(*value.Error); ok {
		errBuf.WriteString(errVal.Message)
		return outBuf.Bytes(), errBuf.Bytes(), 1
	}

	return outBuf.Bytes(), errBuf.Bytes(), 0
}
