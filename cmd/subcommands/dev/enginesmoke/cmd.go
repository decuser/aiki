package enginesmoke

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"aiki/engine"
	"aiki/engine/runtime/hal/substrate"
	"aiki/engine/runtime/help"
	"aiki/engine/runtime/prelude"
	"aiki/engine/semantics/evaluator"
	"aiki/engine/semantics/value"
	"aiki/engine/syntax"
	"aiki/engine/syntax/grammar"
)

type stage string

const (
	stageAll   stage = "all"
	stageLex   stage = "lex"
	stageParse stage = "parse"
	stageEval  stage = "eval"
	stageEbnf  stage = "ebnf"
)

func Run(args []string) int {
	fs := flag.NewFlagSet("enginesmoke", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	var stageName string
	var gold bool
	var check bool
	var grammarPath string

	fs.StringVar(&stageName, "stage", "all", "lex|parse|eval|ebnf|all")
	fs.BoolVar(&gold, "gold", false, "write .gold outputs")
	fs.BoolVar(&check, "check", false, "compare to .gold outputs")
	fs.StringVar(&grammarPath, "grammar", "", "override grammar path")

	if err := fs.Parse(args); err != nil {
		return 2
	}
	if gold && check {
		fmt.Fprintln(os.Stderr, "enginesmoke: cannot use --gold and --check together")
		return 2
	}
	st := stage(stageName)
	if st != stageAll && st != stageLex && st != stageParse && st != stageEval && st != stageEbnf {
		fmt.Fprintf(os.Stderr, "enginesmoke: invalid --stage %q\n", stageName)
		return 2
	}

	inputs, err := collectInputs(fs.Args())
	if err != nil {
		fmt.Fprintln(os.Stderr, "enginesmoke:", err)
		return 1
	}
	if st != stageEbnf && len(inputs) == 0 {
		fmt.Fprintln(os.Stderr, "enginesmoke: no *_engine.ai files found")
		return 1
	}

	g, err := loadGrammar(grammarPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "enginesmoke:", err)
		return 1
	}

	// Preload help registry (required for evaluator and some loaders).
	if err := initHelpRegistry(); err != nil {
		fmt.Fprintln(os.Stderr, "enginesmoke:", err)
		return 1
	}

	// Grammar dump is independent of inputs. Run it once.
	if st == stageEbnf || st == stageAll {
		out := dumpGrammar(g)
		goldPath := grammarGoldPath(grammarPath)
		if gold {
			if err := os.WriteFile(goldPath, out, 0o644); err != nil {
				fmt.Fprintf(os.Stderr, "enginesmoke FAIL: ebnf\n  writing gold: %v\n", err)
				return 1
			}
		} else if check {
			exp, err := os.ReadFile(goldPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "enginesmoke FAIL: ebnf\n  reading gold %s: %v\n", goldPath, err)
				return 1
			}
			if !bytes.Equal(bytes.TrimRight(exp, "\n"), bytes.TrimRight(out, "\n")) {
				fmt.Fprintf(os.Stderr, "enginesmoke FAIL: ebnf\n  gold mismatch: %s\n", goldPath)
				return 1
			}
		} else if st == stageEbnf {
			os.Stdout.Write(out)
			if len(out) > 0 && out[len(out)-1] != '\n' {
				fmt.Fprintln(os.Stdout)
			}
		}
		if st == stageEbnf {
			if gold {
				fmt.Fprintln(os.Stdout, "enginesmoke gold ok (ebnf)")
			}
			if check {
				fmt.Fprintln(os.Stdout, "enginesmoke ok (ebnf)")
			}
			return 0
		}
	}

	for _, inPath := range inputs {
		sourceBytes, err := os.ReadFile(inPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "enginesmoke FAIL: %s\n  reading input: %v\n", inPath, err)
			return 1
		}
		source := string(sourceBytes)

		stages := []stage{st}
		if st == stageAll {
			stages = []stage{stageLex, stageParse, stageEval}
		}

		for _, s := range stages {
			out, err := runStage(s, g, inPath, source)
			if err != nil {
				fmt.Fprintf(os.Stderr, "enginesmoke FAIL: %s (%s)\n  %v\n", inPath, s, err)
				return 1
			}

			goldPath := goldFileFor(inPath, s)
			if gold {
				if err := os.WriteFile(goldPath, out, 0o644); err != nil {
					fmt.Fprintf(os.Stderr, "enginesmoke FAIL: %s (%s)\n  writing gold: %v\n", inPath, s, err)
					return 1
				}
				continue
			}
			if check {
				exp, err := os.ReadFile(goldPath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "enginesmoke FAIL: %s (%s)\n  reading gold %s: %v\n", inPath, s, goldPath, err)
					return 1
				}
				if !bytes.Equal(bytes.TrimRight(exp, "\n"), bytes.TrimRight(out, "\n")) {
					fmt.Fprintf(os.Stderr, "enginesmoke FAIL: %s (%s)\n  gold mismatch: %s\n", inPath, s, goldPath)
					return 1
				}
			}
			if !gold && !check {
				os.Stdout.Write(out)
				if len(out) > 0 && out[len(out)-1] != '\n' {
					fmt.Fprintln(os.Stdout)
				}
			}
		}
	}

	if gold {
		fmt.Fprintf(os.Stdout, "enginesmoke gold ok (%d inputs)\n", len(inputs))
		return 0
	}
	if check {
		fmt.Fprintf(os.Stdout, "enginesmoke ok (%d inputs)\n", len(inputs))
		return 0
	}
	return 0
}

