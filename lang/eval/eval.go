package eval

import (
	"fmt"
	"math/big"
	"os"
	"path/filepath"

	"aiki/hal/core"
	"aiki/lang/ast"
	"aiki/lang/parser"
	"aiki/lang/value"
)

// HAL is the single source of truth for all builtins.
// Specials that need AST access use Fn: nil and are
// handled in evalCallExpression.
var HAL = core.HAL

// Run parses and evaluates code in the given environment.
func Run(input string, env *value.Env) value.Value {
	p := parser.New(input)
	program := p.Parse()

	if len(p.Errors()) > 0 {
		return value.NewError("parse error: %s", p.Errors()[0])
	}

	return Eval(program, env)
}

// RunFile parses and evaluates a file, setting up file/source tracking.
func RunFile(filename string, env *value.Env) value.Value {
	data, err := os.ReadFile(filename)
	if err != nil {
		return value.NewError("cannot read file: %s", err)
	}

	// Set file context for error reporting
	env.SetFile(filepath.Base(filename))
	env.SetSource(string(data))

	p := parser.New(string(data))
	program := p.Parse()

	if len(p.Errors()) > 0 {
		return value.NewError("parse error: %s", p.Errors()[0])
	}

	return Eval(program, env)
}

// Eval evaluates an AST node in the given environment.
func Eval(node ast.Node, env *value.Env) value.Value {
	switch n := node.(type) {
	case *ast.Program:
		return evalProgram(n, env)

	case *ast.LetStatement:
		return evalLetStatement(n, env)

	case *ast.ShapeStatement:
		return evalShapeStatement(n, env)

	case *ast.AssignStatement:
		return evalAssignStatement(n, env)

	case *ast.ReturnStatement:
		val := Eval(n.Value, env)
		if isError(val) {
			return val
		}
		return &value.Return{Value: val}

	case *ast.ExpressionStatement:
		return Eval(n.Expression, env)

	case *ast.BlockStatement:
		return evalBlockStatement(n, env)

	case *ast.IfStatement:
		return evalIfStatement(n, env)

	case *ast.WhileStatement:
		return evalWhileStatement(n, env)

	case *ast.MatchStatement:
		return evalMatchStatement(n, env)

	case *ast.Identifier:
		return evalIdentifier(n, env)

	case *ast.NumberLiteral:
		return evalNumberLiteral(n, env)

	case *ast.BooleanLiteral:
		return value.NativeBoolToBoolean(n.Value)

	case *ast.StringLiteral:
		return &value.String{Value: n.Value}

	case *ast.RuneLiteral:
		return &value.Rune{Value: n.Value}

	case *ast.SymbolLiteral:
		return &value.Symbol{Value: n.Value}

	case *ast.ListLiteral:
		return evalListLiteral(n, env)

	case *ast.ShapedListLiteral:
		return evalShapedListLiteral(n, env)

	case *ast.FunctionLiteral:
		return evalFunctionLiteral(n, env)

	case *ast.CallExpression:
		return evalCallExpression(n, env)

	case *ast.IndexExpression:
		return evalIndexExpression(n, env)

	case *ast.AccessExpression:
		return evalAccessExpression(n, env)

	case *ast.InfixExpression:
		return evalInfixExpression(n, env)

	case *ast.GroupExpression:
		return Eval(n.Inner, env)

	case *ast.PrefixExpression:
		return evalPrefixExpression(n, env)

	case *ast.PipeExpression:
		return evalPipeExpression(n, env)

	case *ast.ImportStatement:
		return evalImportStatement(n, env)

	case *ast.ExportStatement:
		return value.NULL
	}

	return value.NULL
}

func evalProgram(program *ast.Program, env *value.Env) value.Value {
	var result value.Value = value.NULL

	for _, stmt := range program.Statements {
		result = Eval(stmt, env)

		switch r := result.(type) {
		case *value.Return:
			return r.Value
		case *value.Error:
			return r
		}
	}

	return result
}

