// handlers.go defines handler registration and validation.
package evaluator

import (
	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
	"fmt"
)

// Handler is a function that evaluates a specific node type.
type Handler func(e *Evaluator, node *syntax.Node, scope *Scope) (value.Value, error)

// handlers maps node types to their evaluation functions.
var handlers map[string]Handler

// initHandlers initializes the handler map.
// This is called once during package initialization.
func initHandlers() {
	handlers = map[string]Handler{
		// Program and blocks
		"program":   evalProgram,
		"statement": evalStatement,
		"block":     evalBlock,

		// Statements
		"let_stmt":    evalLet,
		"assign_stmt": evalAssign,
		"return_stmt": evalReturn,
		"expr_stmt":   evalExprStmt,
		"if_stmt":     evalIf,
		"while_stmt":  evalWhile,
		"match_stmt":  evalMatch,

		// Expressions
		"expr":         evalExpr,
		"pipe_expr":    evalPipe,
		"pipe_tail":    evalPassthrough,
		"infix_expr":   evalInfix,
		"infix_tail":   evalPassthrough,
		"infix_op":     evalPassthrough,
		"unary_expr":   evalUnary,
		"postfix_expr": evalPostfix,
		"postfix_op":   evalPassthrough,
		"primary":      evalPrimary,

		// Literals
		"lambda":      evalFuncLiteral,
		"func_literal": evalFuncLiteral,
		"list_literal": evalListLiteral,
		"NUMBER":       evalNumber,
		"STRING":       evalString,
		"RUNE":         evalRune,
		"SYMBOL":       evalSymbol,
		"NAME":         evalName,
		
		// Token types from this grammar
		"KEYWORD":      evalKeyword,
		"OPERATOR":     evalPassthrough,
		"PUNCT":        evalPassthrough,
		"DELIMITER":    evalPassthrough,

		// Passthrough nodes (delegate to first child)
		"field":        evalPassthrough,
		"field_list":   evalPassthrough,
		"field_tail":   evalPassthrough,
		"pattern":      evalPassthrough,
		"pattern_list": evalPassthrough,
		"pattern_tail": evalPassthrough,
		"literal":      evalPassthrough,
		"call":         evalPassthrough,
		"index":        evalPassthrough,
		"access":       evalPassthrough,
		"params":       evalPassthrough,
		"param_list":   evalPassthrough,
		"param_tail":   evalPassthrough,
		"rest_param":   evalPassthrough,
		"arg_list":     evalPassthrough,
		"arg_tail":     evalPassthrough,
		"else_clause":  evalPassthrough,
		"match_arm":    evalPassthrough,
		"shape_args":   evalPassthrough,
		"BINOP":        evalPassthrough,
	}
}

// ValidateHandlers checks that all grammar productions have handlers.
// Panics if any production lacks a handler.
func ValidateHandlers(contract syntax.GrammarContract) {
	if contract == nil {
		return
	}

	var missing []string

	// Check all production names from the grammar
	// We need to iterate through known productions
	knownProductions := []string{
		"program", "statement", "block",
		"let_stmt", "assign_stmt", "return_stmt", "expr_stmt",
		"if_stmt", "while_stmt", "match_stmt",
		"expr", "pipe_expr", "infix_expr", "unary_expr", "postfix_expr",
		"primary", "func_literal", "list_literal",
		"field", "pattern", "literal", "call", "index", "access",
		"params", "param_list", "rest_param",
	}

	for _, name := range knownProductions {
		if _, ok := contract.GetProduction(name); ok {
			if _, hasHandler := handlers[name]; !hasHandler {
				missing = append(missing, name)
			}
		}
	}

	if len(missing) > 0 {
		panic(fmt.Sprintf("eval: missing handlers for grammar productions: %v", missing))
	}
}

// evalPassthrough delegates to the first non-terminal child.
func evalPassthrough(e *Evaluator, node *syntax.Node, scope *Scope) (value.Value, error) {
	if len(node.Children) == 1 {
		return e.eval(node.Children[0], scope)
	}
	child := firstNonTerminal(node)
	if child != nil {
		return e.eval(child, scope)
	}
	return value.NullValue(), nil
}

func init() {
	initHandlers()
}
