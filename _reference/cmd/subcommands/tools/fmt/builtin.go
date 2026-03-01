package fmt

import (
	gofmt "fmt"
	"os"

	"aiki/reference/semantics/eval"
	"aiki/reference/semantics/value"
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

	if grammar == nil {
		return value.NewError("fmt: grammar not initialized")
	}

	data, err := os.ReadFile(path.Value)
	if err != nil {
		return value.NewError("fmt: %s", err)
	}

	formatted, err := FormatSource(grammar, string(data))
	if err != nil {
		return value.NewError("fmt: %s", err)
	}

	if formatted != string(data) {
		gofmt.Println(path.Value)
		if err := os.WriteFile(path.Value, []byte(formatted), 0644); err != nil {
			return value.NewError("fmt: %s", err)
		}
	}

	return value.NewNumber(1, 1)
}
