package tests

import (
	"fmt"
	"testing"

	"aiki/lang/ast"
	"aiki/lang/parser"
)

func TestParserLetStatements(t *testing.T) {
	tests := []struct {
		input string
		name  string
	}{
		{"let x = 5", "x"},
		{"let foo = 10", "foo"},
		{"let bar_baz = 100", "bar_baz"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			p := parser.New(tt.input)
			program := p.Parse()
			checkParserErrors(t, p)

			if len(program.Statements) != 1 {
				t.Fatalf("statements: got %d, want 1", len(program.Statements))
			}

			stmt, ok := program.Statements[0].(*ast.LetStatement)
			if !ok {
				t.Fatalf("not LetStatement: got %T", program.Statements[0])
			}

			if stmt.Name.Value != tt.name {
				t.Errorf("name: got %q, want %q", stmt.Name.Value, tt.name)
			}
		})
	}
}

func TestParserShapeStatements(t *testing.T) {
	tests := []struct {
		input  string
		name   string
		fields []string
		embeds []string
	}{
		{"let @point [x, y]", "point", []string{"x", "y"}, nil},
		{"let @user [name, email, age]", "user", []string{"name", "email", "age"}, nil},
		{"let @cat [@pet, color]", "cat", []string{"color"}, []string{"pet"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			p := parser.New(tt.input)
			program := p.Parse()
			checkParserErrors(t, p)

			stmt, ok := program.Statements[0].(*ast.ShapeStatement)
			if !ok {
				t.Fatalf("not ShapeStatement: got %T", program.Statements[0])
			}

			if stmt.Name != tt.name {
				t.Errorf("name: got %q, want %q", stmt.Name, tt.name)
			}

			if len(stmt.Fields) != len(tt.fields) {
				t.Fatalf("fields count: got %d, want %d", len(stmt.Fields), len(tt.fields))
			}
			for i, f := range tt.fields {
				if stmt.Fields[i] != f {
					t.Errorf("field[%d]: got %q, want %q", i, stmt.Fields[i], f)
				}
			}
		})
	}
}

func TestParserInfixExpressions(t *testing.T) {
	tests := []struct {
		input string
		op    string
	}{
		{"1 + 2", "+"},
		{"5 - 3", "-"},
		{"2 * 4", "*"},
		{"8 / 2", "/"},
		{"5 % 3", "%"},
		{"1 < 2", "<"},
		{"2 > 1", ">"},
		{"1 == 1", "=="},
		{"1 != 2", "!="},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			p := parser.New(tt.input)
			program := p.Parse()
			checkParserErrors(t, p)

			stmt := program.Statements[0].(*ast.ExpressionStatement)
			infix := stmt.Expression.(*ast.InfixExpression)

			if infix.Operator != tt.op {
				t.Errorf("operator: got %q, want %q", infix.Operator, tt.op)
			}
		})
	}
}

func TestParserLeftToRight(t *testing.T) {
	// 1 + 2 * 3 should parse as (1 + 2) * 3, not 1 + (2 * 3)
	tests := []struct {
		input       string
		description string
	}{
		{"1 + 2 * 3", "addition then multiplication"},
		{"2 * 3 + 1", "multiplication then addition"},
		{"1 + 2 + 3", "chained addition"},
		{"1 - 2 - 3", "chained subtraction"},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			p := parser.New(tt.input)
			program := p.Parse()
			checkParserErrors(t, p)

			stmt := program.Statements[0].(*ast.ExpressionStatement)
			outer := stmt.Expression.(*ast.InfixExpression)

			_, leftIsInfix := outer.Left.(*ast.InfixExpression)
			if !leftIsInfix {
				t.Errorf("left side should be InfixExpression for left-to-right parsing, got %T", outer.Left)
			}
		})
	}
}

func TestParserGroupedExpressions(t *testing.T) {
	tests := []struct {
		input       string
		description string
	}{
		{"(1 + 2)", "simple group"},
		{"(1 + 2) * 3", "group then op"},
		{"1 + (2 * 3)", "op then group"},
		{"((1 + 2))", "nested groups"},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			p := parser.New(tt.input)
			program := p.Parse()
			checkParserErrors(t, p)

			if len(program.Statements) != 1 {
				t.Fatalf("statements: got %d, want 1", len(program.Statements))
			}
		})
	}
}

func TestParserFunctionLiterals(t *testing.T) {
	tests := []struct {
		input  string
		params []string
	}{
		{"(n) { return n }", []string{"n"}},
		{"(a, b) { return a + b }", []string{"a", "b"}},
		{"(x, y, z) { return x }", []string{"x", "y", "z"}},
		{"() { return 42 }", []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			p := parser.New(tt.input)
			program := p.Parse()
			checkParserErrors(t, p)

			stmt := program.Statements[0].(*ast.ExpressionStatement)
			fn := stmt.Expression.(*ast.FunctionLiteral)

			if len(fn.Parameters) != len(tt.params) {
				t.Fatalf("params count: got %d, want %d", len(fn.Parameters), len(tt.params))
			}

			for i, p := range tt.params {
				if fn.Parameters[i] != p {
					t.Errorf("param[%d]: got %q, want %q", i, fn.Parameters[i], p)
				}
			}
		})
	}
}

