// Package evaluator provides the AST evaluator.
package evaluator

import (
	"fmt"

	"aiki/engine"
	"aiki/engine/runtime/hal"
	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
	"aiki/engine/syntax/grammar"
)

// handlerFunc is the signature for node handlers.
type handlerFunc func(*Evaluator, *syntax.Node, *value.Env) value.Value

// handlers maps node types to their evaluation functions.
// This is the source of truth for grammar-evaluator coupling.
// Initialized in init() to avoid circular references.
var handlers map[string]handlerFunc

func init() {
	handlers = map[string]handlerFunc{
		// Statements
		"program":      (*Evaluator).evalProgram,
		"statement":    (*Evaluator).evalStatement,
		"package_stmt": (*Evaluator).evalPackage,
		"let_stmt":     (*Evaluator).evalLet,
		"assign_stmt":  (*Evaluator).evalAssign,
		"return_stmt":  (*Evaluator).evalReturn,
		"expr_stmt":    (*Evaluator).evalExprStmt,
		"if_stmt":      (*Evaluator).evalIf,
		"while_stmt":   (*Evaluator).evalWhile,
		"match_stmt":   (*Evaluator).evalMatch,
		"block":        (*Evaluator).evalBlock,

		// Expressions
		"expr":         (*Evaluator).evalExpr,
		"pipe_expr":    (*Evaluator).evalExpr,
		"infix_expr":   (*Evaluator).evalExpr,
		"unary_expr":   (*Evaluator).evalExpr,
		"postfix_expr": (*Evaluator).evalExpr,
		"primary":      (*Evaluator).evalPrimary,

		// Literals
		"func_literal": (*Evaluator).evalFunc,
		"list_literal": (*Evaluator).evalList,

		// Tokens
		"NUMBER":    (*Evaluator).evalNumber,
		"STRING":    (*Evaluator).evalString,
		"RUNE":      (*Evaluator).evalRune,
		"SYMBOL":    (*Evaluator).evalSymbol,
		"SHAPE":     (*Evaluator).evalShape,
		"NAME":      (*Evaluator).evalName,
		"TERMINAL":  (*Evaluator).evalTerminal,
		"BINOP":     (*Evaluator).evalTerminal,
		"KEYWORD":   (*Evaluator).evalTerminal,
		"OPERATOR":  (*Evaluator).evalTerminal,
		"DELIMITER": (*Evaluator).evalTerminal,
		"NEWLINE":   (*Evaluator).evalTerminal,

		// Structural (nil = delegate to single child)
		"call":       nil,
		"index":      nil,
		"access":     nil,
		"params":     nil,
		"param_list": nil,
		"rest_param": nil,
		"pattern":    nil,
		"literal":    nil,
		"field":      nil,
	}
}

// Evaluator evaluates AST nodes.
type Evaluator struct {
	observer engine.Observer
	runtime  hal.RuntimeContract
	grammar  *grammar.Grammar
	Counters *Counters // nil = probes disabled
}

// New creates an evaluator with a runtime.
func New(runtime hal.RuntimeContract, observer engine.Observer) *Evaluator {
	if observer == nil {
		observer = engine.SilentObserver{}
	}
	return &Evaluator{
		observer: observer,
		runtime:  runtime,
	}
}

// SetGrammar sets the grammar and validates handler coverage.
func (e *Evaluator) SetGrammar(g *grammar.Grammar) {
	e.grammar = g
	e.validateHandlers()
}

// validateHandlers panics if any grammar production or token lacks a handler.
func (e *Evaluator) validateHandlers() {
	if e.grammar == nil {
		return
	}

	// Check productions
	for name := range e.grammar.Productions {
		if _, ok := handlers[name]; !ok {
			panic(fmt.Sprintf("grammar production has no handler: %s", name))
		}
	}

	// Check non-skip tokens
	for _, tok := range e.grammar.Tokens {
		if !tok.Skip {
			if _, ok := handlers[tok.Name]; !ok {
				panic(fmt.Sprintf("grammar token has no handler: %s", tok.Name))
			}
		}
	}
}

// Eval evaluates an AST node.
func (e *Evaluator) Eval(node *syntax.Node, env *value.Env) value.Value {
	e.observer.OnEval(node.Type, "", 0, node.Pos)

	if e.Counters != nil && node.Pos.Line > 0 {
		e.Counters.CoverHit(env.GetFile(), node.Pos.Line)
	}

	h, ok := handlers[node.Type]
	if !ok {
		// Should never happen after validation
		panic(fmt.Sprintf("no handler for node type: %s", node.Type))
	}

	// nil handler means delegate to single child
	if h == nil {
		if len(node.Children) == 1 {
			return e.Eval(node.Children[0], env)
		}
		return value.EMPTY
	}

	return h(e, node, env)
}
