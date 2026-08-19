package substrate

import (
	"testing"

	"aiki/engine/semantics/value"
)

func TestHashProviderImmutableMapOperations(t *testing.T) {
	h := halHashNew(nil, nil)
	h1 := halHashPut([]value.Value{h, &value.String{Val: "a"}, value.NewNumber(1, 1)}, nil)
	if got := halHashHas([]value.Value{h, &value.String{Val: "a"}}, nil); got != value.FALSE {
		t.Fatalf("put mutated original hash: %s", got.Inspect())
	}
	if got := halHashGet([]value.Value{h1, &value.String{Val: "a"}}, nil); got.Inspect() != "1" {
		t.Fatalf("get = %s, want 1", got.Inspect())
	}
	h2 := halHashDel([]value.Value{h1, &value.String{Val: "a"}}, nil)
	if got := halHashHas([]value.Value{h2, &value.String{Val: "a"}}, nil); got != value.FALSE {
		t.Fatalf("del retained key: %s", got.Inspect())
	}
}

func TestHashProviderHashCodeMatchesReferenceExamples(t *testing.T) {
	if got := halHashCode([]value.Value{&value.String{Val: "alpha"}}, nil); got.Inspect() != "92909918" {
		t.Fatalf("hash_code(alpha) = %s", got.Inspect())
	}
	shaped := &value.List{Shape: "point", Elements: []value.Value{value.NewNumber(3, 1), value.NewNumber(5, 1)}}
	if got := halHashCode([]value.Value{shaped}, nil); got.Inspect() == "0" {
		t.Fatalf("shaped hash unexpectedly zero")
	}
}
