package fmt

import (
	"strings"

	"aiki/ebnf"
)

// Format takes source code and returns formatted source code.
func FormatSource(grammar *ebnf.Grammar, source string) (string, error) {
	node, err := grammar.ParseSource(source)
	if err != nil {
		return "", err
	}
	
	p := &printer{
		indent: 0,
	}
	p.printNode(node)
	return p.buf.String(), nil
}

type printer struct {
	buf    strings.Builder
	indent int
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

func (p *printer) printNode(node *ebnf.Node) {
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
	case "export_stmt":
		p.printExport(node)
	case "import_stmt":
		p.printImport(node)
	case "expr_stmt":
		p.printExprStmt(node)
	case "block":
		p.printBlock(node)
	case "expr", "pipe_expr":
		p.printExpr(node)
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
		// Pass through children
		for _, child := range node.Children {
			p.printNode(child)
		}
	}
}

func (p *printer) printProgram(node *ebnf.Node) {
	for i, child := range node.Children {
		if i > 0 {
			p.newline()
		}
		p.printNode(child)
	}
	// Ensure trailing newline
	if len(node.Children) > 0 {
		p.newline()
	}
}

func (p *printer) printStatement(node *ebnf.Node) {
	p.writeIndent()
	for _, child := range node.Children {
		p.printNode(child)
	}
}

