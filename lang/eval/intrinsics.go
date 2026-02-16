package eval

import (
	"aiki/ebnf"
	"aiki/lang/value"
	"aiki/layers/hal"
	"fmt"
	"os"
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

	return dispatchCall(fn, args, env, node)
}

// dispatchCall applies a function to evaluated arguments.
// Handles intrinsics, builtins, and user functions.
// The node parameter is the call site, used to annotate errors.
func dispatchCall(fn value.Value, args []value.Value, env *value.Env, node *ebnf.Node) value.Value {
	switch f := fn.(type) {
	case *value.Intrinsic:
		return evalIntrinsic(f.Name, args, env, node)
	case *value.Builtin:
		result := f.Fn(args...)
		// Annotate bare errors from builtins with call-site position
		if err, ok := result.(*value.Error); ok && err.Line == 0 {
			return value.AnnotateError(err,
				env.GetFile(),
				node.Line,
				env.GetSourceLine(node.Line),
				env.CopyStack(),
			)
		}
		return result
	case *value.Function:
		return applyNodeFunc(f, args, node.Line)
	default:
		return makeError(env, node, "not callable: %s", fn.Type())
	}
}

// evalIntrinsic dispatches to the appropriate intrinsic implementation.
func evalIntrinsic(name string, args []value.Value, env *value.Env, node *ebnf.Node) value.Value {
	switch name {
	case "apply":
		return evalApply(args, env, node)
	case "load":
		return evalLoad(args, env, node)
	case "import":
		return evalImport(args, env, node)
	case "export":
		return evalExport(args, env, node)
	case "spawn":
		return evalSpawn(args, env, node)
	default:
		return makeError(env, node, "unknown intrinsic: %s", name)
	}
}

// evalApply implements apply(fn, list) — spreads list as args.
func evalApply(args []value.Value, env *value.Env, node *ebnf.Node) value.Value {
	if len(args) != 2 {
		return makeError(env, node, "apply: want 2 arguments, got %d", len(args))
	}

	fn := args[0]
	listVal := args[1]

	list, ok := listVal.(*value.List)
	if !ok {
		return makeError(env, node, "apply: second argument must be list, got %s", listVal.Type())
	}

	return dispatchCall(fn, list.Elements, env, node)
}

// evalLoad implements load(path) — reads and evaluates a file.
func evalLoad(args []value.Value, env *value.Env, node *ebnf.Node) value.Value {
	if len(args) != 1 {
		return makeError(env, node, "load: want 1 argument, got %d", len(args))
	}

	pathStr, ok := args[0].(*value.String)
	if !ok {
		return makeError(env, node, "load: expected string path, got %s", args[0].Type())
	}

	if nodeGrammar == nil {
		return makeError(env, node, "load: grammar not available")
	}

	return RunFileNode(nodeGrammar, pathStr.Value, env)
}

// evalPipeCallSafe dispatches a pipe call, handling Fn: nil specials.
func evalPipeCallSafe(node *ebnf.Node, pipedValue value.Value, env *value.Env) value.Value {
	var fn value.Value
	var callNode *ebnf.Node
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
				callNode = child
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
		return makeError(env, node, "pipe: could not find function")
	}
	if isError(fn) {
		return fn
	}

	// Use callNode if found, otherwise fall back to node
	if callNode == nil {
		callNode = node
	}
	return dispatchCall(fn, args, env, callNode)
}

func evalSpawn(args []value.Value, env *value.Env, node *ebnf.Node) value.Value {
	if len(args) < 1 {
		return makeError(env, node, "spawn: want at least 1 argument (function)")
	}

	fn, ok := args[0].(*value.Function)
	if !ok {
		return makeError(env, node, "spawn: expected function as first argument")
	}

	// Capture arguments passed to spawn: spawn(fn, arg1, arg2)
	fnArgs := args[1:]

	// Use the HAL scheduler to launch a goroutine
	hal.DefaultScheduler.Spawn(func() {
		result := applyNodeFunc(fn, fnArgs, node.Line)
		// Log errors from spawned functions (they can't propagate)
		if err, ok := result.(*value.Error); ok {
			// TODO: Consider a proper error channel or logger
			fmt.Fprintf(os.Stderr, "spawn: %s\n", err.Message)
		}
	})

	return value.True
}
