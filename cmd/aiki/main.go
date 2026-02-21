package main

import (
	"flag"
	"fmt"
	"os"
	"os/user"

	"aiki/cmd/subcommands/ux/repl"
	"aiki/reference/runtime/hal"
	"aiki/reference/runtime/prelude"
	"aiki/reference/semantics/eval"
	"aiki/reference/semantics/value"
	"aiki/reference/syntax"

	aikismoke "aiki/cmd/subcommands/dev/smoke"
	aikidoclint "aiki/cmd/subcommands/tools/doclint"
	aikifmt "aiki/cmd/subcommands/tools/fmt"
	aikilint "aiki/cmd/subcommands/tools/lint"
)

func main() {
	repl.AppVersion = Version
	grammar := syntax.GetGrammar()
	aikifmt.SetGrammar(grammar)
	aikilint.SetGrammar(grammar)
	eval.SetNodeGrammar(grammar)

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
		}

	}

	opts := parseOptions()

	env := value.NewEnv(nil)

	result := eval.RunNode(grammar, prelude.Source, env)
	if e, ok := result.(*value.Error); ok {
		fmt.Fprintf(os.Stderr, "error loading prelude: %s\n", e.Message)
		os.Exit(1)
	}
	env.SnapshotPrelude()

	if opts.Expr != "" {
		runExpr(opts.Expr, env, opts)
		return
	}

	if flag.NArg() == 0 {
		startREPL(env, opts)
	} else {
		runFile(flag.Arg(0), env, opts)
	}
}

func startREPL(env *value.Env, opts Options) {
	u, err := user.Current()
	if err != nil {
		u = &user.User{Username: "user"}
	}

	fmt.Printf("%s %s\n", appName, Version)
	fmt.Printf("Hello %s! The system is live.\n", u.Username)
	fmt.Printf("Type help() for help.\n\n")

	repl.Run(syntax.GetGrammar(), os.Stdin, os.Stdout, env, opts.Debug)
}

func runFile(filename string, env *value.Env, opts Options) {
	result := eval.RunFileNode(syntax.GetGrammar(), filename, env)

	if e, ok := result.(*value.Error); ok {
		fmt.Fprintln(os.Stderr, e.Inspect())
		os.Exit(1)
	}

	if opts.Debug {
		fmt.Println(result.Inspect())
	}
}

func runExpr(expr string, env *value.Env, opts Options) {
	result := eval.RunNode(syntax.GetGrammar(), expr, env)

	if _, ok := result.(*hal.ExitSignal); ok {
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
