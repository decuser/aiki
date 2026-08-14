package lint

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"aiki/engine"
	"aiki/engine/runtime/hal/substrate"
	"aiki/engine/runtime/prelude"
	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
	"aiki/engine/syntax/grammar"
)

// Diagnostic represents a lint finding.
type Diagnostic struct {
	Pos     engine.Position
	Level   string // "error" or "warning"
	Message string
}

// LintSource checks source for undefined names, shadowing, and naming conventions.
// It is structural only: it does not evaluate code.
func LintSource(g *grammar.Grammar, file string, source string, lintScope value.Scope) ([]Diagnostic, error) {
	// Parse using the engine pipeline.
	lx := syntax.NewLexer(g, file, source, nil)
	toks, err := lx.Tokenize()
	if err != nil {
		return nil, err
	}
	p := syntax.NewParser(g, toks, source, nil)
	node, err := p.Parse()
	if err != nil {
		return nil, err
	}

	// Build a set of all HAL builtins that exist in prelude scope.
	// Linting should be grounded only in what the runtime exposes for the active scope.
	c := &checker{
		scopes:      []scopeFrame{makeGlobals(lintScope)},
		diags:       nil,
		grammar:     g,
		currentFile: file,
	}
	c.collectTopLevel(node)
	c.check(node)
	return c.diags, nil
}

type scopeFrame map[string]bool

type checker struct {
	scopes      []scopeFrame
	diags       []Diagnostic
	grammar     *grammar.Grammar
	currentFile string
}

var snakeRe = regexp.MustCompile(`^_?[a-z][a-z0-9_]*$`)
var screamRe = regexp.MustCompile(`^_?[A-Z][A-Z0-9_]*$`)

func isValidCase(name string) bool {
	if name == "_" {
		return true
	}
	return snakeRe.MatchString(name) || screamRe.MatchString(name)
}

func makeGlobals(lintScope value.Scope) scopeFrame {
	globals := make(scopeFrame)
	// Prelude exports are the user visible surface.
	for name := range extractPreludeLets(prelude.Source) {
		globals[name] = true
	}
	// HAL builtins are discoverable from the runtime substrate.
	rt := substrate.NewGoRuntime()
	for _, name := range rt.BuiltinNames(lintScope) {
		globals[name] = true
	}
	// Keywords that appear as identifiers.
	globals["true"] = true
	globals["false"] = true
	globals["not"] = true
	globals["and"] = true
	globals["or"] = true
	return globals
}

func extractPreludeLets(source string) map[string]bool {
	// This intentionally matches the runner's extraction logic: find top level "let name".
	lets := make(map[string]bool)
	lines := strings.Split(source, "\n")
	for _, ln := range lines {
		trim := strings.TrimSpace(ln)
		if !strings.HasPrefix(trim, "let ") {
			continue
		}
		rest := strings.TrimSpace(strings.TrimPrefix(trim, "let "))
		// name ends at first space or '='.
		name := rest
		for i := 0; i < len(rest); i++ {
			if rest[i] == ' ' || rest[i] == '=' {
				name = rest[:i]
				break
			}
		}
		name = strings.TrimSpace(name)
		if name != "" && isNameToken(name) {
			lets[name] = true
		}
	}
	return lets
}

