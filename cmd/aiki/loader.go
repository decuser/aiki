// cmd/loader.go
package main

import "aiki/reference/syntax"

import (
	"fmt"

	"aiki/reference/runtime/prelude"
	"aiki/reference/semantics/eval"
	"aiki/reference/semantics/value"
)

func loadPrelude(grammar *syntax.Grammar, env *value.Env) error {
	result := eval.RunNode(grammar, prelude.Source, env)
	if e, ok := result.(*value.Error); ok {
		return fmt.Errorf("%s", e.Message)
	}
	env.SnapshotPrelude()
	return nil
}
