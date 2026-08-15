package substrate

import (
	"bytes"
	"testing"

	"aiki/engine/runtime/hal"
	"aiki/engine/syntax"
	"aiki/engine/syntax/grammar"
)

func TestStripDocMarkers(t *testing.T) {
	in := "@preamble use(\"turtle\")\npencolor\n@unchecked\nSets the color.\n\npencolor(:red)"
	want := "pencolor\nSets the color.\n\npencolor(:red)"
	if got := stripDocMarkers(in); got != want {
		t.Fatalf("stripDocMarkers() = %q, want %q", got, want)
	}
}

func TestStripDocMarkersLeavesOrdinaryAtLines(t *testing.T) {
	in := "example\n  @unchecked\n@other marker\ntext"
	want := in
	if got := stripDocMarkers(in); got != want {
		t.Fatalf("stripDocMarkers() = %q, want %q", got, want)
	}
}

func TestModuleHelpCarriesPreamble(t *testing.T) {
	mh := &ModuleHelp{Preamble: `use("turtle/simple")`}
	if got, want := mh.Preamble, `use("turtle/simple")`; got != want {
		t.Fatalf("ModuleHelp.Preamble = %q, want %q", got, want)
	}
}

func TestOutputHelpUsesPagerHook(t *testing.T) {
	oldStdout := Stdout
	oldPageOutput := PageOutput
	defer func() {
		Stdout = oldStdout
		PageOutput = oldPageOutput
	}()

	var out bytes.Buffer
	Stdout = &out
	var got string
	PageOutput = func(text string) bool {
		got = text
		return true
	}

	outputHelp("long help\n")
	if got != "long help\n" {
		t.Fatalf("pager got %q", got)
	}
	if out.Len() != 0 {
		t.Fatalf("Stdout received %q despite handled pager output", out.String())
	}
}

func TestOutputHelpFallsBackToStdout(t *testing.T) {
	oldStdout := Stdout
	oldPageOutput := PageOutput
	defer func() {
		Stdout = oldStdout
		PageOutput = oldPageOutput
	}()

	var out bytes.Buffer
	Stdout = &out
	PageOutput = func(string) bool { return false }

	outputHelp("short help\n")
	if got, want := out.String(), "short help\n"; got != want {
		t.Fatalf("Stdout = %q, want %q", got, want)
	}
}

func TestGrammarNewlineHelpUsesDeclaredMetadata(t *testing.T) {
	oldStdout := Stdout
	oldPageOutput := PageOutput
	defer func() {
		Stdout = oldStdout
		PageOutput = oldPageOutput
	}()

	g, err := grammar.Load("grammar.ebnfx", syntax.EbnfxSource, "grammar.help", syntax.HelpSource)
	if err != nil {
		t.Fatal(err)
	}
	if g.Newline.Meta.Help == "" {
		t.Fatal("grammar newline policy has no @help metadata")
	}

	var out bytes.Buffer
	Stdout = &out
	PageOutput = func(string) bool { return false }

	ctx := &hal.EvalContext{Grammar: g}
	showHelp("newline", ctx)

	want := "newline\n  " + g.Newline.Meta.Help + "\n"
	if got := out.String(); got != want {
		t.Fatalf("newline help = %q, want %q", got, want)
	}
}
