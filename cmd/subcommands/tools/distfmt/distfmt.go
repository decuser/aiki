package distfmt

import (
	"bytes"
	"errors"
	"fmt"
	goast "go/ast"
	goformat "go/format"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"unicode/utf8"

	"aiki/cmd/internal/testfixture"
	aikifmt "aiki/cmd/subcommands/tools/fmt"
	"aiki/engine/syntax"
	"aiki/engine/syntax/grammar"
)

const distributionWidth = 100

// Config mirrors the safety/user interaction surface of aiki fmt.
type Config struct {
	Write         bool
	ListOnly      bool
	PrintToStdout bool
	Backup        bool
}

// FormatPath restyles a Go or Aiki file, directory pattern (./...), or path.
func FormatPath(path string, cfg Config) (bool, error) {
	if !isRecursive(path) {
		skip, err := testfixture.IsParseNegative(path)
		if err != nil {
			return false, err
		}
		if skip {
			return false, nil
		}
	}
	if isRecursive(path) {
		dir := strings.TrimSuffix(path, "/...")
		dir = strings.TrimSuffix(dir, string(filepath.Separator)+"...")
		if dir == "" {
			dir = "."
		}
		return formatDir(dir, cfg)
	}
	return formatFile(path, cfg, os.Stdout)
}

func isRecursive(path string) bool {
	return strings.HasSuffix(path, "/...") || strings.HasSuffix(path, string(filepath.Separator)+"...")
}

func formatDir(dir string, cfg Config) (bool, error) {
	changedAny := false
	var formatErrs []error
	walkErr := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
			return filepath.SkipDir
		}
		if info.IsDir() {
			return nil
		}
		ext := filepath.Ext(path)
		if ext != ".ai" && ext != ".go" {
			return nil
		}
		if ext == ".ai" {
			skip, ferr := testfixture.IsParseNegative(path)
			if ferr != nil {
				return ferr
			}
			if skip {
				return nil
			}
		}
		changed, ferr := formatFile(path, cfg, io.Discard)
		if ferr != nil {
			formatErrs = append(formatErrs, fmt.Errorf("%s: %w", path, ferr))
			return nil
		}
		if changed {
			changedAny = true
		}
		return nil
	})
	if walkErr != nil {
		return changedAny, walkErr
	}
	return changedAny, errors.Join(formatErrs...)
}

func formatFile(path string, cfg Config, stdout io.Writer) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}

	var formatted string
	switch filepath.Ext(path) {
	case ".go":
		formatted, err = formatGo(path, string(data))
	case ".ai":
		formatted, err = formatAiki(path, string(data))
	default:
		return false, fmt.Errorf("unsupported file type: %s", path)
	}
	if err != nil {
		return false, err
	}

	changed := formatted != string(data)
	if cfg.PrintToStdout {
		if !changed {
			_, _ = stdout.Write(data)
			return false, nil
		}
		_, _ = io.Copy(stdout, bytes.NewBufferString(formatted))
		return true, nil
	}
	if cfg.ListOnly {
		if changed {
			fmt.Fprintln(os.Stdout, path)
			return true, nil
		}
		return false, nil
	}
	if cfg.Write && changed {
		fmt.Fprintln(os.Stdout, path)
		if cfg.Backup {
			if err := os.WriteFile(path+".bak", data, 0644); err != nil {
				return false, err
			}
		}
		if err := atomicWriteFile(path, []byte(formatted), 0644); err != nil {
			return false, err
		}
	}
	return changed, nil
}

func formatGo(path, source string) (string, error) {
	canonical, err := goformat.Source([]byte(source))
	if err != nil {
		return "", err
	}
	before, err := normalizedGoAST(path, canonical)
	if err != nil {
		return "", err
	}

	restyled := restyleGoMapValues(string(canonical))
	afterCanonical, err := goformat.Source([]byte(restyled))
	if err != nil {
		return "", fmt.Errorf("restyle produced invalid Go: %w", err)
	}
	after, err := normalizedGoAST(path, afterCanonical)
	if err != nil {
		return "", err
	}
	if before != after {
		return "", fmt.Errorf("distfmt changed Go syntax tree; refusing to write")
	}
	return string(afterCanonical), nil
}

