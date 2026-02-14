package repl

import (
	"io"

	"aiki/ebnf"
	"aiki/lang/value"
)

// Run starts the REPL with the given environment.
func Run(grammar *ebnf.Grammar, in io.Reader, out io.Writer, env *value.Env, debug bool) {
	s := NewSession(grammar, out, env, debug)
	s.Run()
}
