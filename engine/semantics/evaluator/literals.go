package eval

import (
	"strconv"
	"strings"

	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
)

func (e *Evaluator) evalNumber(node *syntax.Node, env *value.Env) value.Value {
	num, err := value.NewNumberFromString(node.Value)
	if err != nil {
		return e.makeError(node, env, "invalid number: %s", node.Value)
	}
	return num
}

func (e *Evaluator) evalString(node *syntax.Node, env *value.Env) value.Value {
	s, err := strconv.Unquote(node.Value)
	if err != nil {
		return e.makeError(node, env, "invalid string: %s", node.Value)
	}
	return &value.String{Val: s}
}

func (e *Evaluator) evalRune(node *syntax.Node, env *value.Env) value.Value {
	s := node.Value
	if len(s) >= 2 {
		s = s[1 : len(s)-1]
	}
	if len(s) == 2 && s[0] == '\\' {
		switch s[1] {
		case 'n':
			return &value.Rune{Val: '\n'}
		case 't':
			return &value.Rune{Val: '\t'}
		case 'r':
			return &value.Rune{Val: '\r'}
		case '\\':
			return &value.Rune{Val: '\\'}
		case '\'':
			return &value.Rune{Val: '\''}
		}
	}
	runes := []rune(s)
	if len(runes) > 0 {
		return &value.Rune{Val: runes[0]}
	}
	return value.NULL
}

func (e *Evaluator) evalSymbol(node *syntax.Node, env *value.Env) value.Value {
	return &value.Symbol{Val: strings.TrimPrefix(node.Value, ":")}
}

func (e *Evaluator) evalName(node *syntax.Node, env *value.Env) value.Value {
	name := node.Value

	if builtin, ok := e.hal[name]; ok {
		return builtin
	}

	if val, ok := env.Get(name); ok {
		return val
	}

	return e.makeError(node, env, "undefined: %s", name)
}

func (e *Evaluator) evalTerminal(node *syntax.Node, env *value.Env) value.Value {
	switch node.Value {
	case "true":
		return value.TRUE
	case "false":
		return value.FALSE
	}
	return value.NULL
}

func (e *Evaluator) evalFunc(node *syntax.Node, env *value.Env) value.Value {
	var params []string
	var rest string
	var body *syntax.Node

	for _, child := range node.Children {
		if child.Type == "params" {
			params, rest = e.extractParams(child)
		}
		if child.Type == "block" {
			body = child
		}
	}

	return &value.Function{
		Params: params,
		Rest:   rest,
		Body:   body,
		Env:    env,
	}
}

func (e *Evaluator) evalList(node *syntax.Node, env *value.Env) value.Value {
	var elements []value.Value
	var shape string

	for _, child := range node.Children {
		if child.Type == "SHAPE" {
			shape = strings.TrimPrefix(child.Value, "@")
		} else if child.Type != "TERMINAL" {
			val := e.Eval(child, env)
			if value.IsError(val) {
				return val
			}
			elements = append(elements, val)
		}
	}

	return &value.List{Elements: elements, Shape: shape}
}
