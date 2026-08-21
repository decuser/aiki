package value

import (
	"fmt"
	"math"
	"math/big"
	"math/bits"
)

type numberKind uint8

const (
	numberBigRat numberKind = iota
	numberSmallInt
	numberSmallRat
	numberBinary64
)

// NumberRepresentation is a Go-side observation category for profiling and
// tests. It is not an Aiki type and is never exposed through language
// semantics.
type NumberRepresentation uint8

const (
	NumberBigRational NumberRepresentation = iota
	NumberSmallInteger
	NumberCompactRational
	NumberBinaryCarrier
)

// ProfileRepresentation reports the private realization currently carrying
// this Number. Callers must use it only for observation; semantic behavior must
// never branch on this result outside Number itself.
func (n *Number) ProfileRepresentation() NumberRepresentation {
	if n == nil {
		return NumberSmallInteger
	}
	switch n.kind {
	case numberSmallInt:
		return NumberSmallInteger
	case numberSmallRat:
		return NumberCompactRational
	case numberBinary64:
		return NumberBinaryCarrier
	default:
		return NumberBigRational
	}
}

// Number is Aiki's exact numeric value. Representation is private and may vary;
// all observable behavior is defined by exact rational semantics.
type Number struct {
	kind numberKind
	num  int64
	den  int64
	rat  *big.Rat
}

func (n *Number) Type() Type { return NumberType }

func (n *Number) Inspect() string {
	if n == nil {
		return "0"
	}
	if n.kind == numberSmallInt {
		return fmt.Sprintf("%d", n.num)
	}
	if n.kind == numberSmallRat {
		return fmt.Sprintf("%d/%d", n.num, n.den)
	}
	r := n.bigRatRef()
	if r.IsInt() {
		return r.Num().String()
	}
	return r.RatString()
}

func NewNumber(num, denom int64) *Number {
	if n, ok := newSmallRational(num, denom); ok {
		return n
	}
	return normalizeBigRat(big.NewRat(num, denom))
}

func NewNumberFromString(s string) (*Number, error) {
	r := new(big.Rat)
	if _, ok := r.SetString(s); !ok {
		return nil, fmt.Errorf("invalid number: %s", s)
	}
	return normalizeBigRat(r), nil
}

// NewNumberFromRat constructs a Number from an exact rational value. The input
// is copied so callers cannot mutate Number representation through aliasing.
func NewNumberFromRat(r *big.Rat) *Number {
	if r == nil {
		return NewNumber(0, 1)
	}
	return normalizeBigRat(new(big.Rat).Set(r))
}

// NewNumberFromBigInt constructs an integral Number without narrowing it.
func NewNumberFromBigInt(i *big.Int) *Number {
	if i == nil {
		return NewNumber(0, 1)
	}
	if i.IsInt64() {
		return &Number{kind: numberSmallInt, num: i.Int64()}
	}
	return &Number{kind: numberBigRat, rat: new(big.Rat).SetInt(i)}
}

// NewNumberFromFloat64 retains a finite host binary64 result as an exact dyadic
// rational carrier. This does not admit rounded binary arithmetic into Aiki; it
// only preserves the exact finite bit pattern already returned by the host.
func NewNumberFromFloat64(f float64) (*Number, bool) {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return nil, false
	}
	if f == 0 {
		return &Number{kind: numberSmallInt, num: 0}, true
	}
	if f >= math.MinInt64 && f < 9223372036854775808.0 && math.Trunc(f) == f {
		i := int64(f)
		if float64(i) == f {
			return &Number{kind: numberSmallInt, num: i}, true
		}
	}
	return &Number{kind: numberBinary64, num: int64(math.Float64bits(f))}, true
}

func normalizeBigRat(r *big.Rat) *Number {
	if r.IsInt() && r.Num().IsInt64() {
		return &Number{kind: numberSmallInt, num: r.Num().Int64()}
	}
	if r.Num().IsInt64() && r.Denom().IsInt64() {
		return &Number{kind: numberSmallRat, num: r.Num().Int64(), den: r.Denom().Int64()}
	}
	return &Number{kind: numberBigRat, rat: r}
}

