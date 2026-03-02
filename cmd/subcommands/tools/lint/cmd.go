package lint

import (
    "flag"
    "fmt"
    "os"
)

// Run executes lint.
// Current implementation is format-first: fail if any file is not already
// in canonical fmt form.
//
// Flags:
//   -fmt     enable formatting check (default true)
func Run(args []string) int {
    fs := flag.NewFlagSet("lint", flag.ContinueOnError)
    fs.SetOutput(os.Stderr)
    checkFmt := fs.Bool("fmt", true, "fail if fmt would change files")

    if err := fs.Parse(args); err != nil {
        return 2
    }
    if fs.NArg() == 0 {
        fmt.Fprintln(os.Stderr, "usage: aiki lint <file.ai> [more files] | ./...")
        return 2
    }

    if *checkFmt {
        bad, err := CheckFormatting(fs.Args())
        if err != nil {
            fmt.Fprintln(os.Stderr, "lint:", err)
            return 1
        }
        if len(bad) > 0 {
            for _, p := range bad {
                fmt.Fprintln(os.Stdout, p)
            }
            return 1
        }
    }

    return 0
}
