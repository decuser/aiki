package substrate

import (
	"testing"

	"aiki/engine/semantics/evaluator"
	"aiki/engine/semantics/value"
)

func mustListInts(t *testing.T, v value.Value, want ...int64) *value.List {
	t.Helper()
	l, ok := v.(*value.List)
	if !ok {
		t.Fatalf("got %T, want list", v)
	}
	if l.Len() != len(want) {
		t.Fatalf("length got %d want %d", l.Len(), len(want))
	}
	for i, w := range want {
		n, ok := l.At(i).(*value.Number)
		if !ok || !n.IsInt() || n.Int64Value() != w {
			t.Fatalf("element %d got %v want %d", i, l.At(i), w)
		}
	}
	return l
}

func TestAdaptiveAppendRestDerivedListCopiesBeforePromotion(t *testing.T) {
	base := &value.List{Elements: []value.Value{value.NewNumber(1, 1), value.NewNumber(2, 1), value.NewNumber(3, 1)}}
	frontier := halAppend([]value.Value{base, value.NewNumber(4, 1)}, nil)
	rest := halRest([]value.Value{frontier}, nil)
	derived := halAppend([]value.Value{rest, value.NewNumber(9, 1)}, nil)
	originalNext := halAppend([]value.Value{frontier, value.NewNumber(5, 1)}, nil)

	mustListInts(t, frontier, 1, 2, 3, 4)
	mustListInts(t, rest, 2, 3, 4)
	mustListInts(t, derived, 2, 3, 4, 9)
	mustListInts(t, originalNext, 1, 2, 3, 4, 5)
}

func TestAdaptiveAppendRecordsRealizationWithoutEvalContext(t *testing.T) {
	counters := evaluator.NewCounters()
	var v value.Value = &value.List{Elements: []value.Value{}}
	for i := int64(0); i < 20; i++ {
		v = halAppendWithProbe([]value.Value{v, value.NewNumber(i, 1)}, counters)
	}
	got := counters.ListSnapshot()
	if got.FrontierPromoted != 1 {
		t.Fatalf("promoted got %d want 1", got.FrontierPromoted)
	}
	if got.FrontierExtended != 19 {
		t.Fatalf("extended got %d want 19", got.FrontierExtended)
	}
	if got.FrontierGrown == 0 {
		t.Fatal("expected geometric growth")
	}
	if got.FrontierForked != 0 {
		t.Fatalf("forked got %d want 0", got.FrontierForked)
	}
	if got.ElementsCopied >= 20*19/2 {
		t.Fatalf("copied %d slots; expected subquadratic realization", got.ElementsCopied)
	}
}
