package main

import (
	"flag"
	"fmt"
	"os"
	"os/user"

	"aiki/cmd/repl"
	"aiki/engine/runner"
	"aiki/engine/semantics/value"

	aikidebug "aiki/cmd/subcommands/dev/debug"
	aikismoke "aiki/cmd/subcommands/dev/smoke"
	aikidoclint "aiki/cmd/subcommands/tools/doclint"
	aikifmt "aiki/cmd/subcommands/tools/fmt"
	aikilint "aiki/cmd/subcommands/tools/lint"

	// For fmt/lint - they need reference grammar for now
	"aiki/reference/syntax"
)

func main() {
	repl.AppVersion = Version

	// Set grammar for fmt/lint (still uses reference)
	grammar := syntax.GetGrammar()
	aikifmt.SetGrammar(grammar)
	aikilint.SetGrammar(grammar)

	// Check for subcommands before flag parsing
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "doclint":
			aikidoclint.Run(os.Args[2:])
			return
		case "fmt":
			aikifmt.Run(os.Args[2:])
			return
		case "lint":
			aikilint.Run(os.Args[2:])
			return
		case "smoke":
			aikismoke.Run(os.Args[2:])
			return
		case "debug":
			aikidebug.Run(os.Args[2:])
			return
		}
	}

	opts := parseOptions()

	if opts.Expr != "" {
		runExpr(opts.Expr, opts)
		return
	}

	if flag.NArg() == 0 {
		startREPL(opts)
	} else {
		runFile(flag.Arg(0), opts)
	}
}

func startREPL(opts Options) {
	u, err := user.Current()
	if err != nil {
		u = &user.User{Username: "user"}
	}

	fmt.Printf("%s %s\n", appName, Version)
	fmt.Printf("Hello %s! The system is live.\n", u.Username)
	fmt.Printf("Type help() for help.\n\n")

	sess, err := repl.NewSession(os.Stdout, opts.Debug)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	sess.Run()
}

func runFile(filename string, opts Options) {
	err := runner.Run(filename)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runExpr(expr string, opts Options) {
	sess, err := runner.NewSession()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	result := sess.Eval(expr)

	if _, ok := result.(*value.ExitSignal); ok {
		return
	}

	if e, ok := result.(*value.Error); ok {
		fmt.Fprintln(os.Stderr, e.Inspect())
		os.Exit(1)
	}

	if opts.Debug {
		fmt.Println(result.Inspect())
	}
}
