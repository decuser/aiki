package lint

import (
	"aiki/semantics/eval"
	"aiki/semantics/value"
)

func init() {
	eval.HAL["lint"] = &value.Builtin{
		Name: "lint",
		Fn:   builtinLint,
	}
}

func builtinLint(args ...value.Value) value.Value {
	if len(args) != 1 {
		return value.NewError("lint: want 1 argument, got %d", len(args))
	}

	path, ok := args[0].(*value.String)
	if !ok {
		return value.NewError("lint: expected string argument")
	}

	count, err := Lint(path.Value)
	if err != nil {
		return value.NewError("lint: %s", err)
	}
	return value.NewNumber(int64(count), 1)
}
