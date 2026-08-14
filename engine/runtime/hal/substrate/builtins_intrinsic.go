package substrate

import (
	"aiki/engine"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"aiki/engine/runtime/hal"
	"aiki/engine/runtime/libpath"
	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
)

// halApply implements apply(fn, list) - spreads list as args.
func halApply(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 2 {
		return value.NewFault("apply: want 2 arguments, got %d", len(args))
	}

	fn := args[0]
	listVal := args[1]

	list, ok := listVal.(*value.List)
	if !ok {
		return value.NewFault("apply: second argument must be list, got %s", listVal.Type())
	}

	// Dispatch based on function type
	switch f := fn.(type) {
	case *value.Function:
		semanticHitDetail(ctx, engine.SemanticCall, funcName(f))
		return applyUserFunction(f, list.Elements, ctx)
	case hal.ContextCallable:
		return f.CallWithContext(list.Elements, ctx)
	case value.Callable:
		return f.Call(list.Elements)
	default:
		return value.NewFault("apply: first argument must be function, got %s", fn.Type())
	}
}

// funcName returns a function's name for diagnostics, or a placeholder when
// the function is anonymous.
func funcName(fn *value.Function) string {
	if fn.Name != "" {
		return fn.Name
	}
	return "function"
}

// applyUserFunction calls a user-defined function with args.
func applyUserFunction(fn *value.Function, args []value.Value, ctx *hal.EvalContext) value.Value {
	if ctx == nil || ctx.Eval == nil {
		return value.NewFault("apply: evaluation context not available")
	}

	fnEnv, ok := fn.Env.(*value.Env)
	if !ok {
		return value.NewFault("apply: invalid function environment")
	}

	// Same minimum-arity rule as ordinary function application.
	if len(args) < len(fn.Params) {
		return value.NewFault("%s: want %d arguments, got %d", funcName(fn), len(fn.Params), len(args))
	}

	callEnv := value.NewCallEnv(fnEnv, ctx.Env)
	callEnv.SetSemanticProbe(ctx.Probe)

	for i, param := range fn.Params {
		callEnv.Set(param, args[i])
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
		return value.NewFault("apply: invalid function body")
	}

	var result value.Value
	run := func() { result = ctx.Eval(body, callEnv) }
	if ctx.WithProfileLabels != nil {
		ctx.WithProfileLabels(engine.ProfileLabels{
			Layer:    "semantic",
			Function: funcName(fn),
			File:     callEnv.GetFile(),
		}, ctx.Labels, run)
	} else {
		run()
	}

	if ret, ok := result.(*value.Return); ok {
		return ret.Val
	}

	return result
}

// halExport implements export(:name1, :name2, ...).
// Records exported names on the environment.
func halExport(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) == 0 {
		return value.NewFault("export: want at least 1 argument")
	}

	if ctx == nil || ctx.Env == nil {
		return value.NewFault("export: environment not available")
	}

	var names []string
	for _, arg := range args {
		sym, ok := arg.(*value.Symbol)
		if !ok {
			return value.NewFault("export: expected symbol, got %s", arg.Type())
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
		return value.NewFault("import: want at least module name")
	}

	if ctx == nil || ctx.Env == nil || ctx.Grammar == nil || ctx.Eval == nil {
		return value.NewFault("import: evaluation context not available")
	}

	// First arg: module name or path (string)
	moduleStr, ok := args[0].(*value.String)
	if !ok {
		return value.NewFault("import: expected string module name, got %s", args[0].Type())
	}
	moduleName := moduleStr.Val

	// Remaining args: names to import (symbols)
	var importNames []string
	for _, arg := range args[1:] {
		sym, ok := arg.(*value.Symbol)
		if !ok {
			return value.NewFault("import: expected symbol, got %s", arg.Type())
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
			return value.NewShapedError("import", "import: '%s' is not exported by '%s'", name, mod.Name)
		}
		ctx.Env.Set(name, val)
	}

	return value.EMPTY
}

