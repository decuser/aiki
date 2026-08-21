package evaluator

import (
	"aiki/engine/syntax"
	"testing"
)

func TestBodyHasNoClosureLiteral(t *testing.T) {
	plain := &syntax.Node{Type: "block", Children: []*syntax.Node{{Type: "expr"}}}
	if !bodyHasNoClosureLiteral(plain) {
		t.Fatal("plain body should be reusable")
	}
	nested := &syntax.Node{Type: "block", Children: []*syntax.Node{{Type: "if_stmt", Children: []*syntax.Node{{Type: "func_literal"}}}}}
	if bodyHasNoClosureLiteral(nested) {
		t.Fatal("nested function literal must disable tail env reuse")
	}
}
