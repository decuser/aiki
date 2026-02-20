package main

import (
	"aiki/engine"
	"aiki/engine/semantics/value"
	"fmt"
)

type TraceObserver struct{}

func (t TraceObserver) OnLex(token string, lexeme string, pos engine.Position) {}

func (t TraceObserver) OnParse(production string, depth int, pos engine.Position) {}

func (t TraceObserver) OnReady(substrate string, scope int) {
	fmt.Printf("ready: substrate=%s scope=%d\n", substrate, scope)
}

func (t TraceObserver) OnEval(op string, val value.Value, scope int, pos engine.Position) {
	fmt.Printf("eval: %s (scope %d)\n", op, scope)
}

func (t TraceObserver) OnEffect(action string, substrate string, pos engine.Position) {}
