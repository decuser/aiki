package lint

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"aiki/ebnf"
)

var grammar *ebnf.Grammar

// SetGrammar sets the grammar for linting.
func SetGrammar(g *ebnf.Grammar) {
	grammar = g
}

// Run executes the lint subcommand.
func Run(args []string) {
	fs := flag.NewFlagSet("lint", flag.ExitOnError)
	fs.Parse(args)

	if fs.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "usage: aiki lint <file.ai>")
		fmt.Fprintln(os.Stderr, "       aiki lint ./...")
		os.Exit(1)
	}

	if grammar == nil {
		fmt.Fprintln(os.Stderr, "lint: grammar not initialized")
		os.Exit(1)
	}

	totalDiags := 0
	hasErrors := false

	for _, path := range fs.Args() {
		diags, err := lintPath(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "lint: %s\n", err)
			os.Exit(1)
		}
		for _, d := range diags {
			totalDiags++
			if d.Level == "error" {
				hasErrors = true
			}
		}
	}

	if totalDiags == 0 {
		// Clean - no output, exit 0
		return
	}

	if hasErrors {
		os.Exit(1)
	}
}

func lintPath(pathStr string) ([]Diagnostic, error) {
	if strings.HasSuffix(pathStr, "/...") {
		dir := strings.TrimSuffix(pathStr, "/...")
		if dir == "." || dir == "" {
			dir = "."
		}
		return lintDir(dir)
	}
	return lintFile(pathStr)
}

func lintDir(dir string) ([]Diagnostic, error) {
	var all []Diagnostic
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
			return filepath.SkipDir
		}
		if !info.IsDir() && strings.HasSuffix(path, ".ai") {
			diags, err := lintFile(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
				return nil // continue with other files
			}
			all = append(all, diags...)
		}
		return nil
	})
	return all, err
}

func lintFile(path string) ([]Diagnostic, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	diags, err := LintSource(grammar, string(data))
	if err != nil {
		return nil, fmt.Errorf("%s: %v", path, err)
	}

	for _, d := range diags {
		fmt.Fprintf(os.Stderr, "%s:%d:%d: %s: %s\n",
			path, d.Line, d.Column, d.Level, d.Message)
	}

	return diags, nil
}

// Lint checks a file or directory. Returns diagnostics.
func Lint(pathStr string) ([]Diagnostic, error) {
	if grammar == nil {
		return nil, fmt.Errorf("grammar not initialized")
	}
	return lintPath(pathStr)
}

// LintString checks source code string. Public API for programmatic use.
func LintString(source string) ([]Diagnostic, error) {
	if grammar == nil {
		return nil, fmt.Errorf("grammar not initialized")
	}
	return LintSource(grammar, source)
}
