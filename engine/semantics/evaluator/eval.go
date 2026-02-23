// Package eval provides the AST evaluator.
package eval

import (
	"math/big"
	"strconv"
	"strings"

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
		// Passthrough for single-child nodes
		if len(node.Children) == 1 {
			return e.Eval(node.Children[0], env)
		}
		return value.NULL
	}
}

func (e *Evaluator) evalProgram(node *syntax.Node, env *value.Env) value.Value {
	var result value.Value = value.NULL
	for _, child := range node.Children {
		result = e.Eval(child, env)
		if ret, ok := result.(*value.Return); ok {
			return ret.Val
		}
		if value.IsError(result) {
			return result
		}
	}
	return result
}

func (e *Evaluator) evalStatement(node *syntax.Node, env *value.Env) value.Value {
	if len(node.Children) > 0 {
		return e.Eval(node.Children[0], env)
	}
	return value.NULL
}

func (e *Evaluator) evalExprStmt(node *syntax.Node, env *value.Env) value.Value {
	if len(node.Children) > 0 {
		return e.Eval(node.Children[0], env)
	}
	return value.NULL
}

func (e *Evaluator) evalLet(node *syntax.Node, env *value.Env) value.Value {
	var name string
	var valNode *syntax.Node
	var isShape bool
	var shapeFields []string

	for i, child := range node.Children {
		if child.Type == "NAME" && name == "" {
			name = child.Value
		}
		if child.Type == "SHAPE" {
			isShape = true
			name = strings.TrimPrefix(child.Value, "@")
		}
		if child.Type == "field" {
			for _, f := range child.Children {
				if f.Type == "NAME" {
					shapeFields = append(shapeFields, f.Value)
				}
				if f.Type == "SHAPE" {
					shapeFields = append(shapeFields, f.Value)
				}
			}
		}
		if child.Type == "TERMINAL" && child.Value == "=" && i+1 < len(node.Children) {
			valNode = node.Children[i+1]
		}
	}

	if isShape {
		env.DefineShape(&value.ShapeDef{Name: name, Fields: shapeFields})
		return value.NULL
	}

	if name == "" {
		return e.makeError(node, env, "let: missing name")
	}
	if valNode == nil {
		return e.makeError(node, env, "let: missing value")
	}

	if e.hal[name] != nil {
		return e.makeError(node, env, "cannot shadow builtin: %s", name)
	}

	val := e.Eval(valNode, env)
	if value.IsError(val) {
		return val
	}

	if fn, ok := val.(*value.Function); ok {
		fn.Name = name
	}

	env.Set(name, val)
	return value.NULL
}

func (e *Evaluator) evalAssign(node *syntax.Node, env *value.Env) value.Value {
	var name string
	var valNode *syntax.Node

	for i, child := range node.Children {
		if child.Type == "NAME" && name == "" {
			name = child.Value
		}
		if child.Type == "TERMINAL" && child.Value == "=" && i+1 < len(node.Children) {
			valNode = node.Children[i+1]
		}
	}

	if name == "" || valNode == nil {
		return e.makeError(node, env, "assign: invalid statement")
	}

	if _, ok := env.Get(name); !ok {
		return e.makeError(node, env, "undefined variable: %s", name)
	}

	val := e.Eval(valNode, env)
	if value.IsError(val) {
		return val
	}

	if !env.Update(name, val) {
		return e.makeError(node, env, "cannot assign to: %s", name)
	}

	return value.NULL
}

func (e *Evaluator) evalReturn(node *syntax.Node, env *value.Env) value.Value {
	for _, child := range node.Children {
		if child.Type != "TERMINAL" {
			val := e.Eval(child, env)
			if value.IsError(val) {
				return val
			}
			return &value.Return{Val: val}
		}
	}
	return &value.Return{Val: value.NULL}
}

