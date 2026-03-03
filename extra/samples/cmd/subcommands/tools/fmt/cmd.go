package fmt

import (
	"flag"
	"fmt"
	"os"
)

// Run formats Aiki source files in place.
//
// Default behavior mirrors the user's preferred workflow:
// rewrite files in place and print filenames that changed.
//
// Flags:
//
//	-n     do not write; just list files that would change
//	-p     print formatted source to stdout (single file only; no write)
//	-b     create a .bak file before overwriting (only when changes occur)
func Run(args []string) int {
	fs := flag.NewFlagSet("fmt", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	noWrite := fs.Bool("n", false, "do not write; list files that would change")
	printOut := fs.Bool("p", false, "print formatted output to stdout (single file)")
	backup := fs.Bool("b", false, "create .bak before overwriting")

	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "usage: aiki fmt <file.ai> [more files] | ./...")
		return 2
	}
	if *printOut && fs.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "fmt: -p requires exactly one file")
		return 2
	}

	cfg := Config{
		Write:         !*noWrite && !*printOut,
		ListOnly:      *noWrite,
		PrintToStdout: *printOut,
		Backup:        *backup,
	}

	changedAny := false
	for _, p := range fs.Args() {
		changed, err := FormatPath(p, cfg)
		if err != nil {
			fmt.Fprintln(os.Stderr, "fmt:", err)
			return 1
		}
		if changed {
			changedAny = true
		}
	}

	// In list only mode, return 1 if any file would change (useful for CI).
	if cfg.ListOnly && changedAny {
		return 1
	}
	return 0
}