// loadModule loads a module by name or path, using cache if available.
func loadModule(name string, ctx *hal.EvalContext) (*value.Module, value.Value) {
	var modulePath string
	var pkgName string
	var requestedName string

	if IsPathImport(name) {
		// Path-based import: resolve relative to current file
		modulePath = resolveRelativePath(name, ctx.Env)
		if modulePath == "" {
			return nil, value.NewShapedError("import", "import: cannot find '%s'", name)
		}
		// Package name will be determined from file
		pkgName = ""
	} else {
		// Registry-based import
		if GlobalRegistry == nil {
			return nil, value.NewFault("import: module registry not initialized")
		}

		requestedName = name

		// Resolve public name to its physical module and declared package name.
		// Bare defaults such as "math" may resolve to "math/native".
		var ok bool
		modulePath, pkgName, ok = GlobalRegistry.Resolve(name)
		if !ok {
			return nil, value.NewShapedError("import", "import: unknown package '%s'", name)
		}

		// Check cache after resolution so aliases share the canonical module.
		if mod, ok := GlobalRegistry.GetCached(name); ok {
			if requestedName != pkgName {
				return value.NewModule(requestedName, mod.Exports), nil
			}
			return mod, nil
		}
	}

	// Read module source
	data, err := os.ReadFile(modulePath)
	if err != nil {
		return nil, value.NewShapedError("import", "import: cannot read '%s': %s", modulePath, err)
	}

	// Parse module
	lexer := syntax.NewLexer(ctx.Grammar, modulePath, string(data), nil)
	tokens, err := lexer.Tokenize()
	if err != nil {
		return nil, value.NewShapedError("import", "import: lex error in '%s': %s", modulePath, err)
	}

	parser := syntax.NewParser(ctx.Grammar, tokens, string(data), nil)
	ast, err := parser.Parse()
	if err != nil {
		return nil, value.NewShapedError("import", "import: parse error in '%s': %s", modulePath, err)
	}

	// Create module environment enclosed by prelude env (not caller's env)
	preludeEnv := ctx.Env.GetPreludeEnv()
	if preludeEnv == nil {
		return nil, value.NewFault("import: prelude environment not available")
	}

	// Determine scope based on module path
	// Modules in /lib/ or /contrib/lib/ get ScopePrelude (HAL access)
	// All other modules get ScopeUser (no direct HAL access)
	modScope := value.ScopeUser
	if libpath.IsBlessedLibPath(modulePath) {
		modScope = value.ScopePrelude
	}
	modEnv := value.NewEnclosedEnvWithScope(preludeEnv, modScope)
	modEnv.SetFile(modulePath)
	modEnv.SetSource(string(data))

	// Evaluate module
	result := ctx.Eval(ast, modEnv)
	if value.IsFault(result) {
		return nil, result
	}

	// Get package name from evaluated module
	declaredPkg := modEnv.GetPackageName()
	if pkgName == "" {
		pkgName = declaredPkg
	} else if declaredPkg != "" && declaredPkg != pkgName {
		return nil, value.NewFault("import: package declares '%s' but expected '%s'", declaredPkg, pkgName)
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
		return nil, value.NewFault("import: module '%s' has no exports", pkgName)
	}

	for _, expName := range exportNames {
		val, ok := modEnv.Get(expName)
		if !ok {
			return nil, value.NewFault("import: exported name '%s' not defined in '%s'", expName, pkgName)
		}
		exports[expName] = val
	}

	mod := value.NewModule(pkgName, exports)

	// Cache the canonical module for registry-based imports. Aliases resolve to
	// this cache entry but receive a module value bearing the public name used
	// by the caller.
	if GlobalRegistry != nil && !IsPathImport(name) {
		GlobalRegistry.Cache(pkgName, mod)
		if requestedName != "" && requestedName != pkgName {
			return value.NewModule(requestedName, exports), nil
		}
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

// halUse implements use("module").
// Loads module and binds all exports into current scope.
func halUse(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("use: want 1 argument, got %d", len(args))
	}

	if ctx == nil || ctx.Env == nil || ctx.Grammar == nil || ctx.Eval == nil {
		return value.NewFault("use: evaluation context not available")
	}

	moduleStr, ok := args[0].(*value.String)
	if !ok {
		return value.NewFault("use: expected string module name, got %s", args[0].Type())
	}

	mod, errVal := loadModule(moduleStr.Val, ctx)
	if errVal != nil {
		return errVal
	}

	// Bind all exports into calling environment
	for name, val := range mod.Exports {
		ctx.Env.Set(name, val)
	}

	return value.EMPTY
}

func halLoad(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("load: want 1 argument, got %d", len(args))
	}

	if ctx == nil || ctx.Grammar == nil || ctx.Eval == nil || ctx.Env == nil {
		return value.NewFault("load: evaluation context not available")
	}

	pathStr, ok := args[0].(*value.String)
	if !ok {
		return value.NewFault("load: expected string path, got %s", args[0].Type())
	}

	modulePath := resolveModulePath(pathStr.Val, ctx.Env)
	if modulePath == "" {
		return value.NewShapedError("load", "load: cannot find '%s'", pathStr.Val)
	}

	data, err := os.ReadFile(modulePath)
	if err != nil {
		return value.NewShapedError("load", "load: cannot read '%s': %s", modulePath, err)
	}

	lexer := syntax.NewLexer(ctx.Grammar, modulePath, string(data), nil)
	tokens, err := lexer.Tokenize()
	if err != nil {
		return value.NewShapedError("load", "load: lex error in '%s': %s", modulePath, err)
	}

	parser := syntax.NewParser(ctx.Grammar, tokens, string(data), nil)
	ast, err := parser.Parse()
	if err != nil {
		return value.NewShapedError("load", "load: parse error in '%s': %s", modulePath, err)
	}

	// Evaluate in current environment
	return ctx.Eval(ast, ctx.Env)
}

