package debug

import (
	"flag"
	"fmt"
	"io"
	"os"

	"aiki/engine"
	"aiki/engine/runtime/hal/substrate"
	"aiki/engine/runtime/prelude"
	"aiki/engine/semantics/evaluator"
	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
	"aiki/engine/syntax/grammar"
)

// Run executes the debug subcommand with the given arguments.
// Returns 0 on success, non-zero on error.
func Run(args []string) int {
	fs := flag.NewFlagSet("debug", flag.ContinueOnError)
	stage := fs.String("stage", "all", "lex|parse|eval|all")
	trace := fs.Bool("trace", false, "enable trace observation")
	prelude := fs.Bool("prelude", false, "include prelude in trace")

	if err := fs.Parse(args); err != nil {
		return 1
	}

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: aiki debug [-stage lex|parse|eval|all] [-trace] [-prelude] <filename>")
		return 1
	}

	return run(fs.Arg(0), *stage, *trace, *prelude, os.Stdout, os.Stderr)
}

// run is the internal implementation, separated for testability.
func run(path, stage string, trace, tracePrelude bool, stdout, stderr io.Writer) int {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(stderr, "failed to read file: %v\n", err)
		return 1
	}
	source := string(data)

	g, err := grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
	if err != nil {
		fmt.Fprintf(stderr, "failed to load grammar: %v\n", err)
		return 1
	}

	var obs engine.Observer
	if trace {
		tobs := &TraceObserver{out: stdout}
		if !tracePrelude {
			tobs.fileOnly = path
		}
		obs = tobs
	}

	// Lex
	lexer := syntax.NewLexer(g, path, source, obs)
	tokens, err := lexer.Tokenize()
	if err != nil {
		fmt.Fprintf(stderr, "lex error: %v\n", err)
		return 1
	}

	if stage == "lex" {
		fmt.Fprintln(stdout, "==== Tokens ====")
		for _, tok := range tokens {
			fmt.Fprintf(stdout, "%d:%d\t%s\t%q\n", tok.Pos.Line, tok.Pos.Col, tok.Type, tok.Lexeme)
		}
		return 0
	}

	// Parse
	parser := syntax.NewParser(g, tokens, source, obs)
	tree, err := parser.Parse()
	if err != nil {
		fmt.Fprintf(stderr, "parse error: %v\n", err)
		return 1
	}

	if stage == "parse" {
		fmt.Fprintln(stdout, "==== AST ====")
		printTree(stdout, tree, 0)
		return 0
	}

	// Eval
	if stage == "all" || stage == "eval" {
		fmt.Fprintln(stdout, "==== Eval ====")
		rt := substrate.NewGoRuntime()

		// Create prelude environment
		preludeEnv := value.NewEnvWithScope(value.ScopePrelude)
		preludeEnv.SetFile("<prelude>")
		preludeEnv.SetSource(prelude.Source)

		// Load prelude (with observer only if -prelude flag set)
		var preObs engine.Observer
		if tracePrelude {
			preObs = obs
		}
		preLexer := syntax.NewLexer(g, "<prelude>", prelude.Source, preObs)
		preTokens, err := preLexer.Tokenize()
		if err != nil {
			fmt.Fprintf(stderr, "prelude lex error: %v\n", err)
			return 1
		}
		preParser := syntax.NewParser(g, preTokens, prelude.Source, preObs)
		preTree, err := preParser.Parse()
		if err != nil {
			fmt.Fprintf(stderr, "prelude parse error: %v\n", err)
			return 1
		}
		preEv := evaluator.New(rt, preObs)
		preEv.SetGrammar(g)
		preEv.Eval(preTree, preludeEnv)

		// Create user environment and evaluator
		ev := evaluator.New(rt, obs)
		ev.SetGrammar(g)

		// Create user environment
		userEnv := value.NewEnclosedEnv(preludeEnv)
		userEnv.SetFile(path)
		userEnv.SetSource(source)

		result := ev.Eval(tree, userEnv)
		fmt.Fprintf(stdout, "=> %s\n", result.Inspect())
	}

	return 0
}

func printTree(w io.Writer, node *syntax.Node, depth int) {
	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}
	if node.Value != "" {
		fmt.Fprintf(w, "%s%s %d:%d %q\n", indent, node.Type, node.Pos.Line, node.Pos.Col, node.Value)
	} else {
		fmt.Fprintf(w, "%s%s %d:%d\n", indent, node.Type, node.Pos.Line, node.Pos.Col)
	}
	for _, child := range node.Children {
		printTree(w, child, depth+1)
	}
}
