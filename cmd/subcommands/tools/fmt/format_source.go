package fmt

import (
	"fmt"
	"sort"
	"strings"

	"aiki/engine"
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
	return FormatSourceWithObserver(g, file, source, nil)
}

// FormatSourceWithObserver formats source with an optional observer for tracing.
func FormatSourceWithObserver(g *grammar.Grammar, file string, source string, observer engine.Observer) (string, error) {
	if observer == nil {
		observer = engine.SilentObserver{}
	}

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
		observer:    observer,
	}
	p.printProgram(origAST)
	if p.err != nil {
		return "", p.err
	}
	formatted := p.buf.String()

	// Parse formatted output.
	_, fmtAST, err := parseWithTokens(g, file, formatted)
	if err != nil {
		return "", fmt.Errorf("formatter produced invalid source: %w", err)
	}
	if !astEqual(origAST, fmtAST) {
		// Debug: show where ASTs diverge
		observer.OnFormat("AST_DIFF", astDiffStr(origAST, fmtAST, "root"), "comparison", 0)
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
		if t.Type == "WHITESPACE" || t.Type == "NEWLINE" {
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
	// Filter out implicit semicolon terminals for comparison
	aChildren := filterSemicolons(a.Children)
	bChildren := filterSemicolons(b.Children)
	if len(aChildren) != len(bChildren) {
		return false
	}
	for i := range aChildren {
		if !astEqual(aChildren[i], bChildren[i]) {
			return false
		}
	}
	return true
}

func filterSemicolons(nodes []*syntax.Node) []*syntax.Node {
	result := make([]*syntax.Node, 0, len(nodes))
	for _, n := range nodes {
		if n.Type == "TERMINAL" && n.Value == ";" {
			continue
		}
		result = append(result, n)
	}
	return result
}

func astDiffStr(a, b *syntax.Node, path string) string {
	if a == nil && b == nil {
		return ""
	}
	if a == nil {
		return fmt.Sprintf("%s: ORIG=nil FMT=%s", path, b.Type)
	}
	if b == nil {
		return fmt.Sprintf("%s: ORIG=%s FMT=nil", path, a.Type)
	}
	if a.Type != b.Type {
		return fmt.Sprintf("%s: type ORIG=%s FMT=%s", path, a.Type, b.Type)
	}
	if a.Value != b.Value {
		return fmt.Sprintf("%s: value ORIG=%q FMT=%q", path, a.Value, b.Value)
	}
	aChildren := filterSemicolons(a.Children)
	bChildren := filterSemicolons(b.Children)
	if len(aChildren) != len(bChildren) {
		return fmt.Sprintf("%s: children ORIG=%d FMT=%d", path, len(aChildren), len(bChildren))
	}
	for i := range aChildren {
		if diff := astDiffStr(aChildren[i], bChildren[i], fmt.Sprintf("%s/%s[%d]", path, a.Type, i)); diff != "" {
			return diff
		}
	}
	return ""
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
	observer    engine.Observer
	depth       int
	err         error
}

func (p *printer) observe(method, output, node string) {
	p.observer.OnFormat(method, output, node, p.depth)
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
		text := p.comments[cl]
		p.observe("emitComment", text, "comment")
		p.write(text)
		p.newline()
		delete(p.comments, cl)
	}
}

func (p *printer) emitEOLComment(line int) {
	if text, ok := p.eolComments[line]; ok {
		p.write("  ")
		p.observe("emitEOLComment", text, "comment")
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
		text := p.comments[line]
		p.observe("emitTrailingComment", text, "comment")
		p.write(text)
		p.newline()
	}
}

func (p *printer) printProgram(node *syntax.Node) {
	p.observe("printProgram", "enter", fmt.Sprintf("program(%d children)", len(node.Children)))
	p.depth++
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
	p.depth--
	p.observe("printProgram", "exit", "program")
}

func (p *printer) printStatement(node *syntax.Node) {
	if len(node.Children) == 0 {
		p.observe("printStatement", "skip", "empty")
		return
	}
	child := node.Children[0]
	line := nodeStartLine(node)
	p.observe("printStatement", "enter", child.Type)
	p.depth++

	switch child.Type {
	case "if_stmt":
		p.printIf(child)
	case "while_stmt":
		p.printWhile(child)
	case "match_stmt":
		p.printMatch(child)
	case "select_stmt":
		p.printSelect(child)
	default:
		p.printNode(child)
		p.emitEOLComment(line)
		p.newline()
	}
	p.depth--
	p.observe("printStatement", "exit", child.Type)
}

type printKind uint8

