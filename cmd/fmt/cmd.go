package fmt

import (
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

	for _, path := range args {
		_, err := Format(path)
		if err != nil {
			gofmt.Fprintf(os.Stderr, "fmt: %s\n", err)
			os.Exit(1)
		}
	}
}