func evalBlockStatement(block *ast.BlockStatement, env *value.Env) value.Value {
	var result value.Value = value.NULL
	blockEnv := value.NewEnv(env)

	for _, stmt := range block.Statements {
		result = Eval(stmt, blockEnv)

		if result != nil {
			rt := result.Type()
			if rt == value.ReturnType || rt == value.ErrorType {
				return result
			}
		}
	}

	return result
}

func evalLetStatement(stmt *ast.LetStatement, env *value.Env) value.Value {
	val := Eval(stmt.Value, env)
	if isError(val) {
		return val
	}

	name := stmt.Name.Value

	// Block builtin shadowing - HAL is the single source of truth
	if HAL[name] != nil {
		return makeError(env, stmt.Token.Line, "cannot shadow builtin: %s", name)
	}

	// Check for strict shadowing (warning only)
	snapshot := env.GetSnapshot()
	if snapshot != nil {
		if _, ok := snapshot[name]; ok {
			fmt.Printf("warning: %s shadows strict (use restore(\"%s\") to undo)\n", name, name)
		}
	}

	// If it's a function, give it the binding name
	if fn, ok := val.(*value.Function); ok {
		fn.Name = name
	}

	env.Set(name, val)
	return value.NULL
}

func evalShapeStatement(stmt *ast.ShapeStatement, env *value.Env) value.Value {
	def := &value.ShapeDef{
		Name:   stmt.Name,
		Fields: stmt.Fields,
		Embeds: stmt.Embeds,
	}
	env.DefineShape(def)
	return value.NULL
}

func evalAssignStatement(stmt *ast.AssignStatement, env *value.Env) value.Value {
	val := Eval(stmt.Value, env)
	if isError(val) {
		return val
	}

	name := stmt.Name.Value
	if !env.Update(name, val) {
		return makeError(env, stmt.Token.Line, "undefined: %s", name)
	}
	return value.NULL
}

func evalIfStatement(stmt *ast.IfStatement, env *value.Env) value.Value {
	condition := Eval(stmt.Condition, env)
	if isError(condition) {
		return condition
	}

	if isTruthy(condition) {
		return Eval(stmt.Consequence, env)
	} else if stmt.Alternative != nil {
		return Eval(stmt.Alternative, env)
	}
	return value.NULL
}

func evalWhileStatement(stmt *ast.WhileStatement, env *value.Env) value.Value {
	for {
		condition := Eval(stmt.Condition, env)
		if isError(condition) {
			return condition
		}

		if !isTruthy(condition) {
			break
		}

		result := Eval(stmt.Body, env)
		if result != nil {
			rt := result.Type()
			if rt == value.ReturnType || rt == value.ErrorType {
				return result
			}
		}
	}
	return value.NULL
}

func evalMatchStatement(stmt *ast.MatchStatement, env *value.Env) value.Value {
	val := Eval(stmt.Value, env)
	if isError(val) {
		return val
	}

	for _, arm := range stmt.Arms {
		matchEnv := value.NewEnv(env)
		if matchPattern(arm.Pattern, val, matchEnv) {
			return Eval(arm.Body, matchEnv)
		}
	}

	return value.NULL
}

func matchPattern(pattern ast.Pattern, val value.Value, env *value.Env) bool {
	switch p := pattern.(type) {
	case *ast.WildcardPattern:
		return true

	case *ast.NamePattern:
		env.Set(p.Name, val)
		return true

	case *ast.LiteralPattern:
		patVal := Eval(p.Value, env)
		return valuesEqual(patVal, val)

	case *ast.ListPattern:
		list, ok := val.(*value.List)
		if !ok || len(list.Elements) != len(p.Elements) {
			return false
		}
		for i, elemPat := range p.Elements {
			if !matchPattern(elemPat, list.Elements[i], env) {
				return false
			}
		}
		return true

	case *ast.ShapedListPattern:
		list, ok := val.(*value.List)
		if !ok || list.Shape != p.Shape {
			return false
		}
		if len(list.Elements) != len(p.Elements) {
			return false
		}
		for i, elemPat := range p.Elements {
			if !matchPattern(elemPat, list.Elements[i], env) {
				return false
			}
		}
		return true
	}
	return false
}

