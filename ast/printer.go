package ast

import (
	"fmt"
	"strings"

	"aiki/lexer"
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
func Print(program *Program) string {
	return PrintWithComments(program, nil)
}

// PrintWithComments formats a program with preserved comments.
func PrintWithComments(program *Program, comments []lexer.Comment) string {
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
		p.buf.WriteString("\t")
	}
}

func (p *Printer) newline() {
	p.buf.WriteString("\n")
}

func (p *Printer) printProgram(program *Program) {
	prevWasFunc := false
	for i, stmt := range program.Statements {
		// Get the line number of this statement
		line := p.getStatementLine(stmt)

		// Emit any standalone comments that come before this statement
		p.emitCommentsBefore(line)

		// Blank line before function definitions (except first)
		isFunc := isFunctionDef(stmt)
		if i > 0 && (isFunc || prevWasFunc) {
			p.newline()
		}
		p.printStatementWithEOL(stmt, line)
		p.lastLine = line
		prevWasFunc = isFunc
	}

	// Emit any trailing comments
	p.emitTrailingComments()
}

func (p *Printer) printStatementWithEOL(stmt Statement, startLine int) {
	switch s := stmt.(type) {
	case *LetStatement:
		p.printLetStatement(s)
		p.emitEOLComment(startLine)
		p.newline()
	case *ShapeStatement:
		p.printShapeStatement(s)
		p.emitEOLComment(startLine)
		p.newline()
	case *AssignStatement:
		p.printAssignStatement(s)
		p.emitEOLComment(startLine)
		p.newline()
	case *ReturnStatement:
		p.printReturnStatement(s)
		p.emitEOLComment(startLine)
		p.newline()
	case *ExpressionStatement:
		p.printExpressionStatement(s)
		p.emitEOLComment(startLine)
		p.newline()
	case *IfStatement:
		p.printIfStatement(s)
	case *WhileStatement:
		p.printWhileStatement(s)
	case *MatchStatement:
		p.printMatchStatement(s)
	case *ExportStatement:
		p.printExportStatement(s)
		p.emitEOLComment(startLine)
		p.newline()
	case *ImportStatement:
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

func (p *Printer) getStatementLine(stmt Statement) int {
	switch s := stmt.(type) {
	case *LetStatement:
		return s.Token.Line
	case *ShapeStatement:
		return s.Token.Line
	case *AssignStatement:
		return s.Token.Line
	case *ReturnStatement:
		return s.Token.Line
	case *ExpressionStatement:
		return s.Token.Line
	case *IfStatement:
		return s.Token.Line
	case *WhileStatement:
		return s.Token.Line
	case *MatchStatement:
		return s.Token.Line
	case *ExportStatement:
		return s.Token.Line
	case *ImportStatement:
		return s.Token.Line
	}
	return 0
}

func (p *Printer) emitCommentsBefore(line int) {
	for commentLine, text := range p.comments {
		if commentLine < line && commentLine > p.lastLine {
			p.writeIndent()
			p.write(text)
			p.newline()
			delete(p.comments, commentLine)
		}
	}
}

func (p *Printer) emitTrailingComments() {
	// Collect remaining comments in order
	var lines []int
	for line := range p.comments {
		lines = append(lines, line)
	}
	// Sort
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

func isFunctionDef(stmt Statement) bool {
	if let, ok := stmt.(*LetStatement); ok {
		_, isFunc := let.Value.(*FunctionLiteral)
		return isFunc
	}
	return false
}

func (p *Printer) printStatement(stmt Statement) {
	switch s := stmt.(type) {
	case *LetStatement:
		p.printLetStatement(s)
	case *ShapeStatement:
		p.printShapeStatement(s)
	case *AssignStatement:
		p.printAssignStatement(s)
	case *ReturnStatement:
		p.printReturnStatement(s)
	case *ExpressionStatement:
		p.printExpressionStatement(s)
	case *IfStatement:
		p.printIfStatement(s)
	case *WhileStatement:
		p.printWhileStatement(s)
	case *MatchStatement:
		p.printMatchStatement(s)
	case *ExportStatement:
		p.printExportStatement(s)
	case *ImportStatement:
		p.printImportStatement(s)
	}
}

func (p *Printer) printLetStatement(stmt *LetStatement) {
	p.writeIndent()
	p.write("let ")
	p.write(stmt.Name.Value)
	p.write(" = ")
	p.printExpression(stmt.Value)
}

func (p *Printer) printShapeStatement(stmt *ShapeStatement) {
	p.writeIndent()
	p.write("let @")
	p.write(stmt.Name)
	p.write(" [")
	// Embeds first
	for i, embed := range stmt.Embeds {
		if i > 0 {
			p.write(" ")
		}
		p.write("@")
		p.write(embed)
	}
	// Then fields
	for i, field := range stmt.Fields {
		if i > 0 || len(stmt.Embeds) > 0 {
			p.write(" ")
		}
		p.write(field)
	}
	p.write("]")
}

func (p *Printer) printAssignStatement(stmt *AssignStatement) {
	p.writeIndent()
	p.write(stmt.Name.Value)
	p.write(" = ")
	p.printExpression(stmt.Value)
}

func (p *Printer) printReturnStatement(stmt *ReturnStatement) {
	p.writeIndent()
	p.write("return ")
	p.printExpression(stmt.Value)
}

func (p *Printer) printExpressionStatement(stmt *ExpressionStatement) {
	p.writeIndent()
	p.printExpression(stmt.Expression)
}

func (p *Printer) printIfStatement(stmt *IfStatement) {
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

func (p *Printer) printWhileStatement(stmt *WhileStatement) {
	p.writeIndent()
	p.write("while ")
	p.printExpression(stmt.Condition)
	p.write(" ")
	p.printBlock(stmt.Body)
	p.newline()
}

func (p *Printer) printMatchStatement(stmt *MatchStatement) {
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

func (p *Printer) printMatchArm(arm *MatchArm) {
	p.writeIndent()
	p.printPattern(arm.Pattern)
	p.write(" ")
	p.printBlock(arm.Body)
	p.newline()
}

func (p *Printer) printPattern(pat Pattern) {
	switch pt := pat.(type) {
	case *WildcardPattern:
		p.write("_")
	case *NamePattern:
		p.write(pt.Name)
	case *LiteralPattern:
		p.printExpression(pt.Value)
	case *ListPattern:
		p.write("[")
		for i, elem := range pt.Elements {
			if i > 0 {
				p.write(" ")
			}
			p.printPattern(elem)
		}
		p.write("]")
	case *ShapedListPattern:
		p.write("[@")
		p.write(pt.Shape)
		for _, elem := range pt.Elements {
			p.write(" ")
			p.printPattern(elem)
		}
		p.write("]")
	}
}

func (p *Printer) printExportStatement(stmt *ExportStatement) {
	p.writeIndent()
	p.write("export [")
	for i, name := range stmt.Names {
		if i > 0 {
			p.write(" ")
		}
		p.write(name)
	}
	p.write("]")
}

func (p *Printer) printImportStatement(stmt *ImportStatement) {
	p.writeIndent()
	p.write("from ")
	p.write(stmt.Module)
	p.write(" use [")
	for i, name := range stmt.Names {
		if i > 0 {
			p.write(" ")
		}
		p.write(name)
	}
	p.write("]")
}

func (p *Printer) printBlock(block *BlockStatement) {
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

func (p *Printer) printExpression(expr Expression) {
	switch e := expr.(type) {
	case *Identifier:
		p.write(e.Value)

	case *NumberLiteral:
		p.write(e.Value)

	case *BooleanLiteral:
		if e.Value {
			p.write("true")
		} else {
			p.write("false")
		}

	case *StringLiteral:
		p.write("\"")
		p.write(escapeString(e.Value))
		p.write("\"")

	case *RuneLiteral:
		p.write("'")
		p.write(escapeRune(e.Value))
		p.write("'")

	case *SymbolLiteral:
		p.write(":")
		p.write(e.Value)

	case *ListLiteral:
		p.write("[")
		for i, elem := range e.Elements {
			if i > 0 {
				p.write(" ")
			}
			p.printExpression(elem)
		}
		p.write("]")

	case *ShapedListLiteral:
		p.write("[@")
		p.write(e.Shape)
		for _, elem := range e.Elements {
			p.write(" ")
			p.printExpression(elem)
		}
		p.write("]")

	case *FunctionLiteral:
		p.write("(")
		for i, param := range e.Parameters {
			if i > 0 {
				p.write(" ")
			}
			p.write(param)
		}
		p.write(") ")
		p.printBlock(e.Body)

	case *CallExpression:
		p.printExpression(e.Function)
		p.write("(")
		for i, arg := range e.Arguments {
			if i > 0 {
				p.write(" ")
			}
			p.printExpression(arg)
		}
		p.write(")")

	case *AccessExpression:
		p.printExpression(e.Left)
		p.write(".")
		p.write(e.Key)

	case *InfixExpression:
		p.printExpression(e.Left)
		p.write(" ")
		p.write(e.Operator)
		p.write(" ")
		p.printExpression(e.Right)

	case *PrefixExpression:
		p.write(e.Operator)
		if e.Operator == "not" {
			p.write(" ")
		}
		p.printExpression(e.Right)

	case *PipeExpression:
		p.printPipeExpression(e)
	}
}

func (p *Printer) printPipeExpression(expr *PipeExpression) {
	// Collect all pipe stages
	var stages []Expression
	var current Expression = expr
	for {
		if pipe, ok := current.(*PipeExpression); ok {
			stages = append([]Expression{pipe.Right}, stages...)
			current = pipe.Left
		} else {
			stages = append([]Expression{current}, stages...)
			break
		}
	}

	// Decide: single line or multi-line
	// For now, always single line. Can add length check later.
	for i, stage := range stages {
		if i > 0 {
			p.write(" |> ")
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
		return fmt.Sprintf("%c", r)
	}
}
