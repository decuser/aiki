package debug

import (
	"fmt"
	"io"

	"aiki/engine"
)

// TraceObserver implements engine.Observer for debugging output.
type TraceObserver struct {
	out io.Writer
}

func (t *TraceObserver) OnLex(token string, lexeme string, pos engine.Position) {
	fmt.Fprintf(t.out, "[lex] %s:%d:%d %s %q\n", pos.File, pos.Line, pos.Col, token, lexeme)
}

func (t *TraceObserver) OnParse(production string, depth int, pos engine.Position) {
	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}
	fmt.Fprintf(t.out, "[parse] %s%s at %d:%d\n", indent, production, pos.Line, pos.Col)
}

func (t *TraceObserver) OnEval(node string, result string, scope int, pos engine.Position) {
	fmt.Fprintf(t.out, "[eval] %s => %s (scope %d) at %d:%d\n", node, result, scope, pos.Line, pos.Col)
}

func (t *TraceObserver) OnEffect(action string, target string, pos engine.Position) {
	fmt.Fprintf(t.out, "[effect] %s -> %s at %d:%d\n", action, target, pos.Line, pos.Col)
}
