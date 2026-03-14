package evaluator

import (
	"aiki/engine/runtime/hal"
	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
)

func (e *Evaluator) applyFunction(fn value.Value, args []value.Value, node *syntax.Node, env *value.Env) value.Value {
	switch f := fn.(type) {
	case *value.Function:
		return e.applyUserFunction(f, args, node, env)
	case value.Callable:
		// Set context before calling builtin so intrinsics have access
		ctx := &hal.EvalContext{
			Env:     env,
			Node:    node,
			Grammar: e.grammar,
			Eval:    e.Eval,
		}
		e.runtime.SetContext(ctx)
		result := f.Call(args)
		// Annotate HAL errors with call site location
		if err, ok := result.(*value.Error); ok && err.File == "" {
			err.File = env.GetFile()
			err.Line = node.Pos.Line
			err.Source = env.GetSourceLine(node.Pos.Line)
			err.Stack = env.CopyStack()
		}
		return result
	default:
		return e.makeError(node, env, "not a function: %s", fn.Type())
	}
}

func (e *Evaluator) applyUserFunction(fn *value.Function, args []value.Value, node *syntax.Node, env *value.Env) value.Value {
	currentFn := fn
	currentArgs := args
	callSite := node
	pushed := false

	for {
		fnEnv, ok := currentFn.Env.(*value.Env)
		if !ok {
			return e.makeError(callSite, env, "invalid function environment")
		}

		// Check argument count (excluding rest params)
		if len(currentArgs) < len(currentFn.Params) {
			return e.makeError(callSite, env, "%s: want %d arguments, got %d", currentFn.Name, len(currentFn.Params), len(currentArgs))
		}

		callEnv := value.NewEnclosedEnv(fnEnv)

		for i, param := range currentFn.Params {
			callEnv.Set(param, currentArgs[i])
		}

		if currentFn.Rest != "" {
			restStart := len(currentFn.Params)
			var restArgs []value.Value
			if restStart < len(currentArgs) {
				restArgs = currentArgs[restStart:]
			}
			callEnv.Set(currentFn.Rest, &value.List{Elements: restArgs})
		}

		body, ok := currentFn.Body.(*syntax.Node)
		if !ok {
			return e.makeError(callSite, env, "invalid function body")
		}

		funcName := currentFn.Name
		if funcName == "" {
			funcName = "<anonymous>"
		}

		// Enforce non tail call stack limit on first entry only.
		if !pushed {
			limit := callEnv.GetStackLimit()
			if limit > 0 && callEnv.StackDepth() >= limit {
				return e.makeError(callSite, callEnv, "stack overflow")
			}
			callEnv.PushFrame(funcName, callSite.Pos.Line, callEnv.GetScope())
			pushed = true
		} else {
			// Tail call - reuse the top frame.
			callEnv.ReplaceTopFrame(funcName, callSite.Pos.Line, callEnv.GetScope())
		}

		result := e.evalTail(body, callEnv)

		// Tail call jump
		if tc, ok := result.(*tailCallValue); ok {
			currentFn = tc.Fn
			currentArgs = tc.Args
			callSite = tc.Node
			continue
		}

		// Return control flow
		if ret, ok := result.(*value.Return); ok {
			if tc, ok := ret.Val.(*tailCallValue); ok {
				currentFn = tc.Fn
				currentArgs = tc.Args
				callSite = tc.Node
				continue
			}
			callEnv.PopFrame()
			return ret.Val
		}

		if shouldHalt(result) {
			callEnv.PopFrame()
			return result
		}

		callEnv.PopFrame()
		return result
	}
}

func (e *Evaluator) extractParams(node *syntax.Node) ([]string, string) {
	var params []string
	var rest string

	for _, child := range node.Children {
		if child.Type == "param_list" {
			for _, p := range child.Children {
				if p.Type == "NAME" {
					params = append(params, p.Value)
				}
			}
		}
		if child.Type == "rest_param" {
			for _, p := range child.Children {
				if p.Type == "NAME" {
					rest = p.Value
				}
			}
		}
		if child.Type == "NAME" {
			params = append(params, child.Value)
		}
	}

	return params, rest
}

func (e *Evaluator) evalCallArgs(node *syntax.Node, env *value.Env) []value.Value {
	var args []value.Value
	for _, child := range node.Children {
		if child.Type != "TERMINAL" {
			val := e.Eval(child, env)
			args = append(args, val)
		}
	}
	return args
}
