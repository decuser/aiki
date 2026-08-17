package substrate

import (
	"fmt"
	"sort"
	"strings"

	"aiki/engine/runtime/hal"
	"aiki/engine/semantics/value"
)

func (g *GoRuntime) halQuit(args []value.Value, ctx *hal.EvalContext) value.Value {
	g.CloseAllResources()
	return value.EXIT
}

func (g *GoRuntime) halReset(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 0 {
		return value.NewFault("reset: want 0 arguments, got %d", len(args))
	}
	g.CloseAllResources()
	return value.RESET
}

func (g *GoRuntime) halHelp(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) == 0 {
		return g.showHelpIndex(ctx)
	}

	s, ok := args[0].(*value.String)
	if !ok {
		return value.NewFault("help: expected string, got %s", args[0].Type())
	}

	if s.Val == "" {
		return g.showHelpIndex(ctx)
	}

	return g.showHelp(s.Val, ctx)
}

func (g *GoRuntime) halDoc(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("doc: want 1 argument, got %d", len(args))
	}

	s, ok := args[0].(*value.String)
	if !ok {
		return value.NewFault("doc: expected string, got %s", args[0].Type())
	}

	return g.showDoc(s.Val, ctx)
}

func (g *GoRuntime) outputHelp(text string) {
	if g.pageOutput != nil && g.pageOutput(text) {
		return
	}
	fmt.Fprint(g.stdout, text)
}

func (g *GoRuntime) showHelpIndex(ctx *hal.EvalContext) value.Value {
	var sb strings.Builder
	sb.WriteString("Aiki Help\n\n")

	// Categories from grammar
	sb.WriteString("Syntax:\n")
	sb.WriteString("  Statements: let, if, while, match, select, return\n")
	sb.WriteString("  Expressions: +, -, *, /, <, >, <=, >=, and, or, not, |>\n")
	sb.WriteString("  Values: number, string, rune, symbol, list, function\n\n")

	// Functions from registry
	if g.helpRegistry != nil {
		names := g.helpRegistry.ListFuncs()
		sort.Strings(names)

		// Group by category
		categories := map[string][]string{
			"IO":          {},
			"List":        {},
			"Type":        {},
			"Concurrency": {},
			"REPL":        {},
			"Other":       {},
		}

		for _, name := range names {
			switch {
			case contains([]string{"print", "println", "read", "input"}, name):
				categories["IO"] = append(categories["IO"], name)
			case contains([]string{"first", "rest", "length", "prepend", "append", "empty"}, name):
				categories["List"] = append(categories["List"], name)
			case contains([]string{"type", "inspect", "equal", "shape", "ord", "chr", "is_error", "to_str", "to_number", "to_symbol", "to_decimal", "shaped"}, name):
				categories["Type"] = append(categories["Type"], name)
			case contains([]string{"spawn", "channel", "send", "recv"}, name):
				categories["Concurrency"] = append(categories["Concurrency"], name)
			case contains([]string{"quit", "reset", "delete", "help", "doc", "apply", "load", "stack_limit"}, name):
				categories["REPL"] = append(categories["REPL"], name)
			default:
				categories["Other"] = append(categories["Other"], name)
			}
		}

		sb.WriteString("Prelude Functions:\n")
		for _, cat := range []string{"IO", "List", "Type", "Concurrency", "REPL", "Other"} {
			if len(categories[cat]) > 0 {
				sb.WriteString(fmt.Sprintf("  %s: %s\n", cat, strings.Join(categories[cat], ", ")))
			}
		}
	}

	// Modules from g.moduleRegistry
	if g.moduleRegistry != nil {
		pkgs := g.moduleRegistry.ListPackages()
		if len(pkgs) > 0 {
			sb.WriteString("\nModules:\n")
			sb.WriteString(fmt.Sprintf("  %s\n", strings.Join(pkgs, ", ")))
		}
	}

	sb.WriteString("\nUse help(\"name\") for prelude functions, help(\"module\") for module info,\n")
	sb.WriteString("help(\"module.func\") for module functions.\n")
	sb.WriteString("Use help(\"newline\") for the grammar-declared newline policy.\n")
	sb.WriteString("Use doc(\"name\") for full documentation.\n")

	g.outputHelp(sb.String())
	return value.EMPTY
}

