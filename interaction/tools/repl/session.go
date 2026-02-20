package repl

import (
	"fmt"
	"io"

	"aiki/engine/semantics/evaluator"
	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
)

const (
	promptMain = "> "
	promptCont = "  "
)

// Session manages a REPL session.
type Session struct {
	out     io.Writer
	scope   *evaluator.Scope
	ev      *evaluator.Evaluator
	debug   bool
	reader  LineReader
	tracker *TrackingWriter
	grammar syntax.GrammarContract
}

// NewSession creates a new REPL session.
func NewSession(ev *evaluator.Evaluator, grammar syntax.GrammarContract, out io.Writer, scope *evaluator.Scope, debug bool) *Session {
	reader, err := NewReadlineReader()
	if err != nil {
		reader = NewSimpleReader()
	}
	tracker := &TrackingWriter{Out: out, EndedWithNewline: true}
	return &Session{
		ev:      ev,
		grammar: grammar,
		out:     out,
		scope:   scope,
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
		result := s.evalInput(buffer)
		buffer = ""
		prompt = promptMain

		// Check for special commands
		if result.Type == value.Symbol {
			if sym, ok := result.Data.(string); ok {
				switch sym {
				case "exit":
					fmt.Fprintln(s.out, "Goodbye!")
					return
				case "reset":
					// TODO: Reset environment
					fmt.Fprintln(s.out, "Environment reset.")
					continue
				}
			}
		}

		if result.Type != value.Null {
			fmt.Fprintln(s.out, result.Inspect())
		} else if !s.tracker.EndedWithNewline {
			fmt.Fprintln(s.out)
		}
	}
}

// evalInput parses and evaluates an input string.
func (s *Session) evalInput(input string) value.Value {
	lexer := syntax.NewLexer("repl", input, s.grammar)
	parser, err := syntax.NewParser(lexer, s.grammar)
	if err != nil {
		return value.Value{Type: value.String, Data: fmt.Sprintf("parse error: %v", err)}
	}

	ast, err := parser.Parse()
	if err != nil {
		return value.Value{Type: value.String, Data: fmt.Sprintf("parse error: %v", err)}
	}

	result, err := s.ev.Eval(ast, s.scope)
	if err != nil {
		return value.Value{Type: value.String, Data: fmt.Sprintf("error: %v", err)}
	}

	return result
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
