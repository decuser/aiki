// cmd/loader.go
package main

import "aiki/syntax"

import (
	"fmt"

	"aiki/runtime/prelude"
	"aiki/semantics/eval"
	"aiki/semantics/value"
)

func loadPrelude(grammar *syntax.Grammar, env *value.Env) error {
	result := eval.RunNode(grammar, prelude.Source, env)
	if e, ok := result.(*value.Error); ok {
		return fmt.Errorf("%s", e.Message)
	}
	env.SnapshotPrelude()
	return nil
}
