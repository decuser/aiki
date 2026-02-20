// statements.go implements statement evaluation handlers.
package evaluator

import (
	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
	"strings"
)

// evalProgram evaluates a program (sequence of statements).
func evalProgram(e *Evaluator, node *syntax.Node, scope *Scope) (value.Value, error) {
	var result value.Value = value.NullValue()

	for _, child := range node.Children {
		val, err := e.eval(child, scope)
		if err != nil {
			return val, err
		}

		// Check for return value
		if ret, ok := val.Data.(*ReturnValue); ok {
			return ret.Value, nil
		}

		// Check for error
		if isError(val) {
			return val, nil
		}

		result = val
	}

	return result, nil
}

// evalStatement evaluates a single statement.
func evalStatement(e *Evaluator, node *syntax.Node, scope *Scope) (value.Value, error) {
	if len(node.Children) > 0 {
		return e.eval(node.Children[0], scope)
	}
	return value.NullValue(), nil
}

// evalBlock evaluates a block of statements in a new scope.
func evalBlock(e *Evaluator, node *syntax.Node, scope *Scope) (value.Value, error) {
	blockScope := NewScope(scope)
	var result value.Value = value.NullValue()

	for _, child := range node.Children {
		if child.Type == "TERMINAL" {
			continue
		}
		if child.Type == "statement" {
			val, err := e.eval(child, blockScope)
			if err != nil {
				return val, err
			}

			// Propagate returns and errors
			if val.Type == value.Return || isError(val) {
				return val, nil
			}

			result = val
		}
	}

	return result, nil
}

// evalLet evaluates a let statement (variable or shape definition).
func evalLet(e *Evaluator, node *syntax.Node, scope *Scope) (value.Value, error) {
	// AST structure: let_stmt -> KEYWORD:"let" NAME OPERATOR:"=" expr
	var name string
	var nameNode *syntax.Node
	var valNode *syntax.Node

	for i, child := range node.Children {
		switch child.Type {
		case "NAME":
			if name == "" {
				name = child.Value
				nameNode = child
			}
		case "SHAPE":
			// Shape definition - handle separately
			name = strings.TrimPrefix(child.Value, "@")
			nameNode = child
		case "OPERATOR":
			if child.Value == "=" && i+1 < len(node.Children) {
				valNode = node.Children[i+1]
			}
		case "expr":
			// Direct expr child (after =)
			if valNode == nil {
				valNode = child
			}
		}
	}

	// Variable binding
	if name == "" {
		return value.NullValue(), makeError(scope, node, "let: missing name")
	}
	if valNode == nil {
		return value.NullValue(), makeError(scope, node, "let: missing value")
	}

	// Check for builtin shadowing
	if e.runtime != nil && e.runtime.HasBuiltin(name) {
		return value.NullValue(), makeError(scope, nameNode, "cannot shadow builtin: %s", name)
	}

	val, err := e.eval(valNode, scope)
	if err != nil {
		return val, err
	}
	if isError(val) {
		return val, nil
	}

	// Tag functions with their name
	if val.Type == value.Function {
		if fn, ok := val.Data.(*Function); ok {
			fn.Name = name
		}
	}

	scope.Define(name, val)
	return value.NullValue(), nil
}

// evalAssign evaluates an assignment statement.
// AST: assign_stmt -> NAME OPERATOR:"=" expr
func evalAssign(e *Evaluator, node *syntax.Node, scope *Scope) (value.Value, error) {
	var name string
	var nameNode *syntax.Node
	var valNode *syntax.Node

	for i, child := range node.Children {
		switch child.Type {
		case "NAME":
			if name == "" {
				name = child.Value
				nameNode = child
			}
		case "OPERATOR":
			if child.Value == "=" && i+1 < len(node.Children) {
				valNode = node.Children[i+1]
			}
		case "expr":
			if valNode == nil {
				valNode = child
			}
		}
	}

	if name == "" || valNode == nil {
		return value.NullValue(), makeError(scope, node, "assign: invalid statement")
	}

	val, err := e.eval(valNode, scope)
	if err != nil {
		return val, err
	}
	if isError(val) {
		return val, nil
	}

	if !scope.Update(name, val) {
		return value.NullValue(), makeError(scope, nameNode, "undefined: %s", name)
	}

	return value.NullValue(), nil
}

// evalReturn evaluates a return statement.
// AST: return_stmt -> KEYWORD:"return" expr
func evalReturn(e *Evaluator, node *syntax.Node, scope *Scope) (value.Value, error) {
	for _, child := range node.Children {
		if child.Type == "expr" {
			val, err := e.eval(child, scope)
			if err != nil {
				return val, err
			}
			if isError(val) {
				return val, nil
			}
			return value.Value{
				Type: value.Return,
				Data: &ReturnValue{Value: val},
			}, nil
		}
	}
	return value.Value{
		Type: value.Return,
		Data: &ReturnValue{Value: value.NullValue()},
	}, nil
}

