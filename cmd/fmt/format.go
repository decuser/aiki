package fmt

import (
	gofmt "fmt"
	"strings"

	"aiki/lang/ast"
	"aiki/lang/lexer"
)

// Printer formats an AST back to source code.
type Printer struct {
	buf         strings.Builder
	indent      int
	comments    map[int]string // line -> standalone comment text
	eolComments map[int]string // line -> EOL comment text
	lastLine    int
}

// Print formats a program to canonical source code.
func Print(program *ast.Program) string {
	return PrintWithComments(program, nil)
}

// PrintWithComments formats a program with preserved comments.
func PrintWithComments(program *ast.Program, comments []lexer.Comment) string {
	p := &Printer{
		comments:    make(map[int]string),
		eolComments: make(map[int]string),
	}
	for _, c := range comments {
		if c.IsEOL {
			p.eolComments[c.Line] = c.Text
		} else {
			p.comments[c.Line] = c.Text
		}
	}
	p.printProgram(program)
	return p.buf.String()
}

func (p *Printer) write(s string) {
	p.buf.WriteString(s)
}

func (p *Printer) writeLine(s string) {
	p.writeIndent()
	p.buf.WriteString(s)
	p.buf.WriteString("\n")
}

func (p *Printer) writeIndent() {
	for i := 0; i < p.indent; i++ {
		p.buf.WriteString("    ")
	}
}

func (p *Printer) newline() {
	p.buf.WriteString("\n")
}

func (p *Printer) printProgram(program *ast.Program) {
	prevLine := 0
	for i, stmt := range program.Statements {
		line := p.getStatementLine(stmt)

		if i > 0 && line > prevLine+1 {
			p.newline()
		}

		p.emitCommentsBefore(line)
		p.printStatementWithEOL(stmt, line)
		p.lastLine = line
		prevLine = line
	}

	p.emitTrailingComments()
}

func (p *Printer) printStatementWithEOL(stmt ast.Statement, startLine int) {
	switch s := stmt.(type) {
	case *ast.LetStatement:
		p.printLetStatement(s)
		p.emitEOLComment(startLine)
		p.newline()
	case *ast.ShapeStatement:
		p.printShapeStatement(s)
		p.emitEOLComment(startLine)
		p.newline()
	case *ast.AssignStatement:
		p.printAssignStatement(s)
		p.emitEOLComment(startLine)
		p.newline()
	case *ast.ReturnStatement:
		p.printReturnStatement(s)
		p.emitEOLComment(startLine)
		p.newline()
	case *ast.ExpressionStatement:
		p.printExpressionStatement(s)
		p.emitEOLComment(startLine)
		p.newline()
	case *ast.IfStatement:
		p.printIfStatement(s)
	case *ast.WhileStatement:
		p.printWhileStatement(s)
	case *ast.MatchStatement:
		p.printMatchStatement(s)
	case *ast.ExportStatement:
		p.printExportStatement(s)
		p.emitEOLComment(startLine)
		p.newline()
	case *ast.ImportStatement:
		p.printImportStatement(s)
		p.emitEOLComment(startLine)
		p.newline()
	}
}

func (p *Printer) emitEOLComment(line int) {
	if text, ok := p.eolComments[line]; ok {
		p.write("  ")
		p.write(text)
		delete(p.eolComments, line)
	}
}

func (p *Printer) getStatementLine(stmt ast.Statement) int {
	switch s := stmt.(type) {
	case *ast.LetStatement:
		return s.Token.Line
	case *ast.ShapeStatement:
		return s.Token.Line
	case *ast.AssignStatement:
		return s.Token.Line
	case *ast.ReturnStatement:
		return s.Token.Line
	case *ast.ExpressionStatement:
		return s.Token.Line
	case *ast.IfStatement:
		return s.Token.Line
	case *ast.WhileStatement:
		return s.Token.Line
	case *ast.MatchStatement:
		return s.Token.Line
	case *ast.ExportStatement:
		return s.Token.Line
	case *ast.ImportStatement:
		return s.Token.Line
	}
	return 0
}

func (p *Printer) emitCommentsBefore(line int) {
	var lines []int
	for commentLine := range p.comments {
		if commentLine < line && commentLine > p.lastLine {
			lines = append(lines, commentLine)
		}
	}
	for i := 0; i < len(lines); i++ {
		for j := i + 1; j < len(lines); j++ {
			if lines[j] < lines[i] {
				lines[i], lines[j] = lines[j], lines[i]
			}
		}
	}
	for _, commentLine := range lines {
		p.writeIndent()
		p.write(p.comments[commentLine])
		p.newline()
		delete(p.comments, commentLine)
	}
}

