package debug

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

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
	gold := fs.Bool("gold", false, "write output to gold file")
	check := fs.Bool("check", false, "compare output against gold file")

	if err := fs.Parse(args); err != nil {
		return 1
	}

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: aiki debug [--stage lex|parse|eval|all] [--gold|--check] <filename>")
		return 1
	}

	if *gold && *check {
		fmt.Fprintln(os.Stderr, "cannot use both --gold and --check")
		return 1
	}

	path := fs.Arg(0)
	var buf bytes.Buffer

	code := run(path, *stage, &buf, os.Stderr)
	if code != 0 {
		return code
	}

	goldPath := goldFileName(path, *stage)

	if *gold {
		if err := os.WriteFile(goldPath, buf.Bytes(), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "error writing gold file: %v\n", err)
			return 1
		}
		fmt.Fprintf(os.Stderr, "wrote %s\n", goldPath)
		return 0
	}

	if *check {
		expected, err := os.ReadFile(goldPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading gold file: %v\n", err)
			return 1
		}
		if !bytes.Equal(buf.Bytes(), expected) {
			fmt.Fprintf(os.Stderr, "output differs from %s\n", goldPath)
			fmt.Fprintln(os.Stderr, "=== EXPECTED ===")
			os.Stderr.Write(expected)
			fmt.Fprintln(os.Stderr, "=== GOT ===")
			os.Stderr.Write(buf.Bytes())
			return 1
		}
		fmt.Fprintf(os.Stderr, "ok %s\n", goldPath)
		return 0
	}

	// Normal output
	os.Stdout.Write(buf.Bytes())
	return 0
}

func goldFileName(path, stage string) string {
	ext := filepath.Ext(path)
	base := strings.TrimSuffix(path, ext)
	return fmt.Sprintf("%s.%s.gold", base, stage)
}

// run is the internal implementation, separated for testability.
func run(path, stage string, stdout, stderr io.Writer) int {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(stderr, "error reading file: %v\n", err)
		return 1
	}
	source := string(data)

	grammar := syntax.GetGrammar()

	// === LEX ===
	if stage == "lex" || stage == "all" {
		tokens, err := grammar.Tokenize(source)
		if err != nil {
			fmt.Fprintf(stderr, "lex error: %v\n", err)
			return 1
		}

		if stage == "lex" {
			fmt.Fprintln(stdout, "=== TOKENS ===")
			for _, tok := range tokens {
				fmt.Fprintf(stdout, "%d:%d\t%s\t%q\n", tok.Line, tok.Column, tok.Type, tok.Lexeme)
			}
			return 0
		}
	}

	// === PARSE ===
	ast, err := grammar.ParseSource(source)
	if err != nil {
		fmt.Fprintf(stderr, "parse error: %v\n", err)
		return 1
	}

	if stage == "parse" {
		fmt.Fprintln(stdout, "=== AST ===")
		printAST(stdout, ast, 0)
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

	if stage == "eval" {
		fmt.Fprintln(stdout, "=== RESULT ===")
		fmt.Fprintln(stdout, result.Inspect())
		return 0
	}

	// stage == "all"
	fmt.Fprintln(stdout, "=== TOKENS ===")
	tokens, _ := grammar.Tokenize(source)
	for _, tok := range tokens {
		fmt.Fprintf(stdout, "%d:%d\t%s\t%q\n", tok.Line, tok.Column, tok.Type, tok.Lexeme)
	}
	fmt.Fprintln(stdout, "=== AST ===")
	printAST(stdout, ast, 0)
	fmt.Fprintln(stdout, "=== RESULT ===")
	fmt.Fprintln(stdout, result.Inspect())

	return 0
}

func printAST(w io.Writer, node *syntax.Node, depth int) {
	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}

	if node.Value != "" {
		fmt.Fprintf(w, "%s%s: %q\n", indent, node.Type, node.Value)
	} else {
		fmt.Fprintf(w, "%s%s\n", indent, node.Type)
	}

	for _, child := range node.Children {
		printAST(w, child, depth+1)
	}
}
