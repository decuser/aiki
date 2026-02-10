package main

import (
	"flag"
	"fmt"
	"os"
	"os/user"

	"aiki/eval"
	"aiki/prelude"
	"aiki/repl"
	"aiki/value"
)

func main() {
	// Check for subcommands before flag parsing
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "fmt":
			runFmt(os.Args[2:])
			return
		}
	}

	opts := parseOptions()

	env := value.NewEnv(nil)

	if err := prelude.LoadPrelude(env); err != nil {
		fmt.Fprintf(os.Stderr, "error loading prelude: %s\n", err)
		os.Exit(1)
	}

	if flag.NArg() == 0 {
		startREPL(env, opts)
	} else {
		runFile(flag.Arg(0), env, opts)
	}
}

func runFmt(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: aiki fmt <path>")
		fmt.Fprintln(os.Stderr, "       aiki fmt ./...")
		os.Exit(1)
	}

	for _, path := range args {
		result := eval.Format(path)
		if err, ok := result.(*value.Error); ok {
			fmt.Fprintf(os.Stderr, "%s\n", err.Message)
			os.Exit(1)
		}
	}
}

func startREPL(env *value.Env, opts Options) {
	u, err := user.Current()
	if err != nil {
		u = &user.User{Username: "user"}
	}

	fmt.Printf("%s %s\n", appName, version)
	fmt.Printf("Hello %s! The system is live.\n", u.Username)
	fmt.Printf("Type help() for commands.\n\n")

	repl.Start(os.Stdin, os.Stdout, env, opts.Debug)
}

func runFile(filename string, env *value.Env, opts Options) {
	result := eval.RunFile(filename, env)

	if e, ok := result.(*value.Error); ok {
		fmt.Fprintln(os.Stderr, e.Inspect())
		os.Exit(1)
	}

	if opts.Debug {
		fmt.Println(result.Inspect())
	}
}
