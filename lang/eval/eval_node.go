package eval

import (
	"math/big"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"aiki/ebnf"
	"aiki/hal/core"
	"aiki/lang/value"
)

// HAL is the single source of truth for all builtins.
var HAL = core.HAL

// makeError creates a rich error with file, line, source context, and stack trace.
func makeError(env *value.Env, node *ebnf.Node, format string, args ...interface{}) *value.Error {
	return value.NewErrorAt(
		env.GetFile(),
		node.Line,
		env.GetSourceLine(node.Line),
		env.CopyStack(),
		format,
		args...,
	)
}

// isError checks if a value is an error.
func isError(val value.Value) bool {
	return val != nil && val.Type() == value.ErrorType
}

// isTruthy checks if a value is truthy.
func isTruthy(val value.Value) bool {
	switch v := val.(type) {
	case *value.Boolean:
		return v.Value
	case *value.Null:
		return false
	default:
		return true
	}
}

// EvalNode evaluates a generic AST node from the ebnf parser.
func EvalNode(node *ebnf.Node, env *value.Env) value.Value {
	switch node.Type {
	case "program":
		return evalNodeProgram(node, env)

	case "statement":
		if len(node.Children) > 0 {
			return EvalNode(node.Children[0], env)
		}
		return value.NULL

	case "let_stmt":
		return evalNodeLet(node, env)

	case "assign_stmt":
		return evalNodeAssign(node, env)

	case "return_stmt":
		return evalNodeReturn(node, env)

	case "expr_stmt":
		if len(node.Children) > 0 {
			return EvalNode(node.Children[0], env)
		}
		return value.NULL

	case "if_stmt":
		return evalNodeIf(node, env)

	case "while_stmt":
		return evalNodeWhile(node, env)

	case "match_stmt":
		return evalNodeMatch(node, env)

	case "block":
		return evalNodeBlock(node, env)

	case "expr", "pipe_expr", "infix_expr", "unary_expr", "postfix_expr":
		return evalNodeExpr(node, env)

	case "primary":
		return evalNodePrimary(node, env)

	case "func_literal":
		return evalNodeFunc(node, env)

	case "list_literal":
		return evalNodeList(node, env)

	case "NUMBER":
		return evalNodeNumber(node, env)

	case "STRING":
		s, err := strconv.Unquote(node.Value)
		if err != nil {
			return makeError(env, node, "invalid string: %s", node.Value)
		}
		return &value.String{Value: s}

	case "RUNE":
		s := node.Value
		if len(s) >= 2 {
			s = s[1 : len(s)-1]
		}
		// Handle escape sequences
		if len(s) == 2 && s[0] == '\\' {
			switch s[1] {
			case 'n':
				return &value.Rune{Value: '\n'}
			case 't':
				return &value.Rune{Value: '\t'}
			case 'r':
				return &value.Rune{Value: '\r'}
			case '\\':
				return &value.Rune{Value: '\\'}
			case '\'':
				return &value.Rune{Value: '\''}
			}
		}
		runes := []rune(s)
		if len(runes) > 0 {
			return &value.Rune{Value: runes[0]}
		}
		return value.NULL

	case "SYMBOL":
		return &value.Symbol{Value: strings.TrimPrefix(node.Value, ":")}

	case "NAME":
		return evalNodeIdentWithNode(node, env)

	case "TERMINAL":
		switch node.Value {
		case "true":
			return value.True
		case "false":
			return value.False
		}
		return value.NULL

	default:
		if len(node.Children) == 1 {
			return EvalNode(node.Children[0], env)
		}
		return value.NULL
	}
}

func evalNodeProgram(node *ebnf.Node, env *value.Env) value.Value {
	var result value.Value = value.NULL

	for _, child := range node.Children {
		result = EvalNode(child, env)

		switch r := result.(type) {
		case *value.Return:
			return r.Value
		case *value.Error:
			return r
		}
	}

	return result
}

