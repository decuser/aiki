package value

import (
	"math"
	"testing"
)

func BenchmarkNumberAdaptiveConstructInteger(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewNumber(42, 1)
	}
}

func BenchmarkNumberAdaptiveAddInteger(b *testing.B) {
	left := NewNumber(123456, 1)
	right := NewNumber(789, 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = left.Add(right)
	}
}

func BenchmarkNumberAdaptiveAddRational(b *testing.B) {
	left := NewNumber(1, 3)
	right := NewNumber(2, 7)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = left.Add(right)
	}
}

func BenchmarkNumberAdaptiveMulInteger(b *testing.B) {
	left := NewNumber(123456, 1)
	right := NewNumber(789, 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = left.Mul(right)
	}
}

func BenchmarkNumberAdaptiveCompareInteger(b *testing.B) {
	left := NewNumber(123456, 1)
	right := NewNumber(123457, 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = left.Compare(right)
	}
}

func BenchmarkNumberAdaptiveDivideRational(b *testing.B) {
	left := NewNumber(355, 113)
	right := NewNumber(22, 7)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = left.Quo(right)
	}
}

func BenchmarkNumberAdaptiveBinaryCarrier(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = NewNumberFromFloat64(0.1)
	}
}

func BenchmarkNumberAdaptiveBinaryExactFallbackAdd(b *testing.B) {
	left, _ := NewNumberFromFloat64(0.1)
	right, _ := NewNumberFromFloat64(0.2)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = left.Add(right)
	}
}

func BenchmarkNumberAdaptiveBinaryCertifiedAdd(b *testing.B) {
	left, _ := NewNumberFromFloat64(0.5)
	right, _ := NewNumberFromFloat64(0.25)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = left.Add(right)
	}
}

func BenchmarkNumberAdaptiveBinaryCertifiedMul(b *testing.B) {
	left, _ := NewNumberFromFloat64(0.5)
	right, _ := NewNumberFromFloat64(0.25)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = left.Mul(right)
	}
}

func BenchmarkNumberAdaptiveBinaryBigFallbackAdd(b *testing.B) {
	left, _ := NewNumberFromFloat64(math.SmallestNonzeroFloat64)
	right, _ := NewNumberFromFloat64(0.1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = left.Add(right)
	}
}
