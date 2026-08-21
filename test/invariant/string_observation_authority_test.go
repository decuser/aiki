package invariant

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStringObservationDoesNotMaterializeWholeRuneSlice(t *testing.T) {
	root := distributionRoot(t)

	checks := []struct {
		path  string
		funcs map[string]bool
	}{
		{
			path:  "engine/semantics/evaluator/access.go",
			funcs: map[string]bool{"evalStringIndex": true},
		},
		{
			path:  "engine/runtime/hal/substrate/builtins_list.go",
			funcs: map[string]bool{"halFirst": true, "halLength": true},
		},
		{
			path:  "engine/runtime/hal/substrate/builtins_type.go",
			funcs: map[string]bool{"halOrd": true},
		},
		{
			path: "engine/semantics/value/string_observation.go",
			funcs: map[string]bool{
				"RuneLen":      true,
				"RuneAt":       true,
				"FirstRune":    true,
				"CompareRunes": true,
			},
		},
	}

	for _, check := range checks {
		path := filepath.Join(root, filepath.FromSlash(check.path))
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", check.path, err)
		}
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, path, data, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", check.path, err)
		}

		seen := map[string]bool{}
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil || !check.funcs[fn.Name.Name] {
				continue
			}
			seen[fn.Name.Name] = true
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}
				array, ok := call.Fun.(*ast.ArrayType)
				if !ok {
					return true
				}
				ident, ok := array.Elt.(*ast.Ident)
				if ok && ident.Name == "rune" {
					pos := fset.Position(call.Pos())
					t.Errorf("%s:%d %s materializes []rune in immutable observation path",
						check.path, pos.Line, fn.Name.Name)
				}
				return true
			})
		}

		for name := range check.funcs {
			if !seen[name] {
				t.Errorf("%s: observation function %s not found", check.path, name)
			}
		}
	}
}

func TestNaturalStringOrderingUsesStringObservationAuthority(t *testing.T) {
	root := distributionRoot(t)
	path := filepath.Join(root, "engine", "semantics", "value", "order.go")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	source := string(data)
	if !strings.Contains(source, "return l.CompareRunes(r), true") {
		t.Fatalf("natural string ordering must delegate to String.CompareRunes")
	}
}
