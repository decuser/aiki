// Package eval provides the AST evaluator.
package eval

import (
	"aiki/engine"
	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
)

// Evaluator evaluates AST nodes.
type Evaluator struct {
	observer engine.Observer
	hal      map[string]value.Value // builtins
}

// New creates an evaluator.
func New(observer engine.Observer) *Evaluator {
	if observer == nil {
		observer = engine.SilentObserver{}
	}
	return &Evaluator{
		observer: observer,
		hal:      make(map[string]value.Value),
	}
}

// RegisterBuiltin adds a builtin function.
func (e *Evaluator) RegisterBuiltin(name string, fn value.Value) {
	e.hal[name] = fn
}

// Eval evaluates an AST node.
func (e *Evaluator) Eval(node *syntax.Node, env *value.Env) value.Value {
	e.observer.OnEval(node.Type, "", 0, node.Pos)

	switch node.Type {
	case "program":
		return e.evalProgram(node, env)
	case "statement":
		return e.evalStatement(node, env)
	case "let_stmt":
		return e.evalLet(node, env)
	case "assign_stmt":
		return e.evalAssign(node, env)
	case "return_stmt":
		return e.evalReturn(node, env)
	case "expr_stmt":
		return e.evalExprStmt(node, env)
	case "if_stmt":
		return e.evalIf(node, env)
	case "while_stmt":
		return e.evalWhile(node, env)
	case "match_stmt":
		return e.evalMatch(node, env)
	case "block":
		return e.evalBlock(node, env)
	case "expr", "pipe_expr", "infix_expr", "unary_expr", "postfix_expr":
		return e.evalExpr(node, env)
	case "primary":
		return e.evalPrimary(node, env)
	case "func_literal":
		return e.evalFunc(node, env)
	case "list_literal":
		return e.evalList(node, env)
	case "NUMBER":
		return e.evalNumber(node, env)
	case "STRING":
		return e.evalString(node, env)
	case "RUNE":
		return e.evalRune(node, env)
	case "SYMBOL":
		return e.evalSymbol(node, env)
	case "NAME":
		return e.evalName(node, env)
	case "TERMINAL":
		return e.evalTerminal(node, env)
	default:
		if len(node.Children) == 1 {
			return e.Eval(node.Children[0], env)
		}
		return value.NULL
	}
}