func evalNodeLet(node *ebnf.Node, env *value.Env) value.Value {
	// let_stmt: TERMINAL:"let" NAME TERMINAL:"=" expr
	// Or shape: TERMINAL:"let" SHAPE TERMINAL:"[" fields TERMINAL:"]"

	var name string
	var nameNode *ebnf.Node
	var valNode *ebnf.Node
	var isShape bool
	var shapeFields []string

	for i, child := range node.Children {
		if child.Type == "NAME" && name == "" {
			name = child.Value
			nameNode = child
		}
		if child.Type == "SHAPE" {
			isShape = true
			name = strings.TrimPrefix(child.Value, "@")
			nameNode = child
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
			// Also handle direct NAME/SHAPE in field
			if len(child.Children) == 0 {
				if child.Type == "NAME" {
					shapeFields = append(shapeFields, child.Value)
				}
			}
		}
		if child.Type == "TERMINAL" && child.Value == "=" && i+1 < len(node.Children) {
			valNode = node.Children[i+1]
		}
	}

	if isShape {
		// Shape definition
		def := &value.ShapeDef{
			Name:   name,
			Fields: shapeFields,
		}
		env.DefineShape(def)
		return value.NULL
	}

	if name == "" {
		return makeError(env, node, "let: missing name")
	}
	if valNode == nil {
		return makeError(env, node, "let: missing value")
	}

	if HAL[name] != nil {
		return makeError(env, nameNode, "cannot shadow builtin: %s", name)
	}

	val := EvalNode(valNode, env)
	if isError(val) {
		return val
	}

	if fn, ok := val.(*value.Function); ok {
		fn.Name = name
	}

	env.Set(name, val)
	return value.NULL
}

func evalNodeAssign(node *ebnf.Node, env *value.Env) value.Value {
	var name string
	var nameNode *ebnf.Node
	var valNode *ebnf.Node

	for i, child := range node.Children {
		if child.Type == "NAME" && name == "" {
			name = child.Value
			nameNode = child
		}
		if child.Type == "TERMINAL" && child.Value == "=" && i+1 < len(node.Children) {
			valNode = node.Children[i+1]
		}
	}

	if name == "" || valNode == nil {
		return makeError(env, node, "assign: invalid statement")
	}

	val := EvalNode(valNode, env)
	if isError(val) {
		return val
	}

	if !env.Update(name, val) {
		return makeError(env, nameNode, "undefined: %s", name)
	}
	return value.NULL
}

func evalNodeReturn(node *ebnf.Node, env *value.Env) value.Value {
	for _, child := range node.Children {
		if child.Type != "TERMINAL" {
			val := EvalNode(child, env)
			if isError(val) {
				return val
			}
			return &value.Return{Value: val}
		}
	}
	return &value.Return{Value: value.NULL}
}

func evalNodeIf(node *ebnf.Node, env *value.Env) value.Value {
	var condNode, thenBlock, elseNode *ebnf.Node

	sawIf := false
	sawElse := false
	for _, child := range node.Children {
		if child.Type == "TERMINAL" && child.Value == "if" {
			sawIf = true
			continue
		}
		if child.Type == "TERMINAL" && child.Value == "else" {
			sawElse = true
			continue
		}
		if sawIf && condNode == nil && child.Type != "block" {
			condNode = child
			continue
		}
		if child.Type == "block" && thenBlock == nil {
			thenBlock = child
			continue
		}
		if sawElse {
			elseNode = child
		}
	}

	if condNode == nil || thenBlock == nil {
		return makeError(env, node, "if: invalid statement")
	}

	cond := EvalNode(condNode, env)
	if isError(cond) {
		return cond
	}

	if isTruthy(cond) {
		return EvalNode(thenBlock, env)
	} else if elseNode != nil {
		return EvalNode(elseNode, env)
	}

	return value.NULL
}

func evalNodeWhile(node *ebnf.Node, env *value.Env) value.Value {
	var condNode, bodyNode *ebnf.Node

	for _, child := range node.Children {
		if child.Type == "TERMINAL" {
			continue
		}
		if child.Type == "block" {
			bodyNode = child
		} else if condNode == nil {
			condNode = child
		}
	}

	if condNode == nil || bodyNode == nil {
		return makeError(env, node, "while: invalid statement")
	}

	var result value.Value = value.NULL
	for {
		cond := EvalNode(condNode, env)
		if isError(cond) {
			return cond
		}
		if !isTruthy(cond) {
			break
		}

		result = EvalNode(bodyNode, env)
		if isError(result) {
			return result
		}
		if _, ok := result.(*value.Return); ok {
			return result
		}
	}

	return result
}

