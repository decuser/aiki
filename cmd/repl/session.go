package repl

import (
	"fmt"
	"io"

	"aiki/internal/ebnf"
	"aiki/lang/eval"
	"aiki/lang/value"
	"aiki/runtime/hal"
	"aiki/runtime/prelude"
)

const (
	promptMain = "> "
	promptCont = "  "
)

// Session manages a REPL session.
type Session struct {
	out     io.Writer
	env     *value.Env
	debug   bool
	reader  LineReader
	tracker *TrackingWriter
	grammar *ebnf.Grammar
}

// NewSession creates a new REPL session.
func NewSession(grammar *ebnf.Grammar, out io.Writer, env *value.Env, debug bool) *Session {
	reader, err := NewReadlineReader()
	if err != nil {
		reader = NewSimpleReader()
	}
	tracker := &TrackingWriter{Out: out, EndedWithNewline: true}
	hal.Stdout = tracker
	return &Session{
		grammar: grammar,
		out:     out,
		env:     env,
		debug:   debug,
		reader:  reader,
		tracker: tracker,
	}
}

// Run starts the REPL loop.
func (s *Session) Run() {
	defer s.reader.Close()

	var buffer string
	prompt := promptMain

	for {
		line, err := s.reader.Prompt(prompt)
		if err == ErrInterrupt {
			buffer = ""
			prompt = promptMain
			continue
		}
		if err != nil {
			fmt.Fprintln(s.out, "\nGoodbye!")
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

		s.tracker.EndedWithNewline = true
		result := eval.RunNode(s.grammar, buffer, s.env)
		buffer = ""
		prompt = promptMain

		// Check for reset signal
		if _, ok := result.(*hal.ResetSignal); ok {
			s.env = value.NewEnv(nil)
			eval.RunNode(s.grammar, prelude.Source, s.env)
			s.env.SnapshotStrict()
			fmt.Fprintln(s.out, "Environment reset.")
			continue
		}

		if _, ok := result.(*hal.ExitSignal); ok {
			fmt.Fprintln(s.out, "Goodbye!")
			return
		}
		if result != nil && result != value.NULL {
			fmt.Fprintln(s.out, result.Inspect())
		} else if !s.tracker.EndedWithNewline {
			fmt.Fprintln(s.out)
		}
	}
}

// isComplete checks if the input has balanced delimiters.
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

// isBlank checks if input is empty or only whitespace/comments.
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
