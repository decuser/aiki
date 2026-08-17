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
		"select_stmt":  (*Evaluator).evalSelect,
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
		"NUMBER":   (*Evaluator).evalNumber,
		"STRING":   (*Evaluator).evalString,
		"RUNE":     (*Evaluator).evalRune,
		"SYMBOL":   (*Evaluator).evalSymbol,
		"SHAPE":    (*Evaluator).evalShape,
		"NAME":     (*Evaluator).evalName,
		"TERMINAL": (*Evaluator).evalTerminal,
		"BINOP":    (*Evaluator).evalTerminal,

		// Structural (nil = delegate to single child)
		"call":           nil,
		"index":          nil,
		"access":         nil,
		"params":         nil,
		"param_list":     nil,
		"rest_param":     nil,
		"pattern":        nil,
		"select_case":    nil,
		"select_default": nil,
		"literal":        nil,
		"field":          nil,
	}
}

// Evaluator evaluates AST nodes.
type Evaluator struct {
	observer        engine.Observer
	runtime         hal.RuntimeContract
	grammar         *grammar.Grammar
	binaryOperators map[string]struct{}
	Counters        *Counters // nil = probes disabled
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
	e.binaryOperators = g.Analysis().TerminalAlternatives("BINOP")
	e.validateHandlers()
	if err := validateBinaryOperatorCoverage(e.binaryOperators, eagerBinaryOperatorSemantics, lazyLogicalOperators); err != nil {
		panic(err)
	}
}

// validateHandlers panics if grammar/evaluator dispatch coverage has drifted.
func (e *Evaluator) validateHandlers() {
	if e.grammar == nil {
		return
	}
	if err := validateHandlerCoverage(e.grammar, handlers); err != nil {
		panic(err)
	}
}

// syntheticHandlerNodes are parser-produced node types that do not correspond
// to a named grammar production or TokenRef. Keep this list explicit. TERMINAL
// nodes are synthesized by engine/syntax/parser.go while matching literal
// terminals; grammar analysis intentionally reports only grammar-produced types.
var syntheticHandlerNodes = map[string]struct{}{
	"TERMINAL": {},
}

func validateHandlerCoverage(g *grammar.Grammar, hs map[string]handlerFunc) error {
	want := g.Analysis().ASTNodeTypes()
	for name := range syntheticHandlerNodes {
		want[name] = struct{}{}
	}

	for name := range want {
		if _, ok := hs[name]; !ok {
			return fmt.Errorf("grammar AST node has no evaluator handler: %s", name)
		}
	}
	for name := range hs {
		if _, ok := want[name]; !ok {
			return fmt.Errorf("evaluator handler has no grammar AST node: %s", name)
		}
	}
	return nil
}

func (e *Evaluator) isBinaryOperator(s string) bool {
	_, ok := e.binaryOperators[s]
	return ok
}

func (e *Evaluator) activeProbe(env *value.Env) engine.SemanticProbe {
	if env != nil {
		if probe := env.GetSemanticProbe(); probe != nil {
			return probe
		}
	}
	if e.Counters != nil {
		return e.Counters
	}
	return nil
}

func semanticSite(node *syntax.Node, env *value.Env) engine.SemanticSite {
	if node == nil || env == nil {
		return engine.SemanticSite{}
	}
	site := engine.SemanticSite{
		File: env.GetFile(),
		Line: node.Pos.Line,
		Col:  node.Pos.Col,
	}
	if frame, ok := env.CurrentFrame(); ok {
		site.Function = frame.Name
	}
	if site.Line > 0 {
		site.Source = env.GetSourceLine(site.Line)
	}
	return site
}

func (e *Evaluator) semanticHit(kind engine.SemanticKind, node *syntax.Node, env *value.Env) {
	probe := e.activeProbe(env)
	if probe == nil {
		return
	}
	site := engine.SemanticSite{}
	if attributed, ok := probe.(engine.AttributionProbe); ok && attributed.WantsSites() {
		site = semanticSite(node, env)
	}
	probe.Hit(kind, site)
}

func (e *Evaluator) semanticHitDetail(kind engine.SemanticKind, detail string, node *syntax.Node, env *value.Env) {
	probe := e.activeProbe(env)
	if probe == nil {
		return
	}
	site := engine.SemanticSite{}
	if attributed, ok := probe.(engine.AttributionProbe); ok && attributed.WantsSites() {
		site = semanticSite(node, env)
		site.Detail = detail
	}
	probe.Hit(kind, site)
}

func (e *Evaluator) measure(fn value.Value, args []value.Value, node *syntax.Node, env *value.Env, attributed bool) (value.Value, engine.SemanticMeasurement) {
	var counters *Counters
	if attributed {
		counters = NewAttributedCounters()
	} else {
		counters = NewCounters()
	}
	measureEnv := value.NewEnclosedEnv(env)
	measureEnv.SetSemanticProbe(counters)
	result := e.applyFunction(fn, args, node, measureEnv)
	return result, counters.Measurement()
}

func (e *Evaluator) withProfileLabels(labels engine.ProfileLabels, restore engine.ProfileLabels, fn func()) {
	if labeler, ok := e.runtime.(hal.ProfileLabeler); ok {
		labeler.WithProfileLabels(labels, restore, fn)
		return
	}
	fn()
}

func semanticProfileLabels(env *value.Env) engine.ProfileLabels {
	labels := engine.ProfileLabels{Layer: "semantic"}
	if env == nil {
		return labels
	}
	labels.File = env.GetFile()
	if frame, ok := env.CurrentFrame(); ok {
		labels.Function = frame.Name
	} else {
		labels.Function = "<main>"
	}
	return labels
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
