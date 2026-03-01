package repl

import (
	"github.com/chzyer/readline"
)

type ReadlineReader struct {
	rl *readline.Instance
}

func NewReadlineReader() (*ReadlineReader, error) {
	rl, err := readline.New("> ")
	if err != nil {
		return nil, err
	}
	return &ReadlineReader{rl: rl}, nil
}

func (r *ReadlineReader) Prompt(prompt string) (string, error) {
	r.rl.SetPrompt(prompt)
	line, err := r.rl.Readline()
	if err == readline.ErrInterrupt {
		return "", ErrInterrupt
	}
	return line, err
}

func (r *ReadlineReader) Close() {
	r.rl.Close()
}