func (e *Evaluator) evalIf(node *syntax.Node, env *value.Env) value.Value {
	var cond *syntax.Node
	var thenBlock *syntax.Node
	var elseBlock *syntax.Node

	for i, child := range node.Children {
		if child.Type == "TERMINAL" && child.Value == "if" {
			if i+1 < len(node.Children) {
				cond = node.Children[i+1]
			}
		}
		if child.Type == "block" {
			if thenBlock == nil {
				thenBlock = child
			} else {
				elseBlock = child
			}
		}
		if child.Type == "if_stmt" {
			elseBlock = child
		}
	}

	if cond == nil || thenBlock == nil {
		return e.makeError(node, env, "if: invalid syntax")
	}

	condVal := e.Eval(cond, env)
	if value.IsError(condVal) {
		return condVal
	}

	if value.IsTruthy(condVal) {
		return e.Eval(thenBlock, env)
	} else if elseBlock != nil {
		return e.Eval(elseBlock, env)
	}

	return value.NULL
}

func (e *Evaluator) evalWhile(node *syntax.Node, env *value.Env) value.Value {
	var cond *syntax.Node
	var body *syntax.Node

	for i, child := range node.Children {
		if child.Type == "TERMINAL" && child.Value == "while" {
			if i+1 < len(node.Children) {
				cond = node.Children[i+1]
			}
		}
		if child.Type == "block" {
			body = child
		}
	}

	if cond == nil || body == nil {
		return e.makeError(node, env, "while: invalid syntax")
	}

	var result value.Value = value.NULL
	for {
		condVal := e.Eval(cond, env)
		if value.IsError(condVal) {
			return condVal
		}
		if !value.IsTruthy(condVal) {
			break
		}
		result = e.Eval(body, env)
		if ret, ok := result.(*value.Return); ok {
			return ret
		}
		if value.IsError(result) {
			return result
		}
	}
	return result
}

func (e *Evaluator) evalMatch(node *syntax.Node, env *value.Env) value.Value {
	// Find the expression to match
	var matchExpr *syntax.Node
	var cases []*syntax.Node

	for i, child := range node.Children {
		if child.Type == "TERMINAL" && child.Value == "match" {
			if i+1 < len(node.Children) {
				matchExpr = node.Children[i+1]
			}
		}
		if child.Type == "pattern" || child.Type == "block" {
			cases = append(cases, child)
		}
	}

	if matchExpr == nil {
		return e.makeError(node, env, "match: missing expression")
	}

	matchVal := e.Eval(matchExpr, env)
	if value.IsError(matchVal) {
		return matchVal
	}

	// Process pattern/block pairs
	for i := 0; i+1 < len(cases); i += 2 {
		pattern := cases[i]
		block := cases[i+1]

		bindings := make(map[string]value.Value)
		if e.matchPattern(pattern, matchVal, bindings, env) {
			matchEnv := value.NewEnclosedEnv(env)
			for k, v := range bindings {
				matchEnv.Set(k, v)
			}
			return e.Eval(block, matchEnv)
		}
	}

	return value.NULL
}

func (e *Evaluator) matchPattern(pattern *syntax.Node, val value.Value, bindings map[string]value.Value, env *value.Env) bool {
	// Handle pattern node wrapper
	if pattern.Type == "pattern" && len(pattern.Children) > 0 {
		return e.matchPattern(pattern.Children[0], val, bindings, env)
	}

	// Wildcard
	if pattern.Type == "NAME" && pattern.Value == "_" {
		return true
	}

	// Name binding
	if pattern.Type == "NAME" {
		bindings[pattern.Value] = val
		return true
	}

	// Literal match
	if pattern.Type == "NUMBER" || pattern.Type == "STRING" || pattern.Type == "SYMBOL" {
		patternVal := e.Eval(pattern, env)
		return e.valuesEqual(patternVal, val)
	}

	// List pattern
	if pattern.Type == "list_literal" {
		list, ok := val.(*value.List)
		if !ok {
			return false
		}
		return e.matchListPattern(pattern, list, bindings, env)
	}

	return false
}

