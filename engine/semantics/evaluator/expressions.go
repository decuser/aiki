// expressions.go implements expression evaluation handlers.
package evaluator

import (
	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
)

// evalExpr evaluates an expression (delegates to child).
func evalExpr(e *Evaluator, node *syntax.Node, scope *Scope) (value.Value, error) {
	if len(node.Children) == 1 {
		return e.eval(node.Children[0], scope)
	}

	var result value.Value = value.NullValue()
	for _, child := range node.Children {
		val, err := e.eval(child, scope)
		if err != nil {
			return val, err
		}
		if isError(val) {
			return val, nil
		}
		result = val
	}
	return result, nil
}

// evalPipe evaluates a pipe expression (x |> f()).
func evalPipe(e *Evaluator, node *syntax.Node, scope *Scope) (value.Value, error) {
	var result value.Value
	i := 0

	for i < len(node.Children) {
		child := node.Children[i]

		// Skip pipe operator
		if child.Type == "TERMINAL" && child.Value == "|>" {
			i++
			continue
		}

		// First expression: evaluate normally
		if result.Type == 0 {
			var err error
			result, err = e.eval(child, scope)
			if err != nil {
				return result, err
			}
			if isError(result) {
				return result, nil
			}
			i++
			continue
		}

		// Subsequent expressions: pipe the result as first argument
		result = evalPipeCall(e, child, result, scope)
		if isError(result) {
			return result, nil
		}
		i++
	}

	return result, nil
}

// evalPipeCall evaluates a piped function call with the piped value as first arg.
func evalPipeCall(e *Evaluator, node *syntax.Node, pipedValue value.Value, scope *Scope) value.Value {
	var fn value.Value
	var callNode *syntax.Node
	args := []value.Value{pipedValue}

	// Walk to find the function and call
	var walk func(n *syntax.Node)
	walk = func(n *syntax.Node) {
		for _, child := range n.Children {
			switch child.Type {
			case "primary", "postfix_expr", "infix_expr", "unary_expr", "pipe_expr", "expr":
				walk(child)
			case "NAME":
				fn, _ = scope.Get(child.Value)
				if fn.Type == 0 && e.runtime != nil {
					// Check HAL builtins
					fn = value.Value{
						Type: value.Function,
						Data: child.Value, // Native function name
					}
				}
			case "call":
				callNode = child
				for _, callChild := range child.Children {
					if callChild.Type == "TERMINAL" {
						continue
					}
					val, err := e.eval(callChild, scope)
					if err != nil {
						fn = makeError(scope, callChild, "%s", err.Error()).ToValue()
						return
					}
					if isError(val) {
						fn = val
						return
					}
					args = append(args, val)
				}
			}
		}
	}
	walk(node)

	if fn.Type == 0 {
		return makeError(scope, node, "pipe: could not find function").ToValue()
	}
	if isError(fn) {
		return fn
	}

	// Use callNode if found, otherwise node
	if callNode == nil {
		callNode = node
	}

	result, err := applyCall(e, fn, args, scope, callNode)
	if err != nil {
		return makeError(scope, callNode, "%s", err.Error()).ToValue()
	}
	return result
}

// evalInfix evaluates an infix expression (a + b * c).
// Aiki uses left-to-right evaluation without precedence.
// AST: infix_expr -> unary_expr {infix_tail}
// infix_tail -> infix_op unary_expr
// infix_op -> OPERATOR | "and" | "or"
func evalInfix(e *Evaluator, node *syntax.Node, scope *Scope) (value.Value, error) {
	if len(node.Children) == 0 {
		return value.NullValue(), nil
	}

	// First child should be unary_expr
	result, err := e.eval(node.Children[0], scope)
	if err != nil {
		return result, err
	}
	if isError(result) {
		return result, nil
	}

	// Process infix_tail children
	for i := 1; i < len(node.Children); i++ {
		child := node.Children[i]
		if child.Type == "infix_tail" {
			var op string
			var rightNode *syntax.Node

			for _, tailChild := range child.Children {
				if tailChild.Type == "infix_op" {
					// infix_op contains OPERATOR or KEYWORD
					for _, opChild := range tailChild.Children {
						if opChild.Type == "OPERATOR" || opChild.Type == "KEYWORD" {
							op = opChild.Value
						}
					}
				} else if tailChild.Type == "unary_expr" {
					rightNode = tailChild
				}
			}

			if op == "" || rightNode == nil {
				continue
			}

			right, err := e.eval(rightNode, scope)
			if err != nil {
				return right, err
			}
			if isError(right) {
				return right, nil
			}

			result, err = applyOp(op, result, right, scope, node)
			if err != nil {
				return result, err
			}
		}
	}

	return result, nil
}

// evalUnary evaluates a unary expression (-x, not x).
// AST: unary_expr -> OPERATOR:"-" postfix_expr | postfix_expr
func evalUnary(e *Evaluator, node *syntax.Node, scope *Scope) (value.Value, error) {
	children := node.Children
	if len(children) == 0 {
		return value.NullValue(), nil
	}

	// Check for unary operator (OPERATOR node, not TERMINAL)
	first := children[0]
	if first.Type == "OPERATOR" && len(children) >= 2 {
		op := first.Value
		operand, err := e.eval(children[1], scope)
		if err != nil {
			return operand, err
		}
		if isError(operand) {
			return operand, nil
		}
		return applyUnaryOp(op, operand, scope, node)
	}

	// No operator, delegate to child
	return e.eval(children[0], scope)
}

