package evaluator

import (
	"aiki/engine"
	"aiki/engine/runtime/hal"
	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
	"strconv"
)

func (e *Evaluator) applyFunction(fn value.Value, args []value.Value, node *syntax.Node, env *value.Env) value.Value {
	return e.applyFunctionOwned(fn, args, nil, node, env)
}

// applyFunctionOwned is the internal call boundary for argument-frame ownership.
// argFrame is non-nil only for a proven non-escaping user function.
func (e *Evaluator) applyFunctionOwned(fn value.Value, args []value.Value, argFrame *argFrame, node *syntax.Node, env *value.Env) value.Value {
	probe := e.semanticCallHit(fn, node, env)
	switch f := fn.(type) {
	case *value.Function:
		e.callUserEntry(probe)
		return e.numberCallResult(e.applyUserFunctionOwned(f, args, argFrame, node, env), probe)
	case value.Callable:
		if argFrame != nil {
			defer e.releaseArgFrame(argFrame)
		}
		e.callSubstrate(probe)
		var result value.Value

		needsContext := true
		if requirement, ok := f.(hal.EvalContextRequired); ok {
			needsContext = requirement.NeedsEvalContext()
		}

		if e.profileLabelsEnabled() {
			parentLabels := semanticProfileLabels(env)
			labels := parentLabels
			labels.Layer = "substrate"
			if node != nil && node.Pos.Line > 0 {
				labels.Line = strconv.Itoa(node.Pos.Line)
			}
			if named, ok := f.(hal.ProfileNamed); ok {
				labels.Primitive = named.ProfileName()
			}
			e.withProfileLabels(labels, parentLabels, func() {
				result = e.callSubstrateValue(f, args, node, env, parentLabels, needsContext, probe)
			})
		} else {
			result = e.callSubstrateValue(f, args, node, env, engine.ProfileLabels{}, needsContext, probe)
		}
		// Annotate HAL faults with call site location
		if fault, ok := result.(*value.Fault); ok && fault.File == "" {
			fault.File = env.GetFile()
			fault.Line = node.Pos.Line
			fault.Source = env.GetSourceLine(node.Pos.Line)
			fault.Stack = env.CopyStack()
		}
		return e.numberCallResult(result, probe)
	default:
		if argFrame != nil {
			e.releaseArgFrame(argFrame)
		}
		return e.makeFault(node, env, "not a function: %s", fn.Type())
	}
}

func (e *Evaluator) callSubstrateValue(f value.Callable, args []value.Value, node *syntax.Node, env *value.Env, parentLabels engine.ProfileLabels, needsContext bool, probe engine.SemanticProbe) value.Value {
	cf, contextCallable := f.(hal.ContextCallable)
	if !contextCallable || !needsContext {
		if probe != nil {
			if required, ok := f.(hal.RealizationProbeRequired); ok && required.NeedsRealizationProbe() {
				if probed, ok := f.(hal.ProbeCallable); ok {
					return probed.CallWithProbe(args, probe)
				}
			}
		}
		return f.Call(args)
	}

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
	return cf.CallWithContext(args, ctx)
}

func (e *Evaluator) callUserEntry(probe engine.SemanticProbe) {
	if counters, ok := probe.(*Counters); ok {
		counters.UserCallEntry()
	}
}

func (e *Evaluator) callSubstrate(probe engine.SemanticProbe) {
	if counters, ok := probe.(*Counters); ok {
		counters.SubstrateCall()
	}
}

func (e *Evaluator) callTailReuse(env *value.Env) {
	if counters, ok := e.activeProbe(env).(*Counters); ok {
		counters.TailCallReuse()
	}
}

func (e *Evaluator) callTailEnvReuse(env *value.Env) {
	if counters, ok := e.activeProbe(env).(*Counters); ok {
		counters.TailEnvReuse()
	}
}

func (e *Evaluator) numberCallResult(result value.Value, probe engine.SemanticProbe) value.Value {
	if counters, ok := probe.(*Counters); ok {
		counters.NumberCallResult(result)
	}
	return result
}