// halSpawn implements spawn(fn, args...) - runs function concurrently.
// Spawned functions run with isolated env - only args are visible, no closure capture.
func halSpawn(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) < 1 {
		return value.NewFault("spawn: want at least 1 argument (function)")
	}

	if ctx == nil || ctx.Eval == nil {
		return value.NewFault("spawn: evaluation context not available")
	}

	fn, ok := args[0].(*value.Function)
	if !ok {
		return value.NewFault("spawn: expected function as first argument, got %s", args[0].Type())
	}

	// Capture arguments passed to spawn: spawn(fn, arg1, arg2)
	fnArgs := args[1:]
	semanticHitDetail(ctx, engine.SemanticCall, funcName(fn))

	// Launch goroutine with isolated env
	go func() {
		result := applyUserFunctionIsolated(fn, fnArgs, ctx)
		if fault, ok := result.(*value.Fault); ok {
			if ctx.ReportAsyncFault != nil {
				ctx.ReportAsyncFault(fault)
				return
			}
			// Runtimes without asynchronous fault propagation retain the
			// historical fallback rather than silently swallowing the fault.
			fmt.Fprintf(os.Stderr, "spawn: %s\n", fault.Inspect())
		}
	}()

	return value.TRUE
}

// applyUserFunctionIsolated runs a function with fresh env - no closure capture.
// Only the passed arguments are visible to the function, plus prelude bindings.
func applyUserFunctionIsolated(fn *value.Function, args []value.Value, ctx *hal.EvalContext) value.Value {
	if ctx == nil || ctx.Eval == nil {
		return value.NewFault("spawn: evaluation context not available")
	}

	// Get prelude env from context - spawned fn can see prelude but not user bindings
	preludeEnv := ctx.Env.GetPreludeEnv()
	if preludeEnv == nil {
		return value.NewFault("spawn: could not find prelude environment")
	}

	// Same minimum-arity rule as ordinary function application.
	if len(args) < len(fn.Params) {
		return value.NewFault("%s: want %d arguments, got %d", funcName(fn), len(fn.Params), len(args))
	}

	// Fresh env enclosed by prelude - sees prelude bindings but not outer user scope
	callEnv := value.NewIsolatedEnclosedEnv(preludeEnv)
	callEnv.SetSemanticProbe(ctx.Probe)

	// Copy the shape vocabulary visible where the function was created.
	// A shape definition is an ordered list of field names, not a value: it
	// carries no data out of the spawning environment, and the shape name
	// already crosses the boundary on the value itself. Without this, a shaped
	// argument arrives intact but field access on it faults "unknown shape".
	if fnEnv, ok := fn.Env.(*value.Env); ok && fnEnv != nil {
		for _, def := range fnEnv.CollectShapes() {
			callEnv.DefineShape(def)
		}
	}

	// Bind parameters, under the same minimum-arity rule as ordinary calls.
	for i, param := range fn.Params {
		callEnv.Set(param, args[i])
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
		return value.NewFault("spawn: invalid function body")
	}

	var result value.Value
	run := func() { result = ctx.Eval(body, callEnv) }
	if ctx.WithProfileLabels != nil {
		ctx.WithProfileLabels(engine.ProfileLabels{
			Layer:    "semantic",
			Function: funcName(fn),
			File:     callEnv.GetFile(),
		}, ctx.Labels, run)
	} else {
		run()
	}

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
