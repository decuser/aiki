package fmt

import (
	"aiki/reference/syntax"
	"flag"
	gofmt "fmt"
	"os"
	"path/filepath"
	"strings"
)

var grammar *syntax.Grammar

// SetGrammar sets the grammar for formatting.
func SetGrammar(g *syntax.Grammar) {
	grammar = g
}

// Run executes the fmt subcommand to mirror 'go fmt' behavior.
func Run(args []string) {
	fs := flag.NewFlagSet("fmt", flag.ExitOnError)
	fs.Parse(args)

	if fs.NArg() == 0 {
		gofmt.Fprintln(os.Stderr, "usage: aiki fmt <file.ai>")
		gofmt.Fprintln(os.Stderr, "       aiki fmt ./...")
		os.Exit(1)
	}

	if grammar == nil {
		gofmt.Fprintln(os.Stderr, "fmt: grammar not initialized")
		os.Exit(1)
	}

	for _, path := range fs.Args() {
		// In 'go fmt' style, we always write changes back to the file
		if err := formatPath(path, true); err != nil {
			gofmt.Fprintf(os.Stderr, "fmt: %s\n", err)
			os.Exit(1)
		}
	}
}

func formatPath(pathStr string, write bool) error {
	if strings.HasSuffix(pathStr, "/...") {
		dir := strings.TrimSuffix(pathStr, "/...")
		if dir == "." || dir == "" {
			dir = "."
		}
		return formatDir(dir, write)
	}

	return formatFile(pathStr, write)
}

func formatDir(dir string, write bool) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
			return filepath.SkipDir
		}

		// Format .ai files
		if !info.IsDir() && strings.HasSuffix(path, ".ai") {
			if err := formatFile(path, write); err != nil {
				gofmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
				// Continue with other files
			}
		}
		return nil
	})
}

func formatFile(path string, write bool) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	formatted, err := FormatSource(grammar, string(data))
	if err != nil {
		return gofmt.Errorf("%s: %v", path, err)
	}

	if write {
		// Only write if changed
		if formatted != string(data) {
			gofmt.Println(path)
			if err := os.WriteFile(path, []byte(formatted), 0644); err != nil {
				return err
			}
		}
	} else {
		gofmt.Print(formatted)
	}

	return nil
}

// Format formats source code and returns the result.
// This is the public API for programmatic use.
func Format(source string) (string, error) {
	if grammar == nil {
		return "", gofmt.Errorf("grammar not initialized")
	}
	return FormatSource(grammar, source)
}
