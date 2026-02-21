package debug

import (
	"flag"
	"fmt"
	"io"
	"os"

	"aiki/engine"
	"aiki/reference/runtime/prelude"
	"aiki/reference/semantics/eval"
	"aiki/reference/semantics/value"
	"aiki/reference/syntax"
)

// Run executes the debug subcommand with the given arguments.
// Returns 0 on success, non-zero on error.
func Run(args []string) int {
	fs := flag.NewFlagSet("debug", flag.ContinueOnError)
	stage := fs.String("stage", "all", "output stage: lex, parse, eval, or all")
	trace := fs.Bool("trace", false, "enable trace observation")

	if err := fs.Parse(args); err != nil {
		return 1
	}

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: aiki debug [--stage lex|parse|eval|all] [--trace] <filename>")
		return 1
	}

	return run(fs.Arg(0), *stage, *trace, os.Stdout, os.Stderr)
}

// run is the internal implementation, separated for testability.
func run(path, stage string, trace bool, stdout, stderr io.Writer) int {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(stderr, "error reading file: %v\n", err)
		return 1
	}
	source := string(data)

	grammar := syntax.GetGrammar()

	var obs engine.Observer
	if trace {
		obs = &TraceObserver{out: stdout}
	} else {
		obs = engine.SilentObserver{}
	}

	// === LEX ===
	if stage == "lex" || stage == "all" {
		tokens, err := grammar.Tokenize(source)
		if err != nil {
			fmt.Fprintf(stderr, "lex error: %v\n", err)
			return 1
		}

		if stage == "lex" || trace {
			fmt.Fprintln(stdout, "=== TOKENS ===")
			for _, tok := range tokens {
				obs.OnLex(tok.Type, tok.Lexeme, engine.Position{
					File: path,
					Line: tok.Line,
					Col:  tok.Column,
				})
				fmt.Fprintf(stdout, "%d:%d\t%s\t%q\n", tok.Line, tok.Column, tok.Type, tok.Lexeme)
			}
		}

		if stage == "lex" {
			return 0
		}
	}

	// === PARSE ===
	ast, err := grammar.ParseSource(source)
	if err != nil {
		fmt.Fprintf(stderr, "parse error: %v\n", err)
		return 1
	}

	if stage == "parse" || (stage == "all" && trace) {
		fmt.Fprintln(stdout, "=== AST ===")
		printAST(stdout, ast, 0, obs, path)
	}

	if stage == "parse" {
		return 0
	}

	// === EVAL ===
	env := value.NewEnv(nil)
	eval.RunNode(grammar, prelude.Source, env)
	env.SnapshotPrelude()

	env.SetFile(path)
	env.SetSource(source)

	result := eval.EvalNode(ast, env)

	if e, ok := result.(*value.Error); ok {
		fmt.Fprintln(stderr, e.Inspect())
		return 1
	}

	fmt.Fprintln(stdout, "=== RESULT ===")
	fmt.Fprintln(stdout, result.Inspect())

	return 0
}

func printAST(w io.Writer, node *syntax.Node, depth int, obs engine.Observer, file string) {
	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}

	obs.OnParse(node.Type, depth, engine.Position{
		File: file,
		Line: node.Line,
		Col:  node.Column,
	})

	if node.Value != "" {
		fmt.Fprintf(w, "%s%s: %q\n", indent, node.Type, node.Value)
	} else {
		fmt.Fprintf(w, "%s%s\n", indent, node.Type)
	}

	for _, child := range node.Children {
		printAST(w, child, depth+1, obs, file)
	}
}
