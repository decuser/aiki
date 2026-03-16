package substrate

import (
	"fmt"
	"sort"
	"strings"

	"aiki/engine/runtime/hal"
	"aiki/engine/runtime/help"
	"aiki/engine/semantics/value"
)

// HelpRegistry holds function documentation, set during initialization.
var HelpRegistry *help.Registry

func halQuit(args []value.Value, ctx *hal.EvalContext) value.Value {
	CloseAllCanvases()
	return value.EXIT
}

func halReset(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 0 {
		return value.NewFault("reset: want 0 arguments, got %d", len(args))
	}
	CloseAllCanvases()
	return value.RESET
}

func halHelp(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) == 0 {
		return showHelpIndex(ctx)
	}

	s, ok := args[0].(*value.String)
	if !ok {
		return value.NewFault("help: expected string, got %s", args[0].Type())
	}

	if s.Val == "" {
		return showHelpIndex(ctx)
	}

	return showHelp(s.Val, ctx)
}

func halDoc(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("doc: want 1 argument, got %d", len(args))
	}

	s, ok := args[0].(*value.String)
	if !ok {
		return value.NewFault("doc: expected string, got %s", args[0].Type())
	}

	return showDoc(s.Val, ctx)
}

func showHelpIndex(ctx *hal.EvalContext) value.Value {
	var sb strings.Builder
	sb.WriteString("Aiki Help\n\n")

	// Categories from grammar
	sb.WriteString("Syntax:\n")
	sb.WriteString("  Statements: let, if, while, match, return\n")
	sb.WriteString("  Expressions: +, -, *, /, <, >, <=, >=, and, or, not, |>\n")
	sb.WriteString("  Values: number, string, rune, symbol, list, function\n\n")

	// Functions from registry
	if HelpRegistry != nil {
		names := HelpRegistry.ListFuncs()
		sort.Strings(names)

		// Group by category
		categories := map[string][]string{
			"IO":          {},
			"List":        {},
			"Type":        {},
			"Math":        {},
			"HOF":         {},
			"Canvas":      {},
			"Concurrency": {},
			"REPL":        {},
			"Other":       {},
		}

		for _, name := range names {
			switch {
			case contains([]string{"print", "println", "read", "input"}, name):
				categories["IO"] = append(categories["IO"], name)
			case contains([]string{"first", "rest", "length", "prepend", "append", "empty", "range", "reverse"}, name):
				categories["List"] = append(categories["List"], name)
			case contains([]string{"type", "inspect", "equal", "shape", "ord", "to_str", "to_number", "to_decimal"}, name):
				categories["Type"] = append(categories["Type"], name)
			case contains([]string{"floor", "ceil", "modulo", "sum", "max", "min", "seed", "random"}, name):
				categories["Math"] = append(categories["Math"], name)
			case contains([]string{"map", "filter", "reduce", "each", "find", "any", "all"}, name):
				categories["HOF"] = append(categories["HOF"], name)
			case strings.HasPrefix(name, "canvas") || strings.HasPrefix(name, "dot") || strings.HasPrefix(name, "line") || strings.HasPrefix(name, "rect") || strings.HasPrefix(name, "circle") || strings.HasPrefix(name, "fill") || strings.HasPrefix(name, "clear") || strings.HasPrefix(name, "destroy") || strings.HasPrefix(name, "set_") || strings.HasPrefix(name, "pen"):
				categories["Canvas"] = append(categories["Canvas"], name)
			case contains([]string{"spawn", "channel", "send", "recv"}, name):
				categories["Concurrency"] = append(categories["Concurrency"], name)
			case contains([]string{"quit", "reset", "help", "doc"}, name):
				categories["REPL"] = append(categories["REPL"], name)
			default:
				categories["Other"] = append(categories["Other"], name)
			}
		}

		sb.WriteString("Functions:\n")
		for _, cat := range []string{"IO", "List", "Type", "Math", "HOF", "Canvas", "Concurrency", "REPL", "Other"} {
			if len(categories[cat]) > 0 {
				sb.WriteString(fmt.Sprintf("  %s: %s\n", cat, strings.Join(categories[cat], ", ")))
			}
		}
	}

	sb.WriteString("\nUse help(\"name\") for quick reference, doc(\"name\") for full documentation.\n")

	fmt.Fprint(Stdout, sb.String())
	return value.EMPTY
}

func showHelp(name string, ctx *hal.EvalContext) value.Value {
	// Check function registry first
	if HelpRegistry != nil {
		if entry := HelpRegistry.GetHelp(name); entry != nil {
			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("%s\n", entry.Name))
			sb.WriteString(fmt.Sprintf("  %s\n\n", entry.Help))
			sb.WriteString(fmt.Sprintf("Syntax: %s\n", entry.Template))
			fmt.Fprint(Stdout, sb.String())
			return value.EMPTY
		}
	}

	// Check grammar for syntax help
	if ctx != nil && ctx.Grammar != nil {
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
			fmt.Fprint(Stdout, sb.String())
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
			fmt.Fprint(Stdout, sb.String())
			return value.EMPTY
		}
	}

	fmt.Fprintf(Stdout, "No help for '%s'\n", name)
	return value.EMPTY
}

func showDoc(name string, ctx *hal.EvalContext) value.Value {
	// Check function registry first
	if HelpRegistry != nil {
		if entry := HelpRegistry.GetDoc(name); entry != nil {
			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("%s\n\n", entry.Name))
			sb.WriteString(entry.Doc)
			sb.WriteString("\n")
			fmt.Fprint(Stdout, sb.String())
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
				fmt.Fprint(Stdout, sb.String())
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
				fmt.Fprint(Stdout, sb.String())
				return value.EMPTY
			}
		}
	}

	fmt.Fprintf(Stdout, "No documentation for '%s'\n", name)
	return value.EMPTY
}

func halDelete(args []value.Value, ctx *hal.EvalContext) value.Value {
	if len(args) != 1 {
		return value.NewFault("delete: want 1 argument, got %d", len(args))
	}
	s, ok := args[0].(*value.String)
	if !ok {
		return value.NewFault("delete: expected string, got %s", args[0].Type())
	}
	name := s.Val
	if UserEnv == nil {
		return value.NewFault("delete: no user environment (only available in REPL)")
	}
	if UserEnv.Delete(name) {
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
