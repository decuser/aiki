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
	callFn := fn
	callArgs := args
	callPos := node.Pos

	for {
		fnEnv, ok := callFn.Env.(*value.Env)
		if !ok {
			return e.makeError(node, env, "invalid function environment")
		}

		// Check argument count (excluding rest params)
		if len(callArgs) < len(callFn.Params) {
			return e.makeError(node, env, "%s: want %d arguments, got %d", callFn.Name, len(callFn.Params), len(callArgs))
		}

		callEnv := value.NewEnclosedEnv(fnEnv)

		for i, param := range callFn.Params {
			callEnv.Set(param, callArgs[i])
		}

		if callFn.Rest != "" {
			restStart := len(callFn.Params)
			var restArgs []value.Value
			if restStart < len(callArgs) {
				restArgs = callArgs[restStart:]
			}
			callEnv.Set(callFn.Rest, &value.List{Elements: restArgs})
		}

		body, ok := callFn.Body.(*syntax.Node)
		if !ok {
			return e.makeError(node, env, "invalid function body")
		}

		// Enforce stack limit on non tail call frames.
		limit := callEnv.GetStackLimit()
		if limit < 1 {
			return e.makeError(node, env, "stack_limit: must be integer >= 1")
		}
		if len(callEnv.CopyStack()) >= limit {
			// Use the current call position for attribution.
			return e.makeError(&syntax.Node{Pos: callPos}, callEnv, "stack overflow")
		}

		// Push stack frame for function call
		funcName := callFn.Name
		if funcName == "" {
			funcName = "<anonymous>"
		}
		callEnv.PushFrame(funcName, callPos.Line, callEnv.GetScope())

		result := e.Eval(body, callEnv)

		callEnv.PopFrame()

		if ret, ok := result.(*value.Return); ok {
			if tc, ok := ret.Val.(*value.TailCall); ok {
				nextFn, ok := tc.Fn.(*value.Function)
				if !ok {
					return e.makeError(&syntax.Node{Pos: tc.Pos}, callEnv, "not a function: %s", tc.Fn.Type())
				}
				callFn = nextFn
				callArgs = tc.Args
				callPos = tc.Pos
				continue
			}
			return ret.Val
		}

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
