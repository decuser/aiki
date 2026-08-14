package evaluator

import (
	"reflect"

	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
)

type preparedSelectCase struct {
	block *syntax.Node
	bind  string
	ch    *value.Channel
}

// chooseSelect evaluates each receive channel expression exactly once, then
// waits until one case can proceed. A default case, when present, is chosen
// only when no receive case is immediately ready.
func (e *Evaluator) chooseSelect(node *syntax.Node, env *value.Env) (*preparedSelectCase, value.Value, value.Value) {
	var prepared []*preparedSelectCase
	var cases []reflect.SelectCase

	for _, child := range node.Children {
		switch child.Type {
		case "select_case":
			var bind string
			var expr *syntax.Node
			var block *syntax.Node
			for _, part := range child.Children {
				switch part.Type {
				case "NAME":
					if bind == "" {
						bind = part.Value
					}
				case "expr":
					expr = part
				case "block":
					block = part
				}
			}
			if expr == nil || block == nil {
				return nil, nil, e.makeFault(child, env, "select: invalid receive case")
			}

			chVal := e.Eval(expr, env)
			if shouldHalt(chVal) {
				return nil, nil, chVal
			}
			ch, ok := chVal.(*value.Channel)
			if !ok {
				return nil, nil, e.makeFault(child, env, "select: recv argument must be channel, got %s", chVal.Type())
			}

			prepared = append(prepared, &preparedSelectCase{block: block, bind: bind, ch: ch})
			cases = append(cases, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch.C)})

		case "select_default":
			block := child.ChildByType("block")
			if block == nil {
				return nil, nil, e.makeFault(child, env, "select: invalid default case")
			}
			prepared = append(prepared, &preparedSelectCase{block: block})
			cases = append(cases, reflect.SelectCase{Dir: reflect.SelectDefault})
		}
	}

	if len(cases) == 0 {
		return nil, nil, e.makeFault(node, env, "select: requires at least one case")
	}

	chosen, recv, ok := reflect.Select(cases)
	selected := prepared[chosen]
	if selected.ch == nil { // default
		return selected, value.EMPTY, nil
	}
	if !ok {
		// Aiki channels are not user-closeable. This protects against an
		// internally closed channel without inventing close semantics.
		return nil, nil, e.makeFault(node, env, "select: receive from closed channel")
	}
	v, ok := recv.Interface().(value.Value)
	if !ok {
		return nil, nil, e.makeFault(node, env, "select: invalid channel payload")
	}
	return selected, v, nil
}

func (e *Evaluator) evalSelect(node *syntax.Node, env *value.Env) value.Value {
	selected, received, fault := e.chooseSelect(node, env)
	if fault != nil {
		return fault
	}

	caseEnv := value.NewEnclosedEnv(env)
	if selected.bind != "" {
		caseEnv.Set(selected.bind, received)
	}
	return e.Eval(selected.block, caseEnv)
}

func (e *Evaluator) evalSelectTail(node *syntax.Node, env *value.Env) value.Value {
	selected, received, fault := e.chooseSelect(node, env)
	if fault != nil {
		return fault
	}

	caseEnv := value.NewEnclosedEnv(env)
	if selected.bind != "" {
		caseEnv.Set(selected.bind, received)
	}
	return e.evalTail(selected.block, caseEnv)
}
