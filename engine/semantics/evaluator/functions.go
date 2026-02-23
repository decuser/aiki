package eval

import (
	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
)

func (e *Evaluator) applyFunction(fn value.Value, args []value.Value, node *syntax.Node, env *value.Env) value.Value {
	switch f := fn.(type) {
	case *value.Function:
		return e.applyUserFunction(f, args, node, env)
	case *value.Intrinsic:
		return e.applyIntrinsic(f, args, node, env)
	default:
		return e.makeError(node, env, "not a function: %s", fn.Type())
	}
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

	result := e.Eval(body, callEnv)

	if ret, ok := result.(*value.Return); ok {
		return ret.Val
	}

	return result
}

func (e *Evaluator) applyIntrinsic(fn *value.Intrinsic, args []value.Value, node *syntax.Node, env *value.Env) value.Value {
	if fn.Fn == nil {
		return e.makeError(node, env, "intrinsic not implemented: %s", fn.Name)
	}

	if call, ok := fn.Fn.(func([]value.Value) value.Value); ok {
		return call(args)
	}

	return e.makeError(node, env, "invalid intrinsic: %s", fn.Name)
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