func evalNodeMatch(node *ebnf.Node, env *value.Env) value.Value {
	// match_stmt = "match" expr "{" { pattern block } "}"
	var subject value.Value

	// Find the expression to match
	for _, child := range node.Children {
		if child.Type == "TERMINAL" {
			continue
		}
		if child.Type != "pattern" && child.Type != "block" {
			subject = EvalNode(child, env)
			if isError(subject) {
				return subject
			}
			break
		}
	}

	// Find and evaluate arms
	var currentPattern *ebnf.Node
	for _, child := range node.Children {
		if child.Type == "pattern" {
			currentPattern = child
			continue
		}
		if child.Type == "block" && currentPattern != nil {
			if matchNodePattern(currentPattern, subject, env) {
				return EvalNode(child, env)
			}
			currentPattern = nil
		}
	}

	return value.NULL
}

func matchNodePattern(pattern *ebnf.Node, subject value.Value, env *value.Env) bool {
	// Simple pattern matching
	for _, child := range pattern.Children {
		switch child.Type {
		case "TERMINAL":
			if child.Value == "_" {
				return true // wildcard
			}
			// Literal match
			if child.Value == "true" && isTruthy(subject) {
				return true
			}
			if child.Value == "false" && !isTruthy(subject) {
				return true
			}
		case "NAME":
			// Bind name to subject
			env.Set(child.Value, subject)
			return true
		case "NUMBER":
			if num, ok := subject.(*value.Number); ok {
				patNum := evalNodeNumber(child, env)
				if pn, ok := patNum.(*value.Number); ok {
					return num.Value.Cmp(pn.Value) == 0
				}
			}
		case "SYMBOL":
			if sym, ok := subject.(*value.Symbol); ok {
				return sym.Value == strings.TrimPrefix(child.Value, ":")
			}
		case "literal":
			return matchNodePattern(child, subject, env)
		}
	}
	return false
}

func evalNodeBlock(node *ebnf.Node, env *value.Env) value.Value {
	blockEnv := value.NewEnv(env)
	var result value.Value = value.NULL

	for _, child := range node.Children {
		if child.Type == "TERMINAL" {
			continue
		}
		if child.Type == "statement" {
			result = EvalNode(child, blockEnv)
			if result != nil {
				if result.Type() == value.ReturnType || result.Type() == value.ErrorType {
					return result
				}
			}
		}
	}

	return result
}

func evalNodeExpr(node *ebnf.Node, env *value.Env) value.Value {
	switch node.Type {
	case "pipe_expr":
		return evalNodePipe(node, env)
	case "infix_expr":
		return evalNodeInfix(node, env)
	case "unary_expr":
		return evalNodeUnary(node, env)
	case "postfix_expr":
		return evalNodePostfix(node, env)
	}

	if len(node.Children) == 1 {
		return EvalNode(node.Children[0], env)
	}

	var result value.Value = value.NULL
	for _, child := range node.Children {
		result = EvalNode(child, env)
		if isError(result) {
			return result
		}
	}
	return result
}

func evalNodePipe(node *ebnf.Node, env *value.Env) value.Value {
	var result value.Value

	i := 0
	for i < len(node.Children) {
		child := node.Children[i]

		if child.Type == "TERMINAL" && child.Value == "|>" {
			i++
			continue
		}

		if result == nil {
			result = EvalNode(child, env)
			if isError(result) {
				return result
			}
			i++
			continue
		}

		// Pipe: inject result as first arg to the call
		result = evalPipeCallSafe(child, result, env)
		if isError(result) {
			return result
		}
		i++
	}

	return result
}

func evalNodeInfix(node *ebnf.Node, env *value.Env) value.Value {
	var result value.Value
	var op string
	var opNode *ebnf.Node

	for _, child := range node.Children {
		// Check for operator in BINOP production or as terminal
		if child.Type == "BINOP" {
			for _, c := range child.Children {
				if c.Type == "TERMINAL" {
					op = c.Value
					opNode = c
				}
			}
			continue
		}
		if child.Type == "TERMINAL" && isOp(child.Value) {
			op = child.Value
			opNode = child
			continue
		}

		val := EvalNode(child, env)
		if isError(val) {
			return val
		}

		if result == nil {
			result = val
		} else if op != "" {
			result = applyOp(op, result, val, env, opNode)
			if isError(result) {
				return result
			}
			op = ""
			opNode = nil
		}
	}

	return result
}

