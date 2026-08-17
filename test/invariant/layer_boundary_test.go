package invariant

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestEngineSubstrateDependencyIsConfinedToCompositionRoot(t *testing.T) {
	sources := loadEngineProductionGo(t)
	if err := validateSubstrateImportBoundary(sources); err != nil {
		t.Fatal(err)
	}
}

func TestLayerInvariantRejectsSubstrateImportLeak(t *testing.T) {
	sources := loadEngineProductionGo(t)
	sources["engine/language/forbidden.go"] = "package language\nimport _ \"aiki/engine/runtime/hal/substrate\"\n"
	err := validateSubstrateImportBoundary(sources)
	if err == nil || !strings.Contains(err.Error(), "engine/language/forbidden.go") {
		t.Fatalf("expected substrate layer-boundary failure, got %v", err)
	}
}

func loadEngineProductionGo(t *testing.T) map[string]string {
	t.Helper()
	root := distributionRoot(t)
	base := filepath.Join(root, "engine")
	out := map[string]string{}
	err := filepath.WalkDir(base, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		out[filepath.ToSlash(rel)] = string(data)
		return nil
	})
	if err != nil {
		t.Fatalf("loading engine sources: %v", err)
	}
	return out
}

func validateSubstrateImportBoundary(sources map[string]string) error {
	const substrateImport = "aiki/engine/runtime/hal/substrate"
	var problems []string
	for path, source := range sources {
		f, err := parser.ParseFile(token.NewFileSet(), path, source, parser.ImportsOnly)
		if err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}
		importsSubstrate := false
		for _, imp := range f.Imports {
			if strings.Trim(imp.Path.Value, "\"") == substrateImport {
				importsSubstrate = true
				break
			}
		}
		if !importsSubstrate {
			continue
		}
		if strings.HasPrefix(path, "engine/runtime/hal/substrate/") || strings.HasPrefix(path, "engine/runner/") {
			continue
		}
		problems = append(problems, fmt.Sprintf("%s imports concrete HAL substrate outside composition root", path))
	}
	if len(problems) == 0 {
		return nil
	}
	sort.Strings(problems)
	return fmt.Errorf("engine layer invariant failure:\n  %s", strings.Join(problems, "\n  "))
}
