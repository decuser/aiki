package runner

import (
	"testing"

	"aiki/engine/runtime/hal/substrate"
	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
	"aiki/engine/syntax/grammar"
)

// This is a focused unit test for the same invariant that smoke exercises:
// prelude.help must fully cover the functions defined in prelude.ai.
func TestPreludeHelpRegistryComplete(t *testing.T) {
	g, err := grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
	if err != nil {
		t.Fatalf("loading grammar: %v", err)
	}

	// Make sure help registry init does not error, then use the same runtime
	// to load the prelude.
	rt := substrate.NewGoRuntime()
	if err := initHelpRegistry(g, rt); err != nil {
		t.Fatalf("initHelpRegistry: %v", err)
	}

	preludeEnv := value.NewEnvWithScope(value.ScopePrelude)
	preludeEnv.SetAuthority(rt.AuthorityForSource("engine/runtime/prelude/prelude.ai"))
	if err := loadPrelude(g, rt, preludeEnv); err != nil {
		t.Fatalf("loadPrelude: %v", err)
	}
}
