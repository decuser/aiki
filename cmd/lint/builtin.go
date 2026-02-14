package lint

import (
	"aiki/hal/core"
	"aiki/lang/value"
)

func init() {
	core.HAL["lint"] = &value.Builtin{
		Name: "lint",
		Fn:   builtinLint,
	}
}

func builtinLint(args ...value.Value) value.Value {
	if len(args) != 1 {
		return value.NewError("lint: want 1 argument, got %d", len(args))
	}

	source, ok := args[0].(*value.String)
	if !ok {
		return value.NewError("lint: expected string argument")
	}

	diags, err := LintString(source.Value)
	if err != nil {
		return value.NewError("lint: %s", err)
	}

	if len(diags) == 0 {
		return value.True
	}

	// Return list of diagnostic strings
	elements := make([]value.Value, len(diags))
	for i, d := range diags {
		elements[i] = &value.String{Value: d.Message}
	}
	return &value.List{Elements: elements}
}