func normalizedGoAST(path string, source []byte) (string, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, source, parser.SkipObjectResolution)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	posType := reflect.TypeOf(token.Pos(0))
	filter := func(_ string, value reflect.Value) bool {
		return value.Type() != posType
	}
	if err := goast.Fprint(&buf, fset, file, filter); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// restyleGoMapValues expands long single-line elided composite literals used as
// map values. Selection is AST-driven so source text inside strings/comments is
// never mistaken for Go structure. gofmt preserves the resulting multiline
// composite literal.
func restyleGoMapValues(source string) string {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "distfmt.go", source, parser.SkipObjectResolution)
	if err != nil {
		return source
	}

	type replacement struct {
		start int
		end   int
		text  string
	}
	var replacements []replacement

	goast.Inspect(file, func(node goast.Node) bool {
		kv, ok := node.(*goast.KeyValueExpr)
		if !ok {
			return true
		}
		lit, ok := kv.Value.(*goast.CompositeLit)
		if !ok || lit.Type != nil || len(lit.Elts) < 2 {
			return true
		}

		startPos := fset.Position(lit.Lbrace)
		endPos := fset.Position(lit.Rbrace)
		if startPos.Line != endPos.Line || startPos.Offset < 0 || endPos.Offset < startPos.Offset {
			return true
		}
		if utf8.RuneCountInString(source[startPos.Offset:endPos.Offset+1]) <= distributionWidth {
			return true
		}

		lineStart := strings.LastIndex(source[:startPos.Offset], "\n") + 1
		baseIndent := leadingWhitespace(source[lineStart:startPos.Offset])
		itemIndent := baseIndent + "\t"

		var b strings.Builder
		b.WriteString("{")
		b.WriteString("\n")
		for _, elt := range lit.Elts {
			ep := fset.Position(elt.Pos())
			ee := fset.Position(elt.End())
			if ep.Offset < 0 || ee.Offset < ep.Offset || ee.Offset > len(source) {
				return true
			}
			b.WriteString(itemIndent)
			b.WriteString(strings.TrimSpace(source[ep.Offset:ee.Offset]))
			b.WriteString(",\n")
		}
		b.WriteString(baseIndent)
		b.WriteString("}")

		replacements = append(replacements, replacement{
			start: startPos.Offset,
			end:   endPos.Offset + 1,
			text:  b.String(),
		})
		return true
	})

	if len(replacements) == 0 {
		return source
	}
	sort.Slice(replacements, func(i, j int) bool { return replacements[i].start > replacements[j].start })
	out := source
	for _, r := range replacements {
		out = out[:r.start] + r.text + out[r.end:]
	}
	return out
}

func formatAiki(path, source string) (string, error) {
	g, err := loadGrammar()
	if err != nil {
		return "", err
	}
	canonical, err := aikifmt.FormatSource(g, path, source)
	if err != nil {
		return "", err
	}
	before, err := parseAikiAST(g, path, canonical)
	if err != nil {
		return "", err
	}

	current := canonical
	for pass := 0; pass < 8; pass++ {
		restyled := restyleAiki(current)
		normalized, err := aikifmt.FormatSource(g, path, restyled)
		if err != nil {
			return "", fmt.Errorf("restyle produced invalid Aiki: %w", err)
		}
		after, err := parseAikiAST(g, path, normalized)
		if err != nil {
			return "", err
		}
		if !aikiASTEqual(before, after) {
			return "", fmt.Errorf("distfmt changed Aiki syntax tree; refusing to write")
		}
		if normalized == current {
			return normalized, nil
		}
		current = normalized
	}
	return "", fmt.Errorf("distfmt Aiki layout did not reach a fixed point")
}

