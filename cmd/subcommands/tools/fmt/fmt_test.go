package fmt

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"aiki/engine"
	"aiki/engine/syntax"
	"aiki/engine/syntax/grammar"
)

func TestFormatSource_IdempotentAndParsePreserving(t *testing.T) {
	g, err := grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
	if err != nil {
		t.Fatalf("load grammar: %v", err)
	}

	src := "# header\nlet x=1  #eol\nif x {\n  return x\n}\n"
	out, err := FormatSource(g, "test.ai", src)
	if err != nil {
		t.Fatalf("format: %v", err)
	}

	// Second pass should be identical.
	out2, err := FormatSource(g, "test.ai", out)
	if err != nil {
		t.Fatalf("format second pass: %v", err)
	}
	if out2 != out {
		t.Fatalf("format not idempotent\n--- out ---\n%s\n--- out2 ---\n%s", out, out2)
	}

	// Spot check that it normalized brace placement and spaces.
	if !contains(out, "if x {\n") {
		t.Fatalf("expected same-line brace in output, got:\n%s", out)
	}
	if !contains(out, "let x = 1") {
		t.Fatalf("expected spaced let binding in output, got:\n%s", out)
	}
	if !contains(out, "#eol") {
		t.Fatalf("expected to preserve eol comment, got:\n%s", out)
	}
	if !contains(out, "# header") {
		t.Fatalf("expected to preserve standalone comment, got:\n%s", out)
	}
}

func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && (func() bool {
		for i := 0; i+len(sub) <= len(s); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	})())
}

func TestFormatSourceSelect(t *testing.T) {
	g, err := grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
	if err != nil {
		t.Fatalf("load grammar: %v", err)
	}

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
	g, err := grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
	if err != nil {
		t.Fatalf("load grammar: %v", err)
	}

	for name := range g.Productions {
		if _, ok := productionPrinters[name]; ok {
			continue
		}
		if _, ok := handledByParent[name]; ok {
			continue
		}
		t.Errorf("grammar production %q has no formatter disposition", name)
	}
	for name := range productionPrinters {
		if _, ok := g.Productions[name]; !ok {
			t.Errorf("formatter dispatch %q is not a grammar production", name)
		}
	}
	for name := range handledByParent {
		if _, ok := g.Productions[name]; !ok {
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

func TestRunFailsOnUndeclaredMalformedFile(t *testing.T) {
	dir := t.TempDir()
	bad := filepath.Join(dir, "bad.ai")
	good := filepath.Join(dir, "good.ai")
	if err := os.WriteFile(bad, []byte("let x =\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(good, []byte("let y = 2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if code := Run([]string{dir + "/..."}); code != 1 {
		t.Fatalf("Run exit = %d, want 1", code)
	}
}

func TestFormatPathSkipsDeclaredParseNegative(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "negative.ai")
	if err := os.WriteFile(path, []byte("# @negative parse\nlet x =\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := FormatPath(dir+"/...", Config{}); err != nil {
		t.Fatalf("declared parse-negative was not skipped: %v", err)
	}
}

func TestFormatDirReportsAllUndeclaredMalformedFiles(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"a.ai", "b.ai"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("let x =\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	_, err := FormatPath(dir+"/...", Config{})
	if err == nil {
		t.Fatal("expected malformed files to fail formatting")
	}
	msg := err.Error()
	for _, name := range []string{"a.ai", "b.ai"} {
		if !strings.Contains(msg, name) {
			t.Fatalf("error %q does not mention %s", msg, name)
		}
	}
}
