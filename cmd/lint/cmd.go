package lint

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"aiki/ebnf"
	"aiki/lang/eval"
	"aiki/lang/value"
	"aiki/strict"
)

var grammar *ebnf.Grammar
var strictExports []string

// SetGrammar sets the grammar for linting and evaluates strict to get exports.
func SetGrammar(g *ebnf.Grammar) {
	grammar = g

	// Evaluate strict.ai to get exports
	env := value.NewEnv(nil)
	result := eval.RunNode(g, strict.Source, env)
	if _, ok := result.(*value.Error); !ok {
		strictExports = env.GetExports()
	}
}

// Run executes the lint subcommand.
func Run(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: aiki lint <path>")
		fmt.Fprintln(os.Stderr, "       aiki lint ./...")
		os.Exit(1)
	}

	if grammar == nil {
		fmt.Fprintln(os.Stderr, "lint: grammar not initialized")
		os.Exit(1)
	}

	for _, path := range args {
		_, err := Lint(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "lint: %s\n", err)
			os.Exit(1)
		}
	}
}

// Lint checks a file or directory. Returns count of files checked.
func Lint(pathStr string) (int, error) {
	if grammar == nil {
		return 0, fmt.Errorf("grammar not initialized")
	}

	if strings.HasSuffix(pathStr, "/...") {
		dir := strings.TrimSuffix(pathStr, "/...")
		if dir == "." || dir == "" {
			dir = "."
		}
		return lintDir(dir)
	}

	if err := lintFile(pathStr); err != nil {
		return 0, err
	}
	return 1, nil
}

func lintDir(dir string) (int, error) {
	count := 0
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
			return filepath.SkipDir
		}
		if !info.IsDir() && strings.HasSuffix(path, ".ai") {
			if err := lintFile(path); err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err)
			} else {
				count++
			}
		}
		return nil
	})
	return count, err
}

func lintFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	diags, err := LintSource(grammar, string(data))
	if err != nil {
		return fmt.Errorf("%s: parse error: %s", path, err)
	}

	if len(diags) > 0 {
		fmt.Println(path)
		for _, d := range diags {
			fmt.Printf("\tline %d: [%s] %s\n", d.Line, d.Level, d.Message)
		}
	}

	return nil
}