func (n *Number) bigRatRef() *big.Rat {
	if n == nil {
		return new(big.Rat)
	}
	if n.kind == numberSmallInt {
		return big.NewRat(n.num, 1)
	}
	if n.kind == numberSmallRat {
		return big.NewRat(n.num, n.den)
	}
	if n.kind == numberBinary64 {
		r := new(big.Rat).SetFloat64(math.Float64frombits(uint64(n.num)))
		if r == nil {
			return new(big.Rat)
		}
		return r
	}
	if n.rat == nil {
		return new(big.Rat)
	}
	return n.rat
}

// Rat returns a copy of the exact rational value. It is an explicit boundary
// adapter for algorithms that genuinely require arbitrary-precision rationals.
func (n *Number) Rat() *big.Rat { return new(big.Rat).Set(n.bigRatRef()) }

func (n *Number) RatString() string {
	if n != nil && n.kind == numberSmallInt {
		return fmt.Sprintf("%d", n.num)
	}
	if n != nil && n.kind == numberSmallRat {
		return fmt.Sprintf("%d/%d", n.num, n.den)
	}
	return n.bigRatRef().RatString()
}

func (n *Number) IsInt() bool {
	if n == nil || n.kind == numberSmallInt {
		return true
	}
	if n.kind == numberBinary64 {
		f := math.Float64frombits(uint64(n.num))
		return math.Trunc(f) == f
	}
	return n.bigRatRef().IsInt()
}

func (n *Number) Sign() int {
	if n == nil {
		return 0
	}
	if n.kind == numberBinary64 {
		f := math.Float64frombits(uint64(n.num))
		switch {
		case f < 0:
			return -1
		case f > 0:
			return 1
		default:
			return 0
		}
	}
	if n.kind == numberSmallRat {
		switch {
		case n.num < 0:
			return -1
		case n.num > 0:
			return 1
		default:
			return 0
		}
	}
	if n.kind == numberSmallInt {
		switch {
		case n.num < 0:
			return -1
		case n.num > 0:
			return 1
		default:
			return 0
		}
	}
	return n.bigRatRef().Sign()
}

func (n *Number) IsZero() bool { return n.Sign() == 0 }

func (n *Number) Compare(other *Number) int {
	if n != nil && other != nil && n.kind == numberBinary64 && other.kind == numberBinary64 {
		a := math.Float64frombits(uint64(n.num))
		b := math.Float64frombits(uint64(other.num))
		switch {
		case a < b:
			return -1
		case a > b:
			return 1
		default:
			return 0
		}
	}
	if n != nil && other != nil {
		if an, ad, aok := n.smallFraction(); aok {
			if bn, bd, bok := other.smallFraction(); bok {
				if left, ok := mulInt64(an, bd); ok {
					if right, ok := mulInt64(bn, ad); ok {
						switch {
						case left < right:
							return -1
						case left > right:
							return 1
						default:
							return 0
						}
					}
				}
			}
		}
	}
	if n != nil && other != nil && n.kind == numberSmallInt && other.kind == numberSmallInt {
		switch {
		case n.num < other.num:
			return -1
		case n.num > other.num:
			return 1
		default:
			return 0
		}
	}
	return n.bigRatRef().Cmp(other.bigRatRef())
}

func (n *Number) Equal(other *Number) bool { return n.Compare(other) == 0 }

func (n *Number) IsInt64() bool {
	return n != nil && (n.kind == numberSmallInt || (n.IsInt() && n.bigRatRef().Num().IsInt64()))
}
func (n *Number) IsUint64() bool {
	if n == nil || n.Sign() < 0 || !n.IsInt() {
		return false
	}
	if n.kind == numberSmallInt {
		return true
	}
	return n.bigRatRef().Num().IsUint64()
}

