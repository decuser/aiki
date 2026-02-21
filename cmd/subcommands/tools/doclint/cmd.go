package doclint

import (
	"fmt"
	"os"
)

// Run executes the doclint subcommand.
//
// Behavior
// Reads doclint.ini from current working directory
// Enforces contract headers and allowed tags for all scoped markdown files
func Run(args []string) {
	scanRoot := "."
	if len(args) > 1 {
		fmt.Fprintln(os.Stderr, "usage: aiki doclint [path]")
		os.Exit(1)
	}
	if len(args) == 1 {
		scanRoot = args[0]
	}

	cfg, err := LoadConfig(scanRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "doclint: %s\n", err)
		os.Exit(2)
	}

	violations, err := Check(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "doclint: %s\n", err)
		os.Exit(2)
	}

	if len(violations) > 0 {
		for _, v := range violations {
			fmt.Fprintf(os.Stderr, "%s:%d: %s\n", v.Path, v.Line, v.Message)
		}
		os.Exit(1)
	}
}
