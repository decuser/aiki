package repl
import "aiki/syntax"

import (
	"io"

	"aiki/semantics/value"
)

// Run starts the REPL with the given environment.
func Run(grammar *syntax.Grammar, in io.Reader, out io.Writer, env *value.Env, debug bool) {
	s := NewSession(grammar, out, env, debug)
	s.Run()
}