func (n *Number) Int64Value() int64 {
	if n == nil {
		return 0
	}
	if n.kind == numberSmallInt {
		return n.num
	}
	return n.bigRatRef().Num().Int64()
}

func (n *Number) Uint64Value() uint64 {
	if n == nil {
		return 0
	}
	if n.kind == numberSmallInt {
		return uint64(n.num)
	}
	return n.bigRatRef().Num().Uint64()
}

func (n *Number) Int64() (int64, bool) {
	if !n.IsInt64() {
		return 0, false
	}
	return n.Int64Value(), true
}

func (n *Number) Uint64() (uint64, bool) {
	if !n.IsUint64() {
		return 0, false
	}
	return n.Uint64Value(), true
}

// Float64 is an explicit host-numeric boundary. exact reports whether the
// returned binary64 value exactly denotes the Aiki number.
func (n *Number) Float64() (float64, bool) {
	if n != nil && n.kind == numberBinary64 {
		return math.Float64frombits(uint64(n.num)), true
	}
	return n.bigRatRef().Float64()
}

func (n *Number) Numerator() *big.Int {
	if n != nil && n.kind == numberSmallInt {
		return big.NewInt(n.num)
	}
	if n != nil && n.kind == numberSmallRat {
		return big.NewInt(n.num)
	}
	return new(big.Int).Set(n.bigRatRef().Num())
}

func (n *Number) Denominator() *big.Int {
	if n != nil && (n.kind == numberSmallInt) {
		return big.NewInt(1)
	}
	if n != nil && n.kind == numberSmallRat {
		return big.NewInt(n.den)
	}
	return new(big.Int).Set(n.bigRatRef().Denom())
}

func absUint64(v int64) uint64 {
	if v < 0 {
		return uint64(-(v + 1)) + 1
	}
	return uint64(v)
}

func gcdUint64(a, b uint64) uint64 {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

func newSmallRational(num, den int64) (*Number, bool) {
	if den == 0 {
		return nil, false
	}
	if num == 0 {
		return &Number{kind: numberSmallInt, num: 0}, true
	}
	g := gcdUint64(absUint64(num), absUint64(den))
	n := num / int64(g)
	d := den / int64(g)
	if d < 0 {
		const minInt64 = -1 << 63
		if n == minInt64 || d == minInt64 {
			return nil, false
		}
		n = -n
		d = -d
	}
	if d == 1 {
		return &Number{kind: numberSmallInt, num: n}, true
	}
	return &Number{kind: numberSmallRat, num: n, den: d}, true
}

func binary64SmallFraction(f float64) (int64, int64, bool) {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return 0, 0, false
	}
	if f == 0 {
		return 0, 1, true
	}
	negative := math.Signbit(f)
	mant, exp, ok := binary64DyadicParts(math.Abs(f))
	if !ok || mant == 0 {
		return 0, 0, false
	}
	var mag uint64
	var den int64 = 1
	if exp >= 0 {
		if exp >= 64 || bits.Len64(mant)+exp > 64 {
			return 0, 0, false
		}
		mag = mant << exp
	} else {
		shift := -exp
		// A compact denominator is a positive int64, so 2^63 does not fit.
		if shift > 62 {
			return 0, 0, false
		}
		mag = mant
		den = int64(uint64(1) << shift)
	}
	const maxInt64 = uint64(1<<63 - 1)
	if negative {
		if mag > uint64(1)<<63 {
			return 0, 0, false
		}
		if mag == uint64(1)<<63 {
			return -1 << 63, den, true
		}
		return -int64(mag), den, true
	}
	if mag > maxInt64 {
		return 0, 0, false
	}
	return int64(mag), den, true
}

func (n *Number) smallFraction() (int64, int64, bool) {
	if n == nil {
		return 0, 1, true
	}
	switch n.kind {
	case numberSmallInt:
		return n.num, 1, true
	case numberSmallRat:
		return n.num, n.den, true
	case numberBinary64:
		return binary64SmallFraction(math.Float64frombits(uint64(n.num)))
	default:
		return 0, 0, false
	}
}

