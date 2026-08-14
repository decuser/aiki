package substrate

import "testing"

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
