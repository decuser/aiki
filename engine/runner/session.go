package runner

import (
	"aiki/engine/runtime/hal/substrate"
	"aiki/engine/semantics/evaluator"
	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
	"aiki/engine/syntax/grammar"
	"os"
	"path/filepath"
)

// Session holds the persistent state for a REPL session.
type Session struct {
	Grammar   *grammar.Grammar
	Runtime   *substrate.GoRuntime
	Evaluator *evaluator.Evaluator
	Env       *value.Env
}

// NewSession creates a new REPL session with prelude loaded.
func NewSession() (*Session, error) {
	g, err := grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
	if err != nil {
		return nil, err
	}

	if err := initModuleRegistryForSession(g); err != nil {
		return nil, err
	}
	rt := substrate.NewGoRuntime()

	// Create prelude environment with ScopePrelude
	preludeEnv := value.NewEnvWithScope(value.ScopePrelude)

	// Load prelude
	if err := loadPrelude(g, rt, preludeEnv); err != nil {
		return nil, err
	}

	// Create user environment enclosed by prelude
	userEnv := value.NewEnclosedEnv(preludeEnv)

	ev := evaluator.New(rt, nil)
	ev.SetGrammar(g)

	return &Session{
		Grammar:   g,
		Runtime:   rt,
		Evaluator: ev,
		Env:       userEnv,
	}, nil
}

// Eval evaluates source code and returns the result.
func (s *Session) Eval(source string) value.Value {
	lexer := syntax.NewLexer(s.Grammar, "<repl>", source, nil)
	tokens, err := lexer.Tokenize()
	if err != nil {
		return value.NewError("%s", err)
	}

	parser := syntax.NewParser(s.Grammar, tokens, source, nil)
	ast, err := parser.Parse()
	if err != nil {
		return value.NewError("%s", err)
	}

	return s.Evaluator.Eval(ast, s.Env)
}

// Reset creates a fresh environment with prelude reloaded.
func (s *Session) Reset() error {
	if err := initModuleRegistryForSession(s.Grammar); err != nil {
		return err
	}

	preludeEnv := value.NewEnvWithScope(value.ScopePrelude)
	if err := loadPrelude(s.Grammar, s.Runtime, preludeEnv); err != nil {
		return err
	}
	s.Env = value.NewEnclosedEnv(preludeEnv)
	return nil
}

func initModuleRegistryForSession(g *grammar.Grammar) error {
	if substrate.GlobalRegistry != nil {
		return nil
	}

	homeDir, _ := os.UserHomeDir()
	roots := []string{".", "lib", "vendor"}
	if homeDir != "" {
		roots = append(roots, filepath.Join(homeDir, ".aiki", "lib"))
	}

	reg := substrate.NewModuleRegistry(roots)
	if err := reg.Scan(g); err != nil {
		return err
	}
	substrate.GlobalRegistry = reg
	return nil
}
