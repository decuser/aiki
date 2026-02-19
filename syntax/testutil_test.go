package syntax

import (
	"os"
	"testing"
)

func LoadGrammarForTest(t *testing.T, path string) *Grammar {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read grammar %s: %v", path, err)
	}
	g, err := Parse(string(data))
	if err != nil {
		t.Fatalf("parse grammar %s: %v", path, err)
	}
	return g
}

