package debug

import (
	"fmt"
	"io"
	"strings"

	"aiki/engine"
)

// TraceObserver implements engine.Observer for debugging output.
type TraceObserver struct {
	out      io.Writer
	fileOnly string // if set, only trace this file
}

func (t *TraceObserver) OnLex(token string, lexeme string, pos engine.Position) {
	if t.fileOnly != "" && pos.File != t.fileOnly {
		return
	}
	fmt.Fprintf(t.out, "lex: %s %q at %d:%d\n", token, lexeme, pos.Line, pos.Col)
}

func (t *TraceObserver) OnParse(production string, depth int, pos engine.Position) {
	if t.fileOnly != "" && pos.File != t.fileOnly {
		return
	}
	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}
	fmt.Fprintf(t.out, "parse: %s%s at %d:%d\n", indent, production, pos.Line, pos.Col)
}

func (t *TraceObserver) OnEval(node string, result string, scope int, pos engine.Position) {
	if t.fileOnly != "" && pos.File != t.fileOnly {
		return
	}
	fmt.Fprintf(t.out, "eval: %s -> %s (scope %d) at %d:%d\n", node, result, scope, pos.Line, pos.Col)
}

func (t *TraceObserver) OnEffect(action string, target string, pos engine.Position) {
	if t.fileOnly != "" && pos.File != t.fileOnly {
		return
	}
	fmt.Fprintf(t.out, "effect: %s %s at %d:%d\n", action, target, pos.Line, pos.Col)
}

func (t *TraceObserver) OnFormat(method string, output string, node string, depth int) {
	indent := strings.Repeat("  ", depth)
	// Escape newlines/tabs for readability
	escaped := strings.ReplaceAll(output, "\n", "\\n")
	escaped = strings.ReplaceAll(escaped, "\t", "\\t")
	fmt.Fprintf(t.out, "fmt: %s%s %q [%s]\n", indent, method, escaped, node)
}
