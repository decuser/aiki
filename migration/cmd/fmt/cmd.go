package fmt

import (
	"errors"
	gofmt "fmt"
	"os"
)

// Run executes the fmt subcommand.
func Run(args []string) {
	if len(args) == 0 {
		gofmt.Fprintln(os.Stderr, "usage: aiki fmt <path>")
		gofmt.Fprintln(os.Stderr, "       aiki fmt ./...")
		os.Exit(1)
	}

	gofmt.Fprintln(os.Stderr, "fmt: not yet implemented for new grammar")
	os.Exit(1)
}

// Format formats a file or directory. Returns count of files formatted.
// TODO: Reimplement with EBNF-based AST
func Format(pathStr string) (int, error) {
	return 0, errors.New("fmt not yet implemented for new grammar")
}
