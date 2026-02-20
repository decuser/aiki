package internal

import (
	"aiki/engine"
	"aiki/engine/semantics/value"
)

type SilentObserver struct{}

func (s SilentObserver) OnLex(token string, lexeme string, pos engine.Position)             {}
func (s SilentObserver) OnParse(production string, depth int, pos engine.Position)          {}
func (s SilentObserver) OnReady(substrate string, scope int)                                {}
func (s SilentObserver) OnEval(op string, val value.Value, scope int, pos engine.Position)  {}
func (s SilentObserver) OnEffect(action string, substrate string, pos engine.Position)      {}

var _ engine.Observer = SilentObserver{}
