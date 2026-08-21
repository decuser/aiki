package evaluator

import (
	"aiki/engine"
	"aiki/engine/runtime/hal"
	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
	"strconv"
)

func (e *Evaluator) applyFunction(fn value.Value, args []value.Value, node *syntax.Node, env *value.Env) value.Value {
	callName := fn.Inspect()
	if named, ok := fn.(*value.Function); ok && named.Name != "" {
		callName = named.Name
	}
	e.semanticHitDetail(engine.SemanticCall, callName, node, env)
	switch f := fn.(type) {
	case *value.Function:
		return e.numberCallResult(e.applyUserFunction(f, args, node, env), env)
	case value.Callable:
		probe := e.activeProbe(env)
		parentLabels := semanticProfileLabels(env)
		ctx := &hal.EvalContext{
			Env:     env,
			Node:    node,
			Grammar: e.grammar,
			Eval:    e.Eval,
			Probe:   probe,
			Labels:  parentLabels,
			Measure: func(target value.Value, targetArgs []value.Value, attributed bool) (value.Value, engine.SemanticMeasurement) {
				return e.measure(target, targetArgs, node, env, attributed)
			},
			WithProfileLabels: e.withProfileLabels,
		}
		if faults, ok := e.runtime.(hal.AsyncFaultSource); ok {
			ctx.AsyncFault = faults.AsyncFaults()
			ctx.ReportAsyncFault = faults.ReportAsyncFault
		}
		var result value.Value
		labels := parentLabels
		labels.Layer = "substrate"
		if node != nil && node.Pos.Line > 0 {
			labels.Line = strconv.Itoa(node.Pos.Line)
		}
		if named, ok := f.(hal.ProfileNamed); ok {
			labels.Primitive = named.ProfileName()
		}
		e.withProfileLabels(labels, parentLabels, func() {
			if cf, ok := f.(hal.ContextCallable); ok {
				result = cf.CallWithContext(args, ctx)
			} else {
				result = f.Call(args)
			}
		})
		// Annotate HAL faults with call site location
		if fault, ok := result.(*value.Fault); ok && fault.File == "" {
			fault.File = env.GetFile()
			fault.Line = node.Pos.Line
			fault.Source = env.GetSourceLine(node.Pos.Line)
			fault.Stack = env.CopyStack()
		}
		return e.numberCallResult(result, env)
	default:
		return e.makeFault(node, env, "not a function: %s", fn.Type())
	}
}

func (e *Evaluator) numberCallResult(result value.Value, env *value.Env) value.Value {
	probe := e.activeProbe(env)
	if counters, ok := probe.(*Counters); ok {
		counters.NumberCallResult(result)
	}
	return result
}

func (e *Evaluator) applyUserFunction(fn *value.Function, args []value.Value, node *syntax.Node, env *value.Env) value.Value {
	currentFn := fn
	currentArgs := args
	callSite := node
	pushed := false

	for {
		fnEnv, ok := currentFn.Env.(*value.Env)
		if !ok {
			return e.makeFault(callSite, env, "invalid function environment")
		}

		// Check argument count (excluding rest params)
		if len(currentArgs) < len(currentFn.Params) {
			return e.makeFault(callSite, env, "%s: want %d arguments, got %d", currentFn.Name, len(currentFn.Params), len(currentArgs))
		}

		callEnv := value.NewCallEnv(fnEnv, env)
		callEnv.SetSemanticProbe(e.activeProbe(env))

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
			return e.makeFault(callSite, env, "invalid function body")
		}

		funcName := currentFn.Name
		if funcName == "" {
			funcName = "<anonymous>"
		}

		// Enforce non tail call stack limit on first entry only.
		if !pushed {
			limit := callEnv.GetStackLimit()
			if limit > 0 && callEnv.StackDepth() >= limit {
				return e.makeFault(callSite, callEnv, "stack overflow")
			}
			callEnv.PushFrame(funcName, callSite.Pos.Line, callEnv.GetScope())
			pushed = true
		} else {
			// Tail call - reuse the top frame.
			callEnv.ReplaceTopFrame(funcName, callSite.Pos.Line, callEnv.GetScope())
		}

		var result value.Value
		labels := engine.ProfileLabels{Layer: "semantic", Function: funcName, File: callEnv.GetFile()}
		restore := semanticProfileLabels(env)
		e.withProfileLabels(labels, restore, func() {
			result = e.evalTail(body, callEnv)
		})

		// Tail call jump
		if tc, ok := result.(*tailCallValue); ok {
			e.semanticHitDetail(engine.SemanticCall, funcNameValue(tc.Fn), tc.Node, callEnv)
			currentFn = tc.Fn
			currentArgs = tc.Args
			callSite = tc.Node
			continue
		}

		// Return control flow
		if ret, ok := result.(*value.Return); ok {
			if tc, ok := ret.Val.(*tailCallValue); ok {
				e.semanticHitDetail(engine.SemanticCall, funcNameValue(tc.Fn), tc.Node, callEnv)
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

func funcNameValue(fn *value.Function) string {
	if fn == nil {
		return "<fn>"
	}
	if fn.Name != "" {
		return fn.Name
	}
	return "<anonymous>"
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
