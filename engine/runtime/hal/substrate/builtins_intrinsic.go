package substrate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	if len(args) < 1 {
		return value.NewError("import: want at least module name")
	}

	if ctx == nil || ctx.Env == nil || ctx.Grammar == nil || ctx.Eval == nil {
		return value.NewError("import: evaluation context not available")
	}

	// First arg: module name or path (string)
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

	// Load the module
	mod, errVal := loadModule(moduleName, ctx)
	if errVal != nil {
		return errVal
	}

	// If no names specified, return the Module value
	if len(importNames) == 0 {
		return mod
	}

	// Bind requested names into calling environment
	for _, name := range importNames {
		val, ok := mod.Get(name)
		if !ok {
			return value.NewError("import: '%s' is not exported by '%s'", name, mod.Name)
		}
		ctx.Env.Set(name, val)
	}

	return value.EMPTY
}

// loadModule loads a module by name or path, using cache if available.
func loadModule(name string, ctx *hal.EvalContext) (*value.Module, *value.Error) {
	var modulePath string
	var pkgName string

	if IsPathImport(name) {
		// Path-based import: resolve relative to current file
		modulePath = resolveRelativePath(name, ctx.Env)
		if modulePath == "" {
			return nil, value.NewError("import: cannot find '%s'", name)
		}
		// Package name will be determined from file
		pkgName = ""
	} else {
		// Registry-based import
		if GlobalRegistry == nil {
			return nil, value.NewError("import: module registry not initialized")
		}

		// Check cache first
		if mod, ok := GlobalRegistry.GetCached(name); ok {
			return mod, nil
		}

		// Lookup in registry
		var ok bool
		modulePath, ok = GlobalRegistry.Lookup(name)
		if !ok {
			return nil, value.NewError("import: unknown package '%s'", name)
		}
		pkgName = name
	}

	// Read module source
	data, err := os.ReadFile(modulePath)
	if err != nil {
		return nil, value.NewError("import: cannot read '%s': %s", modulePath, err)
	}

	// Parse module
	lexer := syntax.NewLexer(ctx.Grammar, modulePath, string(data), nil)
	tokens, err := lexer.Tokenize()
	if err != nil {
		return nil, value.NewError("import: lex error in '%s': %s", modulePath, err)
	}

	parser := syntax.NewParser(ctx.Grammar, tokens, string(data), nil)
	ast, err := parser.Parse()
	if err != nil {
		return nil, value.NewError("import: parse error in '%s': %s", modulePath, err)
	}

	// Create module environment enclosed by prelude env (not caller's env)
	preludeEnv := ctx.Env.GetPreludeEnv()
	if preludeEnv == nil {
		return nil, value.NewError("import: prelude environment not available")
	}
	modEnv := value.NewEnclosedEnv(preludeEnv)
	modEnv.SetFile(modulePath)
	modEnv.SetSource(string(data))

	// Evaluate module
	result := ctx.Eval(ast, modEnv)
	if errVal, ok := result.(*value.Error); ok {
		return nil, errVal
	}

	// Get package name from evaluated module
	declaredPkg := modEnv.GetPackageName()
	if pkgName == "" {
		pkgName = declaredPkg
	} else if declaredPkg != "" && declaredPkg != pkgName {
		return nil, value.NewError("import: package declares '%s' but expected '%s'", declaredPkg, pkgName)
	}

	if pkgName == "" {
		// Use filename as fallback
		pkgName = filepath.Base(modulePath)
		pkgName = strings.TrimSuffix(pkgName, ".ai")
	}

	// Build exports map
	exports := make(map[string]value.Value)
	exportNames := modEnv.GetExports()

	if len(exportNames) == 0 {
		return nil, value.NewError("import: module '%s' has no exports", pkgName)
	}

	for _, expName := range exportNames {
		val, ok := modEnv.Get(expName)
		if !ok {
			return nil, value.NewError("import: exported name '%s' not defined in '%s'", expName, pkgName)
		}
		exports[expName] = val
	}

	mod := value.NewModule(pkgName, exports)

	// Cache if registry-based import
	if GlobalRegistry != nil && !IsPathImport(name) {
		GlobalRegistry.Cache(pkgName, mod)
	}

	return mod, nil
}

// resolveRelativePath resolves a path relative to the current file.
func resolveRelativePath(name string, env *value.Env) string {
	currentFile := env.GetFile()

	// Add .ai extension if not present
	if !strings.HasSuffix(name, ".ai") {
		name = name + ".ai"
	}

	if currentFile != "" {
		dir := filepath.Dir(currentFile)
		candidate := filepath.Join(dir, name)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	// Try from current directory
	if _, err := os.Stat(name); err == nil {
		return name
	}

	return ""
}

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
// Spawned functions run with isolated env - only args are visible, no closure capture.
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

	// Launch goroutine with isolated env
	go func() {
		result := applyUserFunctionIsolated(fn, fnArgs, ctx)
		// Log errors from spawned functions (they can't propagate)
		if err, ok := result.(*value.Error); ok {
			fmt.Fprintf(os.Stderr, "spawn: %s\n", err.Inspect())
		}
	}()

	return value.TRUE
}

// applyUserFunctionIsolated runs a function with fresh env - no closure capture.
// Only the passed arguments are visible to the function, plus prelude bindings.
func applyUserFunctionIsolated(fn *value.Function, args []value.Value, ctx *hal.EvalContext) value.Value {
	if ctx == nil || ctx.Eval == nil {
		return value.NewError("spawn: evaluation context not available")
	}

	// Get prelude env from context - spawned fn can see prelude but not user bindings
	preludeEnv := ctx.Env.GetPreludeEnv()
	if preludeEnv == nil {
		return value.NewError("spawn: could not find prelude environment")
	}

	// Fresh env enclosed by prelude - sees prelude bindings but not outer user scope
	callEnv := value.NewEnclosedEnv(preludeEnv)

	// Bind parameters
	for i, param := range fn.Params {
		if i < len(args) {
			callEnv.Set(param, args[i])
		} else {
			callEnv.Set(param, value.EMPTY)
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
		return value.NewError("spawn: invalid function body")
	}

	result := ctx.Eval(body, callEnv)

	if ret, ok := result.(*value.Return); ok {
		return ret.Val
	}
	return result
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

// halStackLimit implements stack_limit(n) via HAL, setting the evaluator stack limit.
func halStackLimit(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewError("stack_limit: want 1 argument")
	}
	n, ok := args[0].(*value.Number)
	if !ok || !n.Val.IsInt() {
		return value.NewError("stack_limit: must be integer >= 1")
	}
	limit64 := n.Val.Num().Int64()
	if limit64 < 1 {
		return value.NewError("stack_limit: must be integer >= 1")
	}
	ctx.Env.SetStackLimit(int(limit64))
	return value.EMPTY
}