func (e *Evaluator) applyUserFunctionOwned(fn *value.Function, args []value.Value, initialArgFrame *argFrame, node *syntax.Node, env *value.Env) value.Value {
	currentFn := fn
	currentArgs := args
	currentArgFrame := initialArgFrame
	callSite := node
	pushed := false
	var reusableCallEnv *value.Env

	for {
		fnEnv, ok := currentFn.Env.(*value.Env)
		if !ok {
			if currentArgFrame != nil {
				e.releaseArgFrame(currentArgFrame)
			}
			return e.makeFault(callSite, env, "invalid function environment")
		}

		// Check argument count (excluding rest params)
		if len(currentArgs) < len(currentFn.Params) {
			if currentArgFrame != nil {
				e.releaseArgFrame(currentArgFrame)
			}
			return e.makeFault(callSite, env, "%s: want %d arguments, got %d", currentFn.Name, len(currentFn.Params), len(currentArgs))
		}

		var callEnv *value.Env
		if reusableCallEnv != nil {
			callEnv = reusableCallEnv
			callEnv.ResetCallEnv(fnEnv, env)
			reusableCallEnv = nil
			e.callTailEnvReuse(callEnv)
		} else {
			callEnv = value.NewCallEnv(fnEnv, env)
		}
		callEnv.BindCallParams(currentFn.Params, currentArgs)

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
			if currentArgFrame != nil {
				callEnv.ClearCallParams()
				e.releaseArgFrame(currentArgFrame)
			}
			return e.makeFault(callSite, env, "invalid function body")
		}

		funcName := currentFn.Name
		if funcName == "" {
			funcName = "<anonymous>"
		}

		// Enforce non tail call stack limit on first entry only.
		line := callSite.Pos.Line
		scope := callEnv.GetScope()
		if !pushed {
			limit := callEnv.GetStackLimit()
			if limit > 0 && callEnv.StackDepth() >= limit {
				if currentArgFrame != nil {
					callEnv.ClearCallParams()
					e.releaseArgFrame(currentArgFrame)
				}
				return e.makeFault(callSite, callEnv, "stack overflow")
			}
			callEnv.PushFrame(funcName, line, scope)
			pushed = true
		} else {
			// Tail call - reuse the top frame.
			callEnv.ReplaceTopFrame(funcName, line, scope)
		}

		var result value.Value
		if e.profileLabelsEnabled() {
			labels := engine.ProfileLabels{Layer: "semantic", Function: funcName, File: callEnv.GetFile()}
			restore := semanticProfileLabels(env)
			e.withProfileLabels(labels, restore, func() {
				result = e.evalTail(body, callEnv)
			})
		} else {
			result = e.evalTail(body, callEnv)
		}

		// Tail call jump
		if tc, ok := result.(*tailCallValue); ok {
			e.semanticHitDetail(engine.SemanticCall, funcNameValue(tc.Fn), tc.Node, callEnv)
			e.callTailReuse(callEnv)
			if currentFn.TailEnvReusable {
				reusableCallEnv = callEnv
			}
			if currentArgFrame != nil {
				callEnv.ClearCallParams()
				e.releaseArgFrame(currentArgFrame)
			}
			currentFn = tc.Fn
			currentArgs = tc.Args
			currentArgFrame = tc.ArgFrame
			callSite = tc.Node
			continue
		}

		// Return control flow
		if ret, ok := result.(*value.Return); ok {
			if tc, ok := ret.Val.(*tailCallValue); ok {
				e.semanticHitDetail(engine.SemanticCall, funcNameValue(tc.Fn), tc.Node, callEnv)
				e.callTailReuse(callEnv)
				if currentFn.TailEnvReusable {
					reusableCallEnv = callEnv
				}
				if currentArgFrame != nil {
					callEnv.ClearCallParams()
					e.releaseArgFrame(currentArgFrame)
				}
				currentFn = tc.Fn
				currentArgs = tc.Args
				currentArgFrame = tc.ArgFrame
				callSite = tc.Node
				continue
			}
			if currentArgFrame != nil {
				callEnv.ClearCallParams()
				e.releaseArgFrame(currentArgFrame)
			}
			callEnv.PopFrame()
			return ret.Val
		}

		if shouldHalt(result) {
			if currentArgFrame != nil {
				callEnv.ClearCallParams()
				e.releaseArgFrame(currentArgFrame)
			}
			callEnv.PopFrame()
			return result
		}

		if currentArgFrame != nil {
			callEnv.ClearCallParams()
			e.releaseArgFrame(currentArgFrame)
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
	count := 0
	for _, child := range node.Children {
		if child.Type == "param_list" {
			for _, p := range child.Children {
				if p.Type == "NAME" {
					count++
				}
			}
		} else if child.Type == "NAME" {
			count++
		}
	}

	var params []string
	if count > 0 {
		params = make([]string, 0, count)
	}
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
