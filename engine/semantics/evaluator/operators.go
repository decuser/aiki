package evaluator

import (
	"aiki/engine"
	"fmt"
	"math/big"

	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
)

type binaryOperatorKind int

const (
	operatorAdd binaryOperatorKind = iota
	operatorSub
	operatorMul
	operatorDiv
	operatorLt
	operatorGt
	operatorLte
	operatorGte
	operatorAnd
	operatorOr
)

// binaryOperatorSemantics is the evaluator's authority on operator meaning.
// Membership is checked bidirectionally against the grammar's BINOP production.
var binaryOperatorSemantics = map[string]binaryOperatorKind{
	"+":   operatorAdd,
	"-":   operatorSub,
	"*":   operatorMul,
	"/":   operatorDiv,
	"<":   operatorLt,
	">":   operatorGt,
	"<=":  operatorLte,
	">=":  operatorGte,
	"and": operatorAnd,
	"or":  operatorOr,
}

func validateBinaryOperatorCoverage(grammarOps map[string]struct{}, semanticOps map[string]binaryOperatorKind) error {
	for op := range grammarOps {
		if _, ok := semanticOps[op]; !ok {
			return fmt.Errorf("grammar BINOP has no evaluator semantics: %s", op)
		}
	}
	for op := range semanticOps {
		if _, ok := grammarOps[op]; !ok {
			return fmt.Errorf("evaluator operator has no grammar BINOP: %s", op)
		}
	}
	return nil
}

func (e *Evaluator) applyOperator(op string, left, right value.Value, node *syntax.Node, env *value.Env) value.Value {
	kind, ok := binaryOperatorSemantics[op]
	if !ok {
		return e.makeFault(node, env, "unknown operator: %s", op)
	}

	switch kind {
	case operatorAdd, operatorSub, operatorMul, operatorDiv:
		e.semanticHit(engine.SemanticArithmetic, node, env)
	case operatorLt, operatorGt, operatorLte, operatorGte:
		e.semanticHit(engine.SemanticComparison, node, env)
	}

	switch kind {
	case operatorAdd:
		return e.opAdd(left, right, node, env)
	case operatorSub:
		return e.opSub(left, right, node, env)
	case operatorMul:
		return e.opMul(left, right, node, env)
	case operatorDiv:
		return e.opDiv(left, right, node, env)
	case operatorLt:
		return e.opLt(left, right, node, env)
	case operatorGt:
		return e.opGt(left, right, node, env)
	case operatorLte:
		return e.opLte(left, right, node, env)
	case operatorGte:
		return e.opGte(left, right, node, env)
	case operatorAnd:
		if !value.IsTruthy(left) {
			return left
		}
		return right
	case operatorOr:
		if value.IsTruthy(left) {
			return left
		}
		return right
	default:
		return e.makeFault(node, env, "unknown operator: %s", op)
	}
}

func (e *Evaluator) opAdd(left, right value.Value, node *syntax.Node, env *value.Env) value.Value {
	if ln, ok := left.(*value.Number); ok {
		if rn, ok := right.(*value.Number); ok {
			result := new(big.Rat).Add(ln.Val, rn.Val)
			return &value.Number{Val: result}
		}
	}
	if ls, ok := left.(*value.String); ok {
		if rs, ok := right.(*value.String); ok {
			return &value.String{Val: ls.Val + rs.Val}
		}
	}
	return e.makeFault(node, env, "cannot add %s and %s", left.Type(), right.Type())
}

func (e *Evaluator) opSub(left, right value.Value, node *syntax.Node, env *value.Env) value.Value {
	if ln, ok := left.(*value.Number); ok {
		if rn, ok := right.(*value.Number); ok {
			result := new(big.Rat).Sub(ln.Val, rn.Val)
			return &value.Number{Val: result}
		}
	}
	return e.makeFault(node, env, "cannot subtract %s and %s", left.Type(), right.Type())
}

func (e *Evaluator) opMul(left, right value.Value, node *syntax.Node, env *value.Env) value.Value {
	if ln, ok := left.(*value.Number); ok {
		if rn, ok := right.(*value.Number); ok {
			result := new(big.Rat).Mul(ln.Val, rn.Val)
			return &value.Number{Val: result}
		}
	}
	return e.makeFault(node, env, "cannot multiply %s and %s", left.Type(), right.Type())
}

func (e *Evaluator) opDiv(left, right value.Value, node *syntax.Node, env *value.Env) value.Value {
	if ln, ok := left.(*value.Number); ok {
		if rn, ok := right.(*value.Number); ok {
			if rn.Val.Sign() == 0 {
				return e.makeFault(node, env, "division by zero")
			}
			result := new(big.Rat).Quo(ln.Val, rn.Val)
			return &value.Number{Val: result}
		}
	}
	return e.makeFault(node, env, "cannot divide %s and %s", left.Type(), right.Type())
}

func (e *Evaluator) opLt(left, right value.Value, node *syntax.Node, env *value.Env) value.Value {
	if ln, ok := left.(*value.Number); ok {
		if rn, ok := right.(*value.Number); ok {
			if ln.Val.Cmp(rn.Val) < 0 {
				return value.TRUE
			}
			return value.FALSE
		}
	}
	return e.makeFault(node, env, "cannot compare %s and %s", left.Type(), right.Type())
}

func (e *Evaluator) opGt(left, right value.Value, node *syntax.Node, env *value.Env) value.Value {
	if ln, ok := left.(*value.Number); ok {
		if rn, ok := right.(*value.Number); ok {
			if ln.Val.Cmp(rn.Val) > 0 {
				return value.TRUE
			}
			return value.FALSE
		}
	}
	return e.makeFault(node, env, "cannot compare %s and %s", left.Type(), right.Type())
}

func (e *Evaluator) opLte(left, right value.Value, node *syntax.Node, env *value.Env) value.Value {
	if ln, ok := left.(*value.Number); ok {
		if rn, ok := right.(*value.Number); ok {
			if ln.Val.Cmp(rn.Val) <= 0 {
				return value.TRUE
			}
			return value.FALSE
		}
	}
	return e.makeFault(node, env, "cannot compare %s and %s", left.Type(), right.Type())
}

func (e *Evaluator) opGte(left, right value.Value, node *syntax.Node, env *value.Env) value.Value {
	if ln, ok := left.(*value.Number); ok {
		if rn, ok := right.(*value.Number); ok {
			if ln.Val.Cmp(rn.Val) >= 0 {
				return value.TRUE
			}
			return value.FALSE
		}
	}
	return e.makeFault(node, env, "cannot compare %s and %s", left.Type(), right.Type())
}
