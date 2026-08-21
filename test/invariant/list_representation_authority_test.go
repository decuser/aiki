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

func TestListElementsRemainReadOnlyOutsideValueAuthority(t *testing.T) {
	root := distributionRoot(t)
	var problems []string
	for _, relRoot := range []string{"engine", "cmd"} {
		base := filepath.Join(root, relRoot)
		err := filepath.WalkDir(base, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			rel = filepath.ToSlash(rel)
			if rel == "engine/semantics/value/value.go" {
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, path, data, 0)
			if err != nil {
				return err
			}
			ast.Inspect(f, func(n ast.Node) bool {
				assign, ok := n.(*ast.AssignStmt)
				if !ok {
					return true
				}
				for _, lhs := range assign.Lhs {
					if mutatesListElements(lhs) {
						pos := fset.Position(lhs.Pos())
						problems = append(problems, fmt.Sprintf("%s:%d direct List.Elements mutation outside value authority", rel, pos.Line))
					}
				}
				return true
			})
			return nil
		})
		if err != nil {
			t.Fatalf("walking %s: %v", relRoot, err)
		}
	}
	if len(problems) > 0 {
		sort.Strings(problems)
		t.Fatalf("list representation authority invariant failure:\n  %s", strings.Join(problems, "\n  "))
	}
}

func mutatesListElements(expr ast.Expr) bool {
	switch x := expr.(type) {
	case *ast.SelectorExpr:
		return x.Sel.Name == "Elements"
	case *ast.IndexExpr:
		selector, ok := x.X.(*ast.SelectorExpr)
		return ok && selector.Sel.Name == "Elements"
	case *ast.SliceExpr:
		selector, ok := x.X.(*ast.SelectorExpr)
		return ok && selector.Sel.Name == "Elements"
	}
	return false
}
