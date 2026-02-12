package core

import "aiki/lang/value"

var CloseAllCanvases func() // set by canvas package at init

type ResetSignal struct{}

func (r *ResetSignal) Type() value.Type { return "reset" }
func (r *ResetSignal) Inspect() string  { return "<reset>" }

var Reset = &ResetSignal{}

type ExitSignal struct{}

func (e *ExitSignal) Type() value.Type { return "exit" }
func (e *ExitSignal) Inspect() string  { return "" }

var Exit = &ExitSignal{}

func init() {
	HAL["reset"] = &value.Builtin{
		Name: "reset",
		Fn: func(args ...value.Value) value.Value {
			if len(args) != 0 {
				return value.NewError("reset: want 0 arguments, got %d", len(args))
			}
			if CloseAllCanvases != nil {
				CloseAllCanvases()
			}
			return Reset
		},
	}

	HAL["quit"] = &value.Builtin{
		Name: "quit",
		Fn: func(args ...value.Value) value.Value {
			if CloseAllCanvases != nil {
				CloseAllCanvases()
			}
			return Exit
		},
	}
}
