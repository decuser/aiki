package fmt

import (
	"aiki/hal/core"
	"aiki/lang/value"
)

func init() {
	core.HAL["fmt"] = &value.Builtin{
		Name: "fmt",
		Fn:   builtinFmt,
	}
}

func builtinFmt(args ...value.Value) value.Value {
	if len(args) != 1 {
		return value.NewError("fmt: want 1 argument, got %d", len(args))
	}

	source, ok := args[0].(*value.String)
	if !ok {
		return value.NewError("fmt: expected string argument")
	}

	formatted, err := Format(source.Value)
	if err != nil {
		return value.NewError("fmt: %s", err)
	}
	return &value.String{Value: formatted}
}