func (g *GoRuntime) showHelp(name string, ctx *hal.EvalContext) value.Value {
	// Check for qualified name (module.func)
	if strings.Contains(name, ".") {
		return g.showModuleFuncHelp(name)
	}

	// Check if it's a module name
	if g.moduleRegistry != nil && g.moduleRegistry.HasPackage(name) {
		return g.showModuleHelp(name)
	}

	// Check prelude function registry
	if g.helpRegistry != nil {
		if entry := g.helpRegistry.GetHelp(name); entry != nil {
			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("%s\n", entry.Name))
			sb.WriteString(fmt.Sprintf("  %s\n\n", entry.Help))
			sb.WriteString(fmt.Sprintf("Syntax: %s\n", entry.Template))
			g.outputHelp(sb.String())
			return value.EMPTY
		}
	}

	// Check grammar for syntax help.
	if ctx != nil && ctx.Grammar != nil {
		if name == "newline" && ctx.Grammar.Newline != nil {
			var sb strings.Builder
			sb.WriteString("newline\n")
			if ctx.Grammar.Newline.Meta.Help != "" {
				sb.WriteString(fmt.Sprintf("  %s\n", ctx.Grammar.Newline.Meta.Help))
			}
			g.outputHelp(sb.String())
			return value.EMPTY
		}

		// Try production name directly
		if prod, ok := ctx.Grammar.GetProduction(name); ok {
			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("%s\n", name))
			if prod.Meta.Help != "" {
				sb.WriteString(fmt.Sprintf("  %s\n\n", prod.Meta.Help))
			}
			if prod.Meta.Template != "" {
				sb.WriteString(fmt.Sprintf("Syntax: %s\n", prod.Meta.Template))
			}
			g.outputHelp(sb.String())
			return value.EMPTY
		}

		// Try with _stmt suffix
		if prod, ok := ctx.Grammar.GetProduction(name + "_stmt"); ok {
			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("%s\n", name))
			if prod.Meta.Help != "" {
				sb.WriteString(fmt.Sprintf("  %s\n\n", prod.Meta.Help))
			}
			if prod.Meta.Template != "" {
				sb.WriteString(fmt.Sprintf("Syntax: %s\n", prod.Meta.Template))
			}
			g.outputHelp(sb.String())
			return value.EMPTY
		}
	}

	fmt.Fprintf(g.stdout, "No help for '%s'\n", name)
	return value.EMPTY
}

func (g *GoRuntime) showModuleHelp(pkgName string) value.Value {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Module: %s\n\n", pkgName))

	mh := g.moduleRegistry.GetModuleHelp(pkgName)
	if mh == nil || len(mh.Funcs) == 0 {
		sb.WriteString("  No help available.\n")
		sb.WriteString(fmt.Sprintf("  Use import(\"%s\") to load this module.\n", pkgName))
	} else {
		sb.WriteString("Exports:\n")
		// Sort function names
		names := make([]string, 0, len(mh.Funcs))
		for name := range mh.Funcs {
			names = append(names, name)
		}
		sort.Strings(names)

		for _, name := range names {
			entry := mh.Funcs[name]
			sb.WriteString(fmt.Sprintf("  %s - %s\n", entry.Template, entry.Help))
		}
		sb.WriteString(fmt.Sprintf("\nUse help(\"%s.func\") for details on a specific function.\n", pkgName))
	}

	g.outputHelp(sb.String())
	return value.EMPTY
}

func (g *GoRuntime) showModuleFuncHelp(qualName string) value.Value {
	parts := strings.SplitN(qualName, ".", 2)
	if len(parts) != 2 {
		fmt.Fprintf(g.stdout, "Invalid qualified name '%s'\n", qualName)
		return value.EMPTY
	}

	pkgName := parts[0]
	funcName := parts[1]

	// Handle nested package names like "hash/ffi.new"
	// The package name might be "hash/ffi" not "hash"
	// Try progressively longer package names
	if g.moduleRegistry != nil {
		// First try exact split
		if !g.moduleRegistry.HasPackage(pkgName) {
			// Maybe the dot is in the middle of a path-like name
			// For "hash/ffi.new", we want pkg="hash/ffi", func="new"
			// This is already handled by SplitN with limit 2
		}
	}

	if g.moduleRegistry == nil || !g.moduleRegistry.HasPackage(pkgName) {
		fmt.Fprintf(g.stdout, "Unknown module '%s'\n", pkgName)
		return value.EMPTY
	}

	mh := g.moduleRegistry.GetModuleHelp(pkgName)
	if mh == nil {
		fmt.Fprintf(g.stdout, "No help available for module '%s'\n", pkgName)
		return value.EMPTY
	}

	if entry, ok := mh.Funcs[funcName]; ok {
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("%s.%s\n", pkgName, entry.Name))
		sb.WriteString(fmt.Sprintf("  %s\n\n", entry.Help))
		sb.WriteString(fmt.Sprintf("Syntax: %s\n", entry.Template))
		g.outputHelp(sb.String())
		return value.EMPTY
	}

	fmt.Fprintf(g.stdout, "No help for '%s' in module '%s'\n", funcName, pkgName)
	return value.EMPTY
}

