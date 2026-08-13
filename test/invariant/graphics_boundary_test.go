package invariant

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Aiki's architecture places host facilities behind the HAL, so that the
// language itself does not depend on any particular host library. Graphics is
// the host facility with the largest dependency and the one most able to leak:
// the ebiten package fails in a package-level init when no display is
// available, which takes down any program that links it, including programs
// that never draw.
//
// These tests hold that boundary. They are also a measurement: as long as the
// graphics dependency reaches exactly one file, a headless configuration
// remains one build tag away rather than a refactor.

// graphicsImportPath is the host graphics library.
const graphicsImportPath = "github.com/hajimehoshi/ebiten"

// graphicsAllowedFile is the single file permitted to import it, relative to
// the distribution root.
const graphicsAllowedFile = "engine/runtime/hal/substrate/ebiten.go"

// languageCorePackages must not depend on the host graphics library at all.
// These are the packages that define and evaluate the language.
var languageCorePackages = []string{
	"engine/semantics/evaluator",
	"engine/semantics/value",
	"engine/syntax",
	"engine/syntax/grammar",
	"engine/runtime/hal",
	"engine/runtime/help",
}

// goFilesImporting returns the repository-relative paths of every Go file
// under dir whose import list mentions prefix. Test files are excluded: a test
// may legitimately reach for a runtime its package does not itself depend on.
// When recurse is false only files directly in dir are considered, so that a
// package is judged on its own imports rather than its subpackages'.
func goFilesImporting(t *testing.T, root, dir, prefix string, recurse bool) []string {
	t.Helper()

	var found []string
	fset := token.NewFileSet()
	base := filepath.Join(root, dir)

	err := filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if info.IsDir() {
			if !recurse && path != base {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		f, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, imp := range f.Imports {
			if strings.Contains(strings.Trim(imp.Path.Value, `"`), prefix) {
				rel, relErr := filepath.Rel(root, path)
				if relErr != nil {
					rel = path
				}
				found = append(found, filepath.ToSlash(rel))
				break
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking %s: %v", dir, err)
	}
	return found
}

// TestGraphicsDependencyConfined checks that the host graphics library is
// imported by one file only. A second importer would spread a dependency that
// cannot be initialized without a display, and would put a headless
// configuration out of reach.
func TestGraphicsDependencyConfined(t *testing.T) {
	root := distributionRoot(t)

	var importers []string
	for _, dir := range []string{"engine", "cmd", "test"} {
		importers = append(importers, goFilesImporting(t, root, dir, graphicsImportPath, true)...)
	}

	switch {
	case len(importers) == 0:
		t.Errorf("no file imports %s; if graphics moved, this test needs updating", graphicsImportPath)
	case len(importers) == 1 && importers[0] == graphicsAllowedFile:
		// The boundary holds.
	default:
		t.Errorf("%s must be imported only by %s, but is imported by: %s",
			graphicsImportPath, graphicsAllowedFile, strings.Join(importers, ", "))
	}
}

// TestLanguageCoreFreeOfGraphics checks that the packages defining and
// evaluating the language do not reach the host graphics library, directly or
// through the substrate package that contains it.
func TestLanguageCoreFreeOfGraphics(t *testing.T) {
	root := distributionRoot(t)

	for _, pkg := range languageCorePackages {
		for _, prefix := range []string{graphicsImportPath, "aiki/engine/runtime/hal/substrate"} {
			if files := goFilesImporting(t, root, pkg, prefix, false); len(files) > 0 {
				t.Errorf("%s must not depend on %s, but %s does",
					pkg, prefix, strings.Join(files, ", "))
			}
		}
	}
}
