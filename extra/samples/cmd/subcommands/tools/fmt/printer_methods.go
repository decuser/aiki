package fmt

import "aiki/engine/syntax"

func (p *printer) printPackage(node *syntax.Node) {
	p.writeIndent()
	// package_stmt = "package" STRING
	p.write("package ")
	for _, child := range node.Children {
		if child.Type == "STRING" {
			p.write(child.Value)
			break
		}
	}
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

	if i < len(children) && children[i].Type == "TERMINAL" && children[i].Value == "if" {
		i++
	}
	if i < len(children) && children[i].Type != "block" && children[i].Type != "TERMINAL" {
		p.printNode(children[i])
		i++
	}
	p.write(" ")
	if i < len(children) && children[i].Type == "block" {
		p.printBlock(children[i])
		i++
	}
	if i < len(children) && children[i].Type == "TERMINAL" && children[i].Value == "else" {
		p.write(" else ")
		i++
		if i < len(children) && children[i].Type == "if_stmt" {
			// else if prints as nested if on same line.
			p.printIf(children[i])
			return
		}
		if i < len(children) && children[i].Type == "block" {
			p.printBlock(children[i])
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
	if i < len(children) && children[i].Type == "TERMINAL" && children[i].Value == "while" {
		i++
	}
	if i < len(children) && children[i].Type != "block" && children[i].Type != "TERMINAL" {
		p.printNode(children[i])
		i++
	}
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
	if i < len(children) && children[i].Type == "TERMINAL" && children[i].Value == "match" {
		i++
	}
	if i < len(children) && children[i].Type != "TERMINAL" && children[i].Type != "pattern" && children[i].Type != "block" {
		p.printNode(children[i])
		i++
	}

	p.write(" {\n")
	p.indent++

	for i < len(children) {
		child := children[i]
		if child.Type == "TERMINAL" {
			i++
			continue
		}
		if child.Type == "pattern" {
			p.writeIndent()
			p.printPattern(child)
			i++
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
			switch child.Value {
			case "_":
				p.write("_")
			case "[":
				p.write("[")
			case "]":
				p.write("]")
			case ",":
				p.write(", ")
			default:
				p.write(child.Value)
			}
		case "NUMBER", "STRING", "SYMBOL", "SHAPE", "NAME":
			p.write(child.Value)
		case "pattern":
			p.printPattern(child)
		case "literal":
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
	for _, child := range node.Children {
		p.printNode(child)
	}
}

func (p *printer) printPipeExpr(node *syntax.Node) {
	children := node.Children
	pipeCount := 0
	for _, child := range children {
		if child.Type == "TERMINAL" && child.Value == "|>" {
			pipeCount++
		}
	}
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
	for _, child := range children {
		if child.Type == "TERMINAL" && child.Value == "|>" {
			p.write(" |>")
			p.newline()
			p.writeIndent()
		} else if child.Type != "TERMINAL" {
			p.printNode(child)
		}
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
			// Keep a conservative set; BINOP is preferred.
			switch child.Value {
			case "and", "or", "==", "!=", "<", ">", "<=", ">=", "+", "-", "*", "/", "%":
				p.write(" ")
				p.write(child.Value)
				p.write(" ")
			}
		} else {
			if !first {
				// spacing handled above
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
