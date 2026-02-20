// literals.go implements literal value evaluation handlers.
package evaluator

import (
	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
	"strconv"
	"strings"
)

// evalNumber evaluates a number literal.
func evalNumber(e *Evaluator, node *syntax.Node, scope *Scope) (value.Value, error) {
	num, err := value.ParseNumber(node.Value)
	if err != nil {
		return value.NullValue(), makeError(scope, node, "invalid number: %s", node.Value)
	}
	return value.Value{Type: value.Number, Data: num}, nil
}

// evalString evaluates a string literal.
func evalString(e *Evaluator, node *syntax.Node, scope *Scope) (value.Value, error) {
	s, err := strconv.Unquote(node.Value)
	if err != nil {
		// Try without quotes for symbols, etc.
		s = node.Value
	}
	return value.Value{Type: value.String, Data: s}, nil
}

// evalRune evaluates a rune literal ('x').
func evalRune(e *Evaluator, node *syntax.Node, scope *Scope) (value.Value, error) {
	s := node.Value

	// Strip quotes
	if len(s) >= 2 {
		s = s[1 : len(s)-1]
	}

	// Handle escape sequences
	if len(s) == 2 && s[0] == '\\' {
		switch s[1] {
		case 'n':
			return value.Value{Type: value.Rune, Data: '\n'}, nil
		case 't':
			return value.Value{Type: value.Rune, Data: '\t'}, nil
		case 'r':
			return value.Value{Type: value.Rune, Data: '\r'}, nil
		case '\\':
			return value.Value{Type: value.Rune, Data: '\\'}, nil
		case '\'':
			return value.Value{Type: value.Rune, Data: '\''}, nil
		}
	}

	runes := []rune(s)
	if len(runes) > 0 {
		return value.Value{Type: value.Rune, Data: runes[0]}, nil
	}

	return value.NullValue(), nil
}

// evalSymbol evaluates a symbol literal (:name).
func evalSymbol(e *Evaluator, node *syntax.Node, scope *Scope) (value.Value, error) {
	sym := strings.TrimPrefix(node.Value, ":")
	return value.Value{Type: value.Symbol, Data: sym}, nil
}

// evalName evaluates a name/identifier.
func evalName(e *Evaluator, node *syntax.Node, scope *Scope) (value.Value, error) {
	name := node.Value

	// Look up in scope
	val, ok := scope.Get(name)
	if ok {
		return val, nil
	}

	// Check intrinsics (these are handled specially at call time)
	if IsIntrinsic(name) {
		return value.Value{
			Type: value.Function,
			Data: "intrinsic:" + name, // Mark as intrinsic for dispatch
		}, nil
	}

	// Check HAL builtins
	if e.runtime != nil && e.runtime.HasBuiltin(name) {
		return value.Value{
			Type: value.Function,
			Data: name, // Store name for native function lookup
		}, nil
	}

	return value.NullValue(), makeError(scope, node, "undefined: %s", name)
}

// evalTerminal evaluates a terminal node (true, false, keywords).
func evalTerminal(e *Evaluator, node *syntax.Node, scope *Scope) (value.Value, error) {
	switch node.Value {
	case "true":
		return value.True(), nil
	case "false":
		return value.False(), nil
	}
	return value.NullValue(), nil
}

// evalKeyword evaluates a keyword node (true, false, etc).
func evalKeyword(e *Evaluator, node *syntax.Node, scope *Scope) (value.Value, error) {
	switch node.Value {
	case "true":
		return value.True(), nil
	case "false":
		return value.False(), nil
	}
	return value.NullValue(), nil
}

// evalFuncLiteral evaluates a function/lambda literal.
// AST: lambda -> PUNCT:"(" [params] PUNCT:")" block
// params -> NAME {"," params} | NAME
func evalFuncLiteral(e *Evaluator, node *syntax.Node, scope *Scope) (value.Value, error) {
	var paramsNode, bodyNode *syntax.Node

	for _, child := range node.Children {
		switch child.Type {
		case "params":
			paramsNode = child
		case "block":
			bodyNode = child
		}
	}

	fn := &Function{
		Params: extractParams(paramsNode),
		Body:   bodyNode,
		Env:    scope,
	}

	return value.Value{Type: value.Function, Data: fn}, nil
}

