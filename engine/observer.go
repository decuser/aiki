package engine

import "aiki/engine/semantics/value"

// Observer is notified of events during lexing, parsing, and evaluation.
type Observer interface {
	OnLex(token string, lexeme string, pos Position)
	OnParse(production string, depth int, pos Position)
	OnReady(substrate string, scope int)
	OnEval(op string, val value.Value, scope int, pos Position)
	OnEffect(action string, substrate string, pos Position)
}
