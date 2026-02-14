package lint

import (
	"errors"
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

	fmt.Fprintln(os.Stderr, "lint: not yet implemented for new grammar")
	os.Exit(1)
}

// Lint checks a file or directory. Returns count of files checked.
// TODO: Reimplement with EBNF-based AST
func Lint(pathStr string) (int, error) {
	return 0, errors.New("lint not yet implemented for new grammar")
}
