package repl

import (
	"io"

	"aiki/lang/value"
)

// Run starts the REPL with the given environment.
func Run(in io.Reader, out io.Writer, env *value.Env, debug bool) {
	s := NewSession(out, env, debug)
	s.Run()
}
