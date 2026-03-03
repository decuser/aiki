package lint

import (
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

	c := &checker{scopes: []scopeFrame{makeGlobals(lintScope)}, diags: nil}
	c.collectTopLevel(node)
	c.check(node)
	return c.diags, nil
}

type scopeFrame map[string]bool

type checker struct {
	scopes []scopeFrame
	diags  []Diagnostic
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
		c.pushScope()
		for _, ch := range node.Children {
			c.check(ch)
		}
		c.popScope()
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

	for _, ch := range node.Children {
		c.check(ch)
	}
}

func (c *checker) bindImportCall(call *syntax.Node) {
	if call == nil {
		return
	}
	// call = "(" [ expr { "," expr } ] ")"
	for _, ch := range call.Children {
		if ch.Type != "expr" {
			continue
		}
		// Import list items are SYMBOL tokens like :sin, :cos.
		if sym := ch.ChildByType("SYMBOL"); sym != nil {
			name := strings.TrimPrefix(sym.Value, ":")
			if name != "" && isNameToken(name) {
				c.define(name)
			}
		}
	}
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