func collectInputs(args []string) ([]string, error) {
	if len(args) == 0 {
		args = []string{"."}
	}
	var out []string
	seen := make(map[string]bool)
	for _, root := range args {
		if root == "./..." {
			root = "."
		}
		info, err := os.Stat(root)
		if err == nil && !info.IsDir() {
			if strings.HasSuffix(root, "_engine.ai") {
				abs := root
				if !seen[abs] {
					seen[abs] = true
					out = append(out, abs)
				}
			}
			continue
		}
		err = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if strings.HasSuffix(path, "_engine.ai") {
				if !seen[path] {
					seen[path] = true
					out = append(out, path)
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	sort.Strings(out)
	return out, nil
}

func goldFileFor(inputPath string, s stage) string {
	return inputPath + "." + string(s) + ".gold"
}

func grammarGoldPath(grammarOverride string) string {
	if grammarOverride == "" {
		return filepath.Join("engine", "syntax", "grammar.ebnfx.ebnf.gold")
	}
	return grammarOverride + ".ebnf.gold"
}

func loadGrammar(overridePath string) (*grammar.Grammar, error) {
	if overridePath == "" {
		return grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
	}
	data, err := os.ReadFile(overridePath)
	if err != nil {
		return nil, fmt.Errorf("reading grammar: %w", err)
	}
	// Keep help stable by using embedded help, but allow overriding help later if needed.
	return grammar.Load(overridePath, string(data), "grammar.help", syntax.HelpSource)
}

func initHelpRegistry() error {
	registry := help.NewRegistry()

	funcs, err := help.ParseHelpFile("prelude.help", prelude.HelpSource)
	if err != nil {
		return err
	}

	docs, err := help.ParseDocFile("prelude.doc", prelude.DocSource)
	if err != nil {
		return err
	}

	registry.Merge(funcs, docs)
	substrate.HelpRegistry = registry
	return nil
}

func runStage(s stage, g *grammar.Grammar, file, source string) ([]byte, error) {
	switch s {
	case stageEbnf:
		return dumpGrammar(g), nil
	case stageLex:
		return dumpLex(g, file, source)
	case stageParse:
		return dumpParse(g, file, source)
	case stageEval:
		return dumpEval(g, file, source)
	default:
		return nil, fmt.Errorf("unknown stage: %s", s)
	}
}

func dumpLex(g *grammar.Grammar, file, source string) ([]byte, error) {
	obs := &tokenDumpObserver{}
	lexer := syntax.NewLexer(g, file, source, obs)
	_, err := lexer.Tokenize()
	if err != nil {
		return nil, err
	}
	return obs.Bytes(), nil
}

func dumpParse(g *grammar.Grammar, file, source string) ([]byte, error) {
	lexer := syntax.NewLexer(g, file, source, nil)
	tokens, err := lexer.Tokenize()
	if err != nil {
		return nil, err
	}
	parser := syntax.NewParser(g, tokens, source, nil)
	ast, err := parser.Parse()
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	dumpNode(&buf, ast, 0)
	return buf.Bytes(), nil
}

func dumpEval(g *grammar.Grammar, file, source string) ([]byte, error) {
	// Redirect runtime stdout for deterministic capture.
	var out bytes.Buffer
	oldStdout := substrate.Stdout
	substrate.Stdout = &out
	defer func() { substrate.Stdout = oldStdout }()

	// Load prelude into a fresh environment.
	rt := substrate.NewGoRuntime()
	preludeEnv := value.NewEnvWithScope(value.ScopePrelude)
	if err := loadPrelude(g, rt, preludeEnv); err != nil {
		return nil, err
	}
	userEnv := value.NewEnclosedEnv(preludeEnv)
	userEnv.SetFile(file)
	userEnv.SetSource(source)

	lexer := syntax.NewLexer(g, file, source, nil)
	tokens, err := lexer.Tokenize()
	if err != nil {
		return nil, err
	}
	parser := syntax.NewParser(g, tokens, source, nil)
	ast, err := parser.Parse()
	if err != nil {
		return nil, err
	}

	ev := evaluator.New(rt, nil)
	ev.SetGrammar(g)
	result := ev.Eval(ast, userEnv)
	if errV, ok := result.(*value.Error); ok {
		return nil, fmt.Errorf("%s", errV.Inspect())
	}
	if faultV, ok := result.(*value.Fault); ok {
		return nil, fmt.Errorf("%s", faultV.Inspect())
	}

	return out.Bytes(), nil
}

func loadPrelude(g *grammar.Grammar, rt *substrate.GoRuntime, env *value.Env) error {
	lexer := syntax.NewLexer(g, "<prelude>", prelude.Source, nil)
	tokens, err := lexer.Tokenize()
	if err != nil {
		return err
	}
	parser := syntax.NewParser(g, tokens, prelude.Source, nil)
	ast, err := parser.Parse()
	if err != nil {
		return err
	}

	env.SetFile("<prelude>")
	env.SetSource(prelude.Source)
	ev := evaluator.New(rt, nil)
	ev.SetGrammar(g)
	result := ev.Eval(ast, env)
	if errV, ok := result.(*value.Error); ok {
		return fmt.Errorf("%s", errV.Inspect())
	}
	if faultV, ok := result.(*value.Fault); ok {
		return fmt.Errorf("%s", faultV.Inspect())
	}
	return nil
}

type tokenDumpObserver struct {
	buf bytes.Buffer
}

func (t *tokenDumpObserver) OnLex(token string, lexeme string, pos engine.Position) {
	// Stable, one token per line.
	// Format: line:col<TAB>TYPE<TAB>%q(lexeme)
	fmt.Fprintf(&t.buf, "%d:%d\t%s\t%s\n", pos.Line, pos.Col, token, strconv.Quote(lexeme))
}

func (t *tokenDumpObserver) OnParse(production string, depth int, pos engine.Position) {
	_ = production
	_ = depth
	_ = pos
}

func (t *tokenDumpObserver) OnEval(node string, result string, scope int, pos engine.Position) {
	_ = node
	_ = result
	_ = scope
	_ = pos
}

func (t *tokenDumpObserver) OnEffect(action string, target string, pos engine.Position) {
	_ = action
	_ = target
	_ = pos
}

func (t *tokenDumpObserver) Bytes() []byte {
	return t.buf.Bytes()
}

func dumpNode(w io.Writer, n *syntax.Node, depth int) {
	if n == nil {
		return
	}
	indent := strings.Repeat("  ", depth)
	if n.Value != "" {
		fmt.Fprintf(w, "%s%s %d:%d %s\n", indent, n.Type, n.Pos.Line, n.Pos.Col, strconv.Quote(n.Value))
	} else {
		fmt.Fprintf(w, "%s%s %d:%d\n", indent, n.Type, n.Pos.Line, n.Pos.Col)
	}
	for _, c := range n.Children {
		dumpNode(w, c, depth+1)
	}
}

func dumpGrammar(g *grammar.Grammar) []byte {
	var buf bytes.Buffer

	// Tokens
	toks := make([]grammar.TokenDef, len(g.Tokens))
	copy(toks, g.Tokens)
	sort.Slice(toks, func(i, j int) bool { return toks[i].Name < toks[j].Name })
	buf.WriteString("TOKENS\n")
	for _, t := range toks {
		kind := "pattern"
		val := ""
		if t.Literal != "" {
			kind = "literal"
			val = t.Literal
		} else if t.Pattern != nil {
			val = t.Pattern.String()
		}
		fmt.Fprintf(&buf, "  %s\tskip=%t\t%s=%s", t.Name, t.Skip, kind, strconv.Quote(val))
		if t.Meta.Error != "" {
			fmt.Fprintf(&buf, "\terror=%s", strconv.Quote(t.Meta.Error))
		}
		if t.Meta.Template != "" {
			fmt.Fprintf(&buf, "\ttemplate=%s", strconv.Quote(t.Meta.Template))
		}
		buf.WriteByte('\n')
	}

	// Productions
	buf.WriteString("\nPRODUCTIONS\n")
	var names []string
	for name := range g.Productions {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		p := g.Productions[name]
		expr := exprString(p.Expr)
		fmt.Fprintf(&buf, "  %s\t= %s", name, expr)
		if p.Meta.Error != "" {
			fmt.Fprintf(&buf, "\terror=%s", strconv.Quote(p.Meta.Error))
		}
		if p.Meta.Template != "" {
			fmt.Fprintf(&buf, "\ttemplate=%s", strconv.Quote(p.Meta.Template))
		}
		buf.WriteByte('\n')
	}

	// Hashes (computed over the normalized content above, excluding the hash lines).
	norm := buf.Bytes()
	sum := sha256.Sum256(norm)
	hexSum := hex.EncodeToString(sum[:])
	buf.WriteString("\nGRAMMARHASH ")
	buf.WriteString(hexSum)
	buf.WriteByte('\n')

	// Per production hashes.
	buf.WriteString("RULEHASHES\n")
	for _, name := range names {
		p := g.Productions[name]
		line := name + "=" + exprString(p.Expr)
		s := sha256.Sum256([]byte(line))
		fmt.Fprintf(&buf, "  %s %s\n", name, hex.EncodeToString(s[:]))
	}

	return buf.Bytes()
}

func exprString(e grammar.Expression) string {
	switch x := e.(type) {
	case *grammar.Sequence:
		parts := make([]string, 0, len(x.Exprs))
		for _, sub := range x.Exprs {
			parts = append(parts, exprString(sub))
		}
		return strings.Join(parts, " ")
	case *grammar.Alternative:
		parts := make([]string, 0, len(x.Exprs))
		for _, sub := range x.Exprs {
			parts = append(parts, exprString(sub))
		}
		return strings.Join(parts, " | ")
	case *grammar.Repetition:
		return "{" + exprString(x.Expr) + "}"
	case *grammar.Option:
		return "[" + exprString(x.Expr) + "]"
	case *grammar.Group:
		return "(" + exprString(x.Expr) + ")"
	case *grammar.Terminal:
		return strconv.Quote(x.Value)
	case *grammar.Reference:
		return x.Name
	case *grammar.TokenRef:
		return x.Name
	default:
		return fmt.Sprintf("<unknown:%T>", e)
	}
}
