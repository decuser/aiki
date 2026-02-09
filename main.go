package main

import (
	"flag"
	"fmt"
	"os"
	"os/user"

	"aiki/eval"
	"aiki/repl"
	"aiki/value"
)

func main() {
	opts := parseOptions()

	env := value.NewEnv(nil)

	if err := loadPrelude(env); err != nil {
		fmt.Fprintf(os.Stderr, "error loading prelude: %s\n", err)
		os.Exit(1)
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

	fmt.Printf("%s %s\n", appName, version)
	fmt.Printf("Hello %s! The system is live.\n", u.Username)
	fmt.Printf("Type help() for commands.\n\n")

	repl.Start(os.Stdin, os.Stdout, env, opts.Debug)
}

func runFile(filename string, env *value.Env, opts Options) {
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading file: %s\n", err)
		os.Exit(1)
	}

	result := eval.Run(string(data), env)

	if e, ok := result.(*value.Error); ok {
		fmt.Fprintf(os.Stderr, "error: %s\n", e.Message)
		os.Exit(1)
	}

	if opts.Debug {
		fmt.Println(result.Inspect())
	}
}
