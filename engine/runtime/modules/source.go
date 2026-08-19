package modules

import (
	"strings"

	"aiki/engine/syntax"
	"aiki/engine/syntax/grammar"
)

// FunctionInfo is the source-declared signature of a named module binding whose
// value is a function literal. It is derived from the Aiki AST rather than from
// a parallel textual grammar.
type FunctionInfo struct {
	Name       string
	Parameters []string
	Rest       string
	Calls      []string
}

// SourceInfo is structural module metadata derived from parsed Aiki source.
// Source remains authoritative; this is a reusable engine-owned view.
type SourceInfo struct {
	Package   string
	Exports   []string
	Imports   []string
	Functions map[string]FunctionInfo
}

// AnalyzeSource parses Aiki module source and derives package, export, and named
// function facts from the real grammar and AST.
func AnalyzeSource(g *grammar.Grammar, file, source string) (SourceInfo, error) {
	lx := syntax.NewLexer(g, file, source, nil)
	tokens, err := lx.Tokenize()
	if err != nil {
		return SourceInfo{}, err
	}
	p := syntax.NewParser(g, tokens, source, nil)
	root, err := p.Parse()
	if err != nil {
		return SourceInfo{}, err
	}

	info := SourceInfo{Functions: make(map[string]FunctionInfo)}
	walkSource(root, &info)
	info.Imports = importCalls(root)
	return info, nil
}

func walkSource(node *syntax.Node, info *SourceInfo) {
	if node == nil || node.Type == "block" {
		return
	}

	switch node.Type {
	case "package_stmt":
		if info.Package == "" {
			if tok := findFirst(node, "STRING"); tok != nil {
				info.Package = unquoteAikiString(tok.Value)
			}
		}
		return
	case "let_stmt":
		if fn, ok := functionBinding(node); ok {
			info.Functions[fn.Name] = fn
		}
		return
	case "postfix_expr":
		if names, ok := exportCall(node); ok {
			info.Exports = append(info.Exports, names...)
			return
		}
	}

	for _, child := range node.Children {
		walkSource(child, info)
	}
}

func importCalls(root *syntax.Node) []string {
	seen := make(map[string]bool)
	var out []string
	var walk func(*syntax.Node)
	walk = func(node *syntax.Node) {
		if node == nil {
			return
		}
		if node.Type == "postfix_expr" && len(node.Children) > 0 {
			name := findFirst(node.Children[0], "NAME")
			if name != nil && (name.Value == "import" || name.Value == "use") {
				for _, child := range node.Children[1:] {
					if child.Type != "call" {
						continue
					}
					if literal := findFirst(child, "STRING"); literal != nil {
						module := unquoteAikiString(literal.Value)
						if module != "" && !seen[module] {
							seen[module] = true
							out = append(out, module)
						}
					}
					break
				}
			}
		}
		for _, child := range node.Children {
			walk(child)
		}
	}
	walk(root)
	return out
}

func functionBinding(node *syntax.Node) (FunctionInfo, bool) {
	var name string
	var valueNode *syntax.Node
	for i, child := range node.Children {
		if child.Type == "NAME" && name == "" {
			name = child.Value
		}
		if child.Type == "TERMINAL" && child.Value == "=" && i+1 < len(node.Children) {
			valueNode = node.Children[i+1]
		}
	}
	if name == "" || valueNode == nil {
		return FunctionInfo{}, false
	}
	fn := findFirst(valueNode, "func_literal")
	if fn == nil {
		return FunctionInfo{}, false
	}

	out := FunctionInfo{Name: name}
	params := findFirst(fn, "params")
	if params == nil {
		out.Calls = functionCalls(fn)
		return out, true
	}
	collectParameters(params, &out)
	out.Calls = functionCalls(fn)
	return out, true
}

func functionCalls(fn *syntax.Node) []string {
	seen := make(map[string]bool)
	var out []string
	var walk func(*syntax.Node)
	walk = func(node *syntax.Node) {
		if node == nil {
			return
		}
		if node.Type == "postfix_expr" && len(node.Children) > 0 {
			hasCall := false
			for _, child := range node.Children[1:] {
				if child.Type == "call" {
					hasCall = true
					break
				}
			}
			if hasCall {
				if name := findFirst(node.Children[0], "NAME"); name != nil && !seen[name.Value] {
					seen[name.Value] = true
					out = append(out, name.Value)
				}
			}
		}
		for _, child := range node.Children {
			walk(child)
		}
	}
	walk(fn)
	return out
}

func collectParameters(node *syntax.Node, out *FunctionInfo) {
	if node == nil {
		return
	}
	if node.Type == "rest_param" {
		if name := findFirst(node, "NAME"); name != nil {
			out.Rest = name.Value
		}
		return
	}
	if node.Type == "NAME" {
		out.Parameters = append(out.Parameters, node.Value)
		return
	}
	for _, child := range node.Children {
		collectParameters(child, out)
	}
}

func exportCall(node *syntax.Node) ([]string, bool) {
	if node == nil || node.Type != "postfix_expr" || len(node.Children) == 0 {
		return nil, false
	}
	name := findFirst(node.Children[0], "NAME")
	if name == nil || name.Value != "export" {
		return nil, false
	}

	var out []string
	for _, child := range node.Children[1:] {
		if child.Type != "call" {
			continue
		}
		for _, sym := range findAll(child, "SYMBOL") {
			name := strings.TrimPrefix(sym.Value, ":")
			if name != "" {
				out = append(out, name)
			}
		}
	}
	return out, true
}

func findFirst(node *syntax.Node, typ string) *syntax.Node {
	if node == nil {
		return nil
	}
	if node.Type == typ {
		return node
	}
	for _, child := range node.Children {
		if found := findFirst(child, typ); found != nil {
			return found
		}
	}
	return nil
}

func findAll(node *syntax.Node, typ string) []*syntax.Node {
	if node == nil {
		return nil
	}
	var out []*syntax.Node
	if node.Type == typ {
		out = append(out, node)
	}
	for _, child := range node.Children {
		out = append(out, findAll(child, typ)...)
	}
	return out
}

func unquoteAikiString(s string) string {
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}

// ExportNames returns the module's exported names for tolerant tooling use.
// Valid source is always interpreted through AnalyzeSource. The narrow textual
// fallback preserves editor behavior for temporarily invalid source; it is not
// used by runtime registration or invariants.
func ExportNames(g *grammar.Grammar, file, source string) []string {
	if g != nil {
		if info, err := AnalyzeSource(g, file, source); err == nil {
			return info.Exports
		}
	}
	return exportNamesFallback(source)
}

func exportNamesFallback(source string) []string {
	var exports []string
	for _, line := range strings.Split(source, "\n") {
		trim := strings.TrimSpace(line)
		if !strings.HasPrefix(trim, "export(") {
			continue
		}
		start := strings.Index(trim, "(")
		end := strings.LastIndex(trim, ")")
		if start < 0 || end <= start {
			continue
		}
		for _, part := range strings.Split(trim[start+1:end], ",") {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, ":") && len(part) > 1 {
				exports = append(exports, strings.TrimPrefix(part, ":"))
			}
		}
	}
	return exports
}