func (e *Evaluator) matchListPattern(pattern *syntax.Node, list *value.List, bindings map[string]value.Value, env *value.Env) bool {
	var patternElems []*syntax.Node
	var shapeName string

	for _, child := range pattern.Children {
		if child.Type == "SHAPE" {
			shapeName = strings.TrimPrefix(child.Value, "@")
		} else if child.Type != "TERMINAL" {
			patternElems = append(patternElems, child)
		}
	}

	if shapeName != "" && list.Shape != shapeName {
		return false
	}

	if len(patternElems) != len(list.Elements) {
		return false
	}

	for i, pe := range patternElems {
		if !e.matchPattern(pe, list.Elements[i], bindings, env) {
			return false
		}
	}

	return true
}

func (e *Evaluator) evalBlock(node *syntax.Node, env *value.Env) value.Value {
	var result value.Value = value.NULL
	for _, child := range node.Children {
		if child.Type == "TERMINAL" {
			continue
		}
		result = e.Eval(child, env)
		if ret, ok := result.(*value.Return); ok {
			return ret
		}
		if value.IsError(result) {
			return result
		}
	}
	return result
}

func (e *Evaluator) evalExpr(node *syntax.Node, env *value.Env) value.Value {
	children := e.nonTerminalChildren(node)

	if len(children) == 0 {
		return value.NULL
	}
	if len(children) == 1 {
		return e.Eval(children[0], env)
	}

	// Check for pipe expression
	for _, child := range node.Children {
		if child.Type == "TERMINAL" && child.Value == "|>" {
			return e.evalPipe(node, env)
		}
	}

	// Check for infix expression
	return e.evalInfix(node, env)
}

func (e *Evaluator) evalPipe(node *syntax.Node, env *value.Env) value.Value {
	children := e.nonTerminalChildren(node)
	if len(children) == 0 {
		return value.NULL
	}

	result := e.Eval(children[0], env)
	if value.IsError(result) {
		return result
	}

	for i := 1; i < len(children); i++ {
		// The right side should be a call; inject result as first arg
		result = e.applyPipe(children[i], result, env)
		if value.IsError(result) {
			return result
		}
	}

	return result
}

func (e *Evaluator) applyPipe(node *syntax.Node, arg value.Value, env *value.Env) value.Value {
	// Find the function and its arguments
	fn := e.evalToFunction(node, env)
	if value.IsError(fn) {
		return fn
	}

	args := []value.Value{arg}
	args = append(args, e.collectCallArgs(node, env)...)

	return e.applyFunction(fn, args, node, env)
}

func (e *Evaluator) evalInfix(node *syntax.Node, env *value.Env) value.Value {
	children := e.nonTerminalChildren(node)
	if len(children) == 0 {
		return value.NULL
	}

	result := e.Eval(children[0], env)
	if value.IsError(result) {
		return result
	}

	// Process operators left to right
	i := 1
	for _, child := range node.Children {
		if child.Type == "TERMINAL" && isOperator(child.Value) {
			if i >= len(children) {
				break
			}
			right := e.Eval(children[i], env)
			if value.IsError(right) {
				return right
			}
			result = e.applyOperator(child.Value, result, right, node, env)
			if value.IsError(result) {
				return result
			}
			i++
		}
	}

	return result
}

func (e *Evaluator) evalPrimary(node *syntax.Node, env *value.Env) value.Value {
	children := e.nonTerminalChildren(node)
	if len(children) == 0 {
		return value.NULL
	}

	// Check for grouped expression
	for _, child := range node.Children {
		if child.Type == "TERMINAL" && child.Value == "(" {
			// Find expr inside parens
			for _, c := range children {
				if c.Type == "expr" {
					return e.Eval(c, env)
				}
			}
		}
	}

	// Evaluate first child and apply postfix operations
	result := e.Eval(children[0], env)
	if value.IsError(result) {
		return result
	}

	// Apply call, index, access
	for i := 1; i < len(children); i++ {
		child := children[i]
		switch child.Type {
		case "call":
			args := e.evalCallArgs(child, env)
			if len(args) == 1 && value.IsError(args[0]) {
				return args[0]
			}
			result = e.applyFunction(result, args, child, env)
		case "index":
			result = e.evalIndex(result, child, env)
		case "access":
			result = e.evalAccess(result, child, env)
		}
		if value.IsError(result) {
			return result
		}
	}

	return result
}

