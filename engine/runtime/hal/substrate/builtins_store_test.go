package substrate

import (
	"aiki/engine/semantics/value"
	"testing"
)

func number(n int64) *value.Number { return value.NewNumber(n, 1) }

func requireFault(t *testing.T, v value.Value) {
	t.Helper()
	if _, ok := v.(*value.Fault); !ok {
		t.Fatalf("expected fault, got %T (%s)", v, v.Inspect())
	}
}

func TestStoreNewZeroInitialized(t *testing.T) {
	got := halStoreNew([]value.Value{number(3)}, nil)
	s, ok := got.(*value.Store)
	if !ok {
		t.Fatalf("expected store, got %T", got)
	}
	if s.StoreLen() != 3 {
		t.Fatalf("length: got %d", s.StoreLen())
	}
	for i := 0; i < 3; i++ {
		v := s.StoreGet(i).(*value.Number)
		if v.Val.Sign() != 0 {
			t.Fatalf("cell %d: got %s", i, v.Inspect())
		}
	}
}

func TestStoreSetAndGet(t *testing.T) {
	s := halStoreNew([]value.Value{number(2)}, nil).(*value.Store)
	if got := halStoreSet([]value.Value{s, number(1), &value.String{Val: "word"}}, nil); got != value.EMPTY {
		t.Fatalf("set: got %s", got.Inspect())
	}
	got := halStoreGet([]value.Value{s, number(1)}, nil)
	if str, ok := got.(*value.String); !ok || str.Val != "word" {
		t.Fatalf("get: got %T %s", got, got.Inspect())
	}
}

func TestStoreLength(t *testing.T) {
	s := halStoreNew([]value.Value{number(4)}, nil).(*value.Store)
	got := halStoreLength([]value.Value{s}, nil).(*value.Number)
	if got.Val.Cmp(number(4).Val) != 0 {
		t.Fatalf("got %s", got.Inspect())
	}
}

func TestStoreRejectsInvalidSize(t *testing.T) {
	requireFault(t, halStoreNew([]value.Value{number(-1)}, nil))
	requireFault(t, halStoreNew([]value.Value{value.NewNumber(3, 2)}, nil))
	requireFault(t, halStoreNew([]value.Value{&value.String{Val: "3"}}, nil))
}

func TestStoreRejectsInvalidIndexes(t *testing.T) {
	s := halStoreNew([]value.Value{number(2)}, nil).(*value.Store)
	for _, idx := range []value.Value{number(-1), number(2), value.NewNumber(1, 2)} {
		requireFault(t, halStoreGet([]value.Value{s, idx}, nil))
		requireFault(t, halStoreSet([]value.Value{s, idx, number(1)}, nil))
	}
}

func TestStoreRejectsWrongReceiver(t *testing.T) {
	requireFault(t, halStoreGet([]value.Value{value.EMPTY, number(0)}, nil))
	requireFault(t, halStoreSet([]value.Value{value.EMPTY, number(0), number(1)}, nil))
	requireFault(t, halStoreLength([]value.Value{value.EMPTY}, nil))
}

func TestStoreSnapshotCopiesPrefix(t *testing.T) {
	s := halStoreNew([]value.Value{number(4)}, nil).(*value.Store)
	halStoreSet([]value.Value{s, number(0), number(10)}, nil)
	halStoreSet([]value.Value{s, number(1), number(20)}, nil)

	got := halStoreSnapshot([]value.Value{s, number(2)}, nil)
	list, ok := got.(*value.List)
	if !ok {
		t.Fatalf("snapshot: got %T", got)
	}
	if len(list.Elements) != 2 || list.Elements[0].Inspect() != "10" || list.Elements[1].Inspect() != "20" {
		t.Fatalf("snapshot: got %s", list.Inspect())
	}

	// Snapshot is immutable with respect to later Store mutation.
	halStoreSet([]value.Value{s, number(0), number(99)}, nil)
	if list.Elements[0].Inspect() != "10" {
		t.Fatalf("snapshot changed after store mutation: %s", list.Inspect())
	}
}

func TestStoreSnapshotRejectsInvalidCount(t *testing.T) {
	s := halStoreNew([]value.Value{number(2)}, nil).(*value.Store)
	requireFault(t, halStoreSnapshot([]value.Value{s, number(-1)}, nil))
	requireFault(t, halStoreSnapshot([]value.Value{s, number(3)}, nil))
	requireFault(t, halStoreSnapshot([]value.Value{s, value.NewNumber(1, 2)}, nil))
}

func TestStoreDigitsToText(t *testing.T) {
	s := &value.Store{Cells: []value.Value{
		value.NewNumber(1, 1),
		value.NewNumber(2, 1),
		value.NewNumber(3, 1),
		value.NewNumber(4, 1),
	}}
	got := halStoreDigitsToText([]value.Value{s, value.NewNumber(3, 1)}, nil)
	str, ok := got.(*value.String)
	if !ok {
		t.Fatalf("digits_to_text returned %T, want *value.String", got)
	}
	if str.Val != "123" {
		t.Fatalf("digits_to_text = %q, want 123", str.Val)
	}

	bad := &value.Store{Cells: []value.Value{value.NewNumber(10, 1)}}
	if _, ok := halStoreDigitsToText([]value.Value{bad}, nil).(*value.Fault); !ok {
		t.Fatalf("digits_to_text accepted non-digit cell")
	}
}

func TestStoreChecksum(t *testing.T) {
	s := &value.Store{Cells: []value.Value{
		value.NewNumber(1, 1),
		value.NewNumber(0, 1),
		value.NewNumber(1, 1),
	}}
	got := halStoreChecksum([]value.Value{s, value.NewNumber(3, 1)}, nil)
	n, ok := got.(*value.Number)
	if !ok {
		t.Fatalf("checksum returned %T, want *value.Number", got)
	}
	if n.Val.Num().Int64() != 38234842 {
		t.Fatalf("checksum = %s, want 38234842", n.Inspect())
	}

	if _, ok := halStoreChecksum([]value.Value{&value.Store{Cells: []value.Value{&value.String{Val: "x"}}}}, nil).(*value.Fault); !ok {
		t.Fatalf("checksum accepted non-integer cell")
	}
}