func valuesEqual(a, b value.Value) bool {
	if a.Type() != b.Type() {
		return false
	}
	switch av := a.(type) {
	case *value.Number:
		bv := b.(*value.Number)
		return av.Value.Cmp(bv.Value) == 0
	case *value.Boolean:
		bv := b.(*value.Boolean)
		return av.Value == bv.Value
	case *value.String:
		bv := b.(*value.String)
		return av.Value == bv.Value
	case *value.Symbol:
		bv := b.(*value.Symbol)
		return av.Value == bv.Value
	case *value.Rune:
		bv := b.(*value.Rune)
		return av.Value == bv.Value
	}
	return false
}

func evalIdentifier(node *ast.Identifier, env *value.Env) value.Value {
	if val, ok := env.Get(node.Value); ok {
		return val
	}

	if builtin, ok := HAL[node.Value]; ok {
		return builtin
	}

	return makeError(env, node.Token.Line, "undefined: %s", node.Value)
}

func evalNumberLiteral(node *ast.NumberLiteral, env *value.Env) value.Value {
	r := new(big.Rat)
	if _, ok := r.SetString(node.Value); !ok {
		return makeError(env, node.Token.Line, "invalid number: %s", node.Value)
	}
	return &value.Number{Value: r}
}

func evalListLiteral(node *ast.ListLiteral, env *value.Env) value.Value {
	elements := make([]value.Value, len(node.Elements))
	for i, el := range node.Elements {
		val := Eval(el, env)
		if isError(val) {
			return val
		}
		elements[i] = val
	}
	return &value.List{Elements: elements}
}

func evalShapedListLiteral(node *ast.ShapedListLiteral, env *value.Env) value.Value {
	fields, ok := env.ResolveFields(node.Shape)
	if !ok {
		return makeError(env, node.Token.Line, "undefined shape: @%s", node.Shape)
	}

	if len(node.Elements) != len(fields) {
		return makeError(env, node.Token.Line, "shape @%s requires %d fields, got %d",
			node.Shape, len(fields), len(node.Elements))
	}

	elements := make([]value.Value, len(node.Elements))
	for i, el := range node.Elements {
		val := Eval(el, env)
		if isError(val) {
			return val
		}
		elements[i] = val
	}

	return &value.List{
		Elements: elements,
		Shape:    node.Shape,
		Fields:   fields,
	}
}

func evalFunctionLiteral(node *ast.FunctionLiteral, env *value.Env) value.Value {
	return &value.Function{
		Parameters: node.Parameters,
		RestParam:  node.RestParam,
		Body:       node.Body,
		Env:        env,
	}
}

func evalCallExpression(node *ast.CallExpression, env *value.Env) value.Value {
	// Handle apply() specially - it needs to spread list as args
	if ident, ok := node.Function.(*ast.Identifier); ok && ident.Value == "apply" {
		return evalApply(node, env)
	}

	// Handle load() specially - it needs file access
	if ident, ok := node.Function.(*ast.Identifier); ok && ident.Value == "load" {
		return evalLoad(node, env)
	}

	fn := Eval(node.Function, env)
	if isError(fn) {
		return fn
	}

	args := make([]value.Value, len(node.Arguments))
	for i, arg := range node.Arguments {
		val := Eval(arg, env)
		if isError(val) {
			return val
		}
		args[i] = val
	}

	return applyFunction(fn, args, env, node.Token.Line)
}

