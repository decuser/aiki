package substrate

import (
	"bytes"
	"testing"
)

func TestCanvasProtoRoundTripCmd(t *testing.T) {
	var b bytes.Buffer
	in := CanvasWireCmd{Op: "line", Args: []int32{1, 2, 3, 4}, HasRGBA: true, RGBA: DefaultFG, Pen: 5}
	if err := CanvasWriteFrame(&b, in); err != nil {
		t.Fatalf("write: %v", err)
	}
	outAny, err := CanvasReadCommand(bytes.NewReader(b.Bytes()))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	out, ok := outAny.(CanvasWireCmd)
	if !ok {
		t.Fatalf("type: %T", outAny)
	}
	if out.Op != in.Op {
		t.Fatalf("op: %q", out.Op)
	}
	if out.Pen != in.Pen {
		t.Fatalf("pen: %v", out.Pen)
	}
	if len(out.Args) != len(in.Args) {
		t.Fatalf("args len: %d", len(out.Args))
	}
	for i := range out.Args {
		if out.Args[i] != in.Args[i] {
			t.Fatalf("args[%d]: %d", i, out.Args[i])
		}
	}
	if !out.HasRGBA {
		t.Fatalf("expected rgba")
	}
	if out.RGBA != in.RGBA {
		t.Fatalf("rgba")
	}
}

func TestCanvasProtoRoundTripSetBG(t *testing.T) {
	var b bytes.Buffer
	in := CanvasWireSetBG{RGBA: DefaultBG}
	if err := CanvasWriteFrame(&b, in); err != nil {
		t.Fatalf("write: %v", err)
	}
	outAny, err := CanvasReadCommand(bytes.NewReader(b.Bytes()))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	out, ok := outAny.(CanvasWireSetBG)
	if !ok {
		t.Fatalf("type: %T", outAny)
	}
	if out.RGBA != in.RGBA {
		t.Fatalf("rgba")
	}
}
