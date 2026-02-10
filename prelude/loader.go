package prelude

import (
	_ "embed"
	"fmt"

	"aiki/eval"
	"aiki/value"
)

//go:embed prelude.ai
var preludeSource string

func LoadPrelude(env *value.Env) error {
	// Load prelude into main env
	result := eval.Run(preludeSource, env)
	if e, ok := result.(*value.Error); ok {
		return fmt.Errorf("%s", e.Message)
	}

	// Snapshot current bindings for restore
	env.SnapshotPrelude()

	return nil
}
