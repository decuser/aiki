package debug

import (
	"aiki/engine"
	"aiki/engine/semantics/value"
	"fmt"
	"io"
)

// TraceObserver implements engine.Observer for debugging.
type TraceObserver struct {
	out io.Writer
}

func (t TraceObserver) OnLex(token string, lexeme string, pos engine.Position) {}

func (t TraceObserver) OnParse(production string, depth int, pos engine.Position) {}

func (t TraceObserver) OnReady(substrate string, scope int) {
	fmt.Fprintf(t.out, "ready: substrate=%s scope=%d\n", substrate, scope)
}

func (t TraceObserver) OnEval(op string, val value.Value, scope int, pos engine.Position) {
	fmt.Fprintf(t.out, "eval: %s (scope %d)\n", op, scope)
}

func (t TraceObserver) OnEffect(action string, substrate string, pos engine.Position) {}
