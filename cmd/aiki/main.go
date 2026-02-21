package main

import (
	"flag"
	"fmt"
	"os"
	"os/user"

	"aiki/cmd/subcommands/dev/debug"
	"aiki/cmd/subcommands/dev/smoke"
	"aiki/cmd/subcommands/ux/repl"
	"aiki/engine/runtime/hal/substrate"
	"aiki/engine/runtime/prelude"
	"aiki/engine/semantics/evaluator"
	"aiki/engine/syntax"
	"aiki/engine/syntax/definition"
)

func main() {
	// Check for subcommands before flag parsing
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "version":
			fmt.Printf("aiki %s\n", Version)
			return
		case "debug":
			os.Exit(debug.Run(os.Args[2:]))
		case "smoke":
			os.Exit(smoke.Run(os.Args[2:]))
		// TODO: Add back when tools are ported
		// case "doclint":
		// 	os.Exit(doclint.Run(os.Args[2:]))
		// case "fmt":
		// 	os.Exit(aikifmt.Run(os.Args[2:]))
		// case "lint":
		// 	os.Exit(lint.Run(os.Args[2:]))
		// case "smoke":
		// 	os.Exit(smoke.Run(os.Args[2:]))
		}
	}

	grammar := definition.New()
	opts := parseOptions()

	// Create runtime and scope
	rt := substrate.NewGoRuntime()
	scope := evaluator.NewScope(nil)

	// Create evaluator
	ev := evaluator.New(rt, nil)
	ev.SetGrammar(grammar)

	_, err := ev.RunSource(prelude.Name, prelude.Source, scope)
	if err != nil {
	    fmt.Fprintf(os.Stderr, "error loading prelude: %v\n", err)
	    os.Exit(1)
	}

	if opts.Expr != "" {
		runExpr(ev, grammar, opts.Expr, scope, opts)
		return
	}

	if flag.NArg() == 0 {
		startREPL(ev, grammar, scope, opts)
	} else {
		runFile(ev, flag.Arg(0), scope, opts)
	}
}

func startREPL(ev *evaluator.Evaluator, grammar syntax.GrammarContract, scope *evaluator.Scope, opts Options) {
	u, err := user.Current()
	if err != nil {
		u = &user.User{Username: "user"}
	}

	fmt.Printf("%s %s\n", appName, Version)
	fmt.Printf("Hello %s! The system is live.\n", u.Username)
	fmt.Printf("Type help() for help.\n\n")

	repl.Run(ev, grammar, os.Stdout, scope, opts.Debug)
}

func runFile(ev *evaluator.Evaluator, filename string, scope *evaluator.Scope, opts Options) {
	scope.SetFile(filename)

	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading file: %v\n", err)
		os.Exit(1)
	}

	scope.SetSource(string(content))
	result, err := ev.RunFile(filename, scope)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if opts.Debug {
		fmt.Println(result.Inspect())
	}
}

func runExpr(ev *evaluator.Evaluator, grammar syntax.GrammarContract, expr string, scope *evaluator.Scope, opts Options) {
	lexer := syntax.NewLexer("expr", expr, grammar)
	parser, err := syntax.NewParser(lexer, grammar)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse error: %v\n", err)
		os.Exit(1)
	}

	ast, err := parser.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse error: %v\n", err)
		os.Exit(1)
	}

	result, err := ev.Eval(ast, scope)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if opts.Debug {
		fmt.Println(result.Inspect())
	}
}
