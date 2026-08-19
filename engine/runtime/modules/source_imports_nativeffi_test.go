package modules

import (
	"strings"
	"testing"
)

func TestAnalyzeSourceCollectsLiteralImportsAndUsesNativeFFIPass(t *testing.T) {
	g := testGrammar(t)
	source := `package "sample"
let dep = import("string/ffi")
use("bytes")
let f = () {
    let nested = import("./local")
    nested
}
export(:f)
`
	info, err := AnalyzeSource(g, "sample.ai", source)
	if err != nil {
		t.Fatalf("AnalyzeSource: %v", err)
	}
	want := []string{"string/ffi", "bytes", "./local"}
	if strings.Join(info.Imports, "\x00") != strings.Join(want, "\x00") {
		t.Fatalf("imports = %v, want %v", info.Imports, want)
	}
}

func TestAnalyzeSourceCollectsFunctionCallsNativeFFIPass(t *testing.T) {
	g := testGrammar(t)
	source := `package "sample"
let helper = (x) { _provider(x) }
let outer = (x) { helper(x) }
export(:outer)
`
	info, err := AnalyzeSource(g, "sample.ai", source)
	if err != nil {
		t.Fatalf("AnalyzeSource: %v", err)
	}
	if strings.Join(info.Functions["helper"].Calls, "\x00") != "_provider" {
		t.Fatalf("helper calls = %v", info.Functions["helper"].Calls)
	}
	if strings.Join(info.Functions["outer"].Calls, "\x00") != "helper" {
		t.Fatalf("outer calls = %v", info.Functions["outer"].Calls)
	}
}
