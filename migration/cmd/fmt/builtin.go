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
	return value.NewError("fmt: not yet implemented for new grammar")
}
