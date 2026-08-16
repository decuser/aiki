package formatter

import (
	"strings"
	"testing"

	"aiki/engine"
	"aiki/engine/syntax"
	"aiki/engine/syntax/grammar"
)

func testGrammar(t *testing.T) *grammar.Grammar {
	t.Helper()
	g, err := grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
	if err != nil {
		t.Fatalf("load grammar: %v", err)
	}
	return g
}

func TestFormatSourceIdempotentAndParsePreserving(t *testing.T) {
	g := testGrammar(t)
	src := "# header\nlet x=1  #eol\nif x {\n  return x\n}\n"
	out, err := FormatSource(g, "test.ai", src)
	if err != nil {
		t.Fatalf("format: %v", err)
	}
	out2, err := FormatSource(g, "test.ai", out)
	if err != nil {
		t.Fatalf("format second pass: %v", err)
	}
	if out2 != out {
		t.Fatalf("format not idempotent\n--- out ---\n%s\n--- out2 ---\n%s", out, out2)
	}
	for _, want := range []string{"if x {\n", "let x = 1", "#eol", "# header"} {
		if !contains(out, want) {
			t.Fatalf("output missing %q:\n%s", want, out)
		}
	}
}

func contains(s, sub string) bool {
	if sub == "" {
		return true
	}
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func TestFormatSourceSelect(t *testing.T) {
	g := testGrammar(t)
	src := "select{let x=recv(a){println(x)}recv(b){println(:b)}default{println(:idle)}}\n"
	out, err := FormatSource(g, "test.ai", src)
	if err != nil {
		t.Fatalf("format: %v", err)
	}
	want := "select {\n\tlet x = recv(a) {\n\t\tprintln(x)\n\t}\n\trecv(b) {\n\t\tprintln(:b)\n\t}\n\tdefault {\n\t\tprintln(:idle)\n\t}\n}\n"
	if out != want {
		t.Fatalf("unexpected select format\n--- got ---\n%s\n--- want ---\n%s", out, want)
	}
}

func TestFormatterProductionCoverageMatchesGrammar(t *testing.T) {
	g := testGrammar(t)
	for _, name := range g.Analysis().ProductionNames() {
		if _, ok := productionPrinters[name]; ok {
			continue
		}
		if _, ok := handledByParent[name]; ok {
			continue
		}
		t.Errorf("grammar production %q has no formatter disposition", name)
	}
	productions := make(map[string]struct{}, len(g.Analysis().ProductionNames()))
	for _, name := range g.Analysis().ProductionNames() {
		productions[name] = struct{}{}
	}
	for name := range productionPrinters {
		if _, ok := productions[name]; !ok {
			t.Errorf("formatter dispatch %q is not a grammar production", name)
		}
	}
	for name := range handledByParent {
		if _, ok := productions[name]; !ok {
			t.Errorf("parent-handled formatter node %q is not a grammar production", name)
		}
		if _, dispatched := productionPrinters[name]; dispatched {
			t.Errorf("formatter node %q is both dispatched and parent-handled", name)
		}
	}
}

func TestFormatterUnknownLeafCannotDisappear(t *testing.T) {
	p := &printer{observer: engine.SilentObserver{}}
	p.printNode(&syntax.Node{Type: "FUTURE_LITERAL", Value: "lost"})
	if p.err == nil {
		t.Fatal("expected unknown leaf to produce formatter error")
	}
	if got := p.buf.String(); got != "" {
		t.Fatalf("unknown leaf emitted output %q", got)
	}
}

func TestFormatSourceConvergesForAdjacentTopLevelDeclarations(t *testing.T) {
	g := testGrammar(t)
	src := `package "test"

let names = ["a", "b"]

let primitive_names = ["export", "import"]

let env_get = (env, name) { env.get(name) }
let env_define = (env, name, value) { env.define(name, value) }
let env_assign = (env, name, value) { env.assign(name, value) }
`
	out, err := FormatSource(g, "test.ai", src)
	if err != nil {
		t.Fatalf("format: %v", err)
	}
	out2, err := FormatSource(g, "test.ai", out)
	if err != nil {
		t.Fatalf("format second pass: %v", err)
	}
	if out2 != out {
		t.Fatalf("format not fixed-point stable\n--- out ---\n%s\n--- out2 ---\n%s", out, out2)
	}
}

func TestFormatPreservesExplicitMultilineCollectionsAndCalls(t *testing.T) {
	g := testGrammar(t)
	src := `let xs = [
	alpha,
	beta,
	gamma
]

let f = (
	a,
	b,
	...rest
) {
	return g(
		a,
		b,
		rest
	)
}
`
	out, err := FormatSource(g, "test.ai", src)
	if err != nil {
		t.Fatalf("format: %v", err)
	}
	for _, want := range []string{
		"let xs = [\n\talpha,\n\tbeta,\n\tgamma\n]\n",
		"let f = (\n\ta,\n\tb,\n\t...rest\n) {\n",
		"return g(\n\t\ta,\n\t\tb,\n\t\trest\n\t)\n",
	} {
		if !contains(out, want) {
			t.Fatalf("multiline layout not preserved; missing %q:\n%s", want, out)
		}
	}
	out2, err := FormatSource(g, "test.ai", out)
	if err != nil {
		t.Fatalf("format second pass: %v", err)
	}
	if out2 != out {
		t.Fatalf("multiline format not fixed-point stable\n--- out ---\n%s\n--- out2 ---\n%s", out, out2)
	}
}

func TestFormatKeepsCompactCollectionsAndCallsCompact(t *testing.T) {
	g := testGrammar(t)
	src := "let xs = [alpha, beta, gamma]\nlet f = (a, b) { return g(a, b) }\n"
	out, err := FormatSource(g, "test.ai", src)
	if err != nil {
		t.Fatalf("format: %v", err)
	}
	for _, want := range []string{
		"let xs = [alpha, beta, gamma]\n",
		"let f = (a, b) {\n",
		"return g(a, b)\n",
	} {
		if !contains(out, want) {
			t.Fatalf("compact layout expanded unexpectedly; missing %q:\n%s", want, out)
		}
	}
}

func TestFormatDoesNotExpandCompactCallWithMultilineFunctionArgument(t *testing.T) {
	g := testGrammar(t)
	src := `run("name", () {
	let x = 1
	return x
})
`
	out, err := FormatSource(g, "test.ai", src)
	if err != nil {
		t.Fatalf("format: %v", err)
	}
	if strings.Contains(out, "run(\n") {
		t.Fatalf("compact call with multiline function argument expanded unexpectedly:\n%s", out)
	}
	if !strings.Contains(out, "run(\"name\", () {") {
		t.Fatalf("compact call head was not preserved:\n%s", out)
	}
}