func evalApply(node *ast.CallExpression, env *value.Env) value.Value {
	if len(node.Arguments) != 2 {
		return makeError(env, node.Token.Line, "apply: want 2 arguments, got %d", len(node.Arguments))
	}

	fn := Eval(node.Arguments[0], env)
	if isError(fn) {
		return fn
	}

	listVal := Eval(node.Arguments[1], env)
	if isError(listVal) {
		return listVal
	}

	list, ok := listVal.(*value.List)
	if !ok {
		return makeError(env, node.Token.Line, "apply: second argument must be list")
	}

	return applyFunction(fn, list.Elements, env, node.Token.Line)
}

func applyFunction(fn value.Value, args []value.Value, env *value.Env, callLine int) value.Value {
	switch f := fn.(type) {
	case *value.Function:
		// Check arity
		if f.RestParam != "" {
			if len(args) < len(f.Parameters) {
				return makeError(env, callLine, "%s: want at least %d arguments, got %d",
					f.Name, len(f.Parameters), len(args))
			}
		} else {
			if len(args) != len(f.Parameters) {
				return makeError(env, callLine, "%s: want %d arguments, got %d",
					f.Name, len(f.Parameters), len(args))
			}
		}

		funcEnv := value.NewEnv(f.Env)

		// Bind regular parameters
		for i, param := range f.Parameters {
			funcEnv.Set(param, args[i])
		}

		// Bind rest parameter if present
		if f.RestParam != "" {
			restArgs := args[len(f.Parameters):]
			restList := &value.List{Elements: restArgs}
			funcEnv.Set(f.RestParam, restList)
		}

		funcName := f.Name
		if funcName == "" {
			funcName = "<lambda>"
		}
		env.PushFrame(funcName, callLine)
		result := Eval(f.Body, funcEnv)
		env.PopFrame()

		if ret, ok := result.(*value.Return); ok {
			return ret.Value
		}
		return result

	case *value.Builtin:
		if f.Fn == nil {
			return makeError(env, callLine, "cannot call %s directly", f.Name)
		}
		return f.Fn(args...)

	default:
		return makeError(env, callLine, "not a function: %s", fn.Type())
	}
}

func evalLoad(node *ast.CallExpression, env *value.Env) value.Value {
	if len(node.Arguments) != 1 {
		return makeError(env, node.Token.Line, "load: want 1 argument, got %d", len(node.Arguments))
	}

	pathVal := Eval(node.Arguments[0], env)
	if isError(pathVal) {
		return pathVal
	}

	pathStr, ok := pathVal.(*value.String)
	if !ok {
		return makeError(env, node.Token.Line, "load: expected string path")
	}

	return RunFile(pathStr.Value, env)
}

func evalIndexExpression(node *ast.IndexExpression, env *value.Env) value.Value {
	left := Eval(node.Left, env)
	if isError(left) {
		return left
	}

	index := Eval(node.Index, env)
	if isError(index) {
		return index
	}

	idxNum, ok := index.(*value.Number)
	if !ok || !idxNum.Value.IsInt() {
		return makeError(env, node.Token.Line, "index must be integer")
	}
	idx := int(idxNum.Value.Num().Int64())

	switch obj := left.(type) {
	case *value.List:
		if idx < 0 || idx >= len(obj.Elements) {
			return makeError(env, node.Token.Line, "index out of bounds: %d", idx)
		}
		return obj.Elements[idx]
	case *value.String:
		runes := []rune(obj.Value)
		if idx < 0 || idx >= len(runes) {
			return makeError(env, node.Token.Line, "index out of bounds: %d", idx)
		}
		return &value.Rune{Value: runes[idx]}
	default:
		return makeError(env, node.Token.Line, "cannot index %s", left.Type())
	}
}

func evalAccessExpression(node *ast.AccessExpression, env *value.Env) value.Value {
	left := Eval(node.Left, env)
	if isError(left) {
		return left
	}

	list, ok := left.(*value.List)
	if !ok {
		return makeError(env, node.Token.Line, "cannot access field on %s", left.Type())
	}

	if list.Shape == "" {
		return makeError(env, node.Token.Line, "cannot access field on raw list, use [index]")
	}

	for i, field := range list.Fields {
		if field == node.Key {
			return list.Elements[i]
		}
	}

	return makeError(env, node.Token.Line, "unknown field: %s", node.Key)
}

