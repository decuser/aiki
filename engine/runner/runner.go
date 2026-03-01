// Package runner provides the runner for executing Aiki programs.
package runner

import (
	"fmt"
	"os"

	"aiki/engine/runtime/hal/substrate"
	"aiki/engine/runtime/prelude"
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

	// Create runtime
	rt := substrate.NewGoRuntime()

	// Create prelude environment with ScopePrelude
	preludeEnv := value.NewEnvWithScope(value.ScopePrelude)

	// Load prelude
	if err := loadPrelude(g, rt, preludeEnv); err != nil {
		return fmt.Errorf("loading prelude: %w", err)
	}

	// Create user environment enclosed by prelude, with ScopeUser
	userEnv := value.NewEnclosedEnv(preludeEnv)

	// Lex user code
	lexer := syntax.NewLexer(g, filename, source, nil)
	tokens, err := lexer.Tokenize()
	if err != nil {
		return fmt.Errorf("lexer: %w", err)
	}

	// Parse user code
	parser := syntax.NewParser(g, tokens, source, nil)
	ast, err := parser.Parse()
	if err != nil {
		return fmt.Errorf("parser: %w", err)
	}

	// Eval user code
	userEnv.SetFile(filename)
	userEnv.SetSource(source)
	ev := evaluator.New(rt, nil)
	result := ev.Eval(ast, userEnv)
	if err, ok := result.(*value.Error); ok {
		return fmt.Errorf("%s", err.Inspect())
	}

	return nil
}

// loadPrelude parses and evaluates the prelude.
func loadPrelude(g *grammar.Grammar, rt *substrate.GoRuntime, env *value.Env) error {
	lexer := syntax.NewLexer(g, "<prelude>", prelude.Source, nil)
	tokens, err := lexer.Tokenize()
	if err != nil {
		return err
	}

	parser := syntax.NewParser(g, tokens, prelude.Source, nil)
	ast, err := parser.Parse()
	if err != nil {
		return err
	}

	env.SetFile("<prelude>")
	env.SetSource(prelude.Source)
	ev := evaluator.New(rt, nil)
	result := ev.Eval(ast, env)
	if err, ok := result.(*value.Error); ok {
		return fmt.Errorf("%s", err.Inspect())
	}

	return nil
}

// RunExpr evaluates a single expression and returns the result.
func RunExpr(expr string) (string, error) {
	g, err := grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
	if err != nil {
		return "", err
	}

	rt := substrate.NewGoRuntime()

	// Create prelude environment with ScopePrelude
	preludeEnv := value.NewEnvWithScope(value.ScopePrelude)

	// Load prelude
	if err := loadPrelude(g, rt, preludeEnv); err != nil {
		return "", fmt.Errorf("loading prelude: %w", err)
	}

	// Create user environment enclosed by prelude
	userEnv := value.NewEnclosedEnv(preludeEnv)

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

	ev := evaluator.New(rt, nil)
	result := ev.Eval(ast, userEnv)
	if err, ok := result.(*value.Error); ok {
		return "", fmt.Errorf("%s", err.Inspect())
	}

	return result.Inspect(), nil
}