const (
	printProgramKind printKind = iota + 1
	printStatementKind
	printPackageKind
	printLetKind
	printAssignKind
	printIfKind
	printWhileKind
	printMatchKind
	printSelectKind
	printReturnKind
	printExprStmtKind
	printBlockKind
	printExprKind
	printPipeExprKind
	printInfixKind
	printUnaryKind
	printPostfixKind
	printPrimaryKind
	printFuncLiteralKind
	printListKind
	printCallKind
	printIndexKind
	printAccessKind
	printParamsKind
	printPatternKind
	printBinopKind
)

// productionPrinters is the formatter's explicit dispatch surface. Its keys are
// checked against the grammar productions in tests; productions intentionally
// emitted by their parent are listed separately in handledByParent.
var productionPrinters = map[string]printKind{
	"program":      printProgramKind,
	"statement":    printStatementKind,
	"package_stmt": printPackageKind,
	"let_stmt":     printLetKind,
	"assign_stmt":  printAssignKind,
	"if_stmt":      printIfKind,
	"while_stmt":   printWhileKind,
	"match_stmt":   printMatchKind,
	"select_stmt":  printSelectKind,
	"return_stmt":  printReturnKind,
	"expr_stmt":    printExprStmtKind,
	"block":        printBlockKind,
	"expr":         printExprKind,
	"pipe_expr":    printPipeExprKind,
	"infix_expr":   printInfixKind,
	"unary_expr":   printUnaryKind,
	"postfix_expr": printPostfixKind,
	"primary":      printPrimaryKind,
	"func_literal": printFuncLiteralKind,
	"list_literal": printListKind,
	"call":         printCallKind,
	"index":        printIndexKind,
	"access":       printAccessKind,
	"params":       printParamsKind,
	"pattern":      printPatternKind,
	"BINOP":        printBinopKind,
}

// handledByParent names grammar productions whose syntax is deliberately
// emitted by the enclosing formatter method rather than through printNode
// dispatch. Keeping the exception set explicit makes production coverage
// mechanically checkable.
var handledByParent = map[string]struct{}{
	"field":          {},
	"select_case":    {},
	"select_default": {},
	"param_list":     {},
	"rest_param":     {},
	"literal":        {},
}

var printableLeafTypes = map[string]struct{}{
	"NUMBER": {},
	"STRING": {},
	"RUNE":   {},
	"SYMBOL": {},
	"SHAPE":  {},
	"NAME":   {},
}

func (p *printer) printNode(node *syntax.Node) {
	if p.err != nil {
		return
	}
	p.observe("printNode", "enter", node.Type)
	p.depth++
	defer func() {
		p.depth--
		p.observe("printNode", "exit", node.Type)
	}()

	if kind, ok := productionPrinters[node.Type]; ok {
		switch kind {
		case printProgramKind:
			p.printProgram(node)
		case printStatementKind:
			p.printStatement(node)
		case printPackageKind:
			p.printPackage(node)
		case printLetKind:
			p.printLet(node)
		case printAssignKind:
			p.printAssign(node)
		case printIfKind:
			p.printIf(node)
		case printWhileKind:
			p.printWhile(node)
		case printMatchKind:
			p.printMatch(node)
		case printSelectKind:
			p.printSelect(node)
		case printReturnKind:
			p.printReturn(node)
		case printExprStmtKind:
			p.printExprStmt(node)
		case printBlockKind:
			p.printBlock(node)
		case printExprKind:
			p.printExpr(node)
		case printPipeExprKind:
			p.printPipeExpr(node)
		case printInfixKind:
			p.printInfix(node)
		case printUnaryKind:
			p.printUnary(node)
		case printPostfixKind:
			p.printPostfix(node)
		case printPrimaryKind:
			p.printPrimary(node)
		case printFuncLiteralKind:
			p.printFuncLiteral(node)
		case printListKind:
			p.printList(node)
		case printCallKind:
			p.printCall(node)
		case printIndexKind:
			p.printIndex(node)
		case printAccessKind:
			p.printAccess(node)
		case printParamsKind:
			p.printParams(node)
		case printPatternKind:
			p.printPattern(node)
		case printBinopKind:
			p.printBinop(node)
		default:
			p.err = fmt.Errorf("formatter has invalid dispatch kind %d for AST node %q", kind, node.Type)
		}
		return
	}
	if _, ok := printableLeafTypes[node.Type]; ok {
		p.observe("printNode", node.Value, node.Type)
		p.write(node.Value)
		return
	}
	if node.Type == "TERMINAL" {
		p.observe("printNode", node.Value, "TERMINAL")
		p.write(node.Value)
		return
	}
	if _, ok := handledByParent[node.Type]; ok {
		p.observe("printNode", "parent-handled-recurse", node.Type)
		for _, child := range node.Children {
			p.printNode(child)
		}
		return
	}

	// Unknown nodes must never disappear silently. A leaf has no child through
	// which the old default-recursion fallback could preserve its text; a new
	// production likewise needs an explicit formatter disposition.
	p.err = fmt.Errorf("formatter has no disposition for AST node %q", node.Type)
}
