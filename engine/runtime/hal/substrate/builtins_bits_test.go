package substrate

import (
	"math/big"
	"testing"

	"aiki/engine/semantics/value"
)

func bitsN(n int64) *value.Number { return value.NewNumber(n, 1) }

func bitsWantNumber(t *testing.T, got value.Value, want int64) {
	t.Helper()
	n, ok := got.(*value.Number)
	if !ok {
		t.Fatalf("expected number, got %T (%s)", got, got.Inspect())
	}
	if n.Val.Cmp(bitsN(want).Val) != 0 {
		t.Fatalf("want %d, got %s", want, n.Inspect())
	}
}

func bitsWantFault(t *testing.T, got value.Value) {
	t.Helper()
	if _, ok := got.(*value.Fault); !ok {
		t.Fatalf("expected fault, got %T (%s)", got, got.Inspect())
	}
}

func TestBitsCoreOperations(t *testing.T) {
	bitsWantNumber(t, halBitsAnd([]value.Value{bitsN(12), bitsN(10)}, nil), 8)
	bitsWantNumber(t, halBitsOr([]value.Value{bitsN(12), bitsN(10)}, nil), 14)
	bitsWantNumber(t, halBitsXor([]value.Value{bitsN(12), bitsN(10)}, nil), 6)
	bitsWantNumber(t, halBitsNot([]value.Value{bitsN(15), bitsN(8)}, nil), 240)
	bitsWantNumber(t, halBitsShl([]value.Value{bitsN(1), bitsN(8)}, nil), 256)
	bitsWantNumber(t, halBitsShr([]value.Value{bitsN(256), bitsN(8)}, nil), 1)
}

func TestBitsNotMasksToWidth(t *testing.T) {
	bitsWantNumber(t, halBitsNot([]value.Value{bitsN(511), bitsN(8)}, nil), 0)
	bitsWantNumber(t, halBitsNot([]value.Value{bitsN(99), bitsN(0)}, nil), 0)
}

func TestBitsArbitraryPrecision(t *testing.T) {
	bigVal := new(value.Number)
	bigVal.Val = new(big.Rat).SetInt(new(big.Int).Lsh(big.NewInt(1), 100))
	got := halBitsShr([]value.Value{bigVal, bitsN(99)}, nil)
	bitsWantNumber(t, got, 2)
}

func TestBitsRejectNegativeAndFractional(t *testing.T) {
	bitsWantFault(t, halBitsAnd([]value.Value{bitsN(-1), bitsN(1)}, nil))
	bitsWantFault(t, halBitsOr([]value.Value{value.NewNumber(1, 2), bitsN(1)}, nil))
	bitsWantFault(t, halBitsShl([]value.Value{bitsN(1), bitsN(-1)}, nil))
	bitsWantFault(t, halBitsShr([]value.Value{bitsN(-1), bitsN(1)}, nil))
	bitsWantFault(t, halBitsNot([]value.Value{bitsN(1), value.NewNumber(1, 2)}, nil))
}

func TestBitsRejectWrongArity(t *testing.T) {
	bitsWantFault(t, halBitsAnd([]value.Value{bitsN(1)}, nil))
	bitsWantFault(t, halBitsNot([]value.Value{bitsN(1)}, nil))
	bitsWantFault(t, halBitsShl([]value.Value{bitsN(1)}, nil))
}
