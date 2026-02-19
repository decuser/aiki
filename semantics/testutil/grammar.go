package testutil

import (
	"os"
	"testing"

	"aiki/syntax"
)

func LoadGrammar(t *testing.T, path string) *syntax.Grammar {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read grammar %s: %v", path, err)
	}
	g, err := syntax.Parse(string(data))
	if err != nil {
		t.Fatalf("parse grammar %s: %v", path, err)
	}
	return g
}