func (p *Printer) emitTrailingComments() {
	var lines []int
	for line := range p.comments {
		lines = append(lines, line)
	}
	for i := 0; i < len(lines); i++ {
		for j := i + 1; j < len(lines); j++ {
			if lines[j] < lines[i] {
				lines[i], lines[j] = lines[j], lines[i]
			}
		}
	}
	for _, line := range lines {
		p.newline()
		p.write(p.comments[line])
		p.newline()
	}
}

func (p *Printer) printStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.LetStatement:
		p.printLetStatement(s)
	case *ast.ShapeStatement:
		p.printShapeStatement(s)
	case *ast.AssignStatement:
		p.printAssignStatement(s)
	case *ast.ReturnStatement:
		p.printReturnStatement(s)
	case *ast.ExpressionStatement:
		p.printExpressionStatement(s)
	case *ast.IfStatement:
		p.printIfStatement(s)
	case *ast.WhileStatement:
		p.printWhileStatement(s)
	case *ast.MatchStatement:
		p.printMatchStatement(s)
	case *ast.ExportStatement:
		p.printExportStatement(s)
	case *ast.ImportStatement:
		p.printImportStatement(s)
	}
}

func (p *Printer) printLetStatement(stmt *ast.LetStatement) {
	p.writeIndent()
	p.write("let ")
	p.write(stmt.Name.Value)
	p.write(" = ")
	p.printExpression(stmt.Value)
}

func (p *Printer) printShapeStatement(stmt *ast.ShapeStatement) {
	p.writeIndent()
	p.write("let @")
	p.write(stmt.Name)
	p.write(" [")
	for i, embed := range stmt.Embeds {
		if i > 0 {
			p.write(", ")
		}
		p.write("@")
		p.write(embed)
	}
	for i, field := range stmt.Fields {
		if i > 0 || len(stmt.Embeds) > 0 {
			p.write(", ")
		}
		p.write(field)
	}
	p.write("]")
}

func (p *Printer) printAssignStatement(stmt *ast.AssignStatement) {
	p.writeIndent()
	p.write(stmt.Name.Value)
	p.write(" = ")
	p.printExpression(stmt.Value)
}

func (p *Printer) printReturnStatement(stmt *ast.ReturnStatement) {
	p.writeIndent()
	p.write("return ")
	p.printExpression(stmt.Value)
}

func (p *Printer) printExpressionStatement(stmt *ast.ExpressionStatement) {
	p.writeIndent()
	p.printExpression(stmt.Expression)
}

func (p *Printer) printIfStatement(stmt *ast.IfStatement) {
	p.writeIndent()
	p.write("if ")
	p.printExpression(stmt.Condition)
	p.write(" ")
	p.printBlock(stmt.Consequence)
	if stmt.Alternative != nil {
		p.write(" else ")
		p.printBlock(stmt.Alternative)
	}
	p.newline()
}

func (p *Printer) printWhileStatement(stmt *ast.WhileStatement) {
	p.writeIndent()
	p.write("while ")
	p.printExpression(stmt.Condition)
	p.write(" ")
	p.printBlock(stmt.Body)
	p.newline()
}

func (p *Printer) printMatchStatement(stmt *ast.MatchStatement) {
	p.writeIndent()
	p.write("match ")
	p.printExpression(stmt.Value)
	p.write(" {\n")
	p.indent++
	for _, arm := range stmt.Arms {
		p.printMatchArm(arm)
	}
	p.indent--
	p.writeIndent()
	p.write("}")
	p.newline()
}

func (p *Printer) printMatchArm(arm *ast.MatchArm) {
	p.writeIndent()
	p.printPattern(arm.Pattern)
	p.write(" ")
	p.printBlock(arm.Body)
	p.newline()
}

func (p *Printer) printPattern(pat ast.Pattern) {
	switch pt := pat.(type) {
	case *ast.WildcardPattern:
		p.write("_")
	case *ast.NamePattern:
		p.write(pt.Name)
	case *ast.LiteralPattern:
		p.printExpression(pt.Value)
	case *ast.ListPattern:
		p.write("[")
		for i, elem := range pt.Elements {
			if i > 0 {
				p.write(", ")
			}
			p.printPattern(elem)
		}
		p.write("]")
	case *ast.ShapedListPattern:
		p.write("[@")
		p.write(pt.Shape)
		for _, elem := range pt.Elements {
			p.write(", ")
			p.printPattern(elem)
		}
		p.write("]")
	}
}

