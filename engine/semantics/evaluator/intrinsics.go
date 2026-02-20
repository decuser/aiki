// intrinsics.go implements intrinsic functions that need evaluator context.
// These cannot be simple HAL primitives because they require access to:
// - The evaluator (to evaluate more code)
// - The scope/environment
// - The grammar (for load/import)
package evaluator

import (
	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
)

// Intrinsic names that are handled specially by the evaluator.
var intrinsics = map[string]bool{
	"spawn":  true,
	"load":   true,
	"import": true,
	"export": true,
	"apply":  true,
}

// IsIntrinsic checks if a name is an intrinsic function.
func IsIntrinsic(name string) bool {
	return intrinsics[name]
}

// evalIntrinsic evaluates an intrinsic function call.
func evalIntrinsic(e *Evaluator, name string, args []value.Value, scope *Scope, node *syntax.Node) (value.Value, error) {
	switch name {
	case "spawn":
		return evalSpawn(e, args, scope, node)
	case "load":
		return evalLoad(e, args, scope, node)
	case "import":
		return evalImport(e, args, scope, node)
	case "export":
		return evalExport(e, args, scope, node)
	case "apply":
		return evalApply(e, args, scope, node)
	default:
		return value.NullValue(), makeError(scope, node, "unknown intrinsic: %s", name)
	}
}

// evalSpawn implements spawn(fn, args...) - launches a concurrent task.
func evalSpawn(e *Evaluator, args []value.Value, scope *Scope, node *syntax.Node) (value.Value, error) {
	if len(args) < 1 {
		return value.NullValue(), makeError(scope, node, "spawn: want at least 1 argument (function)")
	}

	fn := args[0]
	if fn.Type != value.Function {
		return value.NullValue(), makeError(scope, node, "spawn: expected function as first argument")
	}

	// Capture arguments passed to spawn
	fnArgs := args[1:]

	// Use HAL to spawn the goroutine
	e.runtime.Spawn(func() {
		result, err := applyCall(e, fn, fnArgs, scope, node)
		if err != nil {
			// Log errors from spawned functions (they can't propagate)
			e.runtime.LogError(err)
		}
		if isError(result) {
			e.runtime.LogError(asError(result))
		}
	})

	return value.True(), nil
}

// evalLoad implements load(path) - reads and evaluates a file.
func evalLoad(e *Evaluator, args []value.Value, scope *Scope, node *syntax.Node) (value.Value, error) {
	if len(args) != 1 {
		return value.NullValue(), makeError(scope, node, "load: want 1 argument, got %d", len(args))
	}

	pathVal := args[0]
	if pathVal.Type != value.String {
		return value.NullValue(), makeError(scope, node, "load: expected string path, got %s", value.TypeName(pathVal.Type))
	}

	path, _ := pathVal.Data.(string)

	if e.grammar == nil {
		return value.NullValue(), makeError(scope, node, "load: grammar not available")
	}

	return e.RunFile(path, scope)
}

// evalImport implements import("module", :name1, :name2, ...).
func evalImport(e *Evaluator, args []value.Value, scope *Scope, node *syntax.Node) (value.Value, error) {
	if len(args) < 2 {
		return value.NullValue(), makeError(scope, node, "import: want module and at least 1 name")
	}

	// First arg: module name (string)
	moduleVal := args[0]
	if moduleVal.Type != value.String {
		return value.NullValue(), makeError(scope, node, "import: expected string module name, got %s", value.TypeName(moduleVal.Type))
	}
	moduleName, _ := moduleVal.Data.(string)

	// Remaining args: names to import (symbols)
	var importNames []string
	for _, arg := range args[1:] {
		if arg.Type != value.Symbol {
			return value.NullValue(), makeError(scope, node, "import: expected symbol, got %s", value.TypeName(arg.Type))
		}
		name, _ := arg.Data.(string)
		importNames = append(importNames, name)
	}

	// Resolve and load module
	modulePath, err := e.runtime.ResolvePath(moduleName, scope.GetFile())
	if err != nil {
		return value.NullValue(), makeError(scope, node, "import: cannot find module '%s'", moduleName)
	}

	// Read module source
	content, err := e.runtime.ReadFile(modulePath)
	if err != nil {
		return value.NullValue(), makeError(scope, node, "import: cannot read '%s': %s", modulePath, err)
	}

	if e.grammar == nil {
		return value.NullValue(), makeError(scope, node, "import: grammar not available")
	}

	// Parse and evaluate module in its own environment
	modScope := NewScope(nil)
	modScope.SetFile(modulePath)
	modScope.SetSource(string(content))

	lexer := syntax.NewLexer(modulePath, string(content), e.grammar)
	parser, err := syntax.NewParser(lexer, e.grammar)
	if err != nil {
		return value.NullValue(), makeError(scope, node, "import: parse error in '%s': %s", modulePath, err)
	}

	ast, err := parser.Parse()
	if err != nil {
		return value.NullValue(), makeError(scope, node, "import: parse error in '%s': %s", modulePath, err)
	}

	result, err := e.Eval(ast, modScope)
	if err != nil {
		return result, err
	}
	if isError(result) {
		return result, nil
	}

	// Copy requested names into calling environment
	exports := modScope.GetExports()
	for _, name := range importNames {
		// If module declared exports, only allow those
		if len(exports) > 0 && !contains(exports, name) {
			return value.NullValue(), makeError(scope, node, "import: '%s' is not exported by '%s'", name, moduleName)
		}

		val, ok := modScope.Get(name)
		if !ok {
			return value.NullValue(), makeError(scope, node, "import: '%s' not defined in '%s'", name, moduleName)
		}
		scope.Define(name, val)
	}

	return value.NullValue(), nil
}

// evalExport implements export(:name1, :name2, ...).
func evalExport(e *Evaluator, args []value.Value, scope *Scope, node *syntax.Node) (value.Value, error) {
	if len(args) == 0 {
		return value.NullValue(), makeError(scope, node, "export: want at least 1 argument")
	}

	var names []string
	for _, arg := range args {
		if arg.Type != value.Symbol {
			return value.NullValue(), makeError(scope, node, "export: expected symbol, got %s", value.TypeName(arg.Type))
		}
		name, _ := arg.Data.(string)
		names = append(names, name)
	}

	scope.SetExports(names)
	return value.NullValue(), nil
}

// evalApply implements apply(fn, list) - spreads list as arguments.
func evalApply(e *Evaluator, args []value.Value, scope *Scope, node *syntax.Node) (value.Value, error) {
	if len(args) != 2 {
		return value.NullValue(), makeError(scope, node, "apply: want 2 arguments (function, list)")
	}

	fn := args[0]
	if fn.Type != value.Function {
		return value.NullValue(), makeError(scope, node, "apply: first argument must be a function")
	}

	listVal := args[1]
	if listVal.Type != value.List {
		return value.NullValue(), makeError(scope, node, "apply: second argument must be a list")
	}

	// Extract list elements as arguments
	var fnArgs []value.Value
	if list, ok := listVal.Data.([]value.Value); ok {
		fnArgs = list
	} else if shaped, ok := listVal.Data.(*ShapedList); ok {
		fnArgs = shaped.Elements
	}

	return applyCall(e, fn, fnArgs, scope, node)
}

// contains checks if a string slice contains a value.
func contains(slice []string, val string) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}
