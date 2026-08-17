package invariant

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestExactNumberCoreContainsNoFloatPath(t *testing.T) {
	if err := validateNoCoreFloatPath(loadCoreGoSources(t)); err != nil {
		t.Fatal(err)
	}
}

func TestExactNumberInvariantRejectsFloatReintroduction(t *testing.T) {
	sources := loadCoreGoSources(t)
	sources["engine/semantics/value/forbidden.go"] = "package value\nfunc forbiddenRegression(x float64) float64 { return x }\n"
	err := validateNoCoreFloatPath(sources)
	if err == nil || !strings.Contains(err.Error(), "float64") {
		t.Fatalf("expected float-path invariant failure, got %v", err)
	}
}

func loadCoreGoSources(t *testing.T) map[string]string {
	t.Helper()
	root := distributionRoot(t)
	out := map[string]string{}
	for _, relRoot := range []string{"engine/semantics", "engine/syntax"} {
		base := filepath.Join(root, filepath.FromSlash(relRoot))
		err := filepath.WalkDir(base, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			out[filepath.ToSlash(rel)] = string(data)
			return nil
		})
		if err != nil {
			t.Fatalf("loading %s: %v", relRoot, err)
		}
	}
	return out
}

func validateNoCoreFloatPath(sources map[string]string) error {
	var problems []string
	for path, source := range sources {
		f, err := parser.ParseFile(token.NewFileSet(), path, source, 0)
		if err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}
		ast.Inspect(f, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.Ident:
				if x.Name == "float32" || x.Name == "float64" {
					problems = append(problems, fmt.Sprintf("%s: forbidden float type %s", path, x.Name))
				}
			case *ast.SelectorExpr:
				if x.Sel.Name == "ParseFloat" || x.Sel.Name == "SetFloat64" || x.Sel.Name == "Float64" {
					problems = append(problems, fmt.Sprintf("%s: forbidden float conversion %s", path, x.Sel.Name))
				}
			}
			return true
		})
	}
	if len(problems) == 0 {
		return nil
	}
	sort.Strings(problems)
	return fmt.Errorf("exact-number architecture invariant failure:\n  %s", strings.Join(problems, "\n  "))
}
