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

func TestExactNumberCoreKeepsFloatAtNumberBoundary(t *testing.T) {
	if err := validateExactNumberBoundary(loadCoreGoSources(t)); err != nil {
		t.Fatal(err)
	}
}

func TestExactNumberInvariantRejectsFloatOutsideNumberAuthority(t *testing.T) {
	sources := loadCoreGoSources(t)
	sources["engine/semantics/evaluator/forbidden.go"] = "package evaluator\nfunc forbiddenRegression(x float64) float64 { return x + 1 }\n"
	err := validateExactNumberBoundary(sources)
	if err == nil || !strings.Contains(err.Error(), "float64") {
		t.Fatalf("expected float-boundary invariant failure, got %v", err)
	}
}

func TestExactNumberInvariantRejectsRoundedArithmeticPath(t *testing.T) {
	sources := loadCoreGoSources(t)
	sources["engine/semantics/value/number.go"] += `
func (n *Number) forbiddenRoundedAdd(other *Number) *Number {
	a, _ := n.Float64()
	b, _ := other.Float64()
	out, _ := NewNumberFromFloat64(a + b)
	return out
}
`
	err := validateExactNumberBoundary(sources)
	if err == nil || !strings.Contains(err.Error(), "forbiddenRoundedAdd") {
		t.Fatalf("expected rounded-arithmetic invariant failure, got %v", err)
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

func validateExactNumberBoundary(sources map[string]string) error {
	const numberPath = "engine/semantics/value/number.go"
	var problems []string
	for path, source := range sources {
		f, err := parser.ParseFile(token.NewFileSet(), path, source, 0)
		if err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}
		if path != numberPath {
			ast.Inspect(f, func(n ast.Node) bool {
				switch x := n.(type) {
				case *ast.Ident:
					if x.Name == "float32" || x.Name == "float64" {
						problems = append(problems, fmt.Sprintf("%s: float type %s outside Number authority", path, x.Name))
					}
				case *ast.SelectorExpr:
					if x.Sel.Name == "ParseFloat" || x.Sel.Name == "SetFloat64" || x.Sel.Name == "Float64" || x.Sel.Name == "Float64bits" || x.Sel.Name == "Float64frombits" {
						problems = append(problems, fmt.Sprintf("%s: float conversion %s outside Number authority", path, x.Sel.Name))
					}
				}
				return true
			})
			continue
		}

		// The Number authority may carry finite binary64 bits and cross explicit
		// host boundaries. Ordinary arithmetic methods may not use that path.
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			forbiddenArithmetic := fn.Name.Name == "Add" || fn.Name.Name == "Sub" || fn.Name.Name == "Mul" || fn.Name.Name == "Quo" || fn.Name.Name == "Neg"
			if strings.Contains(fn.Name.Name, "forbiddenRounded") {
				forbiddenArithmetic = true
			}
			if !forbiddenArithmetic {
				continue
			}
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				switch x := n.(type) {
				case *ast.Ident:
					if x.Name == "float32" || x.Name == "float64" {
						problems = append(problems, fmt.Sprintf("%s:%s: rounded numeric type %s in exact arithmetic", path, fn.Name.Name, x.Name))
					}
					// Ordinary arithmetic may decode an existing binary64 carrier and
					// pass it through a separately certified exact-operation helper. It
					// may not construct a Number from an arbitrary host float result.
					if x.Name == "NewNumberFromFloat64" {
						problems = append(problems, fmt.Sprintf("%s:%s: host float construction in exact arithmetic", path, fn.Name.Name))
					}
				case *ast.SelectorExpr:
					// Float64frombits/Float64bits are representation operations on the
					// exact dyadic carrier. They do not themselves round. Conversions
					// through Float64/SetFloat64 remain forbidden on ordinary exact
					// arithmetic paths.
					if x.Sel.Name == "Float64" || x.Sel.Name == "SetFloat64" || x.Sel.Name == "ParseFloat" {
						problems = append(problems, fmt.Sprintf("%s:%s: float conversion %s in exact arithmetic", path, fn.Name.Name, x.Sel.Name))
					}
				}
				return true
			})
		}
	}
	if len(problems) == 0 {
		return nil
	}
	sort.Strings(problems)
	return fmt.Errorf("exact-number architecture invariant failure:\n  %s", strings.Join(problems, "\n  "))
}
