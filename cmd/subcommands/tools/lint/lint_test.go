package lint

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
	"aiki/engine/syntax/grammar"
)

func TestLint_CheckFormatting(t *testing.T) {
	dir := t.TempDir()
	p1 := filepath.Join(dir, "a.ai")
	// Unformatted: missing spaces and brace on next line.
	if err := os.WriteFile(p1, []byte("let x=1\nif x {\nreturn x\n}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	p2 := filepath.Join(dir, "b.ai")
	if err := os.WriteFile(p2, []byte("let y = 2\n"), 0644); err != nil {
		t.Fatal(err)
	}

	bad, err := CheckFormatting([]string{dir + "/..."}, false)
	if err != nil {
		t.Fatalf("CheckFormatting: %v", err)
	}
	if len(bad) != 1 {
		t.Fatalf("expected 1 bad file, got %d: %v", len(bad), bad)
	}
	if bad[0] != p1 {
		t.Fatalf("expected bad file %s, got %s", p1, bad[0])
	}
}

func lintSource(t *testing.T, src string) []Diagnostic {
	t.Helper()
	g, err := grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
	if err != nil {
		t.Fatal(err)
	}
	diags, err := LintSource(g, "test.ai", src, value.ScopeUser)
	if err != nil {
		t.Fatalf("LintSource error: %v", err)
	}
	return diags
}

func TestLintUseResolvesRegistryPackage(t *testing.T) {
	dir := t.TempDir()
	moduleDir := filepath.Join(dir, "packages", "list")
	if err := os.MkdirAll(moduleDir, 0755); err != nil {
		t.Fatal(err)
	}
	moduleSource := `package "list"
let map = (xs, f) { return xs }
export(:map)
`
	if err := os.WriteFile(filepath.Join(moduleDir, "list.ai"), []byte(moduleSource), 0644); err != nil {
		t.Fatal(err)
	}

	oldwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldwd) })

	g, err := grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
	if err != nil {
		t.Fatal(err)
	}
	source := `use("list")
let xs = [1, 2, 3]
let ys = map(xs, (x) { return x })
println(ys)
`
	diags, err := LintSource(g, filepath.Join(dir, "program.ai"), source, value.ScopeUser)
	if err != nil {
		t.Fatalf("LintSource error: %v", err)
	}
	for _, d := range diags {
		if d.Level == "error" && strings.Contains(d.Message, "undefined name: 'map'") {
			t.Fatalf("use(\"list\") did not bind registry export map: %v", diags)
		}
	}
}

func TestLintCleanCode(t *testing.T) {
	diags := lintSource(t, "let x = 5\nlet y = x + 1\n")
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
}

func TestLintUndefined(t *testing.T) {
	diags := lintSource(t, "let x = y + 1\n")
	found := false
	for _, d := range diags {
		if d.Level == "error" && strings.Contains(d.Message, "undefined") && strings.Contains(d.Message, "y") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected undefined y, got %v", diags)
	}
}

func TestLintNamingViolation(t *testing.T) {
	diags := lintSource(t, "let badName = 5\n")
	found := false
	for _, d := range diags {
		if d.Level == "warning" && strings.Contains(d.Message, "naming") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected naming warning, got %v", diags)
	}
}

func TestLintForwardReferenceTopLevel(t *testing.T) {
	diags := lintSource(t, "let b = a + 1\nlet a = 5\n")
	for _, d := range diags {
		if d.Level == "error" && strings.Contains(d.Message, "undefined") && strings.Contains(d.Message, "a") {
			t.Fatalf("forward reference should be allowed at top level, got %v", diags)
		}
	}
}

func TestLintShadowWarning(t *testing.T) {
	diags := lintSource(t, "let x = 5\nlet f = () {\n let x = 10\n x\n}\n")
	found := false
	for _, d := range diags {
		if d.Level == "warning" && strings.Contains(d.Message, "shadow") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected shadow warning, got %v", diags)
	}
}

func TestLintPreludeBuiltinsOK(t *testing.T) {
	diags := lintSource(t, "let x = length([1, 2, 3])\n")
	for _, d := range diags {
		if d.Level == "error" && strings.Contains(d.Message, "undefined") && strings.Contains(d.Message, "length") {
			t.Fatalf("prelude builtin length should be defined, got %v", diags)
		}
	}
}

func TestLintParamsNaming(t *testing.T) {
	diags := lintSource(t, "let f = (badParam) { return badParam }\n")
	found := false
	for _, d := range diags {
		if d.Level == "warning" && strings.Contains(d.Message, "badParam") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected param naming warning, got %v", diags)
	}
}

