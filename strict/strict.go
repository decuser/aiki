package strict

import (
	_ "embed"
)

//go:embed strict.ai
var Source string

func Exports() []string {
	return []string{
		"each", "map", "filter", "reduce", "range", "reverse",
		"find", "any", "all", "sum", "max", "min",
		"hash_new", "hash_get", "hash_put", "hash_has", "hash_del",
		"hash_keys", "hash_values", "hash_code", "println",
	}
}
