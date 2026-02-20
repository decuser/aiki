package repl

import (
	"io"

	"aiki/engine/semantics/evaluator"
	"aiki/engine/syntax"
)

// Run starts the REPL with the given environment.
func Run(ev *evaluator.Evaluator, grammar syntax.GrammarContract, out io.Writer, scope *evaluator.Scope, debug bool) {
	s := NewSession(ev, grammar, out, scope, debug)
	s.Run()
}
