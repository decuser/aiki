package repl

import (
	"fmt"
	"io"

	"github.com/peterh/liner"

	"aiki/eval"
	"aiki/value"
)

const (
	promptMain = "> "
	promptCont = "  "
)

func Start(in io.Reader, out io.Writer, env *value.Env, debug bool) {
	line := liner.NewLiner()
	defer line.Close()

	line.SetCtrlCAborts(true)

	var buffer string
	prompt := promptMain

	for {
		input, err := line.Prompt(prompt)
		if err == io.EOF {
			fmt.Fprintln(out, "\nGoodbye!")
			return
		}
		if err == liner.ErrPromptAborted {
			// Ctrl+C: clear buffer and reset
			buffer = ""
			prompt = promptMain
			fmt.Fprintln(out)
			continue
		}
		if err != nil {
			fmt.Fprintln(out, "error:", err)
			return
		}

		// Accumulate input
		if buffer == "" {
			buffer = input
		} else {
			buffer = buffer + "\n" + input
		}

		// Check if input is complete (balanced delimiters)
		if !isComplete(buffer) {
			prompt = promptCont
			continue
		}

		// Skip empty input
		if isBlank(buffer) {
			buffer = ""
			prompt = promptMain
			continue
		}

		// Add to history (complete input only)
		line.AppendHistory(buffer)

		// Parse and evaluate
		result := eval.Run(buffer, env)

		// Reset for next input
		buffer = ""
		prompt = promptMain

		// Display result
		if result != nil && result != value.NULL {
			fmt.Fprintln(out, result.Inspect())
		}
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

		// Handle comments
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

		// Handle strings
		if r == '"' && !inString {
			inString = true
			continue
		}
		if r == '"' && inString {
			// Check for escape
			if i > 0 && runes[i-1] == '\\' {
				continue
			}
			inString = false
			continue
		}
		if inString {
			continue
		}

		// Count delimiters
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

	// Incomplete if still in string or any delimiter is unclosed
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
