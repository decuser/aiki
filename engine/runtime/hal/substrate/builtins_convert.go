package substrate

import (
	"aiki/engine/semantics/value"
)

// halShape returns the shape of a list, or :list if unshaped.
func halShape(args []value.Value) value.Value {
	if len(args) != 1 {
		return value.NewError("shape: want 1 argument, got %d", len(args))
	}
	list, ok := args[0].(*value.List)
	if !ok {
		return &value.Symbol{Val: "list"}
	}
	if list.Shape == "" {
		return &value.Symbol{Val: "list"}
	}
	return &value.Symbol{Val: list.Shape}
}

// halToStr converts any value to a string.
func halToStr(args []value.Value) value.Value {
	if len(args) != 1 {
		return value.NewError("to_str: want 1 argument, got %d", len(args))
	}
	return &value.String{Val: args[0].Inspect()}
}
