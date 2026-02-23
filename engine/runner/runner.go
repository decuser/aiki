// Package runner provides the runner for executing Aiki programs.
package runner

import (
	"fmt"
	"os"

	"aiki/engine/runtime/hal/substrate"
	"aiki/engine/semantics/evaluator"
	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
	"aiki/engine/syntax/grammar"
)

// Run executes an Aiki source file.
func Run(filename string) error {
	source, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	return RunSource(filename, string(source))
}

// RunSource executes Aiki source code.
func RunSource(filename, source string) error {
	// Load grammar
	g, err := grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
	if err != nil {
		return fmt.Errorf("loading grammar: %w", err)
	}

	// Lex
	lexer := syntax.NewLexer(g, filename, source, nil)
	tokens, err := lexer.Tokenize()
	if err != nil {
		return fmt.Errorf("lexer: %w", err)
	}

	// Parse
	parser := syntax.NewParser(g, tokens, source, nil)
	ast, err := parser.Parse()
	if err != nil {
		return fmt.Errorf("parser: %w", err)
	}

	// Create runtime and evaluator
	rt := substrate.NewGoRuntime()
	ev := evaluator.New(rt, nil)

	// Create environment
	env := value.NewEnv()
	env.SetFile(filename)
	env.SetSource(source)

	// Eval
	result := ev.Eval(ast, env)
	if err, ok := result.(*value.Error); ok {
		return fmt.Errorf("%s", err.Message)
	}

	return nil
}

// RunExpr evaluates a single expression and returns the result.
func RunExpr(expr string) (string, error) {
	g, err := grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
	if err != nil {
		return "", err
	}

	lexer := syntax.NewLexer(g, "<expr>", expr, nil)
	tokens, err := lexer.Tokenize()
	if err != nil {
		return "", err
	}

	parser := syntax.NewParser(g, tokens, expr, nil)
	ast, err := parser.Parse()
	if err != nil {
		return "", err
	}

	rt := substrate.NewGoRuntime()
	ev := evaluator.New(rt, nil)
	env := value.NewEnv()

	result := ev.Eval(ast, env)
	if err, ok := result.(*value.Error); ok {
		return "", fmt.Errorf("%s", err.Message)
	}

	return result.Inspect(), nil
}