func evalNodeUnary(node *ebnf.Node, env *value.Env) value.Value {
	var prefixes []string
	var prefixNodes []*ebnf.Node
	var operand value.Value

	for _, child := range node.Children {
		if child.Type == "TERMINAL" {
			if child.Value == "not" || child.Value == "-" {
				prefixes = append(prefixes, child.Value)
				prefixNodes = append(prefixNodes, child)
				continue
			}
		}
		operand = EvalNode(child, env)
		if isError(operand) {
			return operand
		}
	}

	if operand == nil {
		return value.NULL
	}

	// Apply prefixes in reverse order (innermost first)
	for i := len(prefixes) - 1; i >= 0; i-- {
		switch prefixes[i] {
		case "not":
			operand = value.NativeBoolToBoolean(!isTruthy(operand))
		case "-":
			if num, ok := operand.(*value.Number); ok {
				operand = &value.Number{Value: new(big.Rat).Neg(num.Value)}
			} else {
				return makeError(env, prefixNodes[i], "cannot negate %s", operand.Type())
			}
		}
	}

	return operand
}

func evalNodePostfix(node *ebnf.Node, env *value.Env) value.Value {
	var result value.Value

	for _, child := range node.Children {
		if result == nil {
			result = EvalNode(child, env)
			if isError(result) {
				return result
			}
			continue
		}

		switch child.Type {
		case "call":
			result = evalNodeCallSafe(result, child, env)
		case "index":
			result = evalNodeIndex(result, child, env)
		case "access":
			result = evalNodeAccess(result, child, env)
		}

		if isError(result) {
			return result
		}
	}

	return result
}

func evalNodePrimary(node *ebnf.Node, env *value.Env) value.Value {
	for _, child := range node.Children {
		switch child.Type {
		case "NUMBER", "STRING", "RUNE", "SYMBOL", "NAME":
			return EvalNode(child, env)
		case "TERMINAL":
			if child.Value == "true" {
				return value.True
			}
			if child.Value == "false" {
				return value.False
			}
		case "list_literal", "func_literal", "expr", "pipe_expr":
			return EvalNode(child, env)
		}
	}
	return value.NULL
}

func evalNodeFunc(node *ebnf.Node, env *value.Env) value.Value {
	var params []string
	var restParam string
	var body *ebnf.Node

	for _, child := range node.Children {
		if child.Type == "params" {
			params, restParam = extractNodeParams(child)
		}
		if child.Type == "block" {
			body = child
		}
	}

	return &value.Function{
		Parameters: params,
		RestParam:  restParam,
		BodyNode:   body,
		Env:        env,
	}
}

func extractNodeParams(node *ebnf.Node) (params []string, rest string) {
	for _, child := range node.Children {
		switch child.Type {
		case "param_list":
			for _, p := range child.Children {
				if p.Type == "NAME" {
					params = append(params, p.Value)
				}
			}
		case "rest_param":
			for _, p := range child.Children {
				if p.Type == "NAME" {
					rest = p.Value
				}
			}
		case "NAME":
			params = append(params, child.Value)
		}
	}
	return
}

func evalNodeList(node *ebnf.Node, env *value.Env) value.Value {
	var elements []value.Value
	var shape string

	for _, child := range node.Children {
		if child.Type == "TERMINAL" {
			continue
		}
		if child.Type == "SHAPE" {
			shape = strings.TrimPrefix(child.Value, "@")
			continue
		}
		val := EvalNode(child, env)
		if isError(val) {
			return val
		}
		elements = append(elements, val)
	}

	list := &value.List{Elements: elements}
	if shape != "" {
		list.Shape = shape
		if fields, ok := env.ResolveFields(shape); ok {
			list.Fields = fields
		}
	}
	return list
}

