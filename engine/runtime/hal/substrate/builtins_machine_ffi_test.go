package substrate

import (
	"aiki/engine/semantics/value"
	"testing"
)

func num(n int64) value.Value { return value.NewNumber(n, 1) }

func TestMachineFFIBoundaryIsOpaqueAndExplicit(t *testing.T) {
	w := halMachineWord([]value.Value{num(012700)}, nil)
	o, ok := w.(*value.Opaque)
	if !ok || o.Type() != value.OpaqueType || o.Inspect() != "<opaque>" {
		t.Fatalf("word = %#v", w)
	}
	if got := halMachineToNumber([]value.Value{w}, nil).(*value.Number).Val.Num().Int64(); got != 012700 {
		t.Fatalf("number=%d", got)
	}
	if _, ok := halMachineWord([]value.Value{value.NewNumber(1, 3)}, nil).(*value.Fault); !ok {
		t.Fatal("fractional word conversion should fault")
	}
	if _, ok := halMachineWord([]value.Value{num(65536)}, nil).(*value.Fault); !ok {
		t.Fatal("out-of-range word conversion should fault")
	}
}

func TestMachineFFIWordArithmeticWrapsWithoutNumberPromotion(t *testing.T) {
	max := halMachineWord([]value.Value{num(65535)}, nil)
	one := halMachineWord([]value.Value{num(1)}, nil)
	zero := halMachineWordAdd([]value.Value{max, one}, nil)
	if !halMachineWordZero([]value.Value{zero}, nil).(*value.Boolean).Val {
		t.Fatal("65535+1 did not wrap")
	}
	if !halMachineWordAddCarry([]value.Value{max, one}, nil).(*value.Boolean).Val {
		t.Fatal("carry not reported")
	}
	back := halMachineWordSub([]value.Value{zero, one}, nil)
	if !halMachineSame([]value.Value{back, max}, nil).(*value.Boolean).Val {
		t.Fatal("sub wrap mismatch")
	}
}

func TestFixedStoreCarriesOpaqueMachineValues(t *testing.T) {
	s := halFixedStoreNewWord([]value.Value{num(4)}, nil)
	w := halMachineWord([]value.Value{num(012345)}, nil)
	if f := halFixedStoreSet([]value.Value{s, num(2), w}, nil); value.IsFault(f) {
		t.Fatal(f.Inspect())
	}
	got := halFixedStoreGet([]value.Value{s, num(2)}, nil)
	if !halMachineSame([]value.Value{got, w}, nil).(*value.Boolean).Val {
		t.Fatal("store round trip")
	}
	if _, ok := halFixedStoreSet([]value.Value{s, num(1), halMachineByte([]value.Value{num(7)}, nil)}, nil).(*value.Fault); !ok {
		t.Fatal("word store accepted byte")
	}
}

func TestFixedWordStorePhysicalByteLanes(t *testing.T) {
	s := halFixedStoreNewWord([]value.Value{num(8)}, nil)
	a := halMachineAddr18([]value.Value{num(2)}, nil)
	w := halMachineWord([]value.Value{num(0x1234)}, nil)
	if f := halFixedStoreWordWriteAddr([]value.Value{s, a, w}, nil); value.IsFault(f) {
		t.Fatal(f.Inspect())
	}
	lo := halFixedStoreByteReadAddr([]value.Value{s, a}, nil)
	hiAddr := halMachineAddr18([]value.Value{num(3)}, nil)
	hi := halFixedStoreByteReadAddr([]value.Value{s, hiAddr}, nil)
	if halMachineToNumber([]value.Value{lo}, nil).(*value.Number).Val.Num().Int64() != 0x34 {
		t.Fatal("low byte")
	}
	if halMachineToNumber([]value.Value{hi}, nil).(*value.Number).Val.Num().Int64() != 0x12 {
		t.Fatal("high byte")
	}
}

func TestMachineFFIFixedValuesAreInterned(t *testing.T) {
	w1 := halMachineWord([]value.Value{num(012345)}, nil)
	w2 := halMachineWordAdd([]value.Value{halMachineWord([]value.Value{num(012344)}, nil), halMachineWord([]value.Value{num(1)}, nil)}, nil)
	if w1 != w2 {
		t.Fatal("equal word values are not interned")
	}

	s := halFixedStoreNewWord([]value.Value{num(1)}, nil)
	if f := halFixedStoreSet([]value.Value{s, num(0), w1}, nil); value.IsFault(f) {
		t.Fatal(f.Inspect())
	}
	if got := halFixedStoreGet([]value.Value{s, num(0)}, nil); got != w1 {
		t.Fatal("fixed store read did not return interned word")
	}

	b1 := halMachineByte([]value.Value{num(0377)}, nil)
	b2 := newByte(0377)
	if b1 != b2 {
		t.Fatal("equal byte values are not interned")
	}

	a1 := halMachineAddr18([]value.Value{num(0777777)}, nil)
	a2 := newAddr18(0777777)
	if a1 != a2 {
		t.Fatal("equal address values are not interned")
	}
}

func TestFixedCounterStoreAddsWithoutGetSetCycle(t *testing.T) {
	s := halFixedStoreNewCounter([]value.Value{num(2)}, nil)
	if f := halFixedStoreCounterAdd([]value.Value{s, num(0), num(7)}, nil); value.IsFault(f) {
		t.Fatal(f.Inspect())
	}
	if f := halFixedStoreCounterAdd([]value.Value{s, num(0), num(5)}, nil); value.IsFault(f) {
		t.Fatal(f.Inspect())
	}
	got := halFixedStoreCounterGet([]value.Value{s, num(0)}, nil).(*value.Number).Val.Num().Int64()
	if got != 12 {
		t.Fatalf("counter=%d", got)
	}
	if f := halFixedStoreCounterSet([]value.Value{s, num(1), num(3)}, nil); value.IsFault(f) {
		t.Fatal(f.Inspect())
	}
	got = halFixedStoreCounterGet([]value.Value{s, num(1)}, nil).(*value.Number).Val.Num().Int64()
	if got != 3 {
		t.Fatalf("counter set=%d", got)
	}
	if _, ok := halFixedStoreGet([]value.Value{s, num(0)}, nil).(*value.Fault); !ok {
		t.Fatal("generic fixed get should not expose counter storage")
	}
}
