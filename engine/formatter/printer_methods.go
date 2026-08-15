package formatter

import "aiki/engine/syntax"

func (p *printer) printPackage(node *syntax.Node) {
	p.observe("printPackage", "enter", "package_stmt")
	p.writeIndent()
	p.write("package ")
	for _, child := range node.Children {
		if child.Type == "STRING" {
			p.observe("printPackage", child.Value, "STRING")
			p.write(child.Value)
			break
		}
	}
	p.observe("printPackage", "exit", "package_stmt")
}

func (p *printer) printLetBinding(node *syntax.Node) {
	p.observe("printLetBinding", "enter", "let_binding")
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

	p.observe("printLetBinding", name, "NAME")
	p.write(name)
	p.write(" = ")
	if valueNode != nil {
		p.printNode(valueNode)
	}
	p.observe("printLetBinding", "exit", "let_binding")
}

func (p *printer) printAssign(node *syntax.Node) {
	p.observe("printAssign", "enter", "assign_stmt")
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
	p.observe("printAssign", "exit", "assign_stmt")
}

func (p *printer) printIf(node *syntax.Node) {
	p.observe("printIf", "enter", "if_stmt")
	defer p.observe("printIf", "exit", "if_stmt")

	p.writeIndent()
	p.write("if ")

	// Extract structural parts: condition, then-block, else-branch
	var cond *syntax.Node
	var thenBlock *syntax.Node
	var elseNode *syntax.Node // either block or if_stmt

	for _, ch := range node.Children {
		switch ch.Type {
		case "TERMINAL":
			continue // skip "if", "else" terminals
		case "block":
			if thenBlock == nil {
				thenBlock = ch
			} else {
				elseNode = ch
			}
		case "if_stmt":
			elseNode = ch
		default:
			if cond == nil {
				cond = ch
			}
		}
	}

	if cond != nil {
		p.printNode(cond)
	}
	p.write(" ")
	if thenBlock != nil {
		p.printBlock(thenBlock)
	}

	if elseNode != nil {
		p.write(" else ")
		if elseNode.Type == "if_stmt" {
			// else if: print without leading indent
			p.printIfNoIndent(elseNode)
			return // printIfNoIndent handles newline
		} else {
			p.printBlock(elseNode)
		}
	}

	line := nodeStartLine(node)
	p.emitEOLComment(line)
	p.newline()
}

// printIfNoIndent prints an if statement without leading indent (for else if chains)
func (p *printer) printIfNoIndent(node *syntax.Node) {
	p.observe("printIfNoIndent", "enter", "if_stmt")
	defer p.observe("printIfNoIndent", "exit", "if_stmt")

	p.write("if ")

	var cond *syntax.Node
	var thenBlock *syntax.Node
	var elseNode *syntax.Node

	for _, ch := range node.Children {
		switch ch.Type {
		case "TERMINAL":
			continue
		case "block":
			if thenBlock == nil {
				thenBlock = ch
			} else {
				elseNode = ch
			}
		case "if_stmt":
			elseNode = ch
		default:
			if cond == nil {
				cond = ch
			}
		}
	}

	if cond != nil {
		p.printNode(cond)
	}
	p.write(" ")
	if thenBlock != nil {
		p.printBlock(thenBlock)
	}

	if elseNode != nil {
		p.write(" else ")
		if elseNode.Type == "if_stmt" {
			p.printIfNoIndent(elseNode)
			return
		} else {
			p.printBlock(elseNode)
		}
	}

	line := nodeStartLine(node)
	p.emitEOLComment(line)
	p.newline()
}

func (p *printer) printWhile(node *syntax.Node) {
	p.observe("printWhile", "enter", "while_stmt")
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
	p.observe("printWhile", "exit", "while_stmt")
}

func (p *printer) printMatch(node *syntax.Node) {
	p.observe("printMatch", "enter", "match_stmt")
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
	p.observe("printMatch", "exit", "match_stmt")
}

func (p *printer) printSelect(node *syntax.Node) {
	p.observe("printSelect", "enter", "select_stmt")
	p.writeIndent()
	p.write("select {\n")
	p.indent++

	for _, child := range node.Children {
		switch child.Type {
		case "select_case":
			p.writeIndent()
			var bind string
			var expr *syntax.Node
			var block *syntax.Node
			for _, part := range child.Children {
				switch part.Type {
				case "NAME":
					if bind == "" {
						bind = part.Value
					}
				case "expr":
					expr = part
				case "block":
					block = part
				}
			}
			if bind != "" {
				p.write("let ")
				p.write(bind)
				p.write(" = ")
			}
			p.write("recv(")
			if expr != nil {
				p.printNode(expr)
			}
			p.write(") ")
			if block != nil {
				p.printBlock(block)
			}
			p.newline()
		case "select_default":
			p.writeIndent()
			p.write("default ")
			if block := child.ChildByType("block"); block != nil {
				p.printBlock(block)
			}
			p.newline()
		}
	}

	p.indent--
	p.writeIndent()
	p.write("}")
	line := nodeStartLine(node)
	p.emitEOLComment(line)
	p.newline()
	p.observe("printSelect", "exit", "select_stmt")
}