func evalNodeIndex(target value.Value, node *ebnf.Node, env *value.Env) value.Value {
	var idxVal value.Value

	for _, child := range node.Children {
		if child.Type != "TERMINAL" {
			idxVal = EvalNode(child, env)
			break
		}
	}

	if isError(idxVal) {
		return idxVal
	}

	idx, ok := idxVal.(*value.Number)
	if !ok {
		return makeError(env, node, "index must be number, got %s", idxVal.Type())
	}

	i := int(idx.Value.Num().Int64())

	switch t := target.(type) {
	case *value.List:
		if i < 0 || i >= len(t.Elements) {
			return makeError(env, node, "index out of bounds: %d", i)
		}
		return t.Elements[i]
	case *value.String:
		runes := []rune(t.Value)
		if i < 0 || i >= len(runes) {
			return makeError(env, node, "index out of bounds: %d", i)
		}
		return &value.Rune{Value: runes[i]}
	default:
		return makeError(env, node, "cannot index %s", target.Type())
	}
}

func evalNodeAccess(target value.Value, node *ebnf.Node, env *value.Env) value.Value {
	var fieldName string

	for _, child := range node.Children {
		if child.Type == "NAME" {
			fieldName = child.Value
			break
		}
	}

	list, ok := target.(*value.List)
	if !ok {
		return makeError(env, node, "cannot access field on %s", target.Type())
	}

	if list.Fields == nil {
		return makeError(env, node, "list has no shape")
	}

	for i, f := range list.Fields {
		if f == fieldName {
			if i < len(list.Elements) {
				return list.Elements[i]
			}
		}
	}

	return makeError(env, node, "unknown field: %s", fieldName)
}

// evalNodeIdentWithNode looks up an identifier, returning rich error if undefined.
func evalNodeIdentWithNode(node *ebnf.Node, env *value.Env) value.Value {
	name := node.Value
	if val, ok := env.Get(name); ok {
		return val
	}
	if builtin, ok := HAL[name]; ok {
		return builtin
	}
	if intrinsic, ok := value.Intrinsics[name]; ok {
		return intrinsic
	}
	return makeError(env, node, "undefined: %s", name)
}

// evalNodeIdent looks up an identifier (legacy, for internal use without node).
func evalNodeIdent(name string, env *value.Env) value.Value {
	if val, ok := env.Get(name); ok {
		return val
	}
	if builtin, ok := HAL[name]; ok {
		return builtin
	}
	if intrinsic, ok := value.Intrinsics[name]; ok {
		return intrinsic
	}
	return value.NewError("undefined: %s", name)
}

func evalNodeNumber(node *ebnf.Node, env *value.Env) value.Value {
	s := node.Value
	r := new(big.Rat)
	if _, ok := r.SetString(s); !ok {
		return makeError(env, node, "invalid number: %s", s)
	}
	return &value.Number{Value: r}
}

func applyNodeFunc(fn *value.Function, args []value.Value, callLine int) value.Value {
	funcEnv := value.NewEnv(fn.Env)

	// Determine layer from function
	layer := fn.Layer
	if layer == "" {
		layer = value.LayerUser
	}

	// Push stack frame for this call
	funcEnv.PushFrame(fn.Name, callLine, layer)

	for i, param := range fn.Parameters {
		if i < len(args) {
			funcEnv.Set(param, args[i])
		} else {
			funcEnv.Set(param, value.NULL)
		}
	}

	if fn.RestParam != "" {
		start := len(fn.Parameters)
		var rest []value.Value
		if start < len(args) {
			rest = args[start:]
		}
		funcEnv.Set(fn.RestParam, &value.List{Elements: rest})
	}

	var result value.Value
	if fn.BodyNode != nil {
		result = EvalNode(fn.BodyNode.(*ebnf.Node), funcEnv)
	}

	// Pop stack frame
	funcEnv.PopFrame()

	if ret, ok := result.(*value.Return); ok {
		return ret.Value
	}
	return result
}

