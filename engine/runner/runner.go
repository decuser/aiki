// Package runner provides the runner for executing Aiki programs.
package runner

import (
	"fmt"
	"os"

	"aiki/engine"
	"aiki/engine/runtime/hal/substrate"
	"aiki/engine/runtime/libpath"
	"aiki/engine/runtime/modules"
	"aiki/engine/runtime/prelude"
	"aiki/engine/semantics/evaluator"
	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
	"aiki/engine/syntax/grammar"
)

// Run executes an Aiki source file. programArgs are exposed to the program
// through system.args(), excluding the interpreter and source filename.
func Run(filename string, programArgs ...string) error {
	rt := substrate.NewGoRuntime()
	defer rt.CloseAllResources()
	return RunWithRuntime(filename, rt, programArgs...)
}

// RunWithRuntime executes an Aiki source file in the supplied runtime host
// world. The caller owns the runtime's lifetime and cleanup.
func RunWithRuntime(filename string, rt *substrate.GoRuntime, programArgs ...string) error {
	source, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}
	return runSourceWithRuntime(filename, string(source), programArgs, rt)
}

// RunSource executes Aiki source code.
func RunSource(filename, source string) error {
	rt := substrate.NewGoRuntime()
	defer rt.CloseAllResources()
	return runSourceWithRuntime(filename, source, nil, rt)
}

func runSourceWithRuntime(filename, source string, programArgs []string, rt *substrate.GoRuntime) error {
	// Load grammar
	g, err := grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
	if err != nil {
		return fmt.Errorf("loading grammar: %w", err)
	}

	// Initialize the supplied runtime module registry.
	if err := initModuleRegistry(g, rt); err != nil {
		return fmt.Errorf("initializing registry: %w", err)
	}

	rt.SetProgramArgs(programArgs)

	// Create prelude environment with ScopePrelude
	preludeEnv := value.NewEnvWithScope(value.ScopePrelude)
	preludeEnv.SetAuthority(rt.AuthorityForSource("engine/runtime/prelude/prelude.ai"))

	// Load prelude
	if err := loadPrelude(g, rt, preludeEnv); err != nil {
		return fmt.Errorf("loading prelude: %w", err)
	}

	// Create user environment enclosed by prelude
	// Files in /lib/ or /contrib/lib/ retain ScopePrelude as a lexical/tooling role.
	// Raw runtime authority is assigned independently by AuthorityForSource.
	userScope := value.ScopeUser
	if libpath.IsBlessedLibPath(filename) {
		userScope = value.ScopePrelude
	}
	userEnv := value.NewEnclosedEnvWithScope(preludeEnv, userScope)
	userEnv.SetAuthority(rt.AuthorityForSource(filename))

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
	if fault, ok := result.(*value.Fault); ok {
		return fmt.Errorf("%s", fault.Inspect())
	}

	return nil
}

// RunWithCounters executes an Aiki source file with counter probes enabled.
// Returns the counters after execution for coverage/profiling analysis.
func RunWithCounters(filename string, counters *evaluator.Counters) (*evaluator.Counters, error) {
	rt := substrate.NewGoRuntime()
	defer rt.CloseAllResources()
	return RunWithCountersRuntime(filename, counters, rt)
}