func TestParserCallExpressions(t *testing.T) {
	tests := []struct {
		input    string
		argCount int
	}{
		{"f()", 0},
		{"f(1)", 1},
		{"f(1, 2, 3)", 3},
		{"add(1, 2)", 2},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			p := parser.New(tt.input)
			program := p.Parse()
			checkParserErrors(t, p)

			stmt := program.Statements[0].(*ast.ExpressionStatement)
			call := stmt.Expression.(*ast.CallExpression)

			if len(call.Arguments) != tt.argCount {
				t.Errorf("arg count: got %d, want %d", len(call.Arguments), tt.argCount)
			}
		})
	}
}

func TestParserAccessExpressions(t *testing.T) {
	tests := []struct {
		input string
		key   string
	}{
		{"x.0", "0"},
		{"list.5", "5"},
		{"point.x", "x"},
		{"user.name", "name"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			p := parser.New(tt.input)
			program := p.Parse()
			checkParserErrors(t, p)

			stmt := program.Statements[0].(*ast.ExpressionStatement)
			access := stmt.Expression.(*ast.AccessExpression)

			if access.Key != tt.key {
				t.Errorf("key: got %q, want %q", access.Key, tt.key)
			}
		})
	}
}

func TestParserListLiterals(t *testing.T) {
	tests := []struct {
		input   string
		count   int
		isShape bool
		shape   string
	}{
		{"[]", 0, false, ""},
		{"[1]", 1, false, ""},
		{"[1, 2, 3]", 3, false, ""},
		{"[@point, 10, 20]", 2, true, "point"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			p := parser.New(tt.input)
			program := p.Parse()
			checkParserErrors(t, p)

			stmt := program.Statements[0].(*ast.ExpressionStatement)

			if tt.isShape {
				list := stmt.Expression.(*ast.ShapedListLiteral)
				if list.Shape != tt.shape {
					t.Errorf("shape: got %q, want %q", list.Shape, tt.shape)
				}
				if len(list.Elements) != tt.count {
					t.Errorf("count: got %d, want %d", len(list.Elements), tt.count)
				}
			} else {
				list := stmt.Expression.(*ast.ListLiteral)
				if len(list.Elements) != tt.count {
					t.Errorf("count: got %d, want %d", len(list.Elements), tt.count)
				}
			}
		})
	}
}

func TestParserPipeExpressions(t *testing.T) {
	input := "x |> f() |> g()"
	p := parser.New(input)
	program := p.Parse()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	outer := stmt.Expression.(*ast.PipeExpression)
	inner := outer.Left.(*ast.PipeExpression)
	_ = inner.Left.(*ast.Identifier)
}

func TestParserIfStatements(t *testing.T) {
	tests := []struct {
		input   string
		hasElse bool
	}{
		{"if true { return 1 }", false},
		{"if x < 10 { return x } else { return 10 }", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			p := parser.New(tt.input)
			program := p.Parse()
			checkParserErrors(t, p)

			stmt := program.Statements[0].(*ast.IfStatement)

			if stmt.Consequence == nil {
				t.Error("consequence is nil")
			}

			if tt.hasElse && stmt.Alternative == nil {
				t.Error("expected alternative, got nil")
			}
			if !tt.hasElse && stmt.Alternative != nil {
				t.Error("unexpected alternative")
			}
		})
	}
}

func TestParserWhileStatements(t *testing.T) {
	input := "while x > 0 { x = x - 1 }"
	p := parser.New(input)
	program := p.Parse()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.WhileStatement)

	if stmt.Condition == nil {
		t.Error("condition is nil")
	}
	if stmt.Body == nil {
		t.Error("body is nil")
	}
}

func TestParserMatchStatements(t *testing.T) {
	input := `match x {
		[@ok, val] { return val }
		[@error, msg] { return msg }
		_ { return "unknown" }
	}`

	p := parser.New(input)
	program := p.Parse()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.MatchStatement)

	if len(stmt.Arms) != 3 {
		t.Fatalf("arms: got %d, want 3", len(stmt.Arms))
	}
}

func TestParserCallWithGroupedArg(t *testing.T) {
	// This was the ambiguous case that commas fix
	input := "average(guess, (x / guess))"
	p := parser.New(input)
	program := p.Parse()
	checkParserErrors(t, p)

	stmt := program.Statements[0].(*ast.ExpressionStatement)
	call := stmt.Expression.(*ast.CallExpression)

	if len(call.Arguments) != 2 {
		t.Fatalf("arg count: got %d, want 2", len(call.Arguments))
	}
}

func checkParserErrors(t *testing.T, p *parser.Parser) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}
	t.Errorf("parser has %d errors:", len(errors))
	for _, msg := range errors {
		t.Errorf("  %s", msg)
	}
	t.FailNow()
}

// Silence unused import error
var _ = fmt.Sprintf
