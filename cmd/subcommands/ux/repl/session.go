package repl

import (
	"fmt"
	"io"

	"aiki/engine/runner"
	"aiki/engine/semantics/value"
)

const (
	promptMain = "> "
	promptCont = "  "
)

var AppVersion string

// Session manages a REPL session.
type Session struct {
	out     io.Writer
	session *runner.Session
	debug   bool
	reader  LineReader
	tracker *TrackingWriter
}

// NewSession creates a new REPL session.
func NewSession(out io.Writer, debug bool) (*Session, error) {
	sess, err := runner.NewSession()
	if err != nil {
		return nil, err
	}

	var reader LineReader
	reader, err = NewReadlineReader()
	if err != nil {
		reader = NewSimpleReader()
	}

	tracker := &TrackingWriter{Out: out, EndedWithNewline: true}
	sess.Runtime.SetIO(nil, tracker)
	sess.Runtime.SetPageOutput(newPageOutput(out))

	return &Session{
		out:     out,
		session: sess,
		debug:   debug,
		reader:  reader,
		tracker: tracker,
	}, nil
}

// Run starts the REPL loop.
func (s *Session) Run() {
	defer s.reader.Close()
	defer s.session.Runtime.CloseAllResources()
	defer s.session.Runtime.SetPageOutput(nil)

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
		result := s.session.Eval(buffer)
		buffer = ""
		prompt = promptMain

		// Check for reset signal
		if _, ok := result.(*value.ResetSignal); ok {
			if err := s.session.Reset(); err != nil {
				fmt.Fprintf(s.out, "reset error: %v\n", err)
			} else {
				fmt.Fprintln(s.out, "Environment reset.")
			}
			continue
		}

		if _, ok := result.(*value.ExitSignal); ok {
			fmt.Fprintln(s.out, "Goodbye!")
			return
		}

		if result != nil && result != value.EMPTY {
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
