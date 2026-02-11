package lint

import (
	"fmt"
	"os"
)

// Run executes the lint subcommand.
func Run(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: aiki lint <path>")
		fmt.Fprintln(os.Stderr, "       aiki lint ./...")
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