func TestLintMatchPatternBindsName(t *testing.T) {
	diags := lintSource(t, "match 1 { x { println(x) } }\n")
	for _, d := range diags {
		if d.Level == "error" && strings.Contains(d.Message, "undefined") && strings.Contains(d.Message, "x") {
			t.Fatalf("pattern binder x should be defined in arm, got %v", diags)
		}
	}
}

func TestLintIfBlockDoesNotScopeBindings(t *testing.T) {
	// Bindings inside if blocks are visible outside — the evaluator
	// does not create a new environment for if/while blocks.
	src := "if true {\n\tlet y = 42\n}\nlet z = y\n"
	diags := lintSource(t, src)
	for _, d := range diags {
		if d.Level == "error" && strings.Contains(d.Message, "undefined") && strings.Contains(d.Message, "y") {
			t.Fatalf("linter falsely reports y as undefined inside if block: %v", diags)
		}
	}
}

func TestLintWhileBlockDoesNotScopeBindings(t *testing.T) {
	src := "let i = 0\nwhile i < 1 {\n\tlet found = true\n\ti = i + 1\n}\nlet x = found\n"
	diags := lintSource(t, src)
	for _, d := range diags {
		if d.Level == "error" && strings.Contains(d.Message, "undefined") && strings.Contains(d.Message, "found") {
			t.Fatalf("linter falsely reports found as undefined after while block: %v", diags)
		}
	}
}

func TestLintMatchArmScopesBindings(t *testing.T) {
	// Match arms DO create enclosed environments — pattern bindings
	// should not leak outside.
	src := "let val = 42\nmatch val {\n\tx { let inner = x }\n}\nlet z = inner\n"
	diags := lintSource(t, src)
	found := false
	for _, d := range diags {
		if d.Level == "error" && strings.Contains(d.Message, "undefined") && strings.Contains(d.Message, "inner") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected inner to be undefined outside match arm, got %v", diags)
	}
}

func TestLintFunctionBodyScopesBindings(t *testing.T) {
	// Function bodies create their own scope.
	src := "let f = () {\n\tlet local = 1\n}\nlet x = local\n"
	diags := lintSource(t, src)
	found := false
	for _, d := range diags {
		if d.Level == "error" && strings.Contains(d.Message, "undefined") && strings.Contains(d.Message, "local") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected local to be undefined outside function, got %v", diags)
	}
}

func TestLintSelectBindingScopedToArm(t *testing.T) {
	src := "let ch = channel()\nselect {\n\tlet msg = recv(ch) { println(msg) }\n\tdefault { println(:idle) }\n}\nlet x = msg\n"
	diags := lintSource(t, src)
	found := false
	for _, d := range diags {
		if d.Level == "error" && strings.Contains(d.Message, "undefined") && strings.Contains(d.Message, "msg") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected select binding msg to be undefined outside arm, got %v", diags)
	}
}

func TestLintSelectChannelExprUsesOuterScope(t *testing.T) {
	src := "let ch = channel()\nselect {\n\trecv(ch) { println(:ok) }\n}\n"
	diags := lintSource(t, src)
	for _, d := range diags {
		if d.Level == "error" && strings.Contains(d.Message, "undefined") && strings.Contains(d.Message, "ch") {
			t.Fatalf("select channel expression should see outer ch, got %v", diags)
		}
	}
}

// TestLintASTTypeKnowledgeMatchesGrammar prevents the linter from silently
// accumulating stale grammar knowledge. The linter deliberately uses generic
// traversal for most productions, so it does not need one handler per grammar
// production. Every AST node type it does name, however, must be producible by
// the grammar or be an explicit parser-synthesized node.
func TestLintASTTypeKnowledgeMatchesGrammar(t *testing.T) {
	g, err := grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
	if err != nil {
		t.Fatal(err)
	}

	known, err := lintASTTypeLiterals("lint.go")
	if err != nil {
		t.Fatal(err)
	}
	allowed := grammarASTNodeTypes(g)
	allowed["TERMINAL"] = struct{}{} // parser-synthesized punctuation/literal node

	for typ := range known {
		if _, ok := allowed[typ]; !ok {
			t.Errorf("linter names AST node type %q that grammar cannot produce", typ)
		}
	}
}

