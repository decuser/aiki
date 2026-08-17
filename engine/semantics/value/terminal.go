package value

import "fmt"

// TerminalState is an opaque Aiki-visible token for restoring runtime-owned
// terminal state after a raw-mode transition.
type TerminalState struct{ ID uint64 }

func (t *TerminalState) Type() Type { return TerminalStateType }
func (t *TerminalState) Inspect() string {
	if t == nil || t.ID == 0 {
		return "<terminal-state>"
	}
	return fmt.Sprintf("<terminal-state:%d>", t.ID)
}
