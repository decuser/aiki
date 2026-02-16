package lint

import (
	"regexp"

	"aiki/ebnf"
	"aiki/lang/value"
	"aiki/layers/hal"
)

// Diagnostic represents a lint finding.
type Diagnostic struct {
	Line    int
	Column  int
	Level   string // "error" or "warning"
	Message string
}

// LintSource checks source code for lint issues.
func LintSource(grammar *ebnf.Grammar, source string) ([]Diagnostic, error) {
	node, err := grammar.ParseSource(source)
	if err != nil {
		return nil, err
	}

	c := &checker{
		scopes: []map[string]bool{makeGlobals()},
	}

	// Two-pass: first collect top-level let names, then check
	c.collectTopLevel(node)
	c.check(node)

	return c.diags, nil
}

// checker walks the AST and collects diagnostics.
type checker struct {
	scopes []map[string]bool
	diags  []Diagnostic
}

var snakeRe = regexp.MustCompile(`^_?[a-z][a-z0-9_]*$`)
var screamRe = regexp.MustCompile(`^_?[A-Z][A-Z0-9_]*$`)

// isValidCase checks if name is snake_case or SCREAMING_SNAKE_CASE.
func isValidCase(name string) bool {
	return snakeRe.MatchString(name) || screamRe.MatchString(name)
}

// makeGlobals returns a scope containing all HAL builtins, intrinsics, and prelude exports.
func makeGlobals() map[string]bool {
	globals := make(map[string]bool)
	for name := range hal.HAL {
		globals[name] = true
	}
	// Add language intrinsics (apply, load, import, export)
	for name := range value.Intrinsics {
		globals[name] = true
	}
	for _, name := range strictExports {
		globals[name] = true
	}
	// Keywords that appear as identifiers
	globals["true"] = true
	globals["false"] = true
	globals["not"] = true
	globals["and"] = true
	globals["or"] = true
	return globals
}

// collectTopLevel does a first pass over program children to pre-define
// all top-level let bindings. This allows forward references at file scope.
func (c *checker) collectTopLevel(node *ebnf.Node) {
	if node.Type != "program" {
		return
	}
	for _, child := range node.Children {
		if child.Type == "statement" && len(child.Children) > 0 {
			stmt := child.Children[0]
			switch stmt.Type {
			case "let_stmt":
				for _, ch := range stmt.Children {
					if ch.Type == "NAME" {
						c.define(ch.Value)
						break
					}
					if ch.Type == "SHAPE" {
						c.define(ch.Value) // @shape
						break
					}
				}
			}
		}
	}
}

func (c *checker) pushScope() {
	c.scopes = append(c.scopes, make(map[string]bool))
}

func (c *checker) popScope() {
	if len(c.scopes) > 1 {
		c.scopes = c.scopes[:len(c.scopes)-1]
	}
}

func (c *checker) define(name string) {
	c.scopes[len(c.scopes)-1][name] = true
}

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

func (c *checker) addError(line, col int, msg string) {
	c.diags = append(c.diags, Diagnostic{
		Line:    line,
		Column:  col,
		Level:   "error",
		Message: msg,
	})
}

func (c *checker) addWarning(line, col int, msg string) {
	c.diags = append(c.diags, Diagnostic{
		Line:    line,
		Column:  col,
		Level:   "warning",
		Message: msg,
	})
}

func (c *checker) check(node *ebnf.Node) {
	switch node.Type {
	case "program":
		for _, child := range node.Children {
			c.check(child)
		}

	case "statement":
		for _, child := range node.Children {
			c.check(child)
		}

	case "let_stmt":
		c.checkLet(node)

	case "assign_stmt":
		c.checkAssign(node)

	case "if_stmt":
		c.checkIf(node)

	case "while_stmt":
		c.checkWhile(node)

	case "match_stmt":
		c.checkMatch(node)

	case "return_stmt":
		for _, child := range node.Children {
			if child.Type != "TERMINAL" {
				c.check(child)
			}
		}

	case "expr_stmt":
		for _, child := range node.Children {
			c.check(child)
		}

	case "block":
		c.pushScope()
		for _, child := range node.Children {
			c.check(child)
		}
		c.popScope()

	case "func_literal":
		c.checkFunc(node)

	case "expr", "pipe_expr", "infix_expr", "unary_expr", "postfix_expr":
		for _, child := range node.Children {
			c.check(child)
		}

	case "primary":
		for _, child := range node.Children {
			c.check(child)
		}

	case "call", "index":
		for _, child := range node.Children {
			c.check(child)
		}

	case "access":
		c.checkAccess(node)

	case "list_literal":
		for _, child := range node.Children {
			c.check(child)
		}

	case "NAME":
		c.checkNameRef(node)

	default:
		for _, child := range node.Children {
			c.check(child)
		}
	}
}

