package lint

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"aiki/engine/language"
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
	includePrelude := fs.Bool("prelude", false, "include prelude in lint walk")

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
		bad, err := CheckFormatting(fs.Args(), *includePrelude)
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
	if !*includePrelude {
		filtered := make([]string, 0, len(files))
		for _, p := range files {
			if strings.HasSuffix(p, "engine/runtime/prelude/prelude.ai") {
				continue
			}
			filtered = append(filtered, p)
		}
		files = filtered
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
			if !*includePrelude {
				continue
			}
		} else if language.HasPackageDeclaration(string(src)) {
			// Package modules are evaluated in an env enclosed by the prelude env,
			// inheriting ScopePrelude (see substrate import implementation).
			scope = value.ScopePrelude
		}
		diags, err := LintSource(g, path, string(src), scope)
		if err != nil {
			// Parse errors should already be rich; just print.
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		for _, d := range diags {
			if d.Severity == "warning" && !*showWarnings {
				continue
			}
			if d.Severity == "warning" && strings.HasPrefix(d.Message, "shadowing:") && !*showShadow {
				continue
			}
			if d.Severity == "error" {
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
	expandedLine, caret := expandTabsWithCaret(line, d.Pos.Col, 8)
	return fmt.Sprintf("%s:%d:%d:\n%s\n%s\n%s: %s", path, d.Pos.Line, d.Pos.Col, expandedLine, caret, d.Severity, d.Message)
}

// expandTabsWithCaret expands tabs to spaces using a fixed tab width and
// returns an expanded line plus a caret line aligned to the original column.
//
// The engine Position.Col is byte based (1 indexed) and counts a tab as one
// column. When we print the line, terminals expand tabs to tab stops, so a
// naive strings.Repeat(" ", col-1) caret drifts. This function renders a stable
// caret by expanding tabs and mapping the original column into visual columns.
func expandTabsWithCaret(line string, col int, tabWidth int) (string, string) {
	if tabWidth <= 0 {
		tabWidth = 8
	}
	if col < 1 {
		col = 1
	}
	// Clamp to line length + 1 to avoid huge caret runs.
	if col > len(line)+1 {
		col = len(line) + 1
	}

	// Compute visual column for caret by expanding tabs up to col-1.
	visualCol := 0
	for i := 0; i < col-1 && i < len(line); i++ {
		if line[i] == '\t' {
			next := ((visualCol / tabWidth) + 1) * tabWidth
			visualCol = next
		} else {
			visualCol++
		}
	}

	// Expand full line.
	var sb strings.Builder
	sb.Grow(len(line))
	vc := 0
	for i := 0; i < len(line); i++ {
		b := line[i]
		if b == '\t' {
			next := ((vc / tabWidth) + 1) * tabWidth
			spaces := next - vc
			sb.WriteString(strings.Repeat(" ", spaces))
			vc = next
			continue
		}
		sb.WriteByte(b)
		vc++
	}

	caret := strings.Repeat(" ", max(0, visualCol)) + "^"
	return sb.String(), caret
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
