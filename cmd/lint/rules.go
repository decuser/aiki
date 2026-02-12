package lint

import (
	"fmt"
	"reflect"
	"regexp"

	"aiki/lang/ast"
)

var snakeCase = regexp.MustCompile(`^_?([a-z][a-z0-9_]*|[A-Z][A-Z0-9_]*)$`)

// Check verifies the program against a list of allowed global names (builtins).
func Check(node ast.Node, globals []string) []string {
	var errors []string

	scope := make(map[string]bool)
	for _, g := range globals {
		scope[g] = true
	}

	var scopeStack []map[string]bool
	pushScope := func() {
		scopeStack = append(scopeStack, make(map[string]bool))
	}
	popScope := func() {
		scopeStack = scopeStack[:len(scopeStack)-1]
	}
	define := func(name string) {
		if len(scopeStack) > 0 {
			scopeStack[len(scopeStack)-1][name] = true
		}
	}
	isDefined := func(name string) bool {
		for i := len(scopeStack) - 1; i >= 0; i-- {
			if scopeStack[i][name] {
				return true
			}
		}
		return scope[name]
	}

	var walk func(n ast.Node)
	walk = func(n ast.Node) {
		// fmt.Printf("walk: %T %v\n", n, n == nil)
		if n != nil && reflect.ValueOf(n).IsNil() {
			// fmt.Printf("TYPED NIL: %T\n", n)
			return
		}
		// fmt.Printf("walk: %T %v\n", n, n == nil)
		if n == nil {
			return
		}

		switch node := n.(type) {
		case nil:
			return
		case *ast.Program:
			pushScope()
			// First pass: collect all top-level names
			for _, stmt := range node.Statements {
				if let, ok := stmt.(*ast.LetStatement); ok {
					define(let.Name.Value)
				}
			}
			// Second pass: check usage
			for _, stmt := range node.Statements {
				walk(stmt)
			}
			popScope()

		case *ast.FunctionLiteral:
			pushScope()
			for _, param := range node.Parameters {
				if !snakeCase.MatchString(param) {
					errors = append(errors, fmt.Sprintf("line %d: parameter '%s' should be snake_case", node.Token.Line, param))
				}
				define(param)
			}
			if node.RestParam != "" {
				if !snakeCase.MatchString(node.RestParam) {
					errors = append(errors, fmt.Sprintf("line %d: parameter '%s' should be snake_case", node.Token.Line, node.RestParam))
				}
				define(node.RestParam)
			}
			walk(node.Body)
			popScope()

		case *ast.BlockStatement:
			pushScope()
			for _, stmt := range node.Statements {
				if stmt == nil || reflect.ValueOf(stmt).IsNil() {
					continue
				}
				walk(stmt)
			}
			popScope()

		case *ast.LetStatement:
			if node.Name != nil && !snakeCase.MatchString(node.Name.Value) {
				errors = append(errors, fmt.Sprintf("line %d: variable '%s' should be snake_case", node.Token.Line, node.Name.Value))
			}
			walk(node.Value)
			if node.Name != nil {
				define(node.Name.Value)
			}

		case *ast.Identifier:
			if !isDefined(node.Value) {
				errors = append(errors, fmt.Sprintf("line %d: undefined identifier '%s' (missing from language or not defined)", node.Token.Line, node.Value))
			}

		case *ast.CallExpression:
			walk(node.Function)
			for _, arg := range node.Arguments {
				walk(arg)
			}
		case *ast.ExpressionStatement:
			walk(node.Expression)
		case *ast.ReturnStatement:
			walk(node.Value)
		case *ast.IfStatement:
			walk(node.Condition)
			walk(node.Consequence)
			walk(node.Alternative)
		case *ast.WhileStatement:
			walk(node.Condition)
			walk(node.Body)
		case *ast.AssignStatement:
			if node.Name != nil && !isDefined(node.Name.Value) {
				errors = append(errors, fmt.Sprintf("line %d: cannot assign to undefined variable '%s'", node.Token.Line, node.Name.Value))
			}
			walk(node.Value)
		case *ast.InfixExpression:
			walk(node.Left)
			walk(node.Right)
		case *ast.PrefixExpression:
			walk(node.Right)
		case *ast.PipeExpression:
			walk(node.Left)
			walk(node.Right)
		case *ast.ListLiteral:
			for _, el := range node.Elements {
				walk(el)
			}
		case *ast.ShapedListLiteral:
			for _, el := range node.Elements {
				walk(el)
			}
		case *ast.AccessExpression:
			walk(node.Left)
			if len(node.Key) > 0 && (node.Key[0] >= 'a' && node.Key[0] <= 'z') {
				if !snakeCase.MatchString(node.Key) {
					errors = append(errors, fmt.Sprintf("line %d: field access '%s' should be snake_case", node.Token.Line, node.Key))
				}
			}
		case *ast.MatchStatement:
			walk(node.Value)
			for _, arm := range node.Arms {
				walk(arm.Body)
			}
		case *ast.ImportStatement:
			for _, name := range node.Names {
				define(name)
			}
		case *ast.ExportStatement:
			// nothing to walk
		case *ast.ShapeStatement:
			define(node.Name)
		}
	}

	walk(node)
	return errors
}
