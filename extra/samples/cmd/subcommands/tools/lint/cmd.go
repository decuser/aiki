package lint

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
	"aiki/engine/syntax/grammar"
)

// Run executes lint.
// Current implementation is format-first: fail if any file is not already
// in canonical fmt form.
//
// Flags:
//
//	-fmt     enable formatting check (default true)
func Run(args []string) int {
	fs := flag.NewFlagSet("lint", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	checkFmt := fs.Bool("fmt", true, "fail if fmt would change files")
	showWarnings := fs.Bool("w", true, "include warnings")
	showShadow := fs.Bool("shadow", false, "include shadowing warnings")

	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "usage: aiki lint <file.ai> [more files] | ./...")
		return 2
	}

	g, err := grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
	if err != nil {
		fmt.Fprintln(os.Stderr, "lint:", err)
		return 1
	}

	if *checkFmt {
		bad, err := CheckFormatting(fs.Args())
		if err != nil {
			fmt.Fprintln(os.Stderr, "lint:", err)
			return 1
		}
		if len(bad) > 0 {
			for _, p := range bad {
				fmt.Fprintln(os.Stdout, p)
			}
			return 1
		}
	}

	files, err := expandLintPaths(fs.Args())
	if err != nil {
		fmt.Fprintln(os.Stderr, "lint:", err)
		return 1
	}

	anyError := false
	for _, path := range files {
		src, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, "lint:", err)
			return 1
		}
		scope := value.ScopeUser
		if strings.HasSuffix(path, "engine/runtime/prelude/prelude.ai") {
			scope = value.ScopePrelude
		}
		diags, err := LintSource(g, path, string(src), scope)
		if err != nil {
			// Parse errors should already be rich; just print.
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		for _, d := range diags {
			if d.Level == "warning" && !*showWarnings {
				continue
			}
			if d.Level == "warning" && strings.HasPrefix(d.Message, "shadowing:") && !*showShadow {
				continue
			}
			if d.Level == "error" {
				anyError = true
			}
			fmt.Fprintln(os.Stderr, formatDiagnostic(path, string(src), d))
		}
	}

	if anyError {
		return 1
	}
	return 0
}

func formatDiagnostic(path string, source string, d Diagnostic) string {
	// Keep output consistent with engine parse error presentation.
	line := getSourceLine(source, d.Pos.Line)
	caret := strings.Repeat(" ", max(0, d.Pos.Col-1)) + "^"
	return fmt.Sprintf("%s:%d:%d:\n%s\n%s\n%s: %s", path, d.Pos.Line, d.Pos.Col, line, caret, d.Level, d.Message)
}

func getSourceLine(source string, line int) string {
	if line <= 0 {
		return ""
	}
	cur := 1
	start := 0
	for i := 0; i < len(source); i++ {
		if source[i] == '\n' {
			if cur == line {
				return strings.TrimRight(source[start:i], "\r")
			}
			cur++
			start = i + 1
		}
	}
	if cur == line {
		return strings.TrimRight(source[start:], "\r")
	}
	return ""
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