func addSmallRational(an, ad, bn, bd int64) (*Number, bool) {
	g := gcdUint64(uint64(ad), uint64(bd))
	adg := ad / int64(g)
	bdg := bd / int64(g)
	left, ok := mulInt64(an, bdg)
	if !ok {
		return nil, false
	}
	right, ok := mulInt64(bn, adg)
	if !ok {
		return nil, false
	}
	num, ok := addInt64(left, right)
	if !ok {
		return nil, false
	}
	den, ok := mulInt64(adg, bd)
	if !ok || den <= 0 {
		return nil, false
	}
	return newSmallRational(num, den)
}

func mulSmallRational(an, ad, bn, bd int64) (*Number, bool) {
	g1 := gcdUint64(absUint64(an), uint64(bd))
	g2 := gcdUint64(absUint64(bn), uint64(ad))
	an /= int64(g1)
	bd /= int64(g1)
	bn /= int64(g2)
	ad /= int64(g2)
	num, ok := mulInt64(an, bn)
	if !ok {
		return nil, false
	}
	den, ok := mulInt64(ad, bd)
	if !ok || den <= 0 {
		return nil, false
	}
	return newSmallRational(num, den)
}

func addInt64(a, b int64) (int64, bool) {
	r := a + b
	if ((a ^ r) & (b ^ r)) < 0 {
		return 0, false
	}
	return r, true
}

func subInt64(a, b int64) (int64, bool) {
	r := a - b
	if ((a ^ b) & (a ^ r)) < 0 {
		return 0, false
	}
	return r, true
}

func mulInt64(a, b int64) (int64, bool) {
	if a == 0 || b == 0 {
		return 0, true
	}
	const minInt64 = -1 << 63
	if (a == -1 && b == minInt64) || (b == -1 && a == minInt64) {
		return 0, false
	}
	r := a * b
	if r/b != a {
		return 0, false
	}
	return r, true
}

func binary64ExactAdd(a, b float64) (float64, bool) {
	s := a + b
	if math.IsNaN(s) || math.IsInf(s, 0) {
		return 0, false
	}
	// Keep the proof away from the one normal-exponent band where a true
	// half-ulp residue can be smaller than the least subnormal binary64.
	const minNormal = 0x1p-1022
	if s != 0 && math.Abs(s) < 2*minNormal {
		return 0, false
	}
	// TwoSum computes the exact rounding residue of finite binary64 addition.
	// A zero residue proves that the hardware result is the exact dyadic sum.
	bb := s - a
	err := (a - (s - bb)) + (b - bb)
	return s, err == 0
}

func binary64DyadicParts(f float64) (uint64, int, bool) {
	bits64 := math.Float64bits(f)
	expBits := (bits64 >> 52) & 0x7ff
	frac := bits64 & ((uint64(1) << 52) - 1)
	if expBits == 0x7ff {
		return 0, 0, false
	}
	if expBits == 0 && frac == 0 {
		return 0, 0, true
	}
	var mant uint64
	var exp int
	if expBits == 0 {
		mant = frac
		exp = -1074
	} else {
		mant = (uint64(1) << 52) | frac
		exp = int(expBits) - 1023 - 52
	}
	trailing := bits.TrailingZeros64(mant)
	mant >>= trailing
	exp += trailing
	return mant, exp, true
}

func binary64ExactMul(a, b float64) (float64, bool) {
	p := a * b
	if math.IsNaN(p) || math.IsInf(p, 0) {
		return 0, false
	}
	if a == 0 || b == 0 {
		return p, true
	}
	am, ae, aok := binary64DyadicParts(a)
	bm, be, bok := binary64DyadicParts(b)
	if !aok || !bok || am == 0 || bm == 0 {
		return 0, false
	}
	hi, lo := bits.Mul64(am, bm)
	bitLen := bits.Len64(lo)
	if hi != 0 {
		bitLen = 64 + bits.Len64(hi)
	}
	if bitLen > 53 {
		return 0, false
	}
	lowExp := ae + be
	highExp := lowExp + bitLen - 1
	if highExp > 1023 {
		return 0, false
	}
	if highExp >= -1022 {
		return p, true
	}
	// Subnormals are integer multiples of 2^-1074 and carry at most 52 bits.
	if lowExp < -1074 || bitLen > 52 {
		return 0, false
	}
	return p, true
}

