package value

import (
	"math"
	"math/big"
	"math/rand"
	"testing"
)

func requireNumberEqualRat(t *testing.T, got *Number, want *big.Rat) {
	t.Helper()
	if got.Rat().Cmp(want) != 0 {
		t.Fatalf("got %s, want %s", got.RatString(), want.RatString())
	}
}

func TestAdaptiveSmallIntegerExactness(t *testing.T) {
	cases := [][2]int64{
		{0, 0}, {1, 2}, {-7, 11}, {1<<62 - 1, 17}, {-1 << 62, -19},
	}
	for _, tc := range cases {
		a, b := tc[0], tc[1]
		na, nb := NewNumber(a, 1), NewNumber(b, 1)
		requireNumberEqualRat(t, na.Add(nb), new(big.Rat).Add(big.NewRat(a, 1), big.NewRat(b, 1)))
		requireNumberEqualRat(t, na.Sub(nb), new(big.Rat).Sub(big.NewRat(a, 1), big.NewRat(b, 1)))
		requireNumberEqualRat(t, na.Mul(nb), new(big.Rat).Mul(big.NewRat(a, 1), big.NewRat(b, 1)))
	}
}

func TestAdaptiveIntegerOverflowPromotesExactly(t *testing.T) {
	max := int64(^uint64(0) >> 1)
	min := -max - 1
	requireNumberEqualRat(t, NewNumber(max, 1).Add(NewNumber(1, 1)), new(big.Rat).Add(big.NewRat(max, 1), big.NewRat(1, 1)))
	requireNumberEqualRat(t, NewNumber(min, 1).Sub(NewNumber(1, 1)), new(big.Rat).Sub(big.NewRat(min, 1), big.NewRat(1, 1)))
	requireNumberEqualRat(t, NewNumber(max, 1).Mul(NewNumber(2, 1)), new(big.Rat).Mul(big.NewRat(max, 1), big.NewRat(2, 1)))
	requireNumberEqualRat(t, NewNumber(min, 1).Neg(), new(big.Rat).Neg(big.NewRat(min, 1)))
}

func TestAdaptiveCompactRationalDifferential(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	for i := 0; i < 20000; i++ {
		a := rng.Int63n(2_000_001) - 1_000_000
		b := rng.Int63n(999_999) + 1
		c := rng.Int63n(2_000_001) - 1_000_000
		d := rng.Int63n(999_999) + 1
		left, right := NewNumber(a, b), NewNumber(c, d)
		lr, rr := big.NewRat(a, b), big.NewRat(c, d)
		requireNumberEqualRat(t, left.Add(right), new(big.Rat).Add(lr, rr))
		requireNumberEqualRat(t, left.Sub(right), new(big.Rat).Sub(lr, rr))
		requireNumberEqualRat(t, left.Mul(right), new(big.Rat).Mul(lr, rr))
		if c != 0 {
			requireNumberEqualRat(t, left.Quo(right), new(big.Rat).Quo(lr, rr))
		}
		wantCmp := lr.Cmp(rr)
		gotCmp := left.Compare(right)
		if (gotCmp < 0) != (wantCmp < 0) || (gotCmp > 0) != (wantCmp > 0) {
			t.Fatalf("compare %s and %s: got %d want %d", left.RatString(), right.RatString(), gotCmp, wantCmp)
		}
	}
}

func TestBinary64CarrierIsExactDyadicValue(t *testing.T) {
	n, ok := NewNumberFromFloat64(0.1)
	if !ok {
		t.Fatal("finite binary64 rejected")
	}
	want := new(big.Rat).SetFloat64(0.1)
	requireNumberEqualRat(t, n, want)
	if n.Equal(NewNumber(1, 10)) {
		t.Fatal("binary64 0.1 must denote its exact dyadic value, not decimal 1/10")
	}
	f, exact := n.Float64()
	if !exact || math.Float64bits(f) != math.Float64bits(0.1) {
		t.Fatalf("carrier did not round-trip exact host bits: %v %v", f, exact)
	}
}

func TestBinary64CarrierDoesNotAuthorizeRoundedAddition(t *testing.T) {
	a, _ := NewNumberFromFloat64(0.1)
	b, _ := NewNumberFromFloat64(0.2)
	got := a.Add(b)
	want := new(big.Rat).Add(new(big.Rat).SetFloat64(0.1), new(big.Rat).SetFloat64(0.2))
	requireNumberEqualRat(t, got, want)

	hardwareRounded := new(big.Rat).SetFloat64(0.1 + 0.2)
	if got.Rat().Cmp(hardwareRounded) == 0 {
		t.Fatal("ordinary Number addition silently followed rounded binary64 arithmetic")
	}
}

func TestBinary64CarrierRejectsNonFinite(t *testing.T) {
	for _, f := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		if _, ok := NewNumberFromFloat64(f); ok {
			t.Fatalf("accepted non-finite value %v", f)
		}
	}
}

func TestBinary64CertifiedExactArithmetic(t *testing.T) {
	cases := []struct {
		name string
		a    float64
		b    float64
		op   string
	}{
		{"add", 0.5, 0.25, "+"},
		{"sub", 0.75, 0.25, "-"},
		{"mul", 0.5, 0.25, "*"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a, _ := NewNumberFromFloat64(tc.a)
			b, _ := NewNumberFromFloat64(tc.b)
			var got *Number
			var want *big.Rat
			switch tc.op {
			case "+":
				got = a.Add(b)
				want = new(big.Rat).Add(new(big.Rat).SetFloat64(tc.a), new(big.Rat).SetFloat64(tc.b))
			case "-":
				got = a.Sub(b)
				want = new(big.Rat).Sub(new(big.Rat).SetFloat64(tc.a), new(big.Rat).SetFloat64(tc.b))
			case "*":
				got = a.Mul(b)
				want = new(big.Rat).Mul(new(big.Rat).SetFloat64(tc.a), new(big.Rat).SetFloat64(tc.b))
			}
			requireNumberEqualRat(t, got, want)
			if got.kind != numberBinary64 && got.kind != numberSmallInt {
				t.Fatalf("certified exact %s left binary fast path: kind=%v", tc.op, got.kind)
			}
		})
	}
}

