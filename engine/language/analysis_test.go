package language

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"strings"
	"testing"

	"aiki/engine/syntax"
	"aiki/engine/syntax/grammar"
)

func TestGenericTraversalVisitsUnknownWrapper(t *testing.T) {
	c := &checker{scopes: []scopeFrame{make(scopeFrame)}}
	root := &syntax.Node{Type: "future_production", Children: []*syntax.Node{{Type: "NAME", Value: "missing"}}}
	c.check(root)
	if len(c.diags) != 1 || c.diags[0].Severity != "error" || !strings.Contains(c.diags[0].Message, "missing") {
		t.Fatalf("generic traversal did not visit child of unknown wrapper: %v", c.diags)
	}
}

func TestASTTypeKnowledgeMatchesGrammar(t *testing.T) {
	g, err := grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
	if err != nil {
		t.Fatal(err)
	}
	known, err := lintASTTypeLiterals("analysis.go")
	if err != nil {
		t.Fatal(err)
	}
	allowed := g.Analysis().ASTNodeTypes()
	allowed["TERMINAL"] = struct{}{}
	for typ := range known {
		if _, ok := allowed[typ]; !ok {
			t.Errorf("language analysis names AST node type %q that grammar cannot produce", typ)
		}
	}
}

func lintASTTypeLiterals(filename string) (map[string]struct{}, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, nil, 0)
	if err != nil {
		return nil, err
	}
	out := make(map[string]struct{})
	stringLit := func(e ast.Expr) (string, bool) {
		lit, ok := e.(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return "", false
		}
		value, err := strconv.Unquote(lit.Value)
		if err != nil {
			return "", false
		}
		return value, true
	}
	isTypeSelector := func(e ast.Expr) bool { sel, ok := e.(*ast.SelectorExpr); return ok && sel.Sel.Name == "Type" }
	ast.Inspect(file, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.BinaryExpr:
			if x.Op != token.EQL && x.Op != token.NEQ {
				return true
			}
			if isTypeSelector(x.X) {
				if v, ok := stringLit(x.Y); ok {
					out[v] = struct{}{}
				}
			}
			if isTypeSelector(x.Y) {
				if v, ok := stringLit(x.X); ok {
					out[v] = struct{}{}
				}
			}
		case *ast.SwitchStmt:
			if x.Tag == nil || !isTypeSelector(x.Tag) {
				return true
			}
			for _, stmt := range x.Body.List {
				if clause, ok := stmt.(*ast.CaseClause); ok {
					for _, expr := range clause.List {
						if v, ok := stringLit(expr); ok {
							out[v] = struct{}{}
						}
					}
				}
			}
		case *ast.CallExpr:
			name := ""
			switch fun := x.Fun.(type) {
			case *ast.Ident:
				name = fun.Name
			case *ast.SelectorExpr:
				name = fun.Sel.Name
			}
			if name != "ChildByType" && name != "findAllByType" && name != "findFirstByType" {
				return true
			}
			for i := len(x.Args) - 1; i >= 0; i-- {
				if v, ok := stringLit(x.Args[i]); ok {
					out[v] = struct{}{}
					break
				}
			}
		}
		return true
	})
	return out, nil
}