func (p *printer) printLet(node *ebnf.Node) {
	p.write("let ")
	
	// Check if it's a shape definition
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

func (p *printer) printShapeDef(node *ebnf.Node) {
	// let @shape [fields]
	for _, child := range node.Children {
		switch child.Type {
		case "SHAPE":
			p.write(child.Value)
		case "field":
			// handled below
		case "TERMINAL":
			if child.Value == "[" {
				p.write(" [")
			} else if child.Value == "]" {
				p.write("]")
			} else if child.Value == "," {
				p.write(", ")
			}
		}
	}
	
	// Collect and print fields
	first := true
	for _, child := range node.Children {
		if child.Type == "field" {
			for _, f := range child.Children {
				if !first {
					p.write(", ")
				}
				first = false
				p.write(f.Value)
			}
		}
	}
}

func (p *printer) printLetBinding(node *ebnf.Node) {
	// let name = expr
	for _, child := range node.Children {
		switch child.Type {
		case "NAME":
			p.write(child.Value)
		case "TERMINAL":
			if child.Value == "=" {
				p.write(" = ")
			}
		case "expr", "pipe_expr", "infix_expr", "unary_expr", "postfix_expr", "primary", "func_literal", "list_literal":
			p.printNode(child)
		}
	}
}

func (p *printer) printAssign(node *ebnf.Node) {
	for _, child := range node.Children {
		switch child.Type {
		case "NAME":
			p.write(child.Value)
		case "TERMINAL":
			if child.Value == "=" {
				p.write(" = ")
			}
		default:
			p.printNode(child)
		}
	}
}

func (p *printer) printIf(node *ebnf.Node) {
	p.write("if ")
	
	wroteCondition := false
	for _, child := range node.Children {
		switch child.Type {
		case "TERMINAL":
			if child.Value == "else" {
				p.write(" else ")
			}
			// skip "if"
		case "block":
			p.printNode(child)
		case "if_stmt":
			// else if
			p.printIfInner(child)
		default:
			if !wroteCondition {
				p.printNode(child)
				p.write(" ")
				wroteCondition = true
			}
		}
	}
}

func (p *printer) printIfInner(node *ebnf.Node) {
	// For else-if chains, don't add indent
	p.write("if ")
	
	wroteCondition := false
	for _, child := range node.Children {
		switch child.Type {
		case "TERMINAL":
			if child.Value == "else" {
				p.write(" else ")
			}
		case "block":
			p.printNode(child)
		case "if_stmt":
			p.printIfInner(child)
		default:
			if !wroteCondition {
				p.printNode(child)
				p.write(" ")
				wroteCondition = true
			}
		}
	}
}

func (p *printer) printWhile(node *ebnf.Node) {
	p.write("while ")
	
	wroteCondition := false
	for _, child := range node.Children {
		switch child.Type {
		case "TERMINAL":
			// skip "while"
		case "block":
			p.printNode(child)
		default:
			if !wroteCondition {
				p.printNode(child)
				p.write(" ")
				wroteCondition = true
			}
		}
	}
}

func (p *printer) printMatch(node *ebnf.Node) {
	p.write("match ")
	
	wroteValue := false
	var patterns []*ebnf.Node
	var blocks []*ebnf.Node
	
	for _, child := range node.Children {
		switch child.Type {
		case "TERMINAL":
			// skip match, {, }
		case "pattern":
			patterns = append(patterns, child)
		case "block":
			blocks = append(blocks, child)
		default:
			if !wroteValue {
				p.printNode(child)
				p.write(" ")
				wroteValue = true
			}
		}
	}
	
	p.write("{")
	p.newline()
	p.indent++
	
	for i := range patterns {
		p.writeIndent()
		p.printNode(patterns[i])
		p.write(" ")
		if i < len(blocks) {
			p.printNode(blocks[i])
		}
		p.newline()
	}
	
	p.indent--
	p.writeIndent()
	p.write("}")
}

func (p *printer) printPattern(node *ebnf.Node) {
	for _, child := range node.Children {
		switch child.Type {
		case "TERMINAL":
			p.write(child.Value)
		case "NAME", "NUMBER", "STRING", "SYMBOL":
			p.write(child.Value)
		case "pattern":
			p.printNode(child)
		}
	}
}

func (p *printer) printReturn(node *ebnf.Node) {
	p.write("return ")
	for _, child := range node.Children {
		if child.Type != "TERMINAL" {
			p.printNode(child)
		}
	}
}

func (p *printer) printExport(node *ebnf.Node) {
	p.write("export [")
	first := true
	for _, child := range node.Children {
		if child.Type == "NAME" {
			if !first {
				p.write(", ")
			}
			first = false
			p.write(child.Value)
		}
	}
	p.write("]")
}

func (p *printer) printImport(node *ebnf.Node) {
	var module string
	var names []string
	
	for _, child := range node.Children {
		if child.Type == "NAME" {
			if module == "" {
				module = child.Value
			} else {
				names = append(names, child.Value)
			}
		}
	}
	
	p.write("from ")
	p.write(module)
	p.write(" use [")
	for i, name := range names {
		if i > 0 {
			p.write(", ")
		}
		p.write(name)
	}
	p.write("]")
}

func (p *printer) printExprStmt(node *ebnf.Node) {
	for _, child := range node.Children {
		p.printNode(child)
	}
}

func (p *printer) printBlock(node *ebnf.Node) {
	p.write("{")
	p.newline()
	p.indent++
	
	for _, child := range node.Children {
		if child.Type == "statement" {
			p.printNode(child)
			p.newline()
		}
	}
	
	p.indent--
	p.writeIndent()
	p.write("}")
}

func (p *printer) printExpr(node *ebnf.Node) {
	// Check for pipe
	hasPipe := false
	for _, child := range node.Children {
		if child.Type == "TERMINAL" && child.Value == "|>" {
			hasPipe = true
			break
		}
	}
	
	if hasPipe {
		p.printPipeExpr(node)
	} else {
		for _, child := range node.Children {
			p.printNode(child)
		}
	}
}

func (p *printer) printPipeExpr(node *ebnf.Node) {
	first := true
	for _, child := range node.Children {
		if child.Type == "TERMINAL" && child.Value == "|>" {
			p.newline()
			p.writeIndent()
			p.write("|> ")
		} else {
			if !first && child.Type != "TERMINAL" {
				// Already handled by |>
			}
			if first || child.Type != "TERMINAL" {
				p.printNode(child)
				first = false
			}
		}
	}
}

func (p *printer) printInfix(node *ebnf.Node) {
	first := true
	for _, child := range node.Children {
		if child.Type == "BINOP" {
			p.write(" ")
			p.printNode(child)
			p.write(" ")
		} else if child.Type == "TERMINAL" && isOperator(child.Value) {
			p.write(" ")
			p.write(child.Value)
			p.write(" ")
		} else {
			if !first {
				// space already added by operator
			}
			p.printNode(child)
			first = false
		}
	}
}

func (p *printer) printBinop(node *ebnf.Node) {
	for _, child := range node.Children {
		if child.Type == "TERMINAL" {
			p.write(child.Value)
		}
	}
}

func (p *printer) printUnary(node *ebnf.Node) {
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

func (p *printer) printPostfix(node *ebnf.Node) {
	for _, child := range node.Children {
		p.printNode(child)
	}
}

func (p *printer) printPrimary(node *ebnf.Node) {
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

func (p *printer) printFuncLiteral(node *ebnf.Node) {
	p.write("(")
	for _, child := range node.Children {
		if child.Type == "params" {
			p.printParams(child)
		}
	}
	p.write(") ")
	for _, child := range node.Children {
		if child.Type == "block" {
			p.printNode(child)
		}
	}
}

func (p *printer) printParams(node *ebnf.Node) {
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
			for _, rp := range child.Children {
				if rp.Type == "NAME" {
					p.write(rp.Value)
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

func (p *printer) printList(node *ebnf.Node) {
	p.write("[")
	first := true
	for _, child := range node.Children {
		switch child.Type {
		case "TERMINAL":
			// skip [ ] ,
		case "SHAPE":
			p.write(child.Value)
			first = false
		default:
			if !first {
				p.write(", ")
			}
			first = false
			p.printNode(child)
		}
	}
	p.write("]")
}

func (p *printer) printCall(node *ebnf.Node) {
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

func (p *printer) printIndex(node *ebnf.Node) {
	p.write("[")
	for _, child := range node.Children {
		if child.Type != "TERMINAL" {
			p.printNode(child)
		}
	}
	p.write("]")
}

func (p *printer) printAccess(node *ebnf.Node) {
	p.write(".")
	for _, child := range node.Children {
		if child.Type == "NAME" {
			p.write(child.Value)
		}
	}
}

func isOperator(s string) bool {
	ops := map[string]bool{
		"+": true, "-": true, "*": true, "/": true, "%": true,
		"<": true, ">": true, "<=": true, ">=": true,
		"==": true, "!=": true,
		"and": true, "or": true,
	}
	return ops[s]
}
