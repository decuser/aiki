package repl

import (
	"fmt"
	"io"

	"aiki/eval"
	"aiki/prelude"
	"aiki/value"
)

const (
	promptMain = "> "
	promptCont = "  "
)

func Start(in io.Reader, out io.Writer, env *value.Env, debug bool) {
	reader, err := NewReadlineReader()
	if err != nil {
		reader = NewSimpleReader()
	}
	defer reader.Close()

	var buffer string
	prompt := promptMain
	currentEnv := env

	for {
		line, err := reader.Prompt(prompt)
		if err == ErrInterrupt {
			buffer = ""
			prompt = promptMain
			fmt.Fprintln(out)
			continue
		}
		if err != nil {
			fmt.Fprintln(out, "\nGoodbye!")
			return
		}

		if buffer == "" {
			buffer = line
		} else {
			buffer = buffer + "\n" + line
		}

		if !isComplete(buffer) {
			prompt = promptCont
			continue
		}

		if isBlank(buffer) {
			buffer = ""
			prompt = promptMain
			continue
		}

		result := eval.Run(buffer, currentEnv)
		buffer = ""
		prompt = promptMain

		// Check for reset signal
		if _, ok := result.(*eval.ResetSignal); ok {
			currentEnv = value.NewEnv(nil)
			prelude.LoadPrelude(currentEnv)
			fmt.Fprintln(out, "Environment reset.")
			continue
		}

		if result != nil && result != value.NULL {
			fmt.Fprintln(out, result.Inspect())
		} else if !eval.LastPrintEndedWithNewline() {
			fmt.Fprintln(out)
		}
		eval.ResetLastPrint()
	}
}

// isComplete checks if the input has balanced delimiters
func isComplete(input string) bool {
	braces := 0
	brackets := 0
	parens := 0
	inString := false
	inComment := false

	runes := []rune(input)
	for i := 0; i < len(runes); i++ {
		r := runes[i]

		if r == '#' && !inString {
			inComment = true
			continue
		}
		if r == '\n' {
			inComment = false
			continue
		}
		if inComment {
			continue
		}

		if r == '"' && !inString {
			inString = true
			continue
		}
		if r == '"' && inString {
			if i > 0 && runes[i-1] == '\\' {
				continue
			}
			inString = false
			continue
		}
		if inString {
			continue
		}

		switch r {
		case '{':
			braces++
		case '}':
			braces--
		case '[':
			brackets++
		case ']':
			brackets--
		case '(':
			parens++
		case ')':
			parens--
		}
	}

	if inString {
		return false
	}

	return braces <= 0 && brackets <= 0 && parens <= 0
}

// isBlank checks if input is empty or only whitespace/comments
func isBlank(input string) bool {
	inComment := false
	for _, r := range input {
		if r == '#' {
			inComment = true
			continue
		}
		if r == '\n' {
			inComment = false
			continue
		}
		if inComment {
			continue
		}
		if r != ' ' && r != '\t' && r != '\n' && r != '\r' {
			return false
		}
	}
	return true
}