func isNameToken(s string) bool {
	if s == "" {
		return false
	}
	// mirror NAME token shape: [a-zA-Z_][a-zA-Z0-9_]*
	ch := s[0]
	if !(ch == '_' || (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')) {
		return false
	}
	for i := 1; i < len(s); i++ {
		c := s[i]
		if !(c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
			return false
		}
	}
	return true
}

func (c *checker) pushScope() { c.scopes = append(c.scopes, make(scopeFrame)) }

func (c *checker) popScope() {
	if len(c.scopes) > 1 {
		c.scopes = c.scopes[:len(c.scopes)-1]
	}
}

func (c *checker) define(name string) { c.scopes[len(c.scopes)-1][name] = true }

func (c *checker) isDefined(name string) bool {
	for i := len(c.scopes) - 1; i >= 0; i-- {
		if c.scopes[i][name] {
			return true
		}
	}
	return false
}

func (c *checker) isShadowing(name string) bool {
	for i := len(c.scopes) - 2; i >= 0; i-- {
		if c.scopes[i][name] {
			return true
		}
	}
	return false
}

func (c *checker) add(level string, pos engine.Position, msg string) {
	c.diags = append(c.diags, Diagnostic{Pos: pos, Level: level, Message: msg})
}

func (c *checker) collectTopLevel(node *syntax.Node) {
	if node == nil || node.Type != "program" {
		return
	}
	for _, st := range node.Children {
		if st.Type != "statement" || len(st.Children) == 0 {
			continue
		}
		s := st.Children[0]
		if s.Type != "let_stmt" {
			continue
		}
		// let NAME = expr  |  let SHAPE [ ... ]
		for _, ch := range s.Children {
			if ch.Type == "NAME" {
				c.define(ch.Value)
				break
			}
			if ch.Type == "SHAPE" {
				c.define(ch.Value)
				break
			}
		}
	}
}

func (c *checker) check(node *syntax.Node) {
	if node == nil {
		return
	}

	switch node.Type {
	case "program", "statement", "expr_stmt", "expr", "pipe_expr", "infix_expr", "unary_expr", "primary", "list_literal", "call", "index":
		for _, ch := range node.Children {
			c.check(ch)
		}
		return

	case "postfix_expr":
		c.checkPostfix(node)
		return

	case "let_stmt":
		c.checkLet(node)
		return

	case "assign_stmt":
		c.checkAssign(node)
		return

	case "if_stmt":
		c.checkIf(node)
		return

	case "while_stmt":
		c.checkWhile(node)
		return

	case "match_stmt":
		c.checkMatch(node)
		return

	case "return_stmt":
		for _, ch := range node.Children {
			if ch.Type != "TERMINAL" {
				c.check(ch)
			}
		}
		return

	case "block":
		// Blocks do not introduce a new scope in the evaluator.
		// Only match arms create enclosed environments (handled in
		// checkMatch). If/while/plain blocks execute in the
		// surrounding scope, so bindings made inside are visible
		// outside.
		for _, ch := range node.Children {
			c.check(ch)
		}
		return

	case "func_literal":
		c.checkFunc(node)
		return

	case "access":
		c.checkAccess(node)
		return

	case "NAME":
		// Bare NAME usage in expression context.
		if !c.isDefined(node.Value) {
			c.add("error", node.Pos, "undefined name: '"+node.Value+"'")
		}
		return

	default:
		for _, ch := range node.Children {
			c.check(ch)
		}
	}
}

func (c *checker) checkPostfix(node *syntax.Node) {
	// postfix_expr = primary { call | index | access }
	// Handle import(...) as a special form that introduces bindings.
	// Example: import("math/exact", :sin, :cos)
	if node == nil || len(node.Children) == 0 {
		return
	}

	prim := node.Children[0]
	nameTok := prim.ChildByType("NAME")
	if nameTok != nil && nameTok.Value == "import" {
		for _, ch := range node.Children[1:] {
			if ch.Type != "call" {
				continue
			}
			c.bindImportCall(ch)
			break
		}
	}

	// Handle use("module") - bind all exports from module
	if nameTok != nil && nameTok.Value == "use" {
		for _, ch := range node.Children[1:] {
			if ch.Type != "call" {
				continue
			}
			c.bindUseCall(ch)
			break
		}
	}

	for _, ch := range node.Children {
		c.check(ch)
	}
}

func (c *checker) bindImportCall(call *syntax.Node) {
	if call == nil {
		return
	}
	// call = "(" [ expr { "," expr } ] ")"
	// Import list items are SYMBOL tokens like :sin, :cos. Note: SYMBOL is not
	// necessarily a direct child of expr, so we must search recursively.
	for _, ch := range call.Children {
		if ch.Type != "expr" {
			continue
		}
		for _, sym := range findAllByType(ch, "SYMBOL") {
			name := strings.TrimPrefix(sym.Value, ":")
			if name != "" && isNameToken(name) {
				c.define(name)
			}
		}
	}
}

// bindUseCall handles use("module") by loading module exports into scope.
func (c *checker) bindUseCall(call *syntax.Node) {
	if call == nil {
		return
	}
	// call = "(" [ expr { "," expr } ] ")"
	// First argument should be a string literal with the module path.
	for _, ch := range call.Children {
		if ch.Type != "expr" {
			continue
		}
		// Find the STRING token
		strTok := findFirstByType(ch, "STRING")
		if strTok == nil {
			continue
		}
		// Extract module path from string literal (remove quotes)
		modulePath := strings.Trim(strTok.Value, "\"")
		if modulePath == "" {
			continue
		}
		// Resolve and load module exports
		exports := c.resolveModuleExports(modulePath)
		for _, name := range exports {
			c.define(name)
		}
		return // use() takes only one argument
	}
}

// resolveModuleExports finds a module file and extracts its exported names.
func (c *checker) resolveModuleExports(moduleName string) []string {
	// Try to find the module file using the same resolution as runtime
	modulePath := c.resolveModulePath(moduleName)
	if modulePath == "" {
		return nil
	}

	// Read and parse the module to find export() call
	data, err := os.ReadFile(modulePath)
	if err != nil {
		return nil
	}

	return extractExportsFromSource(string(data), c.grammar)
}

// resolveModulePath finds the .ai file for a module name.
// Mirrors the runtime resolution logic.
func (c *checker) resolveModulePath(name string) string {
	// Try relative to current file
	if c.currentFile != "" && c.currentFile != "<unknown>" {
		dir := filepath.Dir(c.currentFile)
		candidate := filepath.Join(dir, name+".ai")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	// Try as-is with .ai extension
	candidate := name + ".ai"
	if _, err := os.Stat(candidate); err == nil {
		return candidate
	}

	// Try lib/ directory relative to cwd
	libCandidate := filepath.Join("lib", name+".ai")
	if _, err := os.Stat(libCandidate); err == nil {
		return libCandidate
	}

	// Try lib/ directory relative to executable
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		libCandidate := filepath.Join(exeDir, "lib", name+".ai")
		if _, err := os.Stat(libCandidate); err == nil {
			return libCandidate
		}
	}

	// Try without extension
	if _, err := os.Stat(name); err == nil {
		return name
	}

	return ""
}

// extractExportsFromSource parses source and finds export(:name, ...) symbols.
func extractExportsFromSource(source string, g *grammar.Grammar) []string {
	if g == nil {
		// Fallback: simple text extraction
		return extractExportsSimple(source)
	}

	// Parse the module
	lx := syntax.NewLexer(g, "<module>", source, nil)
	toks, err := lx.Tokenize()
	if err != nil {
		return extractExportsSimple(source)
	}
	p := syntax.NewParser(g, toks, source, nil)
	node, err := p.Parse()
	if err != nil {
		return extractExportsSimple(source)
	}

	// Find export(...) call and extract symbols
	return findExportSymbols(node)
}

// findExportSymbols walks the AST to find export() calls and extract symbol names.
func findExportSymbols(node *syntax.Node) []string {
	if node == nil {
		return nil
	}

	var exports []string

	// Look for postfix_expr with "export" as the function name
	if node.Type == "postfix_expr" && len(node.Children) > 0 {
		prim := node.Children[0]
		nameTok := prim.ChildByType("NAME")
		if nameTok != nil && nameTok.Value == "export" {
			// Found export call - extract symbols from call arguments
			for _, ch := range node.Children[1:] {
				if ch.Type != "call" {
					continue
				}
				for _, sym := range findAllByType(ch, "SYMBOL") {
					name := strings.TrimPrefix(sym.Value, ":")
					if name != "" && isNameToken(name) {
						exports = append(exports, name)
					}
				}
			}
			return exports
		}
	}

	// Recurse into children
	for _, ch := range node.Children {
		exports = append(exports, findExportSymbols(ch)...)
	}
	return exports
}

// extractExportsSimple is a fallback text-based extraction.
func extractExportsSimple(source string) []string {
	var exports []string
	// Look for export(:name, :name2, ...)
	lines := strings.Split(source, "\n")
	for _, line := range lines {
		trim := strings.TrimSpace(line)
		if !strings.HasPrefix(trim, "export(") {
			continue
		}
		// Extract symbols from the line
		start := strings.Index(trim, "(")
		end := strings.LastIndex(trim, ")")
		if start < 0 || end < 0 || end <= start {
			continue
		}
		args := trim[start+1 : end]
		parts := strings.Split(args, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, ":") {
				name := strings.TrimPrefix(part, ":")
				if isNameToken(name) {
					exports = append(exports, name)
				}
			}
		}
	}
	return exports
}