func TestBinary64RoundedArithmeticStillFallsBack(t *testing.T) {
	a, _ := NewNumberFromFloat64(0.1)
	b, _ := NewNumberFromFloat64(0.2)
	got := a.Add(b)
	if got.kind == numberBinary64 {
		t.Fatal("rounded 0.1 + 0.2 was incorrectly retained as binary64")
	}
	want := new(big.Rat).Add(new(big.Rat).SetFloat64(0.1), new(big.Rat).SetFloat64(0.2))
	requireNumberEqualRat(t, got, want)
}

func TestBinary64CertificationNeverClaimsRoundedResult(t *testing.T) {
	rng := rand.New(rand.NewSource(2))
	checkedAdd := 0
	checkedMul := 0
	for i := 0; i < 50000; i++ {
		a := math.Float64frombits(rng.Uint64())
		b := math.Float64frombits(rng.Uint64())
		if math.IsNaN(a) || math.IsInf(a, 0) || math.IsNaN(b) || math.IsInf(b, 0) {
			continue
		}
		ar := new(big.Rat).SetFloat64(a)
		br := new(big.Rat).SetFloat64(b)
		if ar == nil || br == nil {
			continue
		}
		if got, exact := binary64ExactAdd(a, b); exact {
			gr := new(big.Rat).SetFloat64(got)
			want := new(big.Rat).Add(ar, br)
			if gr == nil || gr.Cmp(want) != 0 {
				t.Fatalf("addition falsely certified exact: a=%g b=%g got=%g", a, b, got)
			}
			checkedAdd++
		}
		if got, exact := binary64ExactMul(a, b); exact {
			gr := new(big.Rat).SetFloat64(got)
			want := new(big.Rat).Mul(ar, br)
			if gr == nil || gr.Cmp(want) != 0 {
				t.Fatalf("multiplication falsely certified exact: a=%g b=%g got=%g", a, b, got)
			}
			checkedMul++
		}
	}
	if checkedAdd == 0 {
		t.Fatalf("certification corpus exercised no exact addition cases: add=%d mul=%d", checkedAdd, checkedMul)
	}
}

func TestBinary64CertificationSubnormalEdges(t *testing.T) {
	min := math.SmallestNonzeroFloat64
	cases := [][2]float64{
		{min, min},
		{min, -min},
		{math.Float64frombits(2), min},
		{math.Float64frombits(0x0010000000000000), -min},
	}
	for _, tc := range cases {
		a, b := tc[0], tc[1]
		if got, exact := binary64ExactAdd(a, b); exact {
			want := new(big.Rat).Add(new(big.Rat).SetFloat64(a), new(big.Rat).SetFloat64(b))
			gr := new(big.Rat).SetFloat64(got)
			if gr == nil || gr.Cmp(want) != 0 {
				t.Fatalf("subnormal addition falsely certified: a=%g b=%g got=%g", a, b, got)
			}
		}
		if got, exact := binary64ExactMul(a, b); exact {
			want := new(big.Rat).Mul(new(big.Rat).SetFloat64(a), new(big.Rat).SetFloat64(b))
			gr := new(big.Rat).SetFloat64(got)
			if gr == nil || gr.Cmp(want) != 0 {
				t.Fatalf("subnormal multiplication falsely certified: a=%g b=%g got=%g", a, b, got)
			}
		}
	}
}

func TestBinary64CompactFallbackConversionIsExact(t *testing.T) {
	cases := []float64{0.1, 0.2, 0.3, -0.1, 0.5, 1.5, 65535.25}
	for _, f := range cases {
		num, den, ok := binary64SmallFraction(f)
		if !ok {
			t.Fatalf("expected common binary64 %g to fit compact exact rational", f)
		}
		got := big.NewRat(num, den)
		want := new(big.Rat).SetFloat64(f)
		if want == nil || got.Cmp(want) != 0 {
			t.Fatalf("compact conversion %g: got %s want %v", f, got.RatString(), want)
		}
	}
	if _, _, ok := binary64SmallFraction(math.SmallestNonzeroFloat64); ok {
		t.Fatal("smallest subnormal unexpectedly fit compact rational denominator")
	}
}

func TestBinary64CompactFallbackNeverChangesValue(t *testing.T) {
	rng := rand.New(rand.NewSource(3))
	checked := 0
	for i := 0; i < 100000; i++ {
		f := math.Float64frombits(rng.Uint64())
		if math.IsNaN(f) || math.IsInf(f, 0) {
			continue
		}
		num, den, ok := binary64SmallFraction(f)
		if !ok {
			continue
		}
		got := big.NewRat(num, den)
		want := new(big.Rat).SetFloat64(f)
		if want == nil || got.Cmp(want) != 0 {
			t.Fatalf("binary64 compact conversion changed value: f=%g got=%s want=%v", f, got.RatString(), want)
		}
		checked++
	}
	if checked == 0 {
		t.Fatal("random corpus exercised no compact binary64 conversions")
	}
}
