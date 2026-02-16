package eval

import (
	"os"
	"path/filepath"

	"aiki/ebnf"
	"aiki/lang/value"
)

// nodeGrammar holds a reference to the grammar for import/load parsing.
// Set by SetNodeGrammar during initialization.
var nodeGrammar *ebnf.Grammar

// SetNodeGrammar stores the grammar for use by import and load.
// Validates that all grammar productions have handlers.
func SetNodeGrammar(g *ebnf.Grammar) {
	nodeGrammar = g
	ValidateHandlers(g)
}

// evalExport implements export(:name1, :name2, ...)
// Records exported names on the environment.
func evalExport(args []value.Value, env *value.Env, node *ebnf.Node) value.Value {
	if len(args) == 0 {
		return makeError(env, node, "export: want at least 1 argument")
	}

	var names []string
	for _, arg := range args {
		sym, ok := arg.(*value.Symbol)
		if !ok {
			return makeError(env, node, "export: expected symbol, got %s", arg.Type())
		}
		names = append(names, sym.Value)
	}

	env.SetExports(names)
	return value.NULL
}

// evalImport implements import("module", :name1, :name2, ...)
// Parses and evaluates the module, then copies exported names into the current environment.
func evalImport(args []value.Value, env *value.Env, node *ebnf.Node) value.Value {
	if len(args) < 2 {
		return makeError(env, node, "import: want module and at least 1 name")
	}

	// First arg: module name (string)
	moduleStr, ok := args[0].(*value.String)
	if !ok {
		return makeError(env, node, "import: expected string module name, got %s", args[0].Type())
	}
	moduleName := moduleStr.Value

	// Remaining args: names to import (symbols)
	var importNames []string
	for _, arg := range args[1:] {
		sym, ok := arg.(*value.Symbol)
		if !ok {
			return makeError(env, node, "import: expected symbol, got %s", arg.Type())
		}
		importNames = append(importNames, sym.Value)
	}

	// Resolve module path
	modulePath := resolveModulePath(moduleName, env)
	if modulePath == "" {
		return makeError(env, node, "import: cannot find module '%s'", moduleName)
	}

	// Read module source
	data, err := os.ReadFile(modulePath)
	if err != nil {
		return makeError(env, node, "import: cannot read '%s': %s", modulePath, err)
	}

	if nodeGrammar == nil {
		return makeError(env, node, "import: grammar not available")
	}

	// Parse and evaluate module in its own environment
	modEnv := value.NewEnv(nil)
	ast, parseErr := nodeGrammar.ParseSource(string(data))
	if parseErr != nil {
		return makeError(env, node, "import: parse error in '%s': %s", modulePath, parseErr)
	}

	result := EvalNode(ast, modEnv)
	if isError(result) {
		return result
	}

	// Copy requested names into calling environment
	exports := modEnv.GetExports()
	for _, name := range importNames {
		// If module declared exports, only allow those
		if len(exports) > 0 && !containsStr(exports, name) {
			return makeError(env, node, "import: '%s' is not exported by '%s'", name, moduleName)
		}
		val, ok := modEnv.Get(name)
		if !ok {
			return makeError(env, node, "import: '%s' not defined in '%s'", name, moduleName)
		}
		env.Set(name, val)
	}

	return value.NULL
}

func resolveModulePath(name string, env *value.Env) string {
	// Try relative to current file
	currentFile := env.GetFile()
	if currentFile != "" {
		dir := filepath.Dir(currentFile)
		candidate := filepath.Join(dir, name+".ai")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	// Try as-is with .ai extension
	candidate := name + ".ai"
	if _, err := os.Stat(candidate); err == nil {
		return candidate
	}

	// Try without extension
	if _, err := os.Stat(name); err == nil {
		return name
	}

	return ""
}

func containsStr(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
