package fmt

import (
	"fmt"
	"sort"
	"strings"

	"aiki/engine/syntax"
	"aiki/engine/syntax/grammar"
)

// FormatSource formats valid Aiki source to the canonical style.
//
// Safety gates:
//  1. Parse original must succeed.
//  2. Parse formatted output must succeed.
//  3. AST structure must match (prevents meaning changes).
func FormatSource(g *grammar.Grammar, file string, source string) (string, error) {
	// Parse original.
	origTokens, origAST, err := parseWithTokens(g, file, source)
	if err != nil {
		return "", err
	}

	// Extract comments by line.
	comments, eol := extractComments(origTokens)

	p := &printer{
		comments:    comments,
		eolComments: eol,
	}
	p.printProgram(origAST)
	formatted := p.buf.String()

	// Parse formatted output.
	_, fmtAST, err := parseWithTokens(g, file, formatted)
	if err != nil {
		return "", fmt.Errorf("formatter produced invalid source: %w", err)
	}
	if !astEqual(origAST, fmtAST) {
		return "", fmt.Errorf("formatter changed AST; refusing to write")
	}
	return formatted, nil
}

func parseWithTokens(g *grammar.Grammar, file, source string) ([]syntax.Token, *syntax.Node, error) {
	lx := syntax.NewLexer(g, file, source, nil)
	toks, err := lx.Tokenize()
	if err != nil {
		return nil, nil, err
	}
	ps := syntax.NewParser(g, toks, source, nil)
	node, err := ps.Parse()
	if err != nil {
		return toks, nil, err
	}
	return toks, node, nil
}

func extractComments(tokens []syntax.Token) (map[int]string, map[int]string) {
	standalone := make(map[int]string)
	eol := make(map[int]string)

	for i, tok := range tokens {
		if tok.Type != "COMMENT" {
			continue
		}
		line := tok.Pos.Line
		if isEOLComment(tokens, i) {
			eol[line] = tok.Lexeme
		} else {
			standalone[line] = tok.Lexeme
		}
	}
	return standalone, eol
}

func isEOLComment(tokens []syntax.Token, idx int) bool {
	c := tokens[idx]
	line := c.Pos.Line
	for j := idx - 1; j >= 0; j-- {
		t := tokens[j]
		if t.Pos.Line != line {
			return false
		}
		if t.Type == "WHITESPACE" {
			continue
		}
		if t.Type == "COMMENT" {
			continue
		}
		// Some non-whitespace token precedes on the same line.
		return true
	}
	return false
}

func astEqual(a, b *syntax.Node) bool {
	if a == nil || b == nil {
		return a == b
	}
	if a.Type != b.Type {
		return false
	}
	if a.Value != b.Value {
		return false
	}
	if len(a.Children) != len(b.Children) {
		return false
	}
	for i := range a.Children {
		if !astEqual(a.Children[i], b.Children[i]) {
			return false
		}
	}
	return true
}

// nodeStartLine returns the line number of a node's first token.
func nodeStartLine(node *syntax.Node) int {
	if node == nil {
		return 0
	}
	if node.Pos.Line > 0 {
		return node.Pos.Line
	}
	for _, child := range node.Children {
		line := nodeStartLine(child)
		if line > 0 {
			return line
		}
	}
	return 0
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

func (p *printer) emitEOLComment(line int) {
	if text, ok := p.eolComments[line]; ok {
		p.write("  ")
		p.write(text)
		delete(p.eolComments, line)
	}
}

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

func (p *printer) printProgram(node *syntax.Node) {
	prevLine := 0
	for i, child := range node.Children {
		line := nodeStartLine(child)
		if i > 0 && line > prevLine+1 {
			p.newline()
		}
		p.emitCommentsBefore(line)
		p.printStatement(child)
		p.lastLine = line
		prevLine = line
	}
	p.emitTrailingComments()

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
	case "package_stmt":
		p.printPackage(node)
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