func (p *Printer) printExportStatement(stmt *ast.ExportStatement) {
	p.writeIndent()
	p.write("export [")
	for i, name := range stmt.Names {
		if i > 0 {
			p.write(", ")
		}
		p.write(name)
	}
	p.write("]")
}

func (p *Printer) printImportStatement(stmt *ast.ImportStatement) {
	p.writeIndent()
	p.write("from ")
	p.write(stmt.Module)
	p.write(" use [")
	for i, name := range stmt.Names {
		if i > 0 {
			p.write(", ")
		}
		p.write(name)
	}
	p.write("]")
}

func (p *Printer) printBlock(block *ast.BlockStatement) {
	p.write("{\n")
	p.indent++
	for _, stmt := range block.Statements {
		line := p.getStatementLine(stmt)
		p.emitCommentsBefore(line)
		p.printStatementWithEOL(stmt, line)
		p.lastLine = line
	}
	p.indent--
	p.writeIndent()
	p.write("}")
}

func (p *Printer) printExpression(expr ast.Expression) {
	switch e := expr.(type) {
	case *ast.Identifier:
		p.write(e.Value)

	case *ast.NumberLiteral:
		p.write(e.Value)

	case *ast.BooleanLiteral:
		if e.Value {
			p.write("true")
		} else {
			p.write("false")
		}

	case *ast.StringLiteral:
		p.write("\"")
		p.write(escapeString(e.Value))
		p.write("\"")

	case *ast.RuneLiteral:
		p.write("'")
		p.write(escapeRune(e.Value))
		p.write("'")

	case *ast.SymbolLiteral:
		p.write(":")
		p.write(e.Value)

	case *ast.ListLiteral:
		p.write("[")
		for i, elem := range e.Elements {
			if i > 0 {
				p.write(", ")
			}
			p.printExpression(elem)
		}
		p.write("]")

	case *ast.ShapedListLiteral:
		p.write("[@")
		p.write(e.Shape)
		for _, elem := range e.Elements {
			p.write(", ")
			p.printExpression(elem)
		}
		p.write("]")

	case *ast.FunctionLiteral:
		p.write("(")
		for i, param := range e.Parameters {
			if i > 0 {
				p.write(", ")
			}
			p.write(param)
		}
		if e.RestParam != "" {
			if len(e.Parameters) > 0 {
				p.write(", ")
			}
			p.write("...")
			p.write(e.RestParam)
		}
		p.write(") ")
		p.printBlock(e.Body)

	case *ast.CallExpression:
		p.printExpression(e.Function)
		p.write("(")
		for i, arg := range e.Arguments {
			if i > 0 {
				p.write(", ")
			}
			p.printExpression(arg)
		}
		p.write(")")

	case *ast.AccessExpression:
		p.printExpression(e.Left)
		p.write(".")
		p.write(e.Key)

	case *ast.IndexExpression:
		p.printExpression(e.Left)
		p.write("[")
		p.printExpression(e.Index)
		p.write("]")

	case *ast.InfixExpression:
		p.printExpression(e.Left)
		p.write(" ")
		p.write(e.Operator)
		p.write(" ")
		p.printExpression(e.Right)

	case *ast.PrefixExpression:
		p.write(e.Operator)
		if e.Operator == "not" {
			p.write(" ")
		}
		p.printExpression(e.Right)

	case *ast.PipeExpression:
		p.printPipeExpression(e)
	}
}

func (p *Printer) printPipeExpression(expr *ast.PipeExpression) {
	var stages []ast.Expression
	var current ast.Expression = expr
	for {
		if pipe, ok := current.(*ast.PipeExpression); ok {
			stages = append([]ast.Expression{pipe.Right}, stages...)
			current = pipe.Left
		} else {
			stages = append([]ast.Expression{current}, stages...)
			break
		}
	}

	for i, stage := range stages {
		if i > 0 {
			p.newline()
			p.writeIndent()
			p.write("|> ")
		}
		p.printExpression(stage)
	}
}

func escapeString(s string) string {
	var buf strings.Builder
	for _, r := range s {
		switch r {
		case '\n':
			buf.WriteString("\\n")
		case '\t':
			buf.WriteString("\\t")
		case '\r':
			buf.WriteString("\\r")
		case '\\':
			buf.WriteString("\\\\")
		case '"':
			buf.WriteString("\\\"")
		default:
			buf.WriteRune(r)
		}
	}
	return buf.String()
}

func escapeRune(r rune) string {
	switch r {
	case '\n':
		return "\\n"
	case '\t':
		return "\\t"
	case '\r':
		return "\\r"
	case '\\':
		return "\\\\"
	case '\'':
		return "\\'"
	default:
		return gofmt.Sprintf("%c", r)
	}
}