func (e *Evaluator) evalFunc(node *syntax.Node, env *value.Env) value.Value {
	var params []string
	var rest string
	var body *syntax.Node

	for _, child := range node.Children {
		if child.Type == "params" {
			params, rest = e.extractParams(child)
		}
		if child.Type == "block" {
			body = child
		}
	}

	return &value.Function{
		Params: params,
		Rest:   rest,
		Body:   body,
		Env:    env,
	}
}

func (e *Evaluator) extractParams(node *syntax.Node) ([]string, string) {
	var params []string
	var rest string

	for _, child := range node.Children {
		if child.Type == "param_list" {
			for _, p := range child.Children {
				if p.Type == "NAME" {
					params = append(params, p.Value)
				}
			}
		}
		if child.Type == "rest_param" {
			for _, p := range child.Children {
				if p.Type == "NAME" {
					rest = p.Value
				}
			}
		}
		if child.Type == "NAME" {
			params = append(params, child.Value)
		}
	}

	return params, rest
}

func (e *Evaluator) evalList(node *syntax.Node, env *value.Env) value.Value {
	var elements []value.Value
	var shape string

	for _, child := range node.Children {
		if child.Type == "SHAPE" {
			shape = strings.TrimPrefix(child.Value, "@")
		} else if child.Type != "TERMINAL" {
			val := e.Eval(child, env)
			if value.IsError(val) {
				return val
			}
			elements = append(elements, val)
		}
	}

	return &value.List{Elements: elements, Shape: shape}
}

func (e *Evaluator) evalNumber(node *syntax.Node, env *value.Env) value.Value {
	num, err := value.NewNumberFromString(node.Value)
	if err != nil {
		return e.makeError(node, env, "invalid number: %s", node.Value)
	}
	return num
}

func (e *Evaluator) evalString(node *syntax.Node, env *value.Env) value.Value {
	s, err := strconv.Unquote(node.Value)
	if err != nil {
		return e.makeError(node, env, "invalid string: %s", node.Value)
	}
	return &value.String{Val: s}
}

func (e *Evaluator) evalRune(node *syntax.Node, env *value.Env) value.Value {
	s := node.Value
	if len(s) >= 2 {
		s = s[1 : len(s)-1]
	}
	if len(s) == 2 && s[0] == '\\' {
		switch s[1] {
		case 'n':
			return &value.Rune{Val: '\n'}
		case 't':
			return &value.Rune{Val: '\t'}
		case 'r':
			return &value.Rune{Val: '\r'}
		case '\\':
			return &value.Rune{Val: '\\'}
		case '\'':
			return &value.Rune{Val: '\''}
		}
	}
	runes := []rune(s)
	if len(runes) > 0 {
		return &value.Rune{Val: runes[0]}
	}
	return value.NULL
}

func (e *Evaluator) evalSymbol(node *syntax.Node, env *value.Env) value.Value {
	return &value.Symbol{Val: strings.TrimPrefix(node.Value, ":")}
}

func (e *Evaluator) evalName(node *syntax.Node, env *value.Env) value.Value {
	name := node.Value

	// Check builtins first
	if builtin, ok := e.hal[name]; ok {
		return builtin
	}

	// Then environment
	if val, ok := env.Get(name); ok {
		return val
	}

	return e.makeError(node, env, "undefined: %s", name)
}

func (e *Evaluator) evalTerminal(node *syntax.Node, env *value.Env) value.Value {
	switch node.Value {
	case "true":
		return value.TRUE
	case "false":
		return value.FALSE
	}
	return value.NULL
}

