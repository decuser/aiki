package doclint

import (
	"flag"
	"fmt"
	"os"
)

// Run executes the doclint subcommand.
//
// Usage:
//
//	aiki doclint [flags] path...
//	aiki doclint extra/doc extra/status
//	aiki doclint ./...
//
// Flags:
//
//	--no-header  allow files without contract headers
func Run(args []string) {
	fs := flag.NewFlagSet("doclint", flag.ContinueOnError)
	noHeader := fs.Bool("no-header", false, "allow files without contract headers")

	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}

	paths := fs.Args()
	if len(paths) == 0 {
		fmt.Fprintln(os.Stderr, "usage: aiki doclint [--no-header] path...")
		os.Exit(2)
	}

	opts := Options{
		RequireHeader: !*noHeader,
	}

	violations, err := Check(paths, opts)
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
