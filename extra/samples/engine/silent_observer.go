package engine

// SilentObserver implements Observer with no-op methods.
type SilentObserver struct{}

func (s SilentObserver) OnLex(token string, lexeme string, pos Position)            {}
func (s SilentObserver) OnParse(production string, depth int, pos Position)         {}
func (s SilentObserver) OnEval(node string, result string, scope int, pos Position) {}
func (s SilentObserver) OnEffect(action string, target string, pos Position)        {}

var _ Observer = SilentObserver{}
