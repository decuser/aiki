package eval

import "aiki/value"

// ResetSignal is returned by reset() for REPL to handle
type ResetSignal struct{}

func (r *ResetSignal) Type() value.Type { return "reset" }
func (r *ResetSignal) Inspect() string  { return "<reset>" }

var Reset = &ResetSignal{}

func init() {
	builtins["reset"] = &value.Builtin{
		Name: "reset",
		Fn: func(args ...value.Value) value.Value {
			if len(args) != 0 {
				return value.NewError("reset: want 0 arguments, got %d", len(args))
			}
			CloseAllCanvases()
			return Reset
		},
	}
}
