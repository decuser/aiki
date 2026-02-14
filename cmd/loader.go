// cmd/loader.go
package main

import (
	"fmt"

	"aiki/ebnf"
	"aiki/lang/eval"
	"aiki/lang/value"
	"aiki/strict"
)

func loadStrict(grammar *ebnf.Grammar, env *value.Env) error {
	result := eval.RunNode(grammar, strict.Source, env)
	if e, ok := result.(*value.Error); ok {
		return fmt.Errorf("%s", e.Message)
	}
	env.SnapshotStrict()
	return nil
}