func applyOp(op string, left, right value.Value, env *value.Env, node *ebnf.Node) value.Value {
	leftNum, leftIsNum := left.(*value.Number)
	rightNum, rightIsNum := right.(*value.Number)

	if leftIsNum && rightIsNum {
		switch op {
		case "+":
			return &value.Number{Value: new(big.Rat).Add(leftNum.Value, rightNum.Value)}
		case "-":
			return &value.Number{Value: new(big.Rat).Sub(leftNum.Value, rightNum.Value)}
		case "*":
			return &value.Number{Value: new(big.Rat).Mul(leftNum.Value, rightNum.Value)}
		case "/":
			if rightNum.Value.Sign() == 0 {
				return makeError(env, node, "division by zero")
			}
			return &value.Number{Value: new(big.Rat).Quo(leftNum.Value, rightNum.Value)}
		case "%":
			if !leftNum.Value.IsInt() || !rightNum.Value.IsInt() {
				return makeError(env, node, "modulo requires integers")
			}
			if rightNum.Value.Sign() == 0 {
				return makeError(env, node, "modulo by zero")
			}
			l := leftNum.Value.Num()
			r := rightNum.Value.Num()
			result := new(big.Int).Mod(l, r)
			return &value.Number{Value: new(big.Rat).SetInt(result)}
		case "<":
			return value.NativeBoolToBoolean(leftNum.Value.Cmp(rightNum.Value) < 0)
		case ">":
			return value.NativeBoolToBoolean(leftNum.Value.Cmp(rightNum.Value) > 0)
		case "<=":
			return value.NativeBoolToBoolean(leftNum.Value.Cmp(rightNum.Value) <= 0)
		case ">=":
			return value.NativeBoolToBoolean(leftNum.Value.Cmp(rightNum.Value) >= 0)
		case "==":
			return value.NativeBoolToBoolean(leftNum.Value.Cmp(rightNum.Value) == 0)
		case "!=":
			return value.NativeBoolToBoolean(leftNum.Value.Cmp(rightNum.Value) != 0)
		}
	}

	// == and != for non-numbers (use equal function logic)
	switch op {
	case "==":
		return value.NativeBoolToBoolean(nodeValuesEqual(left, right))
	case "!=":
		return value.NativeBoolToBoolean(!nodeValuesEqual(left, right))
	}

	if op == "+" {
		leftStr, leftIsStr := left.(*value.String)
		rightStr, rightIsStr := right.(*value.String)
		if leftIsStr && rightIsStr {
			return &value.String{Value: leftStr.Value + rightStr.Value}
		}
	}

	switch op {
	case "and":
		return value.NativeBoolToBoolean(isTruthy(left) && isTruthy(right))
	case "or":
		return value.NativeBoolToBoolean(isTruthy(left) || isTruthy(right))
	}

	return makeError(env, node, "cannot apply %s to %s and %s", op, left.Type(), right.Type())
}

func nodeValuesEqual(a, b value.Value) bool {
	if a.Type() != b.Type() {
		return false
	}
	switch av := a.(type) {
	case *value.Number:
		return av.Value.Cmp(b.(*value.Number).Value) == 0
	case *value.String:
		return av.Value == b.(*value.String).Value
	case *value.Boolean:
		return av.Value == b.(*value.Boolean).Value
	case *value.Symbol:
		return av.Value == b.(*value.Symbol).Value
	case *value.Rune:
		return av.Value == b.(*value.Rune).Value
	case *value.Null:
		return true
	case *value.List:
		bv := b.(*value.List)
		if len(av.Elements) != len(bv.Elements) {
			return false
		}
		for i := range av.Elements {
			if !nodeValuesEqual(av.Elements[i], bv.Elements[i]) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func isOp(s string) bool {
	ops := map[string]bool{
		"+": true, "-": true, "*": true, "/": true, "%": true,
		"<": true, ">": true, "<=": true, ">=": true,
		"==": true, "!=": true,
		"and": true, "or": true,
	}
	return ops[s]
}

// RunNode parses and evaluates code using the EBNF grammar.
func RunNode(grammar *ebnf.Grammar, input string, env *value.Env) value.Value {
	ast, err := grammar.ParseSource(input)
	if err != nil {
		return value.NewError("parse error: %s", err)
	}
	// Push <main> frame for top-level execution
	env.PushFrame("<main>", 1, value.LayerUser)
	result := EvalNode(ast, env)
	env.PopFrame()
	return result
}

// RunFileNode parses and evaluates a file using the EBNF grammar.
func RunFileNode(grammar *ebnf.Grammar, filename string, env *value.Env) value.Value {
	data, err := os.ReadFile(filename)
	if err != nil {
		return value.NewError("cannot read file: %s", err)
	}

	env.SetFile(filepath.Base(filename))
	env.SetSource(string(data))

	ast, err := grammar.ParseSource(string(data))
	if err != nil {
		return value.NewError("parse error: %s", err)
	}
	// Push <main> frame for top-level execution
	env.PushFrame("<main>", 1, value.LayerUser)
	result := EvalNode(ast, env)
	env.PopFrame()
	return result
}