func evalInfixExpression(node *ast.InfixExpression, env *value.Env) value.Value {
	left := Eval(node.Left, env)
	if isError(left) {
		return left
	}

	right := Eval(node.Right, env)
	if isError(right) {
		return right
	}

	switch {
	case left.Type() == value.NumberType && right.Type() == value.NumberType:
		return evalNumberInfix(node.Operator, left.(*value.Number), right.(*value.Number), node.Token.Line, env)
	case left.Type() == value.StringType && right.Type() == value.StringType:
		return evalStringInfix(node.Operator, left.(*value.String), right.(*value.String), node.Token.Line, env)
	case left.Type() == value.BooleanType && right.Type() == value.BooleanType:
		return evalBooleanInfix(node.Operator, left.(*value.Boolean), right.(*value.Boolean), node.Token.Line, env)
	case left.Type() == value.SymbolType && right.Type() == value.SymbolType:
		return evalSymbolInfix(node.Operator, left.(*value.Symbol), right.(*value.Symbol), node.Token.Line, env)
	case node.Operator == "==" || node.Operator == "!=":
		// Cross-type comparison
		eq := left.Type() == right.Type() && left.Inspect() == right.Inspect()
		if node.Operator == "!=" {
			eq = !eq
		}
		return value.NativeBoolToBoolean(eq)
	default:
		return makeError(env, node.Token.Line, "unknown operator: %s %s %s", left.Type(), node.Operator, right.Type())
	}
}

func evalNumberInfix(op string, left, right *value.Number, line int, env *value.Env) value.Value {
	l, r := left.Value, right.Value
	result := new(big.Rat)

	switch op {
	case "+":
		result.Add(l, r)
	case "-":
		result.Sub(l, r)
	case "*":
		result.Mul(l, r)
	case "/":
		if r.Sign() == 0 {
			return makeError(env, line, "division by zero")
		}
		result.Quo(l, r)
	case "%":
		if !l.IsInt() || !r.IsInt() {
			return makeError(env, line, "modulo requires integers")
		}
		if r.Num().Int64() == 0 {
			return makeError(env, line, "modulo by zero")
		}
		li, ri := l.Num().Int64(), r.Num().Int64()
		result.SetInt64(li % ri)
	case "<":
		return value.NativeBoolToBoolean(l.Cmp(r) < 0)
	case ">":
		return value.NativeBoolToBoolean(l.Cmp(r) > 0)
	case "<=":
		return value.NativeBoolToBoolean(l.Cmp(r) <= 0)
	case ">=":
		return value.NativeBoolToBoolean(l.Cmp(r) >= 0)
	case "==":
		return value.NativeBoolToBoolean(l.Cmp(r) == 0)
	case "!=":
		return value.NativeBoolToBoolean(l.Cmp(r) != 0)
	default:
		return makeError(env, line, "unknown operator: %s", op)
	}

	return &value.Number{Value: result}
}

func evalStringInfix(op string, left, right *value.String, line int, env *value.Env) value.Value {
	switch op {
	case "+":
		return &value.String{Value: left.Value + right.Value}
	case "==":
		return value.NativeBoolToBoolean(left.Value == right.Value)
	case "!=":
		return value.NativeBoolToBoolean(left.Value != right.Value)
	case "<":
		return value.NativeBoolToBoolean(left.Value < right.Value)
	case ">":
		return value.NativeBoolToBoolean(left.Value > right.Value)
	case "<=":
		return value.NativeBoolToBoolean(left.Value <= right.Value)
	case ">=":
		return value.NativeBoolToBoolean(left.Value >= right.Value)
	default:
		return makeError(env, line, "unknown operator: string %s string", op)
	}
}

