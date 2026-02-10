package tests

import (
	"testing"

	"aiki/ast"
	"aiki/parser"
)

func TestFmtCommentOrder(t *testing.T) {
	input := `# first comment
# second comment
let x = 1`

	p := parser.New(input)
	program := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	output := ast.PrintWithComments(program, p.Comments())

	expected := `# first comment
# second comment
let x = 1
`

	if output != expected {
		t.Errorf("comment order wrong:\ngot:\n%s\nwant:\n%s", output, expected)
	}
}

func TestFmtCommentOrderMultipleBlocks(t *testing.T) {
	input := `# block one
# still block one
let x = 1

# block two
# still block two
let y = 2`

	p := parser.New(input)
	program := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	output := ast.PrintWithComments(program, p.Comments())

	expected := `# block one
# still block one
let x = 1

# block two
# still block two
let y = 2
`

	if output != expected {
		t.Errorf("comment order wrong:\ngot:\n%s\nwant:\n%s", output, expected)
	}
}

func TestFmtEOLComment(t *testing.T) {
	input := `let x = 1  # end of line comment`

	p := parser.New(input)
	program := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	output := ast.PrintWithComments(program, p.Comments())

	expected := `let x = 1  # end of line comment
`

	if output != expected {
		t.Errorf("EOL comment wrong:\ngot:\n%s\nwant:\n%s", output, expected)
	}
}

func TestFmtIdempotent(t *testing.T) {
	input := `# Hash map implementation
# A hash is a list of 64 buckets

let _HASH_SIZE = 64
`

	p := parser.New(input)
	program := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	output1 := ast.PrintWithComments(program, p.Comments())

	// Parse and format again
	p2 := parser.New(output1)
	program2 := p2.Parse()
	output2 := ast.PrintWithComments(program2, p2.Comments())

	if output1 != output2 {
		t.Errorf("not idempotent:\nfirst:\n%s\nsecond:\n%s", output1, output2)
	}
}

func TestFmtListsWithCommas(t *testing.T) {
	input := `let x = [1, 2, 3]`

	p := parser.New(input)
	program := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	output := ast.PrintWithComments(program, p.Comments())

	expected := `let x = [1, 2, 3]
`

	if output != expected {
		t.Errorf("list format wrong:\ngot:\n%s\nwant:\n%s", output, expected)
	}
}

func TestFmtFunctionWithCommas(t *testing.T) {
	input := `let add = (a, b) { return a + b }`

	p := parser.New(input)
	program := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	output := ast.PrintWithComments(program, p.Comments())

	expected := `let add = (a, b) {
	return a + b
}
`

	if output != expected {
		t.Errorf("function format wrong:\ngot:\n%s\nwant:\n%s", output, expected)
	}
}

func TestFmtCallWithCommas(t *testing.T) {
	input := `add(1, 2)`

	p := parser.New(input)
	program := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	output := ast.PrintWithComments(program, p.Comments())

	expected := `add(1, 2)
`

	if output != expected {
		t.Errorf("call format wrong:\ngot:\n%s\nwant:\n%s", output, expected)
	}
}

func TestFmtShapedListWithCommas(t *testing.T) {
	input := `let p = [@point, 10, 20]`

	p := parser.New(input)
	program := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	output := ast.PrintWithComments(program, p.Comments())

	expected := `let p = [@point, 10, 20]
`

	if output != expected {
		t.Errorf("shaped list format wrong:\ngot:\n%s\nwant:\n%s", output, expected)
	}
}
