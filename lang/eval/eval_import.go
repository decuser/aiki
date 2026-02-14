package eval

import (
	"os"
	"path/filepath"

	"aiki/ebnf"
	"aiki/lang/value"
)

// evalNodeExport handles: export [name1, name2, ...]
// Records exported names on the environment.
// Currently permissive: does not restrict access to non-exported names.
// The Alpha Checklist will replace this keyword with an export() function.
func evalNodeExport(node *ebnf.Node, env *value.Env) value.Value {
	var names []string
	for _, child := range node.Children {
		if child.Type == "NAME" {
			names = append(names, child.Value)
		}
	}
	env.SetExports(names)
	return value.NULL
}

// evalNodeImport handles: from mod use [name1, name2, ...]
// Parses and evaluates the module, then copies exported (or all) names
// into the current environment.
func evalNodeImport(node *ebnf.Node, env *value.Env) value.Value {
	var moduleName string
	var importNames []string
	isFirst := true

	for _, child := range node.Children {
		if child.Type == "NAME" {
			if isFirst {
				moduleName = child.Value
				isFirst = false
			} else {
				importNames = append(importNames, child.Value)
			}
		}
	}

	if moduleName == "" {
		return value.NewError("import: missing module name")
	}

	// Resolve module path
	modulePath := resolveModulePath(moduleName, env)
	if modulePath == "" {
		return value.NewError("import: cannot find module '%s'", moduleName)
	}

	// Read module source
	data, err := os.ReadFile(modulePath)
	if err != nil {
		return value.NewError("import: cannot read '%s': %s", modulePath, err)
	}

	// Parse and evaluate module in its own environment
	if nodeGrammar == nil {
		return value.NewError("import: grammar not available")
	}

	modEnv := value.NewEnv(nil)
	// Copy HAL builtins are available via evalNodeIdent's HAL lookup

	ast, parseErr := nodeGrammar.ParseSource(string(data))
	if parseErr != nil {
		return value.NewError("import: parse error in '%s': %s", modulePath, parseErr)
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
			return value.NewError("import: '%s' is not exported by '%s'", name, moduleName)
		}
		val, ok := modEnv.Get(name)
		if !ok {
			return value.NewError("import: '%s' not defined in '%s'", name, moduleName)
		}
		env.Set(name, val)
	}

	return value.NULL
}

// nodeGrammar holds a reference to the grammar for import parsing.
// Set by SetNodeGrammar during initialization.
var nodeGrammar *ebnf.Grammar

// SetNodeGrammar stores the grammar for use by import.
func SetNodeGrammar(g *ebnf.Grammar) {
	nodeGrammar = g
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

	// Try without extension (already has .ai)
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
