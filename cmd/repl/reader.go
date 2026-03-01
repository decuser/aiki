package repl

import "errors"

var ErrInterrupt = errors.New("interrupt")

// LineReader reads lines for the REPL.
type LineReader interface {
	Prompt(prompt string) (string, error)
	Close()
}