func parseAikiAST(g *grammar.Grammar, path, source string) (*syntax.Node, error) {
	lx := syntax.NewLexer(g, path, source, nil)
	tokens, err := lx.Tokenize()
	if err != nil {
		return nil, err
	}
	ps := syntax.NewParser(g, tokens, source, nil)
	return ps.Parse()
}

func aikiASTEqual(a, b *syntax.Node) bool {
	if a == nil || b == nil {
		return a == b
	}
	if a.Type != b.Type || a.Value != b.Value {
		return false
	}
	ac := nonSemicolonChildren(a.Children)
	bc := nonSemicolonChildren(b.Children)
	if len(ac) != len(bc) {
		return false
	}
	for i := range ac {
		if !aikiASTEqual(ac[i], bc[i]) {
			return false
		}
	}
	return true
}

func nonSemicolonChildren(nodes []*syntax.Node) []*syntax.Node {
	out := make([]*syntax.Node, 0, len(nodes))
	for _, node := range nodes {
		if node.Type == "TERMINAL" && node.Value == ";" {
			continue
		}
		out = append(out, node)
	}
	return out
}

func loadGrammar() (*grammar.Grammar, error) {
	return grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
}

// restyleAiki expands long single-line comma-separated lists and calls. The
// canonical formatter preserves these explicit multiline layouts, so distfmt
// output remains stable under subsequent aiki fmt passes.
func restyleAiki(source string) string {
	lines := strings.Split(source, "\n")
	var out []string
	for _, line := range lines {
		if utf8.RuneCountInString(line) <= distributionWidth {
			out = append(out, line)
			continue
		}
		restyled, ok := expandAikiDelimited(line, '[', ']')
		if !ok {
			restyled, ok = expandAikiDelimited(line, '(', ')')
		}
		if ok {
			out = append(out, restyled...)
		} else {
			out = append(out, line)
		}
	}
	return strings.Join(out, "\n")
}

func expandAikiDelimited(line string, openRune, closeRune rune) ([]string, bool) {
	open := strings.IndexRune(line, openRune)
	close := strings.LastIndex(line, string(closeRune))
	if open < 0 || close <= open {
		return nil, false
	}
	items, ok := splitTopLevelCSV(line[open+1 : close])
	if !ok || len(items) < 2 {
		return nil, false
	}
	prefix := line[:open+1]
	suffix := line[close+1:]
	indent := leadingWhitespace(line) + "\t"
	out := []string{prefix}
	for i, item := range items {
		comma := ""
		if i+1 < len(items) {
			comma = ","
		}
		out = append(out, indent+strings.TrimSpace(item)+comma)
	}
	out = append(out, leadingWhitespace(line)+string(closeRune)+suffix)
	return out, true
}

func splitTopLevelCSV(s string) ([]string, bool) {
	var items []string
	start := 0
	depth := 0
	var quote rune
	escaped := false
	runes := []rune(s)
	for i, r := range runes {
		if quote != 0 {
			if escaped {
				escaped = false
				continue
			}
			if r == '\\' {
				escaped = true
				continue
			}
			if r == quote {
				quote = 0
			}
			continue
		}
		switch r {
		case '\'', '"', '`':
			quote = r
		case '(', '[', '{':
			depth++
		case ')', ']', '}':
			if depth == 0 {
				return nil, false
			}
			depth--
		case ',':
			if depth == 0 {
				items = append(items, string(runes[start:i]))
				start = i + 1
			}
		}
	}
	if quote != 0 || depth != 0 {
		return nil, false
	}
	items = append(items, string(runes[start:]))
	for _, item := range items {
		if strings.TrimSpace(item) == "" {
			return nil, false
		}
	}
	return items, true
}

func leadingWhitespace(s string) string {
	return s[:len(s)-len(strings.TrimLeft(s, " \t"))]
}
