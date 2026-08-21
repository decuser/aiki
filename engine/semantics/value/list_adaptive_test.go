package value

import (
	"math/rand"
	"sync"
	"testing"
)

func listInts(t *testing.T, l *List) []int64 {
	t.Helper()
	out := make([]int64, len(l.Elements))
	for i, v := range l.Elements {
		n, ok := v.(*Number)
		if !ok || !n.IsInt() {
			t.Fatalf("element %d: got %T %v, want integer Number", i, v, v)
		}
		out[i] = n.Int64Value()
	}
	return out
}

func assertListInts(t *testing.T, l *List, want ...int64) {
	t.Helper()
	got := listInts(t, l)
	if len(got) != len(want) {
		t.Fatalf("list length: got %d want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("element %d: got %d want %d (%v)", i, got[i], want[i], got)
		}
	}
}

func TestAdaptiveListLinearAppendPreservesHistory(t *testing.T) {
	root := &List{Elements: []Value{}}
	versions := []*List{root}
	current := root
	for i := int64(0); i < 128; i++ {
		next, r := current.Append(NewNumber(i, 1))
		if i == 0 && !r.Promoted {
			t.Fatal("first append did not promote flat list")
		}
		if i > 0 && !r.Extended {
			t.Fatalf("append %d did not extend frontier: %+v", i, r)
		}
		versions = append(versions, next)
		current = next
	}

	for n, version := range versions {
		if len(version.Elements) != n {
			t.Fatalf("version %d length: got %d", n, len(version.Elements))
		}
		for i, v := range version.Elements {
			if got := v.(*Number).Int64Value(); got != int64(i) {
				t.Fatalf("version %d element %d: got %d", n, i, got)
			}
		}
	}
}

func TestAdaptiveListHistoricalAppendForks(t *testing.T) {
	a := &List{Elements: []Value{NewNumber(1, 1), NewNumber(2, 1)}}
	b, first := a.Append(NewNumber(3, 1))
	if !first.Promoted {
		t.Fatalf("first append: %+v", first)
	}

	// Extend b so a is unquestionably historical relative to b's frontier.
	b2, ext := b.Append(NewNumber(5, 1))
	if !ext.Extended {
		t.Fatalf("frontier extension: %+v", ext)
	}

	c, fork := b.Append(NewNumber(4, 1))
	if !fork.Forked {
		t.Fatalf("historical append did not fork: %+v", fork)
	}
	assertListInts(t, a, 1, 2)
	assertListInts(t, b, 1, 2, 3)
	assertListInts(t, b2, 1, 2, 3, 5)
	assertListInts(t, c, 1, 2, 3, 4)
}

func TestAdaptiveListRestDerivedAppendPromotesCopy(t *testing.T) {
	base := &List{Elements: []Value{NewNumber(1, 1), NewNumber(2, 1), NewNumber(3, 1)}}
	frontier, _ := base.Append(NewNumber(4, 1))

	// This is the exact representation used by rest: an immutable flat subslice
	// with no frontier authority, even though it aliases frontier storage.
	rest := &List{Elements: frontier.Elements[1:]}
	derived, r := rest.Append(NewNumber(9, 1))
	if !r.Promoted || r.Forked {
		t.Fatalf("rest-derived append must promote from flat view: %+v", r)
	}
	assertListInts(t, frontier, 1, 2, 3, 4)
	assertListInts(t, rest, 2, 3, 4)
	assertListInts(t, derived, 2, 3, 4, 9)

	// Appending to the original remains independent of the rest-derived branch.
	originalNext, r2 := frontier.Append(NewNumber(5, 1))
	if !r2.Extended {
		t.Fatalf("original frontier should still extend: %+v", r2)
	}
	assertListInts(t, originalNext, 1, 2, 3, 4, 5)
	assertListInts(t, derived, 2, 3, 4, 9)
}

func TestAdaptiveListShapePreserved(t *testing.T) {
	base := &List{Shape: "pair", Elements: []Value{NewNumber(1, 1)}}
	next, _ := base.Append(NewNumber(2, 1))
	if next.Shape != "pair" {
		t.Fatalf("shape: got %q want pair", next.Shape)
	}
	assertListInts(t, base, 1)
	assertListInts(t, next, 1, 2)
}

func TestAdaptiveListConcurrentAppendSameFrontier(t *testing.T) {
	base, _ := (&List{Elements: []Value{}}).Append(NewNumber(1, 1))
	var wg sync.WaitGroup
	wg.Add(2)
	results := make(chan *List, 2)
	for _, n := range []int64{2, 3} {
		n := n
		go func() {
			defer wg.Done()
			next, _ := base.Append(NewNumber(n, 1))
			results <- next
		}()
	}
	wg.Wait()
	close(results)

	seen := map[int64]bool{}
	for r := range results {
		vals := listInts(t, r)
		if len(vals) != 2 || vals[0] != 1 {
			t.Fatalf("concurrent append result: %v", vals)
		}
		seen[vals[1]] = true
	}
	if !seen[2] || !seen[3] {
		t.Fatalf("concurrent branches missing: %v", seen)
	}
	assertListInts(t, base, 1)
}

func TestAdaptiveListRandomizedDifferential(t *testing.T) {
	rng := rand.New(rand.NewSource(0xA11C1))
	lists := []*List{{Elements: []Value{}}}
	refs := [][]int64{{}}

	for step := 0; step < 5000; step++ {
		idx := rng.Intn(len(lists))
		v := int64(rng.Intn(100000))
		next, _ := lists[idx].Append(NewNumber(v, 1))

		ref := make([]int64, len(refs[idx])+1)
		copy(ref, refs[idx])
		ref[len(ref)-1] = v
		lists = append(lists, next)
		refs = append(refs, ref)

		// Sample old and new versions continuously so a frontier write can never
		// silently corrupt an historical logical prefix.
		for _, check := range []int{idx, len(lists) - 1, rng.Intn(len(lists))} {
			got := listInts(t, lists[check])
			want := refs[check]
			if len(got) != len(want) {
				t.Fatalf("step %d version %d length: got %d want %d", step, check, len(got), len(want))
			}
			for i := range want {
				if got[i] != want[i] {
					t.Fatalf("step %d version %d element %d: got %d want %d", step, check, i, got[i], want[i])
				}
			}
		}
	}
}
