package modules

import (
	"reflect"
	"testing"

	"aiki/engine/syntax"
	"aiki/engine/syntax/grammar"
)

func testGrammar(t *testing.T) *grammar.Grammar {
	t.Helper()
	g, err := grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
	if err != nil {
		t.Fatalf("loading grammar: %v", err)
	}
	return g
}

func TestAnalyzeSourceDerivesModuleSurfaceFromAST(t *testing.T) {
	source := `package "sample"
let alpha = (x, y) { x + y }
let resty = (first, ...rest) { first }
let value = 42
export(:alpha, :resty)
`
	info, err := AnalyzeSource(testGrammar(t), "sample.ai", source)
	if err != nil {
		t.Fatalf("AnalyzeSource: %v", err)
	}
	if info.Package != "sample" {
		t.Fatalf("package = %q, want sample", info.Package)
	}
	if !reflect.DeepEqual(info.Exports, []string{"alpha", "resty"}) {
		t.Fatalf("exports = %v", info.Exports)
	}
	alpha := info.Functions["alpha"]
	if !reflect.DeepEqual(alpha.Parameters, []string{"x", "y"}) || alpha.Rest != "" {
		t.Fatalf("alpha signature = %#v", alpha)
	}
	resty := info.Functions["resty"]
	if !reflect.DeepEqual(resty.Parameters, []string{"first"}) || resty.Rest != "rest" {
		t.Fatalf("resty signature = %#v", resty)
	}
	if _, ok := info.Functions["value"]; ok {
		t.Fatal("non-function binding classified as function")
	}
}

func TestAnalyzeSourceDoesNotPromoteNestedBindings(t *testing.T) {
	source := `package "sample"
let outer = () {
    let nested = (x) { x }
    nested
}
export(:outer)
`
	info, err := AnalyzeSource(testGrammar(t), "sample.ai", source)
	if err != nil {
		t.Fatalf("AnalyzeSource: %v", err)
	}
	if _, ok := info.Functions["outer"]; !ok {
		t.Fatal("outer function missing")
	}
	if _, ok := info.Functions["nested"]; ok {
		t.Fatal("nested function promoted to module surface")
	}
}
