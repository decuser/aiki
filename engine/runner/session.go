package runner

import (
	"aiki/engine/runtime/hal/substrate"
	"aiki/engine/semantics/evaluator"
	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
	"aiki/engine/syntax/grammar"
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

	rt := substrate.NewGoRuntime()
	if err := initModuleRegistry(g, rt); err != nil {
		return nil, err
	}

	// Create prelude environment with ScopePrelude
	preludeEnv := value.NewEnvWithScope(value.ScopePrelude)
	preludeEnv.SetAuthority(rt.AuthorityForSource("engine/runtime/prelude/prelude.ai"))

	// Load prelude
	if err := loadPrelude(g, rt, preludeEnv); err != nil {
		return nil, err
	}

	// Create user environment enclosed by prelude
	userEnv := value.NewEnclosedEnvWithScope(preludeEnv, value.ScopeUser)
	userEnv.SetAuthority(value.NoAuthority())
	rt.SetUserEnv(userEnv)

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
		return value.NewFault("%s", err)
	}

	parser := syntax.NewParser(s.Grammar, tokens, source, nil)
	ast, err := parser.Parse()
	if err != nil {
		return value.NewFault("%s", err)
	}

	s.Env.SetFile("<repl>")
	s.Env.SetSource(source)
	return s.Evaluator.Eval(ast, s.Env)
}

// Reset creates a fresh environment with prelude reloaded.
func (s *Session) Reset() error {
	if err := initModuleRegistry(s.Grammar, s.Runtime); err != nil {
		return err
	}

	preludeEnv := value.NewEnvWithScope(value.ScopePrelude)
	preludeEnv.SetAuthority(s.Runtime.AuthorityForSource("engine/runtime/prelude/prelude.ai"))
	if err := loadPrelude(s.Grammar, s.Runtime, preludeEnv); err != nil {
		return err
	}
	s.Env = value.NewEnclosedEnvWithScope(preludeEnv, value.ScopeUser)
	s.Env.SetAuthority(value.NoAuthority())
	s.Runtime.SetUserEnv(s.Env)
	return nil
}