func findAllByType(n *syntax.Node, typ string) []*syntax.Node {
	if n == nil {
		return nil
	}
	var out []*syntax.Node
	if n.Type == typ {
		out = append(out, n)
	}
	for _, ch := range n.Children {
		out = append(out, findAllByType(ch, typ)...)
	}
	return out
}

func findFirstByType(n *syntax.Node, typ string) *syntax.Node {
	if n == nil {
		return nil
	}
	if n.Type == typ {
		return n
	}
	for _, ch := range n.Children {
		if found := findFirstByType(ch, typ); found != nil {
			return found
		}
	}
	return nil
}

func (c *checker) checkLet(node *syntax.Node) {
	// Two forms: let NAME = expr  | let SHAPE [fields]
	var nameNode *syntax.Node
	var shapeNode *syntax.Node
	var exprNode *syntax.Node
	for _, ch := range node.Children {
		if ch.Type == "NAME" && nameNode == nil {
			nameNode = ch
		}
		if ch.Type == "SHAPE" && shapeNode == nil {
			shapeNode = ch
		}
		if ch.Type == "expr" {
			exprNode = ch
		}
	}

	if nameNode != nil {
		if !isValidCase(nameNode.Value) {
			c.add("warning", nameNode.Pos, "naming: '"+nameNode.Value+"' should be snake_case or SCREAMING_SNAKE_CASE")
		}
		if c.isShadowing(nameNode.Value) {
			c.add("warning", nameNode.Pos, "shadowing: '"+nameNode.Value+"' shadows an outer binding")
		}
		c.define(nameNode.Value)
		if exprNode != nil {
			c.check(exprNode)
		}
		return
	}

	if shapeNode != nil {
		// Shapes are @name. Naming convention applies to suffix.
		suffix := strings.TrimPrefix(shapeNode.Value, "@")
		if suffix != "" && !isValidCase(suffix) {
			c.add("warning", shapeNode.Pos, "naming: '"+shapeNode.Value+"' should use snake_case")
		}
		if c.isShadowing(shapeNode.Value) {
			c.add("warning", shapeNode.Pos, "shadowing: '"+shapeNode.Value+"' shadows an outer binding")
		}
		c.define(shapeNode.Value)
		return
	}
}

