package fmt

import (
	"aiki/syntax"
	"sort"
	"strings"
)

// FormatSource formats source code, preserving comments and blank lines.
func FormatSource(grammar *syntax.Grammar, source string) (string, error) {
	node, comments, err := grammar.ParseSourceWithComments(source)
	if err != nil {
		return "", err
	}

	p := &printer{
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
	p.printProgram(node)
	return p.buf.String(), nil
}

type printer struct {
	buf         strings.Builder
	indent      int
	comments    map[int]string // line -> standalone comment text
	eolComments map[int]string // line -> EOL comment text
	lastLine    int
}

func (p *printer) write(s string) {
	p.buf.WriteString(s)
}

func (p *printer) writeIndent() {
	for i := 0; i < p.indent; i++ {
		p.buf.WriteString("\t")
	}
}

func (p *printer) newline() {
	p.buf.WriteString("\n")
}

// emitCommentsBefore emits standalone comments between lastLine and line.
func (p *printer) emitCommentsBefore(line int) {
	var lines []int
	for commentLine := range p.comments {
		if commentLine < line && commentLine > p.lastLine {
			lines = append(lines, commentLine)
		}
	}
	sort.Ints(lines)
	for _, cl := range lines {
		p.writeIndent()
		p.write(p.comments[cl])
		p.newline()
		delete(p.comments, cl)
	}
}

// emitEOLComment emits an end-of-line comment if present.
func (p *printer) emitEOLComment(line int) {
	if text, ok := p.eolComments[line]; ok {
		p.write("  ")
		p.write(text)
		delete(p.eolComments, line)
	}
}

// emitTrailingComments emits any comments after all statements.
func (p *printer) emitTrailingComments() {
	var lines []int
	for line := range p.comments {
		lines = append(lines, line)
	}
	sort.Ints(lines)
	for _, line := range lines {
		p.newline()
		p.write(p.comments[line])
		p.newline()
	}
}

// nodeStartLine returns the line number of a node's first token.
func nodeStartLine(node *syntax.Node) int {
	if node.Line > 0 {
		return node.Line
	}
	for _, child := range node.Children {
		line := nodeStartLine(child)
		if line > 0 {
			return line
		}
	}
	return 0
}

func (p *printer) printProgram(node *syntax.Node) {
	prevLine := 0
	for i, child := range node.Children {
		line := nodeStartLine(child)

		// Preserve blank lines between statements
		if i > 0 && line > prevLine+1 {
			p.newline()
		}

		p.emitCommentsBefore(line)
		p.printStatement(child)
		p.lastLine = line
		prevLine = line
	}
	p.emitTrailingComments()

	// Ensure trailing newline
	s := p.buf.String()
	if len(s) > 0 && s[len(s)-1] != '\n' {
		p.newline()
	}
}

func (p *printer) printStatement(node *syntax.Node) {
	if len(node.Children) == 0 {
		return
	}
	child := node.Children[0]
	line := nodeStartLine(node)

	switch child.Type {
	case "if_stmt":
		p.printIf(child)
	case "while_stmt":
		p.printWhile(child)
	case "match_stmt":
		p.printMatch(child)
	default:
		// Single-line statements: let, assign, return, expr, export, import
		p.printNode(child)
		p.emitEOLComment(line)
		p.newline()
	}
}

func (p *printer) printNode(node *syntax.Node) {
	switch node.Type {
	case "program":
		p.printProgram(node)
	case "statement":
		p.printStatement(node)
	case "let_stmt":
		p.printLet(node)
	case "assign_stmt":
		p.printAssign(node)
	case "if_stmt":
		p.printIf(node)
	case "while_stmt":
		p.printWhile(node)
	case "match_stmt":
		p.printMatch(node)
	case "return_stmt":
		p.printReturn(node)
	case "expr_stmt":
		p.printExprStmt(node)
	case "block":
		p.printBlock(node)
	case "expr":
		p.printExpr(node)
	case "pipe_expr":
		p.printPipeExpr(node)
	case "infix_expr":
		p.printInfix(node)
	case "unary_expr":
		p.printUnary(node)
	case "postfix_expr":
		p.printPostfix(node)
	case "primary":
		p.printPrimary(node)
	case "func_literal":
		p.printFuncLiteral(node)
	case "list_literal":
		p.printList(node)
	case "call":
		p.printCall(node)
	case "index":
		p.printIndex(node)
	case "access":
		p.printAccess(node)
	case "params":
		p.printParams(node)
	case "pattern":
		p.printPattern(node)
	case "NUMBER", "STRING", "RUNE", "SYMBOL", "SHAPE", "NAME":
		p.write(node.Value)
	case "TERMINAL":
		p.write(node.Value)
	case "BINOP":
		p.printBinop(node)
	default:
		for _, child := range node.Children {
			p.printNode(child)
		}
	}
}

func (p *printer) printLet(node *syntax.Node) {
	p.writeIndent()
	p.write("let ")

	isShape := false
	for _, child := range node.Children {
		if child.Type == "SHAPE" {
			isShape = true
			break
		}
	}

	if isShape {
		p.printShapeDef(node)
	} else {
		p.printLetBinding(node)
	}
}

func (p *printer) printShapeDef(node *syntax.Node) {
	for _, child := range node.Children {
		if child.Type == "SHAPE" {
			p.write(child.Value)
			break
		}
	}

	p.write(" [")
	first := true
	for _, child := range node.Children {
		if child.Type == "field" {
			for _, f := range child.Children {
				if f.Type == "NAME" || f.Type == "SHAPE" {
					if !first {
						p.write(", ")
					}
					first = false
					p.write(f.Value)
				}
			}
		}
	}
	p.write("]")
}

func (p *printer) printLetBinding(node *syntax.Node) {
	var name string
	var valueNode *syntax.Node

	foundEquals := false
	for _, child := range node.Children {
		if child.Type == "NAME" && !foundEquals {
			name = child.Value
		}
		if child.Type == "TERMINAL" && child.Value == "=" {
			foundEquals = true
			continue
		}
		if foundEquals {
			valueNode = child
			break
		}
	}

	p.write(name)
	p.write(" = ")
	if valueNode != nil {
		p.printNode(valueNode)
	}
}

func (p *printer) printAssign(node *syntax.Node) {
	p.writeIndent()
	foundEquals := false
	for _, child := range node.Children {
		if child.Type == "TERMINAL" && child.Value == "=" {
			foundEquals = true
			p.write(" = ")
			continue
		}
		if !foundEquals {
			p.printNode(child)
		} else {
			p.printNode(child)
		}
	}
}

func (p *printer) printIf(node *syntax.Node) {
	p.writeIndent()
	p.write("if ")

	children := node.Children
	i := 0

	// Skip "if" terminal
	if i < len(children) && children[i].Type == "TERMINAL" && children[i].Value == "if" {
		i++
	}

	// Condition
	if i < len(children) && children[i].Type != "block" && children[i].Type != "TERMINAL" {
		p.printNode(children[i])
		i++
	}

	// Consequence block
	p.write(" ")
	if i < len(children) && children[i].Type == "block" {
		p.printBlock(children[i])
		i++
	}

	// else
	if i < len(children) && children[i].Type == "TERMINAL" && children[i].Value == "else" {
		p.write(" else ")
		i++
		if i < len(children) && children[i].Type == "block" {
			p.printBlock(children[i])
			i++
		}
	}

	line := nodeStartLine(node)
	p.emitEOLComment(line)
	p.newline()
}

func (p *printer) printWhile(node *syntax.Node) {
	p.writeIndent()
	p.write("while ")

	children := node.Children
	i := 0

	// Skip "while" terminal
	if i < len(children) && children[i].Type == "TERMINAL" && children[i].Value == "while" {
		i++
	}

	// Condition
	if i < len(children) && children[i].Type != "block" && children[i].Type != "TERMINAL" {
		p.printNode(children[i])
		i++
	}

	// Body
	p.write(" ")
	if i < len(children) && children[i].Type == "block" {
		p.printBlock(children[i])
	}

	line := nodeStartLine(node)
	p.emitEOLComment(line)
	p.newline()
}

func (p *printer) printMatch(node *syntax.Node) {
	p.writeIndent()
	p.write("match ")

	children := node.Children
	i := 0

	// Skip "match" terminal
	if i < len(children) && children[i].Type == "TERMINAL" && children[i].Value == "match" {
		i++
	}

	// Value expression - skip terminals and patterns
	if i < len(children) && children[i].Type != "TERMINAL" && children[i].Type != "pattern" && children[i].Type != "block" {
		p.printNode(children[i])
		i++
	}

	p.write(" {\n")
	p.indent++

	// Process arms: pattern followed by block
	// Grammar: match_stmt = "match" expr "{" { pattern block } "}"
	for i < len(children) {
		child := children[i]

		// Skip terminals like "{" and "}"
		if child.Type == "TERMINAL" {
			i++
			continue
		}

		// Found a pattern - next should be block
		if child.Type == "pattern" {
			p.writeIndent()
			p.printPattern(child)
			i++

			// Look for the block
			if i < len(children) && children[i].Type == "block" {
				p.write(" ")
				p.printBlock(children[i])
				p.newline()
				i++
			}
			continue
		}

		i++
	}

	p.indent--
	p.writeIndent()
	p.write("}")

	line := nodeStartLine(node)
	p.emitEOLComment(line)
	p.newline()
}

func (p *printer) printPattern(node *syntax.Node) {
	for _, child := range node.Children {
		switch child.Type {
		case "TERMINAL":
			if child.Value == "_" {
				p.write("_")
			} else if child.Value == "[" {
				p.write("[")
			} else if child.Value == "]" {
				p.write("]")
			} else if child.Value == "," {
				p.write(", ")
			} else {
				p.write(child.Value)
			}
		case "NUMBER", "STRING", "SYMBOL", "SHAPE":
			p.write(child.Value)
		case "NAME":
			p.write(child.Value)
		case "pattern":
			p.printPattern(child)
		case "literal":
			// Handle nested literal in pattern
			p.printPatternLiteral(child)
		}
	}
}

func (p *printer) printPatternLiteral(node *syntax.Node) {
	for _, child := range node.Children {
		switch child.Type {
		case "NUMBER", "STRING", "SYMBOL", "SHAPE":
			p.write(child.Value)
		case "TERMINAL":
			p.write(child.Value)
		}
	}
}

func (p *printer) printReturn(node *syntax.Node) {
	p.writeIndent()
	p.write("return ")
	for _, child := range node.Children {
		if child.Type != "TERMINAL" {
			p.printNode(child)
		}
	}
}

func (p *printer) printExprStmt(node *syntax.Node) {
	p.writeIndent()
	for _, child := range node.Children {
		p.printNode(child)
	}
}

func (p *printer) printBlock(node *syntax.Node) {
	p.write("{\n")
	p.indent++

	for _, child := range node.Children {
		if child.Type == "statement" {
			line := nodeStartLine(child)
			p.emitCommentsBefore(line)
			p.printStatement(child)
			p.lastLine = line
		}
	}

	p.indent--
	p.writeIndent()
	p.write("}")
}

func (p *printer) printExpr(node *syntax.Node) {
	// expr contains pipe_expr which contains the actual |> terminals
	// Just delegate to children - printNode will route pipe_expr correctly
	for _, child := range node.Children {
		p.printNode(child)
	}
}

func (p *printer) printPipeExpr(node *syntax.Node) {
	// Collect all parts: expressions and pipe operators
	// Output: expr |>\n\tindent expr |>\n\tindent expr
	children := node.Children

	// Count pipe segments to decide if we should break
	pipeCount := 0
	for _, child := range children {
		if child.Type == "TERMINAL" && child.Value == "|>" {
			pipeCount++
		}
	}

	// Short pipes (1 pipe) stay on one line
	if pipeCount <= 1 {
		for _, child := range children {
			if child.Type == "TERMINAL" && child.Value == "|>" {
				p.write(" |> ")
			} else if child.Type != "TERMINAL" {
				p.printNode(child)
			}
		}
		return
	}

	// Long pipes break with |> at end of line
	// Pattern: expr |>\n\texpr |>\n\texpr
	for i, child := range children {
		if child.Type == "TERMINAL" && child.Value == "|>" {
			p.write(" |>")
			p.newline()
			p.writeIndent()
		} else if child.Type != "TERMINAL" {
			p.printNode(child)
		}
		_ = i
	}
}

func (p *printer) printInfix(node *syntax.Node) {
	first := true
	for _, child := range node.Children {
		if child.Type == "BINOP" {
			p.write(" ")
			p.printBinop(child)
			p.write(" ")
		} else if child.Type == "TERMINAL" {
			if child.Value == "and" || child.Value == "or" ||
				child.Value == "==" || child.Value == "!=" ||
				child.Value == "<" || child.Value == ">" ||
				child.Value == "<=" || child.Value == ">=" ||
				child.Value == "+" || child.Value == "-" ||
				child.Value == "*" || child.Value == "/" || child.Value == "%" {
				p.write(" ")
				p.write(child.Value)
				p.write(" ")
			}
		} else {
			if !first {
				// already handled spacing
			}
			p.printNode(child)
			first = false
		}
	}
}

func (p *printer) printBinop(node *syntax.Node) {
	for _, child := range node.Children {
		if child.Type == "TERMINAL" {
			p.write(child.Value)
		}
	}
}

func (p *printer) printUnary(node *syntax.Node) {
	for _, child := range node.Children {
		if child.Type == "TERMINAL" {
			if child.Value == "not" {
				p.write("not ")
			} else if child.Value == "-" {
				p.write("-")
			}
		} else {
			p.printNode(child)
		}
	}
}

func (p *printer) printPostfix(node *syntax.Node) {
	for _, child := range node.Children {
		p.printNode(child)
	}
}

func (p *printer) printPrimary(node *syntax.Node) {
	for _, child := range node.Children {
		switch child.Type {
		case "TERMINAL":
			if child.Value == "(" {
				p.write("(")
			} else if child.Value == ")" {
				p.write(")")
			} else if child.Value == "true" || child.Value == "false" {
				p.write(child.Value)
			}
		default:
			p.printNode(child)
		}
	}
}

func (p *printer) printFuncLiteral(node *syntax.Node) {
	p.write("(")
	for _, child := range node.Children {
		if child.Type == "params" {
			p.printParams(child)
		}
	}
	p.write(") ")
	for _, child := range node.Children {
		if child.Type == "block" {
			p.printBlock(child)
		}
	}
}

func (p *printer) printParams(node *syntax.Node) {
	first := true
	for _, child := range node.Children {
		switch child.Type {
		case "param_list":
			for _, param := range child.Children {
				if param.Type == "NAME" {
					if !first {
						p.write(", ")
					}
					first = false
					p.write(param.Value)
				}
			}
		case "rest_param":
			if !first {
				p.write(", ")
			}
			first = false
			p.write("...")
			for _, param := range child.Children {
				if param.Type == "NAME" {
					p.write(param.Value)
				}
			}
		case "NAME":
			if !first {
				p.write(", ")
			}
			first = false
			p.write(child.Value)
		}
	}
}

func (p *printer) printList(node *syntax.Node) {
	p.write("[")
	first := true
	for _, child := range node.Children {
		if child.Type == "TERMINAL" {
			continue
		}
		if !first {
			p.write(", ")
		}
		first = false
		p.printNode(child)
	}
	p.write("]")
}

func (p *printer) printCall(node *syntax.Node) {
	p.write("(")
	first := true
	for _, child := range node.Children {
		if child.Type == "TERMINAL" {
			continue
		}
		if !first {
			p.write(", ")
		}
		first = false
		p.printNode(child)
	}
	p.write(")")
}

func (p *printer) printIndex(node *syntax.Node) {
	p.write("[")
	for _, child := range node.Children {
		if child.Type != "TERMINAL" {
			p.printNode(child)
		}
	}
	p.write("]")
}

func (p *printer) printAccess(node *syntax.Node) {
	p.write(".")
	for _, child := range node.Children {
		if child.Type == "NAME" {
			p.write(child.Value)
		}
	}
}
