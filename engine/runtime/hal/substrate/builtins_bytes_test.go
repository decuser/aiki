package substrate

import (
	"testing"

	"aiki/engine/semantics/value"
)

func TestBytesDigitsFromText(t *testing.T) {
	got := halBytesDigitsFromText([]value.Value{&value.String{Val: "01234"}}, nil)
	b, ok := got.(*value.Bytes)
	if !ok {
		t.Fatalf("digits_from_text: got %T", got)
	}
	want := []byte{0, 1, 2, 3, 4}
	if len(b.Val) != len(want) {
		t.Fatalf("length: got %d want %d", len(b.Val), len(want))
	}
	for i := range want {
		if b.Val[i] != want[i] {
			t.Fatalf("byte %d: got %d want %d", i, b.Val[i], want[i])
		}
	}
}

func TestBytesDigitsToText(t *testing.T) {
	got := halBytesDigitsToText([]value.Value{&value.Bytes{Val: []byte{0, 1, 2, 3, 4}}}, nil)
	s, ok := got.(*value.String)
	if !ok {
		t.Fatalf("digits_to_text: got %T", got)
	}
	if s.Val != "01234" {
		t.Fatalf("got %q", s.Val)
	}
}

func TestBytesDigitsRejectInvalidInput(t *testing.T) {
	requireFault(t, halBytesDigitsFromText([]value.Value{&value.String{Val: "01x"}}, nil))
	requireFault(t, halBytesDigitsToText([]value.Value{&value.Bytes{Val: []byte{0, 10}}}, nil))
}
