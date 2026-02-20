package evaluator_test

import (
	"fmt"
	"testing"
	
	"aiki/engine/internal"
	"aiki/engine/syntax"
	"aiki/engine/syntax/definition"
)

func printAST(node *syntax.Node, indent int) {
	prefix := ""
	for i := 0; i < indent; i++ {
		prefix += "  "
	}
	if node.Value != "" {
		fmt.Printf("%s%s: %q\n", prefix, node.Type, node.Value)
	} else {
		fmt.Printf("%s%s\n", prefix, node.Type)
	}
	for _, child := range node.Children {
		printAST(child, indent+1)
	}
}

func TestDebugAST(t *testing.T) {
	grammar := definition.New()
	grammar.SetObserver(internal.SilentObserver{})
	
	tests := []string{
		"42",
		"1 + 2",
		"let x = 5",
	}
	
	for _, src := range tests {
		fmt.Printf("\n=== %s ===\n", src)
		lexer := syntax.NewLexer("test", src, grammar)
		parser, err := syntax.NewParser(lexer, grammar)
		if err != nil {
			t.Fatalf("parser error: %v", err)
		}
		ast, err := parser.Parse()
		if err != nil {
			t.Fatalf("parse error: %v", err)
		}
		printAST(ast, 0)
	}
}

func TestDebugLambda(t *testing.T) {
	grammar := definition.New()
	grammar.SetObserver(internal.SilentObserver{})
	
	src := "let f = (x) { x + 1 }"
	fmt.Printf("\n=== %s ===\n", src)
	lexer := syntax.NewLexer("test", src, grammar)
	parser, err := syntax.NewParser(lexer, grammar)
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}
	ast, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	printAST(ast, 0)
}

func TestDebugCall(t *testing.T) {
	grammar := definition.New()
	grammar.SetObserver(internal.SilentObserver{})
	
	src := "f(5)"
	fmt.Printf("\n=== %s ===\n", src)
	lexer := syntax.NewLexer("test", src, grammar)
	parser, err := syntax.NewParser(lexer, grammar)
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}
	ast, err := parser.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	printAST(ast, 0)
}
