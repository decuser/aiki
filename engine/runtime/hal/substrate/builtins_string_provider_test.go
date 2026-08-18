package substrate

import (
	"testing"

	"aiki/engine/semantics/value"
)

func TestStringProviderUsesRunePositions(t *testing.T) {
	got := halStringSubstring([]value.Value{
		&value.String{Val: "aé猫z"}, value.NewNumber(1, 1), value.NewNumber(3, 1),
	}, nil)
	if got.Inspect() != `"é猫"` {
		t.Fatalf("substring = %s, want %q", got.Inspect(), "é猫")
	}

	got = halStringIndexOf([]value.Value{
		&value.String{Val: "café latte"}, &value.String{Val: "latte"},
	}, nil)
	if got.Inspect() != "5" {
		t.Fatalf("index_of = %s, want 5", got.Inspect())
	}
}

func TestStringProviderSplitPreservesUnicode(t *testing.T) {
	got := halStringSplit([]value.Value{
		&value.String{Val: "café猫latte猫fin"}, &value.String{Val: "猫"},
	}, nil)
	want := `["café", "latte", "fin"]`
	if got.Inspect() != want {
		t.Fatalf("split = %s, want %s", got.Inspect(), want)
	}
}

func TestStringProviderWhitespaceMatchesAikiContract(t *testing.T) {
	got := halStringTrim([]value.Value{&value.String{Val: " \tfoo\r\n"}}, nil)
	if got.Inspect() != `"foo"` {
		t.Fatalf("trim = %s, want %q", got.Inspect(), "foo")
	}
	got = halStringTrim([]value.Value{&value.String{Val: "\u00a0foo\u00a0"}}, nil)
	trimmed, ok := got.(*value.String)
	if !ok {
		t.Fatalf("trim unicode space returned %T, want *value.String", got)
	}
	if trimmed.Val != "\u00a0foo\u00a0" {
		t.Fatalf("trim unicode space = %q; Aiki trim must remain ASCII-whitespace-only", trimmed.Val)
	}
}
