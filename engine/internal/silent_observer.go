package internal

import "aiki/engine"

// SilentObserver implements Observer with no-op methods.
type SilentObserver struct{}

func (s SilentObserver) OnLex(token string, lexeme string, pos engine.Position)            {}
func (s SilentObserver) OnParse(production string, depth int, pos engine.Position)         {}
func (s SilentObserver) OnEval(node string, result string, scope int, pos engine.Position) {}
func (s SilentObserver) OnEffect(action string, target string, pos engine.Position)        {}

var _ engine.Observer = SilentObserver{}
