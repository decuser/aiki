package value

import "testing"

func TestStringRuneObservation(t *testing.T) {
	s := &String{Val: "aé猫🙂"}
	if got := s.RuneLen(); got != 4 {
		t.Fatalf("RuneLen = %d, want 4", got)
	}
	want := []rune{'a', 'é', '猫', '🙂'}
	for i, w := range want {
		got, ok := s.RuneAt(i)
		if !ok || got != w {
			t.Fatalf("RuneAt(%d) = %q,%v want %q,true", i, got, ok, w)
		}
	}
	for _, i := range []int{-1, 4, 20} {
		if got, ok := s.RuneAt(i); ok {
			t.Fatalf("RuneAt(%d) = %q,true want false", i, got)
		}
	}
}

func TestStringRuneObservationInvalidUTF8MatchesRuneConversion(t *testing.T) {
	raw := string([]byte{'a', 0xff, 'b'})
	s := &String{Val: raw}
	want := []rune(raw)
	if got := s.RuneLen(); got != len(want) {
		t.Fatalf("RuneLen = %d, want %d", got, len(want))
	}
	for i, w := range want {
		got, ok := s.RuneAt(i)
		if !ok || got != w {
			t.Fatalf("RuneAt(%d) = %U,%v want %U,true", i, got, ok, w)
		}
	}
}

func TestStringCompareRunes(t *testing.T) {
	tests := []struct {
		left, right string
		want        int
	}{
		{"alpha", "beta", -1},
		{"beta", "beta", 0},
		{"猫", "é", 1},
		{"é", "猫", -1},
		{"a", "aa", -1},
		{"aa", "a", 1},
	}
	for _, tt := range tests {
		got := (&String{Val: tt.left}).CompareRunes(&String{Val: tt.right})
		if got != tt.want {
			t.Fatalf("CompareRunes(%q,%q) = %d want %d", tt.left, tt.right, got, tt.want)
		}
	}
}

func TestStringRuneAtDoesNotAllocate(t *testing.T) {
	s := &String{Val: "0123456789é猫🙂abcdefghijklmnopqrstuvwxyz"}
	allocs := testing.AllocsPerRun(1000, func() {
		r, ok := s.RuneAt(20)
		if !ok || r == 0 {
			panic("unexpected RuneAt result")
		}
	})
	if allocs != 0 {
		t.Fatalf("RuneAt allocations = %v, want 0", allocs)
	}
}
