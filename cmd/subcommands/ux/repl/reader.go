package repl

import "errors"

// ErrInterrupt is returned when the user presses Ctrl-C.
var ErrInterrupt = errors.New("interrupt")

// LineReader abstracts line input with editing and history.
type LineReader interface {
	// Prompt displays the prompt and reads a line.
	// Returns the line (without newline) or error on EOF/interrupt.
	Prompt(prompt string) (string, error)

	// Close cleans up resources.
	Close() error
}
