package substrate

import (
	"fmt"
	"os"

	"aiki/engine/runtime/hal"
	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
)

// halApply implements apply(fn, list) - spreads list as args.
func halApply(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewError("apply: want 2 arguments, got %d", len(args))
	}

	fn := args[0]
	listVal := args[1]

	list, ok := listVal.(*value.List)
	if !ok {
		return value.NewError("apply: second argument must be list, got %s", listVal.Type())
	}

	// Dispatch based on function type
	switch f := fn.(type) {
	case *value.Function:
		return applyUserFunction(f, list.Elements, ctx)
	case value.Callable:
		return f.Call(list.Elements)
	default:
		return value.NewError("apply: first argument must be function, got %s", fn.Type())
	}
}

// applyUserFunction calls a user-defined function with args.
func applyUserFunction(fn *value.Function, args []value.Value, ctx *hal.EvalContext) value.Value {
	if ctx == nil || ctx.Eval == nil {
		return value.NewError("apply: evaluation context not available")
	}

	fnEnv, ok := fn.Env.(*value.Env)
	if !ok {
		return value.NewError("apply: invalid function environment")
	}

	callEnv := value.NewEnclosedEnv(fnEnv)

	for i, param := range fn.Params {
		if i < len(args) {
			callEnv.Set(param, args[i])
		} else {
			callEnv.Set(param, value.EMPTY)
		}
	}

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
		return value.NewError("apply: invalid function body")
	}

	result := ctx.Eval(body, callEnv)

	if ret, ok := result.(*value.Return); ok {
		return ret.Val
	}

	return result
}

// halExport implements export(:name1, :name2, ...).
// Records exported names on the environment.
func halExport(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) == 0 {
		return value.NewError("export: want at least 1 argument")
	}

	if ctx == nil || ctx.Env == nil {
		return value.NewError("export: environment not available")
	}

	var names []string
	for _, arg := range args {
		sym, ok := arg.(*value.Symbol)
		if !ok {
			return value.NewError("export: expected symbol, got %s", arg.Type())
		}
		names = append(names, sym.Val)
	}

	ctx.Env.SetExports(names)
	return value.EMPTY
}

// halImport implements import("module", :name1, :name2, ...).
// Parses and evaluates the module, then copies exported names into the current environment.
func halImport(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) < 2 {
		return value.NewError("import: want module and at least 1 name")
	}

	if ctx == nil || ctx.Env == nil || ctx.Grammar == nil || ctx.Eval == nil {
		return value.NewError("import: evaluation context not available")
	}

	// First arg: module name (string)
	moduleStr, ok := args[0].(*value.String)
	if !ok {
		return value.NewError("import: expected string module name, got %s", args[0].Type())
	}
	moduleName := moduleStr.Val

	// Remaining args: names to import (symbols)
	var importNames []string
	for _, arg := range args[1:] {
		sym, ok := arg.(*value.Symbol)
		if !ok {
			return value.NewError("import: expected symbol, got %s", arg.Type())
		}
		importNames = append(importNames, sym.Val)
	}

	// Resolve module path
	modulePath := resolveModulePath(moduleName, ctx.Env)
	if modulePath == "" {
		return value.NewError("import: cannot find module '%s'", moduleName)
	}

	// Read module source
	data, err := os.ReadFile(modulePath)
	if err != nil {
		return value.NewError("import: cannot read '%s': %s", modulePath, err)
	}

	// Parse module
	lexer := syntax.NewLexer(ctx.Grammar, modulePath, string(data), nil)
	tokens, err := lexer.Tokenize()
	if err != nil {
		return value.NewError("import: lex error in '%s': %s", modulePath, err)
	}

	parser := syntax.NewParser(ctx.Grammar, tokens, string(data), nil)
	ast, err := parser.Parse()
	if err != nil {
		return value.NewError("import: parse error in '%s': %s", modulePath, err)
	}

	// Create module environment enclosed by caller's env chain.
	// This gives the module access to prelude bindings.
	modEnv := value.NewEnclosedEnv(ctx.Env)
	modEnv.SetFile(modulePath)
	modEnv.SetSource(string(data))

	// Evaluate module
	result := ctx.Eval(ast, modEnv)
	if err, ok := result.(*value.Error); ok {
		return err
	}

	// Copy requested names into calling environment
	exports := modEnv.GetExports()
	for _, name := range importNames {
		// If module declared exports, only allow those
		if len(exports) > 0 && !containsStr(exports, name) {
			return value.NewError("import: '%s' is not exported by '%s'", name, moduleName)
		}
		val, ok := modEnv.Get(name)
		if !ok {
			return value.NewError("import: '%s' not defined in '%s'", name, moduleName)
		}
		ctx.Env.Set(name, val)
	}

	return value.EMPTY
}

// halLoad implements load(path) - reads and evaluates a file.
func halLoad(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewError("load: want 1 argument, got %d", len(args))
	}

	if ctx == nil || ctx.Grammar == nil || ctx.Eval == nil || ctx.Env == nil {
		return value.NewError("load: evaluation context not available")
	}

	pathStr, ok := args[0].(*value.String)
	if !ok {
		return value.NewError("load: expected string path, got %s", args[0].Type())
	}

	modulePath := resolveModulePath(pathStr.Val, ctx.Env)
	if modulePath == "" {
		return value.NewError("load: cannot find '%s'", pathStr.Val)
	}

	data, err := os.ReadFile(modulePath)
	if err != nil {
		return value.NewError("load: cannot read '%s': %s", modulePath, err)
	}

	lexer := syntax.NewLexer(ctx.Grammar, modulePath, string(data), nil)
	tokens, err := lexer.Tokenize()
	if err != nil {
		return value.NewError("load: lex error in '%s': %s", modulePath, err)
	}

	parser := syntax.NewParser(ctx.Grammar, tokens, string(data), nil)
	ast, err := parser.Parse()
	if err != nil {
		return value.NewError("load: parse error in '%s': %s", modulePath, err)
	}

	// Evaluate in current environment
	return ctx.Eval(ast, ctx.Env)
}

// halSpawn implements spawn(fn, args...) - runs function concurrently.
func halSpawn(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) < 1 {
		return value.NewError("spawn: want at least 1 argument (function)")
	}

	if ctx == nil || ctx.Eval == nil {
		return value.NewError("spawn: evaluation context not available")
	}

	fn, ok := args[0].(*value.Function)
	if !ok {
		return value.NewError("spawn: expected function as first argument, got %s", args[0].Type())
	}

	// Capture arguments passed to spawn: spawn(fn, arg1, arg2)
	fnArgs := args[1:]

	// Launch goroutine
	go func() {
		result := applyUserFunction(fn, fnArgs, ctx)
		// Log errors from spawned functions (they can't propagate)
		if err, ok := result.(*value.Error); ok {
			fmt.Fprintf(os.Stderr, "spawn: %s\n", err.Inspect())
		}
	}()

	return value.TRUE
}

// containsStr checks if a string slice contains a value.
func containsStr(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