// TestLintGenericTraversalVisitsUnknownWrapper establishes the other half of
// the linter coupling: a newly introduced production does not require a lint
// handler merely to make its subtree visible. Unknown wrapper nodes recurse by
// default, so ordinary checks still reach their children.
func TestLintGenericTraversalVisitsUnknownWrapper(t *testing.T) {
	c := &checker{scopes: []scopeFrame{make(scopeFrame)}}
	root := &syntax.Node{
		Type:     "future_production",
		Children: []*syntax.Node{{Type: "NAME", Value: "missing"}},
	}

	c.check(root)
	if len(c.diags) != 1 || c.diags[0].Level != "error" || !strings.Contains(c.diags[0].Message, "missing") {
		t.Fatalf("generic traversal did not visit child of unknown wrapper: %v", c.diags)
	}
}

func grammarASTNodeTypes(g *grammar.Grammar) map[string]struct{} {
	out := make(map[string]struct{}, len(g.Productions)+8)
	for name := range g.Productions {
		out[name] = struct{}{}
	}

	var walk func(grammar.Expression)
	walk = func(expr grammar.Expression) {
		switch e := expr.(type) {
		case *grammar.TokenRef:
			out[e.Name] = struct{}{}
		case *grammar.Sequence:
			for _, child := range e.Exprs {
				walk(child)
			}
		case *grammar.Alternative:
			for _, child := range e.Exprs {
				walk(child)
			}
		case *grammar.Repetition:
			walk(e.Expr)
		case *grammar.Option:
			walk(e.Expr)
		case *grammar.Group:
			walk(e.Expr)
		}
	}
	for _, prod := range g.Productions {
		walk(prod.Expr)
	}
	return out
}

// lintASTTypeLiterals extracts the string literals that lint.go uses as AST
// node-type knowledge: comparisons/switches on .Type, ChildByType calls, and
// the local recursive type-search helpers. Keeping the extraction in the test
// means adding a new hardcoded node name to the linter automatically enters
// the grammar-coupling check; there is no second manually maintained list.
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
	isTypeSelector := func(e ast.Expr) bool {
		sel, ok := e.(*ast.SelectorExpr)
		return ok && sel.Sel.Name == "Type"
	}

	ast.Inspect(file, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.BinaryExpr:
			if x.Op != token.EQL && x.Op != token.NEQ {
				return true
			}
			if isTypeSelector(x.X) {
				if s, ok := stringLit(x.Y); ok {
					out[s] = struct{}{}
				}
			}
			if isTypeSelector(x.Y) {
				if s, ok := stringLit(x.X); ok {
					out[s] = struct{}{}
				}
			}

		case *ast.SwitchStmt:
			if x.Tag == nil || !isTypeSelector(x.Tag) {
				return true
			}
			for _, stmt := range x.Body.List {
				clause, ok := stmt.(*ast.CaseClause)
				if !ok {
					continue
				}
				for _, expr := range clause.List {
					if s, ok := stringLit(expr); ok {
						out[s] = struct{}{}
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
				if s, ok := stringLit(x.Args[i]); ok {
					out[s] = struct{}{}
					break
				}
			}
		}
		return true
	})
	return out, nil
}

func TestRunFailsOnUndeclaredMalformedFile(t *testing.T) {
	dir := t.TempDir()
	bad := filepath.Join(dir, "bad.ai")
	good := filepath.Join(dir, "good.ai")
	if err := os.WriteFile(bad, []byte("let x =\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(good, []byte("let y = 2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if code := Run([]string{dir + "/..."}); code != 1 {
		t.Fatalf("Run exit = %d, want 1", code)
	}
}

func TestCheckFormattingSkipsDeclaredParseNegative(t *testing.T) {
	dir := t.TempDir()
	negative := filepath.Join(dir, "negative.ai")
	good := filepath.Join(dir, "good.ai")
	if err := os.WriteFile(negative, []byte("# @negative parse\nlet x =\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(good, []byte("let y = 2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	bad, err := CheckFormatting([]string{dir + "/..."}, false)
	if err != nil {
		t.Fatalf("declared parse-negative was not skipped: %v", err)
	}
	if len(bad) != 0 {
		t.Fatalf("unexpected formatting failures: %v", bad)
	}
	files, err := expandLintPaths([]string{dir + "/..."})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || files[0] != good {
		t.Fatalf("lint files = %v, want only %s", files, good)
	}
}

func TestCheckFormattingReportsAllUndeclaredMalformedFiles(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"a.ai", "b.ai"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("let x =\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	_, err := CheckFormatting([]string{dir + "/..."}, false)
	if err == nil {
		t.Fatal("expected malformed files to fail formatting preflight")
	}
	msg := err.Error()
	for _, name := range []string{"a.ai", "b.ai"} {
		if !strings.Contains(msg, name) {
			t.Fatalf("error %q does not mention %s", msg, name)
		}
	}
}
