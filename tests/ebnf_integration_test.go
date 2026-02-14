package tests

import (
	"fmt"
	"testing"

	"aiki/ebnf"
)

// TestIntegration tests the full pipeline without depending on lang/value
// This is a standalone test - for real integration, eval.go needs lang/value

func TestIntegrationPrintAST(t *testing.T) {
	g, err := ebnf.ParseFile("../cmd/grammar.ebnf")
	if err != nil {
		t.Fatalf("parse grammar error: %v", err)
	}

	tests := []struct {
		name   string
		source string
	}{
		{"number", "42"},
		{"let", "let x = 42"},
		{"add", "1 + 2"},
		{"function", "let f = (x) { return x + 1 }"},
		{"call", "f(5)"},
		{"if", "if x { return 1 }"},
		{"if-else", "if x { return 1 } else { return 2 }"},
		{"while", "while x { x = x - 1 }"},
		{"list", "[1, 2, 3]"},
		{"index", "list[0]"},
		{"pipe", "x |> f()"},
		{"multi", "let x = 1\nlet y = 2\nx + y"},
		{"factorial", `let factorial = (n) {
			if n <= 1 {
				return 1
			}
			return n * factorial(n - 1)
		}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := g.ParseSource(tt.source)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			fmt.Printf("\n=== %s ===\n", tt.name)
			fmt.Printf("Source: %s\n", tt.source)
			fmt.Println("AST:")
			printASTCompact(ast, 0)
		})
	}
}

func printASTCompact(n *ebnf.Node, indent int) {
	prefix := ""
	for i := 0; i < indent; i++ {
		prefix += "  "
	}

	if n.IsTerminal() {
		fmt.Printf("%s%s:%q\n", prefix, n.Type, n.Value)
	} else {
		fmt.Printf("%s%s\n", prefix, n.Type)
		for _, c := range n.Children {
			printASTCompact(c, indent+1)
		}
	}
}