// RunWithCountersRuntime executes a counter-instrumented Aiki source file in
// the supplied runtime. The caller owns the runtime's lifetime and cleanup.
func RunWithCountersRuntime(filename string, counters *evaluator.Counters, rt *substrate.GoRuntime) (*evaluator.Counters, error) {
	source, err := os.ReadFile(filename)
	if err != nil {
		return counters, fmt.Errorf("reading file: %w", err)
	}

	g, err := grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
	if err != nil {
		return counters, fmt.Errorf("loading grammar: %w", err)
	}

	if err := initModuleRegistry(g, rt); err != nil {
		return counters, fmt.Errorf("initializing registry: %w", err)
	}

	preludeEnv := value.NewEnvWithScope(value.ScopePrelude)
	preludeEnv.SetAuthority(rt.AuthorityForSource("engine/runtime/prelude/prelude.ai"))
	if err := loadPrelude(g, rt, preludeEnv); err != nil {
		return counters, fmt.Errorf("loading prelude: %w", err)
	}

	userScope := value.ScopeUser
	if libpath.IsBlessedLibPath(filename) {
		userScope = value.ScopePrelude
	}
	userEnv := value.NewEnclosedEnvWithScope(preludeEnv, userScope)
	userEnv.SetAuthority(rt.AuthorityForSource(filename))

	lexer := syntax.NewLexer(g, filename, string(source), nil)
	tokens, err := lexer.Tokenize()
	if err != nil {
		return counters, fmt.Errorf("lexer: %w", err)
	}

	parser := syntax.NewParser(g, tokens, string(source), nil)
	ast, err := parser.Parse()
	if err != nil {
		return counters, fmt.Errorf("parser: %w", err)
	}

	userEnv.SetFile(filename)
	userEnv.SetSource(string(source))
	ev := evaluator.New(rt, nil)
	ev.SetGrammar(g)
	ev.Counters = counters
	result := ev.Eval(ast, userEnv)
	if fault, ok := result.(*value.Fault); ok {
		return counters, fmt.Errorf("%s", fault.Inspect())
	}

	return counters, nil
}

// loadPrelude parses and evaluates the prelude.
func loadPrelude(g *grammar.Grammar, rt *substrate.GoRuntime, env *value.Env) error {
	// Initialize help registry
	if err := initHelpRegistry(g, rt); err != nil {
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
	if fault, ok := result.(*value.Fault); ok {
		return fmt.Errorf("%s", fault.Inspect())
	}

	return nil
}

// initHelpRegistry loads prelude help and doc files with validation.
// initModuleRegistry creates and scans the module registry.
func initModuleRegistry(g *grammar.Grammar, rt *substrate.GoRuntime) error {
	homeDir, _ := os.UserHomeDir()
	registry := modules.NewModuleRegistry(modules.DefaultModuleRoots(homeDir))
	if err := registry.Scan(g); err != nil {
		return err
	}
	rt.SetModuleRegistry(registry)
	return nil
}

func initHelpRegistry(g *grammar.Grammar, rt *substrate.GoRuntime) error {
	catalog, err := prelude.LoadCatalog(g)
	if err != nil {
		return err
	}
	rt.SetHelpRegistry(catalog.Registry)
	return nil
}

// RunExpr evaluates a single expression and returns the result.
func RunExpr(expr string) (string, error) {
	g, err := grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
	if err != nil {
		return "", err
	}

	rt := substrate.NewGoRuntime()
	defer rt.CloseAllResources()
	if err := initModuleRegistry(g, rt); err != nil {
		return "", fmt.Errorf("initializing registry: %w", err)
	}

	// Create prelude environment with ScopePrelude
	preludeEnv := value.NewEnvWithScope(value.ScopePrelude)
	preludeEnv.SetAuthority(rt.AuthorityForSource("engine/runtime/prelude/prelude.ai"))

	// Load prelude
	if err := loadPrelude(g, rt, preludeEnv); err != nil {
		return "", fmt.Errorf("loading prelude: %w", err)
	}

	// Create user environment enclosed by prelude
	userEnv := value.NewEnclosedEnvWithScope(preludeEnv, value.ScopeUser)
	userEnv.SetAuthority(value.NoAuthority())

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
	if fault, ok := result.(*value.Fault); ok {
		return "", fmt.Errorf("%s", fault.Inspect())
	}

	return result.Inspect(), nil
}

// RunProfile executes an Aiki source file with semantic profiling enabled.
func RunProfile(filename string, attributed bool) (engine.SemanticMeasurement, error) {
	run, err := RunProfileDetailed(filename, ProfileOptions{Attributed: attributed})
	return run.Semantic, err
}