func evalBooleanInfix(op string, left, right *value.Boolean, line int, env *value.Env) value.Value {
	switch op {
	case "==":
		return value.NativeBoolToBoolean(left.Value == right.Value)
	case "!=":
		return value.NativeBoolToBoolean(left.Value != right.Value)
	case "and":
		return value.NativeBoolToBoolean(left.Value && right.Value)
	case "or":
		return value.NativeBoolToBoolean(left.Value || right.Value)
	default:
		return makeError(env, line, "unknown operator: boolean %s boolean", op)
	}
}

func evalSymbolInfix(op string, left, right *value.Symbol, line int, env *value.Env) value.Value {
	switch op {
	case "==":
		return value.NativeBoolToBoolean(left.Value == right.Value)
	case "!=":
		return value.NativeBoolToBoolean(left.Value != right.Value)
	default:
		return makeError(env, line, "unknown operator: symbol %s symbol", op)
	}
}

func evalPrefixExpression(node *ast.PrefixExpression, env *value.Env) value.Value {
	right := Eval(node.Right, env)
	if isError(right) {
		return right
	}

	switch node.Operator {
	case "not":
		return evalNotOperator(right)
	case "-":
		return evalMinusPrefix(right, node.Token.Line, env)
	default:
		return makeError(env, node.Token.Line, "unknown operator: %s", node.Operator)
	}
}

func evalNotOperator(right value.Value) value.Value {
	switch r := right.(type) {
	case *value.Boolean:
		return value.NativeBoolToBoolean(!r.Value)
	default:
		return value.False
	}
}

func evalMinusPrefix(right value.Value, line int, env *value.Env) value.Value {
	num, ok := right.(*value.Number)
	if !ok {
		return makeError(env, line, "cannot negate %s", right.Type())
	}
	result := new(big.Rat).Neg(num.Value)
	return &value.Number{Value: result}
}

func evalPipeExpression(node *ast.PipeExpression, env *value.Env) value.Value {
	left := Eval(node.Left, env)
	if isError(left) {
		return left
	}

	// Check for [@error ...] short-circuit
	if list, ok := left.(*value.List); ok {
		if list.Shape == "error" {
			return left
		}
		// Auto-unwrap [@ok val]
		if list.Shape == "ok" && len(list.Elements) > 0 {
			left = list.Elements[0]
		}
	}

	// Prepend left to call arguments
	fn := Eval(node.Right.Function, env)
	if isError(fn) {
		return fn
	}

	args := []value.Value{left}
	for _, arg := range node.Right.Arguments {
		val := Eval(arg, env)
		if isError(val) {
			return val
		}
		args = append(args, val)
	}

	return applyFunction(fn, args, env, node.Token.Line)
}

func evalImportStatement(node *ast.ImportStatement, env *value.Env) value.Value {
	modPath := node.Module + ".ai"

	data, err := os.ReadFile(modPath)
	if err != nil {
		return makeError(env, node.Token.Line, "cannot load module: %s", modPath)
	}

	modEnv := value.NewEnv(nil)

	p := parser.New(string(data))
	program := p.Parse()

	if len(p.Errors()) > 0 {
		return makeError(env, node.Token.Line, "parse error in %s: %s", modPath, p.Errors()[0])
	}

	result := Eval(program, modEnv)
	if isError(result) {
		return result
	}

	// Import requested names
	for _, name := range node.Names {
		val, ok := modEnv.Get(name)
		if !ok {
			return makeError(env, node.Token.Line, "module %s does not export: %s", node.Module, name)
		}
		env.Set(name, val)
	}

	return value.NULL
}

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

func isError(val value.Value) bool {
	return val != nil && val.Type() == value.ErrorType
}

func makeError(env *value.Env, line int, format string, args ...interface{}) *value.Error {
	return value.NewErrorAt(
		env.GetFile(),
		line,
		env.GetSourceLine(line),
		env.CopyStack(),
		format,
		args...,
	)
}
