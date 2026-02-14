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
	return value.NewError("lint: not yet implemented for new grammar")
}
