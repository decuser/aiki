package invariant

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"aiki/test/support/docexamples"
)

func TestLibModulesHaveDoc(t *testing.T) {
	for name, aiPath := range shippedModulePaths(t) {
		docPath := strings.TrimSuffix(aiPath, ".ai") + ".doc"
		if _, err := os.Stat(docPath); err != nil {
			t.Errorf("%s: no doc file at %s", name, filepath.Base(docPath))
		}
	}
}

func TestDocEntryDisposition(t *testing.T) {
	examples, err := docexamples.Load(distributionRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(examples) == 0 {
		t.Fatal("no doc examples found")
	}
	uncovered := 0
	for _, ex := range examples {
		if len(ex.Expects) == 0 && !ex.Unchecked {
			t.Errorf("%s.%s: no # expected comment and not marked @unchecked", ex.Module, ex.Name)
			uncovered++
		}
	}
	if uncovered == 0 {
		t.Logf("all %d doc entries have a disposition (checked or @unchecked)", len(examples))
	}
}