// Helper methods

func (e *Evaluator) nonTerminalChildren(node *syntax.Node) []*syntax.Node {
	var result []*syntax.Node
	for _, child := range node.Children {
		if child.Type != "TERMINAL" {
			result = append(result, child)
		}
	}
	return result
}

func (e *Evaluator) makeError(node *syntax.Node, env *value.Env, format string, args ...interface{}) *value.Error {
	return value.NewErrorAt(
		env.GetFile(),
		node.Pos.Line,
		env.GetSourceLine(node.Pos.Line),
		format,
		args...,
	)
}

func (e *Evaluator) applyOperator(op string, left, right value.Value, node *syntax.Node, env *value.Env) value.Value {
	switch op {
	case "+":
		return e.opAdd(left, right, node, env)
	case "-":
		return e.opSub(left, right, node, env)
	case "*":
		return e.opMul(left, right, node, env)
	case "/":
		return e.opDiv(left, right, node, env)
	case "<":
		return e.opLt(left, right, node, env)
	case ">":
		return e.opGt(left, right, node, env)
	case "<=":
		return e.opLte(left, right, node, env)
	case ">=":
		return e.opGte(left, right, node, env)
	case "and":
		if !value.IsTruthy(left) {
			return left
		}
		return right
	case "or":
		if value.IsTruthy(left) {
			return left
		}
		return right
	default:
		return e.makeError(node, env, "unknown operator: %s", op)
	}
}

func (e *Evaluator) opAdd(left, right value.Value, node *syntax.Node, env *value.Env) value.Value {
	if ln, ok := left.(*value.Number); ok {
		if rn, ok := right.(*value.Number); ok {
			result := new(big.Rat).Add(ln.Val, rn.Val)
			return &value.Number{Val: result}
		}
	}
	if ls, ok := left.(*value.String); ok {
		if rs, ok := right.(*value.String); ok {
			return &value.String{Val: ls.Val + rs.Val}
		}
	}
	return e.makeError(node, env, "cannot add %s and %s", left.Type(), right.Type())
}

func (e *Evaluator) opSub(left, right value.Value, node *syntax.Node, env *value.Env) value.Value {
	if ln, ok := left.(*value.Number); ok {
		if rn, ok := right.(*value.Number); ok {
			result := new(big.Rat).Sub(ln.Val, rn.Val)
			return &value.Number{Val: result}
		}
	}
	return e.makeError(node, env, "cannot subtract %s and %s", left.Type(), right.Type())
}

func (e *Evaluator) opMul(left, right value.Value, node *syntax.Node, env *value.Env) value.Value {
	if ln, ok := left.(*value.Number); ok {
		if rn, ok := right.(*value.Number); ok {
			result := new(big.Rat).Mul(ln.Val, rn.Val)
			return &value.Number{Val: result}
		}
	}
	return e.makeError(node, env, "cannot multiply %s and %s", left.Type(), right.Type())
}

func (e *Evaluator) opDiv(left, right value.Value, node *syntax.Node, env *value.Env) value.Value {
	if ln, ok := left.(*value.Number); ok {
		if rn, ok := right.(*value.Number); ok {
			if rn.Val.Sign() == 0 {
				return e.makeError(node, env, "division by zero")
			}
			result := new(big.Rat).Quo(ln.Val, rn.Val)
			return &value.Number{Val: result}
		}
	}
	return e.makeError(node, env, "cannot divide %s and %s", left.Type(), right.Type())
}

func (e *Evaluator) opLt(left, right value.Value, node *syntax.Node, env *value.Env) value.Value {
	if ln, ok := left.(*value.Number); ok {
		if rn, ok := right.(*value.Number); ok {
			if ln.Val.Cmp(rn.Val) < 0 {
				return value.TRUE
			}
			return value.FALSE
		}
	}
	return e.makeError(node, env, "cannot compare %s and %s", left.Type(), right.Type())
}

