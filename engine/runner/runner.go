// Package runner provides the runner for executing Aiki programs.
package runner

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"aiki/engine/runtime/hal/substrate"
	"aiki/engine/runtime/help"
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

	// Initialize module registry
	if err := initModuleRegistry(g); err != nil {
		return fmt.Errorf("initializing registry: %w", err)
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
	ev.SetGrammar(g)
	result := ev.Eval(ast, userEnv)
	if err, ok := result.(*value.Error); ok {
		return fmt.Errorf("%s", err.Inspect())
	}

	return nil
}

// loadPrelude parses and evaluates the prelude.
func loadPrelude(g *grammar.Grammar, rt *substrate.GoRuntime, env *value.Env) error {
	// Initialize help registry
	if err := initHelpRegistry(); err != nil {
		return fmt.Errorf("loading help: %w", err)
	}

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
	ev.SetGrammar(g)
	result := ev.Eval(ast, env)
	if err, ok := result.(*value.Error); ok {
		return fmt.Errorf("%s", err.Inspect())
	}

	return nil
}

// initHelpRegistry loads prelude help and doc files with validation.
// initModuleRegistry creates and scans the module registry.
func initModuleRegistry(g *grammar.Grammar) error {
	// Default roots: current dir, lib/, vendor/, user lib
	homeDir, _ := os.UserHomeDir()
	roots := []string{
		".",
		"lib",
		"vendor",
	}
	if homeDir != "" {
		roots = append(roots, filepath.Join(homeDir, ".aiki", "lib"))
	}

	registry := substrate.NewModuleRegistry(roots)
	if err := registry.Scan(g); err != nil {
		return err
	}
	substrate.GlobalRegistry = registry
	return nil
}

func initHelpRegistry() error {
	registry := help.NewRegistry()

	funcs, err := help.ParseHelpFile("prelude.help", prelude.HelpSource)
	if err != nil {
		return err
	}

	docs, err := help.ParseDocFile("prelude.doc", prelude.DocSource)
	if err != nil {
		return err
	}

	// Extract function names from prelude source
	preludeFuncs := extractPreludeFuncs(prelude.Source)

	// Validate 1:1 match
	if err := validateHelpCoverage(preludeFuncs, funcs, "prelude"); err != nil {
		return err
	}

	registry.Merge(funcs, docs)
	substrate.HelpRegistry = registry

	return nil
}

// extractPreludeFuncs extracts top-level "let name = " bindings from prelude source.
// Only captures lets at column 0 (not indented).
func extractPreludeFuncs(source string) map[string]bool {
	names := make(map[string]bool)
	lines := strings.Split(source, "\n")
	
	for _, line := range lines {
		// Only top-level: must start with "let " (no leading whitespace)
		if !strings.HasPrefix(line, "let ") {
			continue
		}
		
		// Extract name: "let name = ..." or "let name = (..."
		rest := strings.TrimPrefix(line, "let ")
		// Find the name (ends at space or =)
		var name strings.Builder
		for _, r := range rest {
			if r == ' ' || r == '=' {
				break
			}
			name.WriteRune(r)
		}
		n := name.String()
		if n != "" && !strings.HasPrefix(n, "_") {
			names[n] = true
		}
	}
	return names
}

// validateHelpCoverage checks that every function has help and vice versa.
func validateHelpCoverage(funcs map[string]bool, helpEntries map[string]help.FuncEntry, source string) error {
	var errors []string

	// Check every function has help
	for name := range funcs {
		if _, ok := helpEntries[name]; !ok {
			errors = append(errors, fmt.Sprintf("missing help for '%s'", name))
		}
	}

	// Check no orphan help entries
	for name := range helpEntries {
		if !funcs[name] {
			errors = append(errors, fmt.Sprintf("orphan help entry '%s'", name))
		}
	}

	if len(errors) > 0 {
		sort.Strings(errors)
		return fmt.Errorf("%s help mismatch:\n  %s", source, strings.Join(errors, "\n  "))
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
	ev.SetGrammar(g)
	result := ev.Eval(ast, userEnv)
	if err, ok := result.(*value.Error); ok {
		return "", fmt.Errorf("%s", err.Inspect())
	}

	return result.Inspect(), nil
}
