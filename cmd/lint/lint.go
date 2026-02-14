package lint

import (
	"regexp"
	"strings"

	"aiki/ebnf"
	"aiki/hal/core"
	"aiki/strict"
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

// makeGlobals returns a scope containing all HAL builtins and strict exports.
func makeGlobals() map[string]bool {
	globals := make(map[string]bool)
	for name := range core.HAL {
		globals[name] = true
	}
	for _, name := range strict.Exports() {
		globals[name] = true
	}
	// Keywords used as identifiers in special forms
	globals["true"] = true
	globals["false"] = true
	globals["not"] = true
	globals["and"] = true
	globals["or"] = true
	return globals
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
	// Check all scopes except the current one
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

	case "export_stmt":
		c.checkExport(node)

	case "import_stmt":
		c.checkImport(node)

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

	case "call", "index", "access":
		for _, child := range node.Children {
			c.check(child)
		}

	case "list_literal":
		for _, child := range node.Children {
			c.check(child)
		}

	case "NAME":
		c.checkNameRef(node)

	default:
		// Recurse into children for any unhandled node types
		for _, child := range node.Children {
			c.check(child)
		}
	}
}

func (c *checker) checkLet(node *ebnf.Node) {
	// let_stmt has children: TERMINAL:"let" NAME TERMINAL:"=" expr
	// or shape: TERMINAL:"let" SHAPE TERMINAL:"[" fields TERMINAL:"]"
	var name *ebnf.Node
	isShape := false

	for _, child := range node.Children {
		if child.Type == "SHAPE" {
			isShape = true
			break
		}
		if child.Type == "NAME" && name == nil {
			name = child
		}
	}

	if isShape {
		// Shape definitions - define the shape name
		for _, child := range node.Children {
			if child.Type == "SHAPE" {
				shapeName := strings.TrimPrefix(child.Value, "@")
				c.define("@" + shapeName)
			}
		}
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

		// Define the name before checking the value (for recursion)
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
	// assign_stmt: NAME "=" expr  or  postfix_expr "=" expr
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
	// Check the expression being matched
	for _, child := range node.Children {
		if child.Type == "match_arm" {
			c.checkMatchArm(child)
		} else if child.Type != "TERMINAL" {
			c.check(child)
		}
	}
}

func (c *checker) checkMatchArm(node *ebnf.Node) {
	// match_arm: pattern block
	// Pattern can introduce bindings in the block scope
	c.pushScope()
	for _, child := range node.Children {
		if child.Type == "pattern" {
			c.checkPattern(child)
		} else if child.Type == "block" {
			// Don't push another scope - we already pushed one
			for _, stmt := range child.Children {
				c.check(stmt)
			}
		} else if child.Type != "TERMINAL" {
			c.check(child)
		}
	}
	c.popScope()
}

func (c *checker) checkPattern(node *ebnf.Node) {
	for _, child := range node.Children {
		if child.Type == "NAME" {
			// Pattern names are bindings, not references
			c.define(child.Value)
		} else if child.Type == "pattern" {
			c.checkPattern(child)
		}
		// Wildcards (_), literals, terminals are fine
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
			// Don't push another scope - func already pushed one
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
			// Direct NAME child (simple params)
			if !isValidCase(child.Value) {
				c.addError(child.Line, child.Column,
					"naming: '"+child.Value+"' must be snake_case or SCREAMING_SNAKE_CASE")
			}
			c.define(child.Value)
		}
	}
}

func (c *checker) checkExport(node *ebnf.Node) {
	// export [name1, name2, ...]
	for _, child := range node.Children {
		if child.Type == "NAME" {
			if !c.isDefined(child.Value) {
				c.addWarning(child.Line, child.Column,
					"export: '"+child.Value+"' not defined at export point")
			}
			if strings.HasPrefix(child.Value, "_") {
				c.addWarning(child.Line, child.Column,
					"export: '"+child.Value+"' has _prefix (internal convention)")
			}
		}
	}
}

func (c *checker) checkImport(node *ebnf.Node) {
	// from mod use [name1, name2, ...]
	// Define imported names in current scope
	isFirst := true
	for _, child := range node.Children {
		if child.Type == "NAME" {
			if isFirst {
				// First NAME is the module name, skip
				isFirst = false
				continue
			}
			c.define(child.Value)
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
