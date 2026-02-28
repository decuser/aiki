package evaluator

import (
	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
)

func (e *Evaluator) applyFunction(fn value.Value, args []value.Value, node *syntax.Node, env *value.Env) value.Value {
	// Check for intrinsic functions that need evaluator context
	if intrinsic, ok := fn.(*Intrinsic); ok {
		return e.callIntrinsic(intrinsic.Name, args, node, env)
	}

	switch f := fn.(type) {
	case *value.Function:
		return e.applyUserFunction(f, args, node, env)
	case value.Callable:
		return f.Call(args)
	default:
		return e.makeError(node, env, "not a function: %s", fn.Type())
	}
}

// Intrinsic represents a function that needs evaluator context.
type Intrinsic struct {
	Name string
}

func (i *Intrinsic) Type() value.Type { return value.FunctionType }
func (i *Intrinsic) Inspect() string  { return "<intrinsic: " + i.Name + ">" }

// callIntrinsic dispatches to intrinsic implementations.
func (e *Evaluator) callIntrinsic(name string, args []value.Value, node *syntax.Node, env *value.Env) value.Value {
	switch name {
	case "apply":
		return e.intrinsicApply(args, node, env)
	default:
		return e.makeError(node, env, "unknown intrinsic: %s", name)
	}
}

// intrinsicApply implements apply(fn, list) - spreads list as args.
func (e *Evaluator) intrinsicApply(args []value.Value, node *syntax.Node, env *value.Env) value.Value {
	if len(args) != 2 {
		return e.makeError(node, env, "apply: want 2 arguments, got %d", len(args))
	}

	fn := args[0]
	listVal := args[1]

	list, ok := listVal.(*value.List)
	if !ok {
		return e.makeError(node, env, "apply: second argument must be list, got %s", listVal.Type())
	}

	return e.applyFunction(fn, list.Elements, node, env)
}

func (e *Evaluator) applyUserFunction(fn *value.Function, args []value.Value, node *syntax.Node, env *value.Env) value.Value {
	fnEnv, ok := fn.Env.(*value.Env)
	if !ok {
		return e.makeError(node, env, "invalid function environment")
	}

	callEnv := value.NewEnclosedEnv(fnEnv)

	for i, param := range fn.Params {
		if i < len(args) {
			callEnv.Set(param, args[i])
		} else {
			callEnv.Set(param, value.NULL)
		}
	}

	if fn.Rest != "" {
		restStart := len(fn.Params)
		var restArgs []value.Value
		if restStart < len(args) {
			restArgs = args[restStart:]
		}
		callEnv.Set(fn.Rest, &value.List{Elements: restArgs})
	}

	body, ok := fn.Body.(*syntax.Node)
	if !ok {
		return e.makeError(node, env, "invalid function body")
	}

	// Push stack frame for function call
	funcName := fn.Name
	if funcName == "" {
		funcName = "<anonymous>"
	}
	callEnv.PushFrame(funcName, node.Pos.Line, callEnv.GetScope())

	result := e.Eval(body, callEnv)

	callEnv.PopFrame()

	if ret, ok := result.(*value.Return); ok {
		return ret.Val
	}

	return result
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
