package eval

import (
	"aiki/ebnf"
	"aiki/lang/value"
)

// evalNodeCallSafe dispatches a function call, handling Fn: nil specials.
// This replaces direct f.Fn(args...) calls throughout the evaluator.
func evalNodeCallSafe(fn value.Value, node *ebnf.Node, env *value.Env) value.Value {
	var args []value.Value

	for _, child := range node.Children {
		if child.Type == "TERMINAL" {
			continue
		}
		val := EvalNode(child, env)
		if isError(val) {
			return val
		}
		args = append(args, val)
	}

	return dispatchCall(fn, args, env)
}

// dispatchCall applies a function to evaluated arguments.
// Handles specials (apply, load), builtin nil-Fn guard, and normal dispatch.
func dispatchCall(fn value.Value, args []value.Value, env *value.Env) value.Value {
	switch f := fn.(type) {
	case *value.Builtin:
		// Intercept specials by name before calling Fn
		switch f.Name {
		case "apply":
			return evalApply(args, env)
		case "load":
			return evalLoad(args, env)
		}
		// Guard: Fn: nil means special-only, shouldn't reach here
		if f.Fn == nil {
			return value.NewError("cannot call %s directly", f.Name)
		}
		return f.Fn(args...)
	case *value.Function:
		return applyNodeFunc(f, args)
	default:
		return value.NewError("not callable: %s", fn.Type())
	}
}

// evalApply implements apply(fn, list) — spreads list as args.
func evalApply(args []value.Value, env *value.Env) value.Value {
	if len(args) != 2 {
		return value.NewError("apply: want 2 arguments, got %d", len(args))
	}

	fn := args[0]
	listVal := args[1]

	list, ok := listVal.(*value.List)
	if !ok {
		return value.NewError("apply: second argument must be list, got %s", listVal.Type())
	}

	return dispatchCall(fn, list.Elements, env)
}

// evalLoad implements load(path) — reads and evaluates a file.
func evalLoad(args []value.Value, env *value.Env) value.Value {
	if len(args) != 1 {
		return value.NewError("load: want 1 argument, got %d", len(args))
	}

	pathStr, ok := args[0].(*value.String)
	if !ok {
		return value.NewError("load: expected string path, got %s", args[0].Type())
	}

	if nodeGrammar == nil {
		return value.NewError("load: grammar not available")
	}

	return RunFileNode(nodeGrammar, pathStr.Value, env)
}

// evalPipeCallSafe dispatches a pipe call, handling Fn: nil specials.
func evalPipeCallSafe(node *ebnf.Node, pipedValue value.Value, env *value.Env) value.Value {
	var fn value.Value
	var args []value.Value
	args = append(args, pipedValue)

	var walk func(n *ebnf.Node)
	walk = func(n *ebnf.Node) {
		for _, child := range n.Children {
			switch child.Type {
			case "primary", "postfix_expr", "infix_expr", "unary_expr", "pipe_expr", "expr":
				walk(child)
			case "NAME":
				fn = evalNodeIdent(child.Value, env)
			case "call":
				for _, callChild := range child.Children {
					if callChild.Type == "TERMINAL" {
						continue
					}
					val := EvalNode(callChild, env)
					if isError(val) {
						fn = val
						return
					}
					args = append(args, val)
				}
			}
		}
	}
	walk(node)

	if fn == nil {
		return value.NewError("pipe: could not find function")
	}
	if isError(fn) {
		return fn
	}

	return dispatchCall(fn, args, env)
}