// evalExprStmt evaluates an expression statement.
func evalExprStmt(e *Evaluator, node *syntax.Node, scope *Scope) (value.Value, error) {
	if len(node.Children) > 0 {
		return e.eval(node.Children[0], scope)
	}
	return value.NullValue(), nil
}

// evalIf evaluates an if statement.
// AST: if_stmt -> KEYWORD:"if" expr block [KEYWORD:"else" (if_stmt | block)]
func evalIf(e *Evaluator, node *syntax.Node, scope *Scope) (value.Value, error) {
	var condNode, thenBlock, elseNode *syntax.Node

	i := 0
	children := node.Children

	// Skip KEYWORD:"if"
	if i < len(children) && children[i].Type == "KEYWORD" && children[i].Value == "if" {
		i++
	}

	// Condition expression
	if i < len(children) && children[i].Type == "expr" {
		condNode = children[i]
		i++
	}

	// Then block
	if i < len(children) && children[i].Type == "block" {
		thenBlock = children[i]
		i++
	}

	// Optional else
	if i < len(children) && children[i].Type == "KEYWORD" && children[i].Value == "else" {
		i++
		if i < len(children) {
			elseNode = children[i]
		}
	}

	if condNode == nil || thenBlock == nil {
		return value.NullValue(), makeError(scope, node, "if: invalid statement")
	}

	cond, err := e.eval(condNode, scope)
	if err != nil {
		return cond, err
	}
	if isError(cond) {
		return cond, nil
	}

	if isTruthy(cond) {
		return e.eval(thenBlock, scope)
	} else if elseNode != nil {
		return e.eval(elseNode, scope)
	}

	return value.NullValue(), nil
}

// evalWhile evaluates a while statement.
// AST: while_stmt -> KEYWORD:"while" expr block
func evalWhile(e *Evaluator, node *syntax.Node, scope *Scope) (value.Value, error) {
	var condNode, bodyNode *syntax.Node

	for _, child := range node.Children {
		switch child.Type {
		case "expr":
			if condNode == nil {
				condNode = child
			}
		case "block":
			bodyNode = child
		}
	}

	if condNode == nil || bodyNode == nil {
		return value.NullValue(), makeError(scope, node, "while: invalid statement")
	}

	var result value.Value = value.NullValue()
	for {
		cond, err := e.eval(condNode, scope)
		if err != nil {
			return cond, err
		}
		if isError(cond) {
			return cond, nil
		}
		if !isTruthy(cond) {
			break
		}

		result, err = e.eval(bodyNode, scope)
		if err != nil {
			return result, err
		}
		if isError(result) {
			return result, nil
		}
		// Check for return
		if result.Type == value.Return {
			return result, nil
		}
	}

	return result, nil
}

// evalMatch evaluates a match statement.
func evalMatch(e *Evaluator, node *syntax.Node, scope *Scope) (value.Value, error) {
	var subject value.Value

	// Find the subject expression (first expr child)
	for _, child := range node.Children {
		if child.Type == "expr" {
			var err error
			subject, err = e.eval(child, scope)
			if err != nil {
				return subject, err
			}
			if isError(subject) {
				return subject, nil
			}
			break
		}
	}

	// Process match_arm children
	for _, child := range node.Children {
		if child.Type == "match_arm" {
			var patternNode, blockNode *syntax.Node
			for _, armChild := range child.Children {
				if armChild.Type == "pattern" {
					patternNode = armChild
				} else if armChild.Type == "block" {
					blockNode = armChild
				}
			}
			
			if patternNode != nil && blockNode != nil {
				if matchPattern(patternNode, subject, scope) {
					return e.eval(blockNode, scope)
				}
			}
		}
	}

	return value.NullValue(), nil
}

// matchPattern checks if a pattern matches a value, binding names if needed.
func matchPattern(pattern *syntax.Node, subject value.Value, scope *Scope) bool {
	for _, child := range pattern.Children {
		switch child.Type {
		case "TERMINAL":
			if child.Value == "_" {
				return true
			}
			if child.Value == "true" && isTruthy(subject) {
				return true
			}
			if child.Value == "false" && !isTruthy(subject) {
				return true
			}
		case "NAME":
			// _ is wildcard (matches anything without binding)
			if child.Value == "_" {
				return true
			}
			// Other names are binding patterns - always match and bind
			scope.Define(child.Value, subject)
			return true
		case "NUMBER":
			if subject.Type == value.Number {
				// Compare numbers
				patNum, _ := value.ParseNumber(child.Value)
				if subNum, ok := subject.Data.(float64); ok {
					return subNum == patNum
				}
			}
		case "SYMBOL":
			if subject.Type == value.Symbol {
				patSym := strings.TrimPrefix(child.Value, ":")
				if subSym, ok := subject.Data.(string); ok {
					return subSym == patSym
				}
			}
		case "literal":
			return matchPattern(child, subject, scope)
		}
	}
	return false
}

// ReturnValue wraps a return value for propagation.
type ReturnValue struct {
	Value value.Value
}
