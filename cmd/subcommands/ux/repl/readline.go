package repl

import (
	"io"

	"github.com/chzyer/readline"
)

type readlineReader struct {
	rl *readline.Instance
}

// NewReadlineReader creates a LineReader with readline support.
func NewReadlineReader() (LineReader, error) {
	rl, err := readline.NewEx(&readline.Config{
		HistoryFile:     "/tmp/aiki_history",
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		return nil, err
	}
	return &readlineReader{rl: rl}, nil
}

func (r *readlineReader) Prompt(prompt string) (string, error) {
	r.rl.SetPrompt(prompt)
	line, err := r.rl.Readline()
	if err == readline.ErrInterrupt {
		return "", ErrInterrupt
	}
	if err == io.EOF {
		return "", io.EOF
	}
	return line, err
}

func (r *readlineReader) Close() error {
	return r.rl.Close()
}