func (g *GoRuntime) showDoc(name string, ctx *hal.EvalContext) value.Value {
	// Check for qualified name (module.func)
	if strings.Contains(name, ".") {
		return g.showModuleFuncDoc(name)
	}

	// Check if it's a module name
	if g.moduleRegistry != nil && g.moduleRegistry.HasPackage(name) {
		return g.showModuleDoc(name)
	}

	// Check prelude function registry
	if g.helpRegistry != nil {
		if entry := g.helpRegistry.GetDoc(name); entry != nil {
			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("%s\n\n", entry.Name))
			sb.WriteString(entry.Doc)
			sb.WriteString("\n")
			g.outputHelp(sb.String())
			return value.EMPTY
		}
	}

	// Check grammar for full doc
	if ctx != nil && ctx.Grammar != nil {
		// Try production name directly
		if prod, ok := ctx.Grammar.GetProduction(name); ok {
			if prod.Meta.Doc != "" {
				var sb strings.Builder
				sb.WriteString(fmt.Sprintf("%s\n\n", name))
				sb.WriteString(prod.Meta.Doc)
				sb.WriteString("\n")
				g.outputHelp(sb.String())
				return value.EMPTY
			}
		}

		// Try with _stmt suffix
		if prod, ok := ctx.Grammar.GetProduction(name + "_stmt"); ok {
			if prod.Meta.Doc != "" {
				var sb strings.Builder
				sb.WriteString(fmt.Sprintf("%s\n", name))
				sb.WriteString(prod.Meta.Doc)
				sb.WriteString("\n")
				g.outputHelp(sb.String())
				return value.EMPTY
			}
		}
	}

	fmt.Fprintf(g.stdout, "No documentation for '%s'\n", name)
	return value.EMPTY
}

func stripDocMarkers(doc string) string {
	lines := strings.Split(doc, "\n")
	kept := lines[:0]
	for _, line := range lines {
		if strings.HasPrefix(line, "@preamble") || strings.HasPrefix(line, "@unchecked") {
			continue
		}
		kept = append(kept, line)
	}
	return strings.Join(kept, "\n")
}

func (g *GoRuntime) showModuleDoc(pkgName string) value.Value {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Module: %s\n\n", pkgName))

	mh := g.moduleRegistry.GetModuleHelp(pkgName)
	if mh == nil || len(mh.Docs) == 0 {
		sb.WriteString("No documentation available.\n")
	} else {
		if mh.Preamble != "" {
			sb.WriteString(mh.Preamble)
			sb.WriteString("\n\n")
		}

		// Show all function docs
		names := make([]string, 0, len(mh.Docs))
		for name := range mh.Docs {
			names = append(names, name)
		}
		sort.Strings(names)

		for i, name := range names {
			entry := mh.Docs[name]
			sb.WriteString(fmt.Sprintf("%s\n%s\n", entry.Name, stripDocMarkers(entry.Doc)))
			if i < len(names)-1 {
				sb.WriteString("\n---\n\n")
			}
		}
	}

	g.outputHelp(sb.String())
	return value.EMPTY
}

func (g *GoRuntime) showModuleFuncDoc(qualName string) value.Value {
	parts := strings.SplitN(qualName, ".", 2)
	if len(parts) != 2 {
		fmt.Fprintf(g.stdout, "Invalid qualified name '%s'\n", qualName)
		return value.EMPTY
	}

	pkgName := parts[0]
	funcName := parts[1]

	if g.moduleRegistry == nil || !g.moduleRegistry.HasPackage(pkgName) {
		fmt.Fprintf(g.stdout, "Unknown module '%s'\n", pkgName)
		return value.EMPTY
	}

	mh := g.moduleRegistry.GetModuleHelp(pkgName)
	if mh == nil {
		fmt.Fprintf(g.stdout, "No documentation available for module '%s'\n", pkgName)
		return value.EMPTY
	}

	if entry, ok := mh.Docs[funcName]; ok {
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("%s.%s\n\n", pkgName, entry.Name))
		if mh.Preamble != "" {
			sb.WriteString(mh.Preamble)
			sb.WriteString("\n\n")
		}
		sb.WriteString(stripDocMarkers(entry.Doc))
		sb.WriteString("\n")
		g.outputHelp(sb.String())
		return value.EMPTY
	}

	fmt.Fprintf(g.stdout, "No documentation for '%s' in module '%s'\n", funcName, pkgName)
	return value.EMPTY
}

func (g *GoRuntime) halDelete(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("delete: want 1 argument, got %d", len(args))
	}
	s, ok := args[0].(*value.String)
	if !ok {
		return value.NewFault("delete: expected string, got %s", args[0].Type())
	}
	name := s.Val
	if g.userEnv == nil {
		return value.NewFault("delete: no user environment (only available in REPL)")
	}
	if g.userEnv.Delete(name) {
		return value.TRUE
	}
	return value.FALSE
}

func contains(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