func (e *Evaluator) opGt(left, right value.Value, node *syntax.Node, env *value.Env) value.Value {
	if ln, ok := left.(*value.Number); ok {
		if rn, ok := right.(*value.Number); ok {
			if ln.Val.Cmp(rn.Val) > 0 {
				return value.TRUE
			}
			return value.FALSE
		}
	}
	return e.makeError(node, env, "cannot compare %s and %s", left.Type(), right.Type())
}

func (e *Evaluator) opLte(left, right value.Value, node *syntax.Node, env *value.Env) value.Value {
	if ln, ok := left.(*value.Number); ok {
		if rn, ok := right.(*value.Number); ok {
			if ln.Val.Cmp(rn.Val) <= 0 {
				return value.TRUE
			}
			return value.FALSE
		}
	}
	return e.makeError(node, env, "cannot compare %s and %s", left.Type(), right.Type())
}

func (e *Evaluator) opGte(left, right value.Value, node *syntax.Node, env *value.Env) value.Value {
	if ln, ok := left.(*value.Number); ok {
		if rn, ok := right.(*value.Number); ok {
			if ln.Val.Cmp(rn.Val) >= 0 {
				return value.TRUE
			}
			return value.FALSE
		}
	}
	return e.makeError(node, env, "cannot compare %s and %s", left.Type(), right.Type())
}

func (e *Evaluator) applyFunction(fn value.Value, args []value.Value, node *syntax.Node, env *value.Env) value.Value {
	switch f := fn.(type) {
	case *value.Function:
		return e.applyUserFunction(f, args, node, env)
	case *value.Intrinsic:
		return e.applyIntrinsic(f, args, node, env)
	default:
		return e.makeError(node, env, "not a function: %s", fn.Type())
	}
}

func (e *Evaluator) applyUserFunction(fn *value.Function, args []value.Value, node *syntax.Node, env *value.Env) value.Value {
	fnEnv, ok := fn.Env.(*value.Env)
	if !ok {
		return e.makeError(node, env, "invalid function environment")
	}

	callEnv := value.NewEnclosedEnv(fnEnv)

	// Bind parameters
	for i, param := range fn.Params {
		if i < len(args) {
			callEnv.Set(param, args[i])
		} else {
			callEnv.Set(param, value.NULL)
		}
	}

	// Bind rest parameter
	if fn.Rest != "" {
		restStart := len(fn.Params)
		var restArgs []value.Value
		if restStart < len(args) {
			restArgs = args[restStart:]
		}
		callEnv.Set(fn.Rest, &value.List{Elements: restArgs})
	}

	body, ok := fn.Body.(*syntax.Node)
	if !ok {
		return e.makeError(node, env, "invalid function body")
	}

	result := e.Eval(body, callEnv)

	if ret, ok := result.(*value.Return); ok {
		return ret.Val
	}

	return result
}

func (e *Evaluator) applyIntrinsic(fn *value.Intrinsic, args []value.Value, node *syntax.Node, env *value.Env) value.Value {
	// Intrinsic.Fn should be set by HAL registration
	if fn.Fn == nil {
		return e.makeError(node, env, "intrinsic not implemented: %s", fn.Name)
	}

	// Cast and call - HAL provides the actual function
	if call, ok := fn.Fn.(func([]value.Value) value.Value); ok {
		return call(args)
	}

	return e.makeError(node, env, "invalid intrinsic: %s", fn.Name)
}

func (e *Evaluator) evalCallArgs(node *syntax.Node, env *value.Env) []value.Value {
	var args []value.Value
	for _, child := range node.Children {
		if child.Type != "TERMINAL" {
			val := e.Eval(child, env)
			args = append(args, val)
		}
	}
	return args
}

func (e *Evaluator) collectCallArgs(node *syntax.Node, env *value.Env) []value.Value {
	// Find call node in children
	for _, child := range node.Children {
		if child.Type == "call" {
			return e.evalCallArgs(child, env)
		}
	}
	return nil
}