func (c *checker) checkLet(node *ebnf.Node) {
	var name *ebnf.Node
	isShape := false

	for _, child := range node.Children {
		if child.Type == "SHAPE" {
			isShape = true
			// Define the shape name
			c.define(child.Value)
			break
		}
		if child.Type == "NAME" && name == nil {
			name = child
		}
	}

	if isShape {
		// Shape definitions — already handled above
		return
	}

	if name != nil {
		// Check case convention
		if !isValidCase(name.Value) {
			c.addError(name.Line, name.Column,
				"naming: '"+name.Value+"' must be snake_case or SCREAMING_SNAKE_CASE")
		}

		// Check for shadowing
		if c.isShadowing(name.Value) {
			c.addWarning(name.Line, name.Column,
				"shadow: '"+name.Value+"' shadows existing binding")
		}

		// Define the name in current scope (this was missing!)
		c.define(name.Value)
	}

	// Check the expression (right side of =)
	foundEquals := false
	for _, child := range node.Children {
		if child.Type == "TERMINAL" && child.Value == "=" {
			foundEquals = true
			continue
		}
		if foundEquals {
			c.check(child)
		}
	}
}

func (c *checker) checkAssign(node *ebnf.Node) {
	for _, child := range node.Children {
		if child.Type == "NAME" {
			if !c.isDefined(child.Value) {
				c.addError(child.Line, child.Column,
					"undefined: '"+child.Value+"'")
			}
		}
		if child.Type != "TERMINAL" {
			c.check(child)
		}
	}
}

func (c *checker) checkIf(node *ebnf.Node) {
	for _, child := range node.Children {
		if child.Type != "TERMINAL" {
			c.check(child)
		}
	}
}

func (c *checker) checkWhile(node *ebnf.Node) {
	for _, child := range node.Children {
		if child.Type != "TERMINAL" {
			c.check(child)
		}
	}
}

func (c *checker) checkMatch(node *ebnf.Node) {
	for _, child := range node.Children {
		if child.Type == "pattern" {
			// Pattern followed by block — handle as arm
			continue
		}
		if child.Type == "block" {
			// Find the preceding pattern
			c.checkMatchArm(node, child)
		} else if child.Type != "TERMINAL" {
			c.check(child)
		}
	}
}

func (c *checker) checkMatchArm(matchNode *ebnf.Node, block *ebnf.Node) {
	c.pushScope()

	// Find pattern that precedes this block
	var pattern *ebnf.Node
	for i, child := range matchNode.Children {
		if child == block && i > 0 {
			for j := i - 1; j >= 0; j-- {
				if matchNode.Children[j].Type == "pattern" {
					pattern = matchNode.Children[j]
					break
				}
			}
			break
		}
	}

	if pattern != nil {
		c.checkPattern(pattern)
	}

	// Check block contents
	for _, stmt := range block.Children {
		c.check(stmt)
	}

	c.popScope()
}

func (c *checker) checkPattern(node *ebnf.Node) {
	for _, child := range node.Children {
		if child.Type == "NAME" {
			// Binding in pattern — define it
			c.define(child.Value)
		} else if child.Type == "pattern" {
			c.checkPattern(child)
		}
	}
}

func (c *checker) checkFunc(node *ebnf.Node) {
	c.pushScope()

	// Define parameters
	for _, child := range node.Children {
		if child.Type == "params" {
			c.defineParams(child)
		}
	}

	// Check body
	for _, child := range node.Children {
		if child.Type == "block" {
			for _, stmt := range child.Children {
				c.check(stmt)
			}
		}
	}

	c.popScope()
}

func (c *checker) defineParams(node *ebnf.Node) {
	for _, child := range node.Children {
		switch child.Type {
		case "param_list":
			for _, param := range child.Children {
				if param.Type == "NAME" {
					if !isValidCase(param.Value) {
						c.addError(param.Line, param.Column,
							"naming: '"+param.Value+"' must be snake_case or SCREAMING_SNAKE_CASE")
					}
					c.define(param.Value)
				}
			}
		case "rest_param":
			for _, param := range child.Children {
				if param.Type == "NAME" {
					if !isValidCase(param.Value) {
						c.addError(param.Line, param.Column,
							"naming: '"+param.Value+"' must be snake_case or SCREAMING_SNAKE_CASE")
					}
					c.define(param.Value)
				}
			}
		case "NAME":
			if !isValidCase(child.Value) {
				c.addError(child.Line, child.Column,
					"naming: '"+child.Value+"' must be snake_case or SCREAMING_SNAKE_CASE")
			}
			c.define(child.Value)
		}
	}
}

func (c *checker) checkAccess(node *ebnf.Node) {
	// Check the left side normally
	for _, child := range node.Children {
		if child.Type == "NAME" {
			// Field name — check snake_case
			if len(child.Value) > 0 && child.Value[0] >= 'a' && child.Value[0] <= 'z' {
				if !isValidCase(child.Value) {
					c.addError(child.Line, child.Column,
						"naming: field '"+child.Value+"' should be snake_case")
				}
			}
			// Don't check isDefined — field names are resolved at runtime
		}
	}
}

func (c *checker) checkNameRef(node *ebnf.Node) {
	name := node.Value
	// Skip keywords that appear as NAME tokens
	if name == "true" || name == "false" || name == "not" ||
		name == "and" || name == "or" || name == "let" ||
		name == "if" || name == "else" || name == "while" ||
		name == "match" || name == "return" || name == "export" ||
		name == "from" || name == "use" {
		return
	}
	if !c.isDefined(name) {
		c.addError(node.Line, node.Column, "undefined: '"+name+"'")
	}
}
