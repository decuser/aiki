package fmt

import (
	"aiki/lang/eval"
	"aiki/lang/value"
)

func init() {
	eval.HAL["fmt"] = &value.Builtin{
		Name: "fmt",
		Fn:   builtinFmt,
	}
}

func builtinFmt(args ...value.Value) value.Value {
	if len(args) != 1 {
		return value.NewError("fmt: want 1 argument, got %d", len(args))
	}

	path, ok := args[0].(*value.String)
	if !ok {
		return value.NewError("fmt: expected string argument")
	}

	count, err := Format(path.Value)
	if err != nil {
		return value.NewError("fmt: %s", err)
	}
	return value.NewNumber(int64(count), 1)
}
