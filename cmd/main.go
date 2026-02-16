package main

import (
	_ "embed"
	"flag"
	"fmt"
	"os"
	"os/user"

	"aiki/cmd/repl"
	"aiki/ebnf"
	"aiki/lang/eval"
	"aiki/lang/value"
	"aiki/layers/hal"
	"aiki/layers/prelude"
	"aiki/version"

	aikifmt "aiki/cmd/fmt"
	aikilint "aiki/cmd/lint"
)

//go:embed grammar.ebnf
var grammarSource string

var grammar *ebnf.Grammar

func main() {
	var err error
	grammar, err = ebnf.Parse(grammarSource) // Parse, not ParseFile
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading grammar: %s\n", err)
		os.Exit(1)
	}
	aikifmt.SetGrammar(grammar)
	aikilint.SetGrammar(grammar)
	eval.SetNodeGrammar(grammar)

	// Check for subcommands before flag parsing
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "fmt":
			aikifmt.Run(os.Args[2:])
			return
		case "lint": // New case
			aikilint.Run(os.Args[2:])
			return
		}
	}

	opts := parseOptions()

	env := value.NewEnv(nil)

	result := eval.RunNode(grammar, strict.Source, env)
	if e, ok := result.(*value.Error); ok {
		fmt.Fprintf(os.Stderr, "error loading strict: %s\n", e.Message)
		os.Exit(1)
	}
	env.SnapshotStrict()

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

	fmt.Printf("%s %s\n", appName, version.Version)
	fmt.Printf("Hello %s! The system is live.\n", u.Username)
	fmt.Printf("Type help() for help.\n\n")

	repl.Run(grammar, os.Stdin, os.Stdout, env, opts.Debug)
}

func runFile(filename string, env *value.Env, opts Options) {
	result := eval.RunFileNode(grammar, filename, env)

	if e, ok := result.(*value.Error); ok {
		fmt.Fprintln(os.Stderr, e.Inspect())
		os.Exit(1)
	}

	if opts.Debug {
		fmt.Println(result.Inspect())
	}
}

func runExpr(expr string, env *value.Env, opts Options) {
	result := eval.RunNode(grammar, expr, env)

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