// evalPostfix evaluates a postfix expression (f(x), a[0], obj.field).
// AST: postfix_expr -> primary { postfix_op }
// postfix_op -> call | index | access
func evalPostfix(e *Evaluator, node *syntax.Node, scope *Scope) (value.Value, error) {
	if len(node.Children) == 0 {
		return value.NullValue(), nil
	}

	// Evaluate the primary expression
	result, err := e.eval(node.Children[0], scope)
	if err != nil {
		return result, err
	}
	if isError(result) {
		return result, nil
	}

	// Process postfix operators
	for i := 1; i < len(node.Children); i++ {
		child := node.Children[i]

		// Handle postfix_op wrapper
		if child.Type == "postfix_op" && len(child.Children) > 0 {
			child = child.Children[0] // Get the actual call/index/access
		}

		switch child.Type {
		case "call":
			// Function call
			args, err := evalArgsFromCall(e, child, scope)
			if err != nil {
				return value.NullValue(), err
			}
			result, err = applyCall(e, result, args, scope, child)
			if err != nil {
				return result, err
			}
			if isError(result) {
				return result, nil
			}

		case "index":
			// Array/list indexing [n]
			idx, err := evalIndex(e, child, scope)
			if err != nil {
				return idx, err
			}
			if isError(idx) {
				return idx, nil
			}
			result, err = applyIndex(result, idx, scope, child)
			if err != nil {
				return result, err
			}

		case "access":
			// Field access .name
			field := extractFieldName(child)
			result, err = applyAccess(result, field, scope, child)
			if err != nil {
				return result, err
			}
		}
	}

	return result, nil
}

// evalPrimary evaluates a primary expression.
func evalPrimary(e *Evaluator, node *syntax.Node, scope *Scope) (value.Value, error) {
	for _, child := range node.Children {
		// Skip parentheses
		if child.Type == "TERMINAL" && (child.Value == "(" || child.Value == ")") {
			continue
		}
		return e.eval(child, scope)
	}
	return value.NullValue(), nil
}

// evalArgsFromCall evaluates function call arguments from a call node.
// AST: call -> "(" [ arg_list ] ")"
// arg_list -> expr { "," expr }
func evalArgsFromCall(e *Evaluator, callNode *syntax.Node, scope *Scope) ([]value.Value, error) {
	var args []value.Value

	// Find the arg_list node inside call
	var argListNode *syntax.Node
	for _, child := range callNode.Children {
		if child.Type == "arg_list" {
			argListNode = child
			break
		}
	}

	if argListNode == nil {
		return args, nil // No arguments
	}

	// Collect all exprs from arg_list (including arg_tails)
	exprs := collectExprsFromArgList(argListNode)
	for _, exprNode := range exprs {
		val, err := e.eval(exprNode, scope)
		if err != nil {
			return nil, err
		}
		if isError(val) {
			return nil, asError(val)
		}
		args = append(args, val)
	}

	return args, nil
}

// evalArgs evaluates function call arguments (legacy - for args node).
// AST: args -> expr "," args | expr
func evalArgs(e *Evaluator, callNode *syntax.Node, scope *Scope) ([]value.Value, error) {
	var args []value.Value

	// Find the args node inside call
	var argsNode *syntax.Node
	for _, child := range callNode.Children {
		if child.Type == "args" {
			argsNode = child
			break
		}
	}

	if argsNode == nil {
		return args, nil // No arguments
	}

	// Recursively collect args from the args node
	collectArgs(e, argsNode, &args, scope)
	return args, nil
}

// collectArgs recursively collects arguments from an args node.
func collectArgs(e *Evaluator, node *syntax.Node, args *[]value.Value, scope *Scope) error {
	for _, child := range node.Children {
		switch child.Type {
		case "expr":
			val, err := e.eval(child, scope)
			if err != nil {
				return err
			}
			if isError(val) {
				return asError(val)
			}
			*args = append(*args, val)
		case "args":
			// Recursive args
			if err := collectArgs(e, child, args, scope); err != nil {
				return err
			}
		case "PUNCT":
			// Skip punctuation
		}
	}
	return nil
}

// evalIndex evaluates an index expression [n].
func evalIndex(e *Evaluator, indexNode *syntax.Node, scope *Scope) (value.Value, error) {
	for _, child := range indexNode.Children {
		if child.Type == "TERMINAL" || child.Type == "DELIMITER" {
			continue
		}
		if child.Type == "expr" {
			return e.eval(child, scope)
		}
	}
	return value.NullValue(), nil
}

// extractOp extracts the operator string from a BINOP node.
func extractOp(node *syntax.Node) string {
	if node.Value != "" {
		return node.Value
	}
	for _, child := range node.Children {
		if child.Type == "TERMINAL" {
			return child.Value
		}
	}
	return ""
}

// extractFieldName extracts the field name from an access node.
func extractFieldName(node *syntax.Node) string {
	for _, child := range node.Children {
		if child.Type == "NAME" {
			return child.Value
		}
		if child.Type == "NUMBER" {
			return child.Value
		}
	}
	return ""
}
