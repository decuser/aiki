package evaluator

import (
	"strconv"
	"strings"

	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
)

func (e *Evaluator) evalNumber(node *syntax.Node, env *value.Env) value.Value {
	num, err := value.NewNumberFromString(node.Value)
	if err != nil {
		return e.makeFault(node, env, "invalid number: %s", node.Value)
	}
	return num
}

func (e *Evaluator) evalString(node *syntax.Node, env *value.Env) value.Value {
	s, err := strconv.Unquote(node.Value)
	if err != nil {
		return e.makeFault(node, env, "invalid string: %s", node.Value)
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
	return value.EMPTY
}

func (e *Evaluator) evalSymbol(node *syntax.Node, env *value.Env) value.Value {
	return &value.Symbol{Val: strings.TrimPrefix(node.Value, ":")}
}

func (e *Evaluator) evalShape(node *syntax.Node, env *value.Env) value.Value {
	// SHAPE token evaluates to symbol (used in list literals)
	return &value.Symbol{Val: strings.TrimPrefix(node.Value, "@")}
}

func (e *Evaluator) evalName(node *syntax.Node, env *value.Env) value.Value {
	name := node.Value

	// Check environment first (allows shadowing)
	if val, ok := env.Get(name); ok {
		return val
	}

	// Then check builtins via runtime (using env's scope)
	if e.runtime != nil {
		if builtin, ok := e.runtime.GetBuiltin(name, env.GetAuthority()); ok {
			return builtin
		}
	}

	return e.makeFault(node, env, "undefined: %s", name)
}

func (e *Evaluator) evalTerminal(node *syntax.Node, env *value.Env) value.Value {
	switch node.Value {
	case "true":
		return value.TRUE
	case "false":
		return value.FALSE
	}
	return value.EMPTY
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
		Params:          params,
		Rest:            rest,
		Body:            body,
		Env:             env,
		TailEnvReusable: bodyHasNoClosureLiteral(body),
	}
}

func bodyHasNoClosureLiteral(node *syntax.Node) bool {
	if node == nil {
		return true
	}
	for _, child := range node.Children {
		if child.Type == "func_literal" {
			return false
		}
		if !bodyHasNoClosureLiteral(child) {
			return false
		}
	}
	return true
}

func (e *Evaluator) evalList(node *syntax.Node, env *value.Env) value.Value {
	elemCount := 0
	for _, child := range node.Children {
		if child.Type != "TERMINAL" && child.Type != "SHAPE" {
			elemCount++
		}
	}

	var elements []value.Value
	if elemCount > 0 {
		elements = make([]value.Value, elemCount)
	}
	var shape string
	i := 0

	for _, child := range node.Children {
		if child.Type == "SHAPE" {
			shape = strings.TrimPrefix(child.Value, "@")
		} else if child.Type != "TERMINAL" {
			val := e.Eval(child, env)
			if shouldHalt(val) {
				return val
			}
			elements[i] = val
			i++
		}
	}

	return &value.List{Elements: elements, Shape: shape}
}