func newExactBinary64Result(f float64) *Number {
	if f == 0 {
		return &Number{kind: numberSmallInt, num: 0}
	}
	return &Number{kind: numberBinary64, num: int64(math.Float64bits(f))}
}

func (n *Number) Add(other *Number) *Number {
	if n != nil && other != nil && n.kind == numberBinary64 && other.kind == numberBinary64 {
		a := math.Float64frombits(uint64(n.num))
		b := math.Float64frombits(uint64(other.num))
		if r, exact := binary64ExactAdd(a, b); exact {
			return newExactBinary64Result(r)
		}
	}
	if an, ad, ok := n.smallFraction(); ok {
		if bn, bd, ok := other.smallFraction(); ok {
			if r, ok := addSmallRational(an, ad, bn, bd); ok {
				return r
			}
		}
	}
	return normalizeBigRat(new(big.Rat).Add(n.bigRatRef(), other.bigRatRef()))
}

func (n *Number) Sub(other *Number) *Number {
	if n != nil && other != nil && n.kind == numberBinary64 && other.kind == numberBinary64 {
		a := math.Float64frombits(uint64(n.num))
		b := math.Float64frombits(uint64(other.num))
		if r, exact := binary64ExactAdd(a, -b); exact {
			return newExactBinary64Result(r)
		}
	}
	if an, ad, ok := n.smallFraction(); ok {
		if bn, bd, ok := other.smallFraction(); ok {
			if bn != -1<<63 {
				if r, ok := addSmallRational(an, ad, -bn, bd); ok {
					return r
				}
			}
		}
	}
	return normalizeBigRat(new(big.Rat).Sub(n.bigRatRef(), other.bigRatRef()))
}

func (n *Number) Mul(other *Number) *Number {
	if n != nil && other != nil && n.kind == numberBinary64 && other.kind == numberBinary64 {
		a := math.Float64frombits(uint64(n.num))
		b := math.Float64frombits(uint64(other.num))
		if r, exact := binary64ExactMul(a, b); exact {
			return newExactBinary64Result(r)
		}
	}
	if an, ad, ok := n.smallFraction(); ok {
		if bn, bd, ok := other.smallFraction(); ok {
			if r, ok := mulSmallRational(an, ad, bn, bd); ok {
				return r
			}
		}
	}
	return normalizeBigRat(new(big.Rat).Mul(n.bigRatRef(), other.bigRatRef()))
}

func (n *Number) Quo(other *Number) *Number {
	if an, ad, ok := n.smallFraction(); ok {
		if bn, bd, ok := other.smallFraction(); ok && bn != 0 {
			// a/b ÷ c/d = a/b × d/c. Normalize the divisor sign before
			// using the multiplication path so denominators stay positive.
			if bn < 0 {
				if bn != -1<<63 && an != -1<<63 {
					bn = -bn
					an = -an
				} else {
					return normalizeBigRat(new(big.Rat).Quo(n.bigRatRef(), other.bigRatRef()))
				}
			}
			if r, ok := mulSmallRational(an, ad, bd, bn); ok {
				return r
			}
		}
	}
	return normalizeBigRat(new(big.Rat).Quo(n.bigRatRef(), other.bigRatRef()))
}

func (n *Number) Neg() *Number {
	const minInt64 = -1 << 63
	if n.kind == numberSmallInt && n.num != minInt64 {
		return &Number{kind: numberSmallInt, num: -n.num}
	}
	if n.kind == numberSmallRat && n.num != minInt64 {
		return &Number{kind: numberSmallRat, num: -n.num, den: n.den}
	}
	return normalizeBigRat(new(big.Rat).Neg(n.bigRatRef()))
}
