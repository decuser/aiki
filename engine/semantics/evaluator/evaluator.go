// evaluator.go is the main evaluation entry point and dispatcher.
package evaluator

import (
	"aiki/engine"
	"aiki/engine/runtime/hal"
	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
)

// Evaluator executes Aiki programs by walking the AST.
type Evaluator struct {
	runtime  hal.RuntimeContract
	observer engine.Observer
	grammar  syntax.GrammarContract
}

// New creates a new Evaluator with the given runtime and observer.
func New(rt hal.RuntimeContract, obs engine.Observer) *Evaluator {
	return &Evaluator{
		runtime:  rt,
		observer: obs,
	}
}

// SetGrammar stores the grammar for use by intrinsics (import, load).
func (e *Evaluator) SetGrammar(g syntax.GrammarContract) {
	e.grammar = g
	ValidateHandlers(g)
}

// Eval evaluates an AST node in the given scope.
func (e *Evaluator) Eval(node *syntax.Node, scope *Scope) (value.Value, error) {
	if e.observer != nil {
		e.observer.OnReady("go-substrate", 0)
	}
	return e.eval(node, scope)
}

// eval is the internal evaluation dispatcher.
func (e *Evaluator) eval(node *syntax.Node, scope *Scope) (value.Value, error) {
	if node == nil {
		return value.NullValue(), nil
	}

	// Notify observer
	if e.observer != nil {
		e.observer.OnEval(node.Type, value.NullValue(), 0, node.Pos)
	}

	// Dispatch to handler
	if handler, ok := handlers[node.Type]; ok {
		return handler(e, node, scope)
	}

	// Default: if single child, delegate
	if len(node.Children) == 1 {
		return e.eval(node.Children[0], scope)
	}

	return value.NullValue(), nil
}

// RunSource parses and evaluates source code.
func (e *Evaluator) RunSource(name, source string, scope *Scope) (value.Value, error) {
	if e.grammar == nil {
		return value.NullValue(), makeError(scope, nil, "grammar not initialized")
	}
	scope.SetFile(name)
	lexer := syntax.NewLexer(name, source, e.grammar)
	parser, err := syntax.NewParser(lexer, e.grammar)
	if err != nil {
		return value.NullValue(), err
	}
	ast, err := parser.Parse()
	if err != nil {
		return value.NullValue(), err
	}
	scope.SetSource(source)
	return e.Eval(ast, scope)
}

// RunFile reads, parses, and evaluates a file.
func (e *Evaluator) RunFile(filename string, scope *Scope) (value.Value, error) {
	// File reading goes through HAL
	content, err := e.runtime.ReadFile(filename)
	if err != nil {
		return value.NullValue(), err
	}

	scope.SetFile(filename)
	return e.RunSource(filename, string(content), scope)
}

// isTruthy determines if a value is truthy.
func isTruthy(val value.Value) bool {
	switch val.Type {
	case value.Boolean:
		if b, ok := val.Data.(bool); ok {
			return b
		}
		return false
	case value.Null:
		return false
	default:
		return true
	}
}