func (e *Evaluator) evalToFunction(node *syntax.Node, env *value.Env) value.Value {
	// Recursively find the function value
	for _, child := range node.Children {
		if child.Type == "NAME" {
			return e.evalName(child, env)
		}
		if child.Type != "TERMINAL" && child.Type != "call" {
			return e.evalToFunction(child, env)
		}
	}
	return e.Eval(node, env)
}

func (e *Evaluator) evalIndex(val value.Value, node *syntax.Node, env *value.Env) value.Value {
	list, ok := val.(*value.List)
	if !ok {
		if s, ok := val.(*value.String); ok {
			return e.evalStringIndex(s, node, env)
		}
		return e.makeError(node, env, "cannot index %s", val.Type())
	}

	// Find index expression
	for _, child := range node.Children {
		if child.Type != "TERMINAL" {
			idx := e.Eval(child, env)
			if value.IsError(idx) {
				return idx
			}
			num, ok := idx.(*value.Number)
			if !ok {
				return e.makeError(node, env, "index must be a number")
			}
			if !num.Val.IsInt() {
				return e.makeError(node, env, "index must be an integer")
			}
			i := int(num.Val.Num().Int64())
			if i < 0 || i >= len(list.Elements) {
				return e.makeError(node, env, "index out of bounds: %d", i)
			}
			return list.Elements[i]
		}
	}

	return value.NULL
}

func (e *Evaluator) evalStringIndex(s *value.String, node *syntax.Node, env *value.Env) value.Value {
	for _, child := range node.Children {
		if child.Type != "TERMINAL" {
			idx := e.Eval(child, env)
			if value.IsError(idx) {
				return idx
			}
			num, ok := idx.(*value.Number)
			if !ok {
				return e.makeError(node, env, "index must be a number")
			}
			if !num.Val.IsInt() {
				return e.makeError(node, env, "index must be an integer")
			}
			i := int(num.Val.Num().Int64())
			runes := []rune(s.Val)
			if i < 0 || i >= len(runes) {
				return e.makeError(node, env, "index out of bounds: %d", i)
			}
			return &value.Rune{Val: runes[i]}
		}
	}
	return value.NULL
}

func (e *Evaluator) evalAccess(val value.Value, node *syntax.Node, env *value.Env) value.Value {
	list, ok := val.(*value.List)
	if !ok || list.Shape == "" {
		return e.makeError(node, env, "cannot access field on %s", val.Type())
	}

	// Find field name
	for _, child := range node.Children {
		if child.Type == "NAME" {
			fieldName := child.Value
			shapeDef, ok := env.GetShape(list.Shape)
			if !ok {
				return e.makeError(node, env, "unknown shape: %s", list.Shape)
			}
			for i, f := range shapeDef.Fields {
				if f == fieldName && i < len(list.Elements) {
					return list.Elements[i]
				}
			}
			return e.makeError(node, env, "no field %s in @%s", fieldName, list.Shape)
		}
	}

	return value.NULL
}

func (e *Evaluator) valuesEqual(a, b value.Value) bool {
	if a.Type() != b.Type() {
		return false
	}
	switch av := a.(type) {
	case *value.Number:
		bv := b.(*value.Number)
		return av.Val.Cmp(bv.Val) == 0
	case *value.String:
		bv := b.(*value.String)
		return av.Val == bv.Val
	case *value.Symbol:
		bv := b.(*value.Symbol)
		return av.Val == bv.Val
	case *value.Boolean:
		bv := b.(*value.Boolean)
		return av.Val == bv.Val
	case *value.Rune:
		bv := b.(*value.Rune)
		return av.Val == bv.Val
	case *value.Null:
		return true
	default:
		return false
	}
}

func isOperator(s string) bool {
	switch s {
	case "+", "-", "*", "/", "<", ">", "<=", ">=", "and", "or":
		return true
	}
	return false
}