// evalListLiteral evaluates a list literal [a, b, c].
func evalListLiteral(e *Evaluator, node *syntax.Node, scope *Scope) (value.Value, error) {
	var elements []value.Value
	var shape string

	for _, child := range node.Children {
		switch child.Type {
		case "DELIMITER":
			continue
		case "SHAPE":
			shape = child.Value
		case "arg_list":
			// Collect all exprs from arg_list and arg_tails
			exprs := collectExprsFromArgList(child)
			for _, exprNode := range exprs {
				val, err := e.eval(exprNode, scope)
				if err != nil {
					return val, err
				}
				if isError(val) {
					return val, nil
				}
				elements = append(elements, val)
			}
		case "shape_args":
			// Shaped list: [@shape, expr, expr, ...]
			exprs := collectExprsFromShapeArgs(child)
			for _, exprNode := range exprs {
				val, err := e.eval(exprNode, scope)
				if err != nil {
					return val, err
				}
				if isError(val) {
					return val, nil
				}
				elements = append(elements, val)
			}
		case "expr":
			// Direct expr child (shouldn't happen with new grammar but be safe)
			val, err := e.eval(child, scope)
			if err != nil {
				return val, err
			}
			if isError(val) {
				return val, nil
			}
			elements = append(elements, val)
		}
	}

	list := value.Value{Type: value.List, Data: elements}

	// If shaped, wrap with shape info
	if shape != "" {
		list.Data = &ShapedList{
			Shape:    shape,
			Elements: elements,
		}
	}

	return list, nil
}

// collectExprsFromArgList extracts all expr nodes from arg_list.
// AST: arg_list -> expr { arg_tail }
// arg_tail -> "," expr
func collectExprsFromArgList(argList *syntax.Node) []*syntax.Node {
	var exprs []*syntax.Node
	for _, child := range argList.Children {
		if child.Type == "expr" {
			exprs = append(exprs, child)
		} else if child.Type == "arg_tail" {
			// arg_tail contains "," and expr
			for _, tailChild := range child.Children {
				if tailChild.Type == "expr" {
					exprs = append(exprs, tailChild)
				}
			}
		}
	}
	return exprs
}

// collectExprsFromShapeArgs extracts expr nodes from shape_args.
// AST: shape_args -> { "," expr }
func collectExprsFromShapeArgs(shapeArgs *syntax.Node) []*syntax.Node {
	var exprs []*syntax.Node
	for _, child := range shapeArgs.Children {
		if child.Type == "expr" {
			exprs = append(exprs, child)
		}
	}
	return exprs
}

// extractParams extracts parameter names from a params node.
// AST: params -> NAME "," params | NAME
func extractParams(node *syntax.Node) []string {
	if node == nil {
		return nil
	}

	var params []string
	collectParams(node, &params)
	return params
}

// collectParams recursively collects parameter names.
func collectParams(node *syntax.Node, params *[]string) {
	for _, child := range node.Children {
		switch child.Type {
		case "NAME":
			*params = append(*params, child.Value)
		case "params":
			collectParams(child, params)
		case "param_list":
			collectParams(child, params)
		case "param_tail":
			// param_tail -> "," NAME
			collectParams(child, params)
		case "rest_param":
			// Handle ...rest
			for _, c := range child.Children {
				if c.Type == "NAME" {
					*params = append(*params, "..."+c.Value)
				}
			}
		}
	}
}

// Function represents an Aiki function value.
type Function struct {
	Name   string
	Params []string
	Body   *syntax.Node
	Env    *Scope
	Layer  Layer
}

// ShapedList represents a list with a shape tag.
type ShapedList struct {
	Shape    string
	Elements []value.Value
}
