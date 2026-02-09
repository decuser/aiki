package main

import (
	_ "embed"
	"fmt"

	"aiki/eval"
	"aiki/value"
)

//go:embed prelude/prelude.ai
var preludeSource string

func loadPrelude(env *value.Env) error {
	result := eval.Run(preludeSource, env)
	if e, ok := result.(*value.Error); ok {
		return fmt.Errorf("%s", e.Message)
	}
	return nil
}
