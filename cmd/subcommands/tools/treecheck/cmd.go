package treecheck

import (
	"flag"
	"fmt"
	"os"
)

func Run(args []string) int {
	fs := flag.NewFlagSet("treecheck", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	root := fs.String("root", ".", "repository root")
	allow := fs.String("allow", "treecheck.allow", "root exception manifest")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "usage: aiki treecheck [-root dir] [-allow file]")
		return 2
	}

	result, err := Check(*root, *allow)
	if err != nil {
		fmt.Fprintln(os.Stderr, "treecheck:", err)
		return 1
	}
	if len(result.Errors) == 0 && len(result.Orphans) == 0 {
		fmt.Printf("treecheck ok (%d files, %d structurally justified, %d explicitly allowed)\n", result.Files, result.Justified, result.Allowed)
		return 0
	}

	fmt.Fprintln(os.Stderr, "treecheck FAIL")
	for _, f := range result.Errors {
		fmt.Fprintf(os.Stderr, "  error: %s: %s\n", f.Path, f.Reason)
	}
	for _, f := range result.Orphans {
		fmt.Fprintf(os.Stderr, "  orphan: %s: %s\n", f.Path, f.Reason)
	}
	fmt.Fprintf(os.Stderr, "treecheck: %d error(s), %d orphan candidate(s) across %d files\n", len(result.Errors), len(result.Orphans), result.Files)
	return 1
}