func (c *checker) checkAssign(node *syntax.Node) {
	// assign_stmt = NAME '=' expr
	var nameNode *syntax.Node
	var exprNode *syntax.Node
	for _, ch := range node.Children {
		if ch.Type == "NAME" && nameNode == nil {
			nameNode = ch
		}
		if ch.Type == "expr" {
			exprNode = ch
		}
	}
	if nameNode != nil {
		if !c.isDefined(nameNode.Value) {
			c.add("error", nameNode.Pos, "assign to undefined name: '"+nameNode.Value+"'")
		}
	}
	if exprNode != nil {
		c.check(exprNode)
	}
}

func (c *checker) checkIf(node *syntax.Node) {
	// if_stmt = 'if' expr block [ 'else' ( if_stmt | block ) ]
	for _, ch := range node.Children {
		if ch.Type == "expr" {
			c.check(ch)
		}
		if ch.Type == "block" {
			c.check(ch)
		}
		if ch.Type == "if_stmt" {
			c.check(ch)
		}
	}
}

func (c *checker) checkWhile(node *syntax.Node) {
	for _, ch := range node.Children {
		if ch.Type == "expr" {
			c.check(ch)
		}
		if ch.Type == "block" {
			c.check(ch)
		}
	}
}

func (c *checker) checkMatch(node *syntax.Node) {
	// match_stmt = 'match' expr '{' { pattern block } '}'
	// First child expr is the matched value.
	for _, ch := range node.Children {
		if ch.Type == "expr" {
			c.check(ch)
			break
		}
	}

	// Remaining pairs pattern block
	for i := 0; i < len(node.Children); i++ {
		if node.Children[i].Type != "pattern" {
			continue
		}
		pat := node.Children[i]
		blk := node.Child(i + 1)
		if blk == nil || blk.Type != "block" {
			continue
		}
		c.pushScope()
		c.bindPattern(pat)
		c.check(blk)
		c.popScope()
	}
}

func (c *checker) bindPattern(pat *syntax.Node) {
	if pat == nil {
		return
	}
	// pattern may contain NAME binders.
	if pat.Type == "NAME" {
		// "_" in patterns is terminal, not NAME token.
		if pat.Value != "_" {
			c.define(pat.Value)
		}
		return
	}
	for _, ch := range pat.Children {
		c.bindPattern(ch)
	}
}

func (c *checker) checkFunc(node *syntax.Node) {
	// func_literal = '(' params ')' block
	var params *syntax.Node
	var blk *syntax.Node
	for _, ch := range node.Children {
		if ch.Type == "params" {
			params = ch
		}
		if ch.Type == "block" {
			blk = ch
		}
	}

	c.pushScope()
	if params != nil {
		c.bindParams(params)
	}
	if blk != nil {
		c.check(blk)
	}
	c.popScope()
}

func (c *checker) bindParams(params *syntax.Node) {
	// params contains NAME and rest_param, possibly via param_list.
	var bind func(n *syntax.Node)
	bind = func(n *syntax.Node) {
		if n == nil {
			return
		}
		if n.Type == "NAME" {
			if !isValidCase(n.Value) {
				c.add("warning", n.Pos, "naming: '"+n.Value+"' should be snake_case or SCREAMING_SNAKE_CASE")
			}
			if c.isShadowing(n.Value) {
				c.add("warning", n.Pos, "shadowing: '"+n.Value+"' shadows an outer binding")
			}
			c.define(n.Value)
			return
		}
		for _, ch := range n.Children {
			bind(ch)
		}
	}
	bind(params)
}

func (c *checker) checkAccess(node *syntax.Node) {
	// access = '.' NAME
	for _, ch := range node.Children {
		if ch.Type == "NAME" {
			if !snakeRe.MatchString(ch.Value) {
				c.add("warning", ch.Pos, "naming: field '"+ch.Value+"' should be snake_case")
			}
		}
	}
	// Do not treat field NAME as a variable reference.
	for _, ch := range node.Children {
		if ch.Type == "NAME" {
			continue
		}
		c.check(ch)
	}
}
