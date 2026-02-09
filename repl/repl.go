package repl

import (
	"fmt"
	"io"

	"github.com/peterh/liner"

	"aiki/eval"
	"aiki/value"
)

const prompt = "> "

func Start(in io.Reader, out io.Writer, env *value.Env, debug bool) {
	line := liner.NewLiner()
	defer line.Close()

	for {
		input, err := line.Prompt(prompt)
		if err == io.EOF {
			fmt.Fprintln(out, "\nGoodbye!")
			return
		}
		if err == liner.ErrPromptAborted {
			fmt.Fprintln(out)
			continue
		}
		if err != nil {
			fmt.Fprintln(out, "error:", err)
			return
		}

		if input == "" {
			continue
		}

		line.AppendHistory(input)

		result := eval.Run(input, env)

		if result != nil && result != value.NULL {
			fmt.Fprintln(out, result.Inspect())
		}
	}
}

