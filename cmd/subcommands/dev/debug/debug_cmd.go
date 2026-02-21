package debug

import (
	"aiki/engine"
	"aiki/engine/runtime/hal/substrate"
	"aiki/engine/runtime/prelude"
	"aiki/engine/semantics/evaluator"
	"aiki/engine/syntax"
	"aiki/engine/syntax/definition"
	"flag"
	"fmt"
	"io"
	"os"
)

// Run executes the debug subcommand with the given arguments.
// Returns 0 on success, non-zero on error.
func Run(args []string) int {
	fs := flag.NewFlagSet("debug", flag.ContinueOnError)
	boundary := fs.String("boundary", "all", "filter output by boundary")
	trace := fs.Bool("trace", false, "enable trace observation")

	if err := fs.Parse(args); err != nil {
		return 1
	}

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: aiki debug <filename>")
		return 1
	}

	return run(fs.Arg(0), *boundary, *trace, os.Stdout, os.Stderr)
}

// run is the internal implementation, separated for testability.
func run(path, boundary string, trace bool, stdout, stderr io.Writer) int {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(stderr, "failed to read file: %v\n", err)
		return 1
	}

	grammar := definition.New()
	var obs engine.Observer
	if trace {
		obs = TraceObserver{out: stdout}
		grammar.SetObserver(obs)
	}

	lexer := syntax.NewLexer(path, string(data), grammar)
	parser, err := syntax.NewParser(lexer, grammar)
	if err != nil {
		fmt.Fprintf(stderr, "parser init failed: %v\n", err)
		return 1
	}
	tree, err := parser.Parse()
	if err != nil {
		fmt.Fprintf(stderr, "parse error: %v\n", err)
		return 1
	}

	if boundary == "all" || boundary == "eval" {
		fmt.Fprintln(stdout, "\n==== Entering Evaluator ====")
		rt := substrate.NewGoRuntime()
		global := evaluator.NewScope(nil)
		global.SetFile(path)
		global.SetSource(string(data))
		ev := evaluator.New(rt, obs)
		ev.SetGrammar(grammar)
		
		// Load prelude
		_, err = ev.RunSource(prelude.Name, prelude.Source, global)
		if err != nil {
			fmt.Fprintf(stderr, "error loading prelude: %v\n", err)
			return 1
		}
		
		result, err := ev.Eval(tree, global)
		if err != nil {
			fmt.Fprintf(stderr, "eval error: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "=> %s\n", result.Inspect())
	}

	return 0
}
