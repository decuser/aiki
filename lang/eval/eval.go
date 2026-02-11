package eval

import (
	"fmt"
	"math/big"
	"os"
	"path/filepath"

	"aiki/lang/ast"
	"aiki/hal/core"
	"aiki/lang/parser"
	"aiki/lang/value"
)

var HAL = core.HAL

// BuiltinNames exports builtin names for shadow checking
var BuiltinNames = make(map[string]bool)

func init() {
	for name := range HAL {
		BuiltinNames[name] = true
	}
}

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

	// Push <main> frame
	env.PushFrame("<main>", 1)
	defer env.PopFrame()

	p := parser.New(string(data))
	program := p.Parse()

	if len(p.Errors()) > 0 {
		return value.NewError("parse error: %s", p.Errors()[0])
	}

	return Eval(program, env)
}

// makeError creates an error with file, line, source context, and stack trace.
func makeError(env *value.Env, line int, format string, a ...interface{}) *value.Error {
	return value.NewErrorAt(
		env.GetFile(),
		line,
		env.GetSourceLine(line),
		env.CopyStack(),
		format,
		a...,
	)
}

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
		return evalNumberLiteral(n)

	case *ast.BooleanLiteral:
		if n.Value {
			return value.True
		}
		return value.False

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
		return &value.Function{
			Parameters: n.Parameters,
			Body:       n.Body,
			Env:        env,
		}

	case *ast.CallExpression:
		return evalCallExpression(n, env)

	case *ast.AccessExpression:
		return evalAccessExpression(n, env)

	case *ast.InfixExpression:
		return evalInfixExpression(n, env)

	case *ast.PrefixExpression:
		return evalPrefixExpression(n, env)

	case *ast.PipeExpression:
		return evalPipeExpression(n, env)
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

	// Block builtin shadowing
	if BuiltinNames[name] {
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
	if !env.Update(stmt.Name.Value, val) {
		return makeError(env, stmt.Token.Line, "undefined variable: %s", stmt.Name.Value)
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
	var result value.Value = value.NULL

	for {
		condition := Eval(stmt.Condition, env)
		if isError(condition) {
			return condition
		}

		if !isTruthy(condition) {
			break
		}

		result = Eval(stmt.Body, env)
		if result != nil {
			rt := result.Type()
			if rt == value.ReturnType || rt == value.ErrorType {
				return result
			}
		}
	}

	return result
}

func evalMatchStatement(stmt *ast.MatchStatement, env *value.Env) value.Value {
	val := Eval(stmt.Value, env)
	if isError(val) {
		return val
	}

	for _, arm := range stmt.Arms {
		bindings := make(map[string]value.Value)
		if matchPattern(arm.Pattern, val, bindings) {
			matchEnv := value.NewEnv(env)
			for name, v := range bindings {
				matchEnv.Set(name, v)
			}
			return Eval(arm.Body, matchEnv)
		}
	}

	return value.NULL
}

func matchPattern(pattern ast.Pattern, val value.Value, bindings map[string]value.Value) bool {
	switch p := pattern.(type) {
	case *ast.WildcardPattern:
		return true

	case *ast.NamePattern:
		bindings[p.Name] = val
		return true

	case *ast.LiteralPattern:
		return false

	case *ast.ListPattern:
		list, ok := val.(*value.List)
		if !ok || len(list.Elements) != len(p.Elements) {
			return false
		}
		for i, elem := range p.Elements {
			if !matchPattern(elem, list.Elements[i], bindings) {
				return false
			}
		}
		return true

	case *ast.ShapedListPattern:
		list, ok := val.(*value.List)
		if !ok || list.Shape != p.Shape || len(list.Elements) != len(p.Elements) {
			return false
		}
		for i, elem := range p.Elements {
			if !matchPattern(elem, list.Elements[i], bindings) {
				return false
			}
		}
		return true
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

func evalNumberLiteral(node *ast.NumberLiteral) value.Value {
	n, err := value.NewNumberFromString(node.Value)
	if err != nil {
		return value.NewError("invalid number: %s", node.Value)
	}
	return n
}

func evalListLiteral(node *ast.ListLiteral, env *value.Env) value.Value {
	elements := evalExpressions(node.Elements, env)
	if len(elements) == 1 && isError(elements[0]) {
		return elements[0]
	}
	return &value.List{Elements: elements}
}

func evalShapedListLiteral(node *ast.ShapedListLiteral, env *value.Env) value.Value {
	fields, ok := env.ResolveFields(node.Shape)
	if !ok {
		return makeError(env, node.Token.Line, "undefined shape: @%s", node.Shape)
	}

	elements := evalExpressions(node.Elements, env)
	if len(elements) == 1 && isError(elements[0]) {
		return elements[0]
	}

	if len(elements) != len(fields) {
		return makeError(env, node.Token.Line, "shape @%s requires %d fields, got %d", node.Shape, len(fields), len(elements))
	}

	return &value.List{
		Elements: elements,
		Shape:    node.Shape,
		Fields:   fields,
	}
}

func evalCallExpression(node *ast.CallExpression, env *value.Env) value.Value {
	// Handle restore() specially - it needs env access
	if ident, ok := node.Function.(*ast.Identifier); ok && ident.Value == "restore" {
		return evalRestore(node, env)
	}

	// Handle load() specially - it needs env access
	if ident, ok := node.Function.(*ast.Identifier); ok && ident.Value == "load" {
		return evalLoad(node, env)
	}

	// Handle spawn() specially - it needs to run function in goroutine
	if ident, ok := node.Function.(*ast.Identifier); ok && ident.Value == "spawn" {
		return evalSpawn(node, env)
	}

	fn := Eval(node.Function, env)
	if isError(fn) {
		return fn
	}

	args := evalExpressions(node.Arguments, env)
	if len(args) == 1 && isError(args[0]) {
		return args[0]
	}

	return applyFunction(fn, args, env, node.Token.Line)
}

func evalRestore(node *ast.CallExpression, env *value.Env) value.Value {
	if len(node.Arguments) != 1 {
		return makeError(env, node.Token.Line, "restore: want 1 argument, got %d", len(node.Arguments))
	}

	arg := Eval(node.Arguments[0], env)
	if isError(arg) {
		return arg
	}

	str, ok := arg.(*value.String)
	if !ok {
		return makeError(env, node.Token.Line, "restore: expected string argument")
	}

	name := str.Value

	if env.Restore(name) {
		return value.NULL
	}

	return makeError(env, node.Token.Line, "restore: %s not found in strict", name)
}

func evalLoad(node *ast.CallExpression, env *value.Env) value.Value {
	if len(node.Arguments) != 1 {
		return makeError(env, node.Token.Line, "load: want 1 argument, got %d", len(node.Arguments))
	}

	arg := Eval(node.Arguments[0], env)
	if isError(arg) {
		return arg
	}

	path, ok := arg.(*value.String)
	if !ok {
		return makeError(env, node.Token.Line, "load: expected string argument")
	}

	// Save current file context
	oldFile := env.GetFile()

	// Run the loaded file
	result := RunFile(path.Value, env)

	// Restore file context
	env.SetFile(oldFile)

	return result
}

func evalSpawn(node *ast.CallExpression, env *value.Env) value.Value {
	if len(node.Arguments) != 1 {
		return makeError(env, node.Token.Line, "spawn: want 1 argument, got %d", len(node.Arguments))
	}

	arg := Eval(node.Arguments[0], env)
	if isError(arg) {
		return arg
	}

	fn, ok := arg.(*value.Function)
	if !ok {
		return makeError(env, node.Token.Line, "spawn: argument must be a function")
	}

	if len(fn.Parameters) != 0 {
		return makeError(env, node.Token.Line, "spawn: function must take no arguments")
	}

	core.DefaultScheduler.Spawn(func() {
		Eval(fn.Body, value.NewEnv(fn.Env))
	})

	return value.True
}

func applyFunction(fn value.Value, args []value.Value, env *value.Env, callLine int) value.Value {
	switch f := fn.(type) {
	case *value.Function:
		if len(args) != len(f.Parameters) {
			return makeError(env, callLine, "wrong number of arguments: want %d, got %d", len(f.Parameters), len(args))
		}

		name := f.Name
		if name == "" {
			name = "<anonymous>"
		}
		env.PushFrame(name, callLine)
		defer env.PopFrame()

		extEnv := value.NewEnv(f.Env)
		for i, param := range f.Parameters {
			extEnv.Set(param, args[i])
		}
		result := Eval(f.Body, extEnv)
		if ret, ok := result.(*value.Return); ok {
			return ret.Value
		}
		return result

	case *value.Builtin:
		return f.Fn(args...)

	default:
		return makeError(env, callLine, "not a function: %s", fn.Type())
	}
}

func evalAccessExpression(node *ast.AccessExpression, env *value.Env) value.Value {
	left := Eval(node.Left, env)
	if isError(left) {
		return left
	}

	switch l := left.(type) {
	case *value.List:
		// Try numeric index first
		if idx, ok := parseIndex(node.Key); ok {
			if idx < 0 || idx >= len(l.Elements) {
				return makeError(env, node.Token.Line, "index out of bounds: %d", idx)
			}
			return l.Elements[idx]
		}

		// Named field access for shaped lists
		if l.Shape != "" {
			for i, field := range l.Fields {
				if field == node.Key {
					return l.Elements[i]
				}
			}
			return makeError(env, node.Token.Line, "unknown field: %s", node.Key)
		}

		return makeError(env, node.Token.Line, "cannot access field on raw list")

	case *value.String:
		if idx, ok := parseIndex(node.Key); ok {
			runes := []rune(l.Value)
			if idx < 0 || idx >= len(runes) {
				return makeError(env, node.Token.Line, "index out of bounds: %d", idx)
			}
			return &value.Rune{Value: runes[idx]}
		}
		return makeError(env, node.Token.Line, "string access requires numeric index")

	default:
		return makeError(env, node.Token.Line, "cannot access on %s", left.Type())
	}
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
	case node.Operator == "==":
		return value.NativeBoolToBoolean(left == right)
	case node.Operator == "!=":
		return value.NativeBoolToBoolean(left != right)
	default:
		return makeError(env, node.Token.Line, "unknown operator: %s %s %s", left.Type(), node.Operator, right.Type())
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

func evalNumberInfix(op string, left, right *value.Number, line int, env *value.Env) value.Value {
	switch op {
	case "+":
		r := new(big.Rat).Add(left.Value, right.Value)
		return &value.Number{Value: r}
	case "-":
		r := new(big.Rat).Sub(left.Value, right.Value)
		return &value.Number{Value: r}
	case "*":
		r := new(big.Rat).Mul(left.Value, right.Value)
		return &value.Number{Value: r}
	case "/":
		if right.Value.Sign() == 0 {
			return &value.List{
				Elements: []value.Value{
					&value.Symbol{Value: "error"},
					&value.String{Value: "division by zero"},
				},
				Shape:  "error",
				Fields: []string{"tag", "reason"},
			}
		}
		r := new(big.Rat).Quo(left.Value, right.Value)
		return &value.Number{Value: r}
	case "%":
		if !left.Value.IsInt() || !right.Value.IsInt() {
			return makeError(env, line, "modulo requires integers")
		}
		if right.Value.Sign() == 0 {
			return makeError(env, line, "division by zero")
		}
		l := left.Value.Num()
		r := right.Value.Num()
		m := new(big.Int).Mod(l, r)
		return &value.Number{Value: new(big.Rat).SetInt(m)}
	case "<":
		return value.NativeBoolToBoolean(left.Value.Cmp(right.Value) < 0)
	case ">":
		return value.NativeBoolToBoolean(left.Value.Cmp(right.Value) > 0)
	case "<=":
		return value.NativeBoolToBoolean(left.Value.Cmp(right.Value) <= 0)
	case ">=":
		return value.NativeBoolToBoolean(left.Value.Cmp(right.Value) >= 0)
	case "==":
		return value.NativeBoolToBoolean(left.Value.Cmp(right.Value) == 0)
	case "!=":
		return value.NativeBoolToBoolean(left.Value.Cmp(right.Value) != 0)
	default:
		return makeError(env, line, "unknown operator: %s", op)
	}
}

func evalStringInfix(op string, left, right *value.String, line int, env *value.Env) value.Value {
	switch op {
	case "+":
		return &value.String{Value: left.Value + right.Value}
	case "==":
		return value.NativeBoolToBoolean(left.Value == right.Value)
	case "!=":
		return value.NativeBoolToBoolean(left.Value != right.Value)
	default:
		return makeError(env, line, "unknown operator: string %s string", op)
	}
}

func evalBooleanInfix(op string, left, right *value.Boolean, line int, env *value.Env) value.Value {
	switch op {
	case "and":
		return value.NativeBoolToBoolean(left.Value && right.Value)
	case "or":
		return value.NativeBoolToBoolean(left.Value || right.Value)
	case "==":
		return value.NativeBoolToBoolean(left.Value == right.Value)
	case "!=":
		return value.NativeBoolToBoolean(left.Value != right.Value)
	default:
		return makeError(env, line, "unknown operator: boolean %s boolean", op)
	}
}

func evalPrefixExpression(node *ast.PrefixExpression, env *value.Env) value.Value {
	right := Eval(node.Right, env)
	if isError(right) {
		return right
	}

	switch node.Operator {
	case "not":
		return value.NativeBoolToBoolean(!isTruthy(right))
	case "-":
		if n, ok := right.(*value.Number); ok {
			r := new(big.Rat).Neg(n.Value)
			return &value.Number{Value: r}
		}
		return makeError(env, node.Token.Line, "cannot negate %s", right.Type())
	default:
		return makeError(env, node.Token.Line, "unknown operator: %s", node.Operator)
	}
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

func evalExpressions(exprs []ast.Expression, env *value.Env) []value.Value {
	var result []value.Value
	for _, e := range exprs {
		val := Eval(e, env)
		if isError(val) {
			return []value.Value{val}
		}
		result = append(result, val)
	}
	return result
}

func isError(v value.Value) bool {
	return v != nil && v.Type() == value.ErrorType
}

func isTruthy(v value.Value) bool {
	switch v := v.(type) {
	case *value.Boolean:
		return v.Value
	case *value.Null:
		return false
	default:
		return true
	}
}

func parseIndex(s string) (int, bool) {
	idx := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, false
		}
		idx = idx*10 + int(r-'0')
	}
	return idx, true
}