func (p *printer) printPattern(node *syntax.Node) {
	p.observe("printPattern", "enter", "pattern")
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
		case "NUMBER", "STRING", "RUNE", "SYMBOL", "SHAPE", "NAME":
			p.write(child.Value)
		case "pattern":
			p.printPattern(child)
		case "literal":
			p.printPatternLiteral(child)
		}
	}
	p.observe("printPattern", "exit", "pattern")
}

func (p *printer) printPatternLiteral(node *syntax.Node) {
	for _, child := range node.Children {
		switch child.Type {
		case "NUMBER", "STRING", "RUNE", "SYMBOL", "SHAPE":
			p.write(child.Value)
		case "TERMINAL":
			p.write(child.Value)
		}
	}
}

func (p *printer) printReturn(node *syntax.Node) {
	p.observe("printReturn", "enter", "return_stmt")
	p.writeIndent()
	p.write("return ")
	for _, child := range node.Children {
		if child.Type != "TERMINAL" {
			p.printNode(child)
		}
	}
	p.observe("printReturn", "exit", "return_stmt")
}

func (p *printer) printExprStmt(node *syntax.Node) {
	p.observe("printExprStmt", "enter", "expr_stmt")
	p.writeIndent()
	for _, child := range node.Children {
		p.printNode(child)
	}
	p.observe("printExprStmt", "exit", "expr_stmt")
}

func (p *printer) printBlock(node *syntax.Node) {
	p.observe("printBlock", "enter", "block")
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
	p.observe("printBlock", "exit", "block")
}

func (p *printer) printExpr(node *syntax.Node) {
	p.observe("printExpr", "enter", "expr")
	for _, child := range node.Children {
		p.printNode(child)
	}
	p.observe("printExpr", "exit", "expr")
}

func (p *printer) printPipeExpr(node *syntax.Node) {
	p.observe("printPipeExpr", "enter", "pipe_expr")
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
		p.observe("printPipeExpr", "exit", "pipe_expr")
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
	p.observe("printPipeExpr", "exit", "pipe_expr")
}

func (p *printer) printInfix(node *syntax.Node) {
	p.observe("printInfix", "enter", "infix_expr")
	first := true
	for _, child := range node.Children {
		if child.Type == "BINOP" {
			p.write(" ")
			p.printBinop(child)
			p.write(" ")
		} else if child.Type == "TERMINAL" {
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
	p.observe("printInfix", "exit", "infix_expr")
}

func (p *printer) printBinop(node *syntax.Node) {
	for _, child := range node.Children {
		if child.Type == "TERMINAL" {
			p.observe("printBinop", child.Value, "TERMINAL")
			p.write(child.Value)
		}
	}
}

func (p *printer) printUnary(node *syntax.Node) {
	p.observe("printUnary", "enter", "unary_expr")
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
	p.observe("printUnary", "exit", "unary_expr")
}

func (p *printer) printPostfix(node *syntax.Node) {
	p.observe("printPostfix", "enter", "postfix_expr")
	for _, child := range node.Children {
		p.printNode(child)
	}
	p.observe("printPostfix", "exit", "postfix_expr")
}

func (p *printer) printPrimary(node *syntax.Node) {
	p.observe("printPrimary", "enter", "primary")
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
	p.observe("printPrimary", "exit", "primary")
}

func (p *printer) printFuncLiteral(node *syntax.Node) {
	p.observe("printFuncLiteral", "enter", "func_literal")
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
	p.observe("printFuncLiteral", "exit", "func_literal")
}

func (p *printer) printParams(node *syntax.Node) {
	p.observe("printParams", "enter", "params")
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
					p.observe("printParams", param.Value, "NAME")
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
					p.observe("printParams", "..."+param.Value, "rest")
					p.write(param.Value)
				}
			}
		case "NAME":
			if !first {
				p.write(", ")
			}
			first = false
			p.observe("printParams", child.Value, "NAME")
			p.write(child.Value)
		}
	}
	p.observe("printParams", "exit", "params")
}

func (p *printer) printList(node *syntax.Node) {
	p.observe("printList", "enter", "list_literal")
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
	p.observe("printList", "exit", "list_literal")
}

func (p *printer) printCall(node *syntax.Node) {
	p.observe("printCall", "enter", "call")
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
	p.observe("printCall", "exit", "call")
}

func (p *printer) printIndex(node *syntax.Node) {
	p.observe("printIndex", "enter", "index")
	p.write("[")
	for _, child := range node.Children {
		if child.Type != "TERMINAL" {
			p.printNode(child)
		}
	}
	p.write("]")
	p.observe("printIndex", "exit", "index")
}

func (p *printer) printAccess(node *syntax.Node) {
	p.observe("printAccess", "enter", "access")
	p.write(".")
	for _, child := range node.Children {
		if child.Type == "NAME" {
			p.observe("printAccess", child.Value, "NAME")
			p.write(child.Value)
		}
	}
	p.observe("printAccess", "exit", "access")
}

func (p *printer) printLet(node *syntax.Node) {
	p.observe("printLet", "enter", "let_stmt")
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
	p.observe("printLet", "exit", "let_stmt")
}

func (p *printer) printShapeDef(node *syntax.Node) {
	p.observe("printShapeDef", "enter", "shape_def")
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
	p.observe("printShapeDef", "exit", "shape_def")
}
