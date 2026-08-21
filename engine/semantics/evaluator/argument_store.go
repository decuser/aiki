package evaluator

import (
	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
)

const (
	argFrameInlineCapacity   = 2
	argFrameMaxRetainedSpill = 64
)

// argFrame is an exclusively-owned temporary carrier for evaluated arguments.
// The frame is external to Env so calls that cannot safely reuse arguments do
// not pay for it. Small arities stay inline; larger arities promote within the
// frame to a reusable slice.
type argFrame struct {
	inline [argFrameInlineCapacity]value.Value
	spill  []value.Value
	used   int
}

type evaluatedArgs struct {
	Values []value.Value
	Frame  *argFrame
}

func reusableArgTarget(fn value.Value) bool {
	f, ok := fn.(*value.Function)
	return ok && f.TailEnvReusable && f.Rest == ""
}

func (e *Evaluator) acquireArgValues(fn value.Value, arity int, env *value.Env) evaluatedArgs {
	if arity == 0 {
		return evaluatedArgs{}
	}

	if !reusableArgTarget(fn) {
		if counters, ok := e.activeProbe(env).(*Counters); ok {
			counters.RecordArgDurable()
		}
		return evaluatedArgs{Values: make([]value.Value, arity)}
	}

	var frame *argFrame
	reused := false
	if cached := e.argFrames.Get(); cached != nil {
		frame = cached.(*argFrame)
		reused = true
	} else {
		frame = &argFrame{}
	}

	frame.used = arity
	promoted := arity > argFrameInlineCapacity
	var values []value.Value
	if !promoted {
		values = frame.inline[:arity]
	} else {
		if cap(frame.spill) < arity {
			frame.spill = make([]value.Value, arity)
		} else {
			frame.spill = frame.spill[:arity]
			clear(frame.spill)
		}
		values = frame.spill
	}

	if counters, ok := e.activeProbe(env).(*Counters); ok {
		if reused {
			counters.RecordArgFrameReused(promoted)
		} else {
			counters.RecordArgFrameNew(promoted)
		}
	}
	return evaluatedArgs{Values: values, Frame: frame}
}

func (e *Evaluator) releaseArgFrame(frame *argFrame) {
	if frame == nil {
		return
	}
	if frame.used <= argFrameInlineCapacity {
		clear(frame.inline[:frame.used])
	} else if frame.spill != nil {
		clear(frame.spill[:frame.used])
	}
	if frame.spill != nil {
		if cap(frame.spill) > argFrameMaxRetainedSpill {
			frame.spill = nil
		} else {
			frame.spill = frame.spill[:0]
		}
	}
	frame.used = 0
	e.argFrames.Put(frame)
}

func (e *Evaluator) releaseEvaluatedArgs(args evaluatedArgs) {
	if args.Frame != nil {
		e.releaseArgFrame(args.Frame)
	}
}

func (e *Evaluator) applyEvaluatedFunction(fn value.Value, args evaluatedArgs, node *syntax.Node, env *value.Env) value.Value {
	return e.applyFunctionOwned(fn, args.Values, args.Frame, node, env)
}

func callNode(node *syntax.Node) *syntax.Node {
	if node == nil {
		return nil
	}
	if node.Type == "call" {
		return node
	}
	for _, child := range node.Children {
		if child.Type == "call" {
			return child
		}
	}
	return nil
}

// evalCallArgsFor evaluates the explicit arguments in node into storage chosen
// after the callee is known. leading is used by pipeline calls, whose piped
// value occupies argument slot zero but is not a syntactic argument child.
func (e *Evaluator) evalCallArgsFor(fn value.Value, node *syntax.Node, env *value.Env, leading value.Value, hasLeading bool) evaluatedArgs {
	explicit := 0
	if node != nil {
		for _, child := range node.Children {
			if child.Type != "TERMINAL" {
				explicit++
			}
		}
	}

	total := explicit
	if hasLeading {
		total++
	}
	if counters, ok := e.activeProbe(env).(*Counters); ok {
		counters.RecordCallArity(total)
	}
	args := e.acquireArgValues(fn, total, env)
	if total == 0 {
		return args
	}

	i := 0
	if hasLeading {
		args.Values[0] = leading
		i = 1
	}
	if node != nil {
		for _, child := range node.Children {
			if child.Type == "TERMINAL" {
				continue
			}
			args.Values[i] = e.Eval(child, env)
			i++
		}
	}
	return args
}

func (e *Evaluator) tailCallWithArgs(fn *value.Function, args evaluatedArgs, node *syntax.Node, env *value.Env) *tailCallValue {
	if args.Frame != nil {
		if counters, ok := e.activeProbe(env).(*Counters); ok {
			counters.RecordArgTailTransfer()
		}
	}
	return &tailCallValue{Fn: fn, Args: args.Values, ArgFrame: args.Frame, Node: node}
}
